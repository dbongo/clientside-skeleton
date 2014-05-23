package main

import (
	"bytes"
	"code.google.com/p/go.crypto/bcrypt"
	"crypto/rand"
	"fmt"
	"github.com/dalu/jwt"
	"github.com/dalu/mail"
	"github.com/dchest/authcookie"
	"github.com/dchest/uniuri"
	"github.com/gomango/context"
	"io"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
	"strings"
	"text/template"
	"time"
)

const cUser string = "users"

var cuser *mgo.Collection

type (
	UserClaim struct {
		Id   string
		Role string
	}

	UserContext struct {
		context.Context
		Data *User
	}

	AccountContext struct {
		context.Context
	}

	AccountRegisterContext struct {
		context.Context
		Data *AccountRegister
	}

	AccountLoginContext struct {
		context.Context
		Data *AccountLogin
	}

	AccountResetContext struct {
		context.Context
		Data *AccountReset
	}

	AccountRegister struct {
		Name    string `json:"name"`
		Surname string `json:"surname"`
		Email   string `json:"email"`
		Email2  string `json:"email2"`
	}

	AccountLogin struct {
		Email      string `json:"email"`
		Password   string `json:"password"`
		Rememberme bool   `json:"rememberme"`
	}

	AccountReset struct {
		Email string `json:"email"`
		Token string `json:"token"`
	}

	User struct {
		Id            bson.ObjectId `bson:"_id,omitempty" json:"id"`
		Email         string        `json:"email"`
		Password      string        `json:"password"`
		Created       time.Time     `json:"created"`
		Role          string        `json:"role"`
		AccountStatus AccountStatus `json:"status"`
		LoginStatus   LoginStatus   `json:"loginstatus"`
		ResetToken    string        `bson:"resettoken,omitempty" json:"resettoken"`
	}

	AccountStatus struct {
		Condition string    `json:"condition"`
		Bancount  int       `bson:",omitempty" json:bancount`
		Since     time.Time `bson:",omitempty" json:"since"`
		Until     time.Time `bson:",omitempty" json:"until"`
	}

	LoginStatus struct {
		FailedAttempts int8         `json:"failedattempts"`
		LastSuccessful time.Time    `json:"lastlogin"`
		LastFailed     time.Time    `json:"lastfailed"`
		LoginHistory   []LoginEntry `json:"loginhistory"`
	}

	LoginEntry struct {
		Timestamp time.Time `bson:"timestamp" json:"timestamp"`
		UserAgent string    `bson:"useragent" json:"useragent"`
		Ip        string    `bson:"ip" json:"ip"`
	}
)

func accountRoutes() {
	router.Get("/users", (*UserContext).Find)
	router.Get("/user/{id:bsonid}", (*UserContext).FindId)

	router.Get("/account/status", (*AccountContext).UserStatus)
	router.Post("/account/ban", (*AccountContext).BanUser)

	router.Put("/account", (*AccountRegisterContext).Create)
	router.Post("/account", (*AccountLoginContext).Authenticate)
	router.Post("/account/logout", (*AccountContext).Logout)
	router.Post("/account/reset", (*AccountResetContext).ResetRequest)
	router.Put("/account/reset", (*AccountResetContext).ResetPassword)
}

func accountCollections() {
	cuser = mdb.C(cUser)
}

func (c *UserContext) DecodeJSON() {
	c.Data = new(User)
	c.Context.Data = c.Data
	c.Context.DecodeJSON()
}

func (c *AccountRegisterContext) DecodeJSON() {
	c.Data = new(AccountRegister)
	c.Context.Data = c.Data
	c.Context.DecodeJSON()
}

func (c *AccountLoginContext) DecodeJSON() {
	c.Data = new(AccountLogin)
	c.Context.Data = c.Data
	c.Context.DecodeJSON()
}

func (c *AccountResetContext) DecodeJSON() {
	c.Data = new(AccountReset)
	c.Context.Data = c.Data
	c.Context.DecodeJSON()
}

// User

func (ctx *UserContext) Find() {
	if u := userFromToken(ctx.Request); u != nil {
		if u.Role == "admin" {
			result := []User{}
			cuser.Find(nil).Sort("email").All(&result)
			ctx.JSON(result)
			return
		} else {
			ctx.Status(403)
			return
		}
	} else {
		ctx.Status(401)
		return
	}
}

func (ctx *UserContext) FindId(id string) {
	if u := userFromToken(ctx.Request); u != nil {
		if u.Role == "admin" || u.Id == id {
			c := new(User)
			bid := bson.ObjectIdHex(id)
			if err := cuser.FindId(bid).One(&c); err != nil {
				if err == mgo.ErrNotFound {
					ctx.Status(400)
					return
				} else {
					if lerr, ok := err.(*mgo.LastError); ok {
						ctx.Status(500)
						ctx.Text("Code: " + string(lerr.Code))
						ctx.Text("Message:" + lerr.Err)
					}
					return
				}
			}
			ctx.JSON(c)
		} else {
			ctx.Status(403)
		}
	} else {
		ctx.Status(401)
	}
}

// Account
func (ctx *AccountContext) BanUser() {
	if u := userFromToken(ctx.Request); u != nil {
		if u.Role != "admin" {
			ctx.Status(403)
			return
		}
		if input, ok := ctx.Data.(map[string]interface{}); ok {
			if userid, ok := input["userid"].(string); ok {
				data := make(map[string]interface{})
				alerts := make([]Alert, 0)
				user := new(User)
				if err := cuser.FindId(bson.ObjectIdHex(userid)).One(user); err != nil {
					ctx.Status(500)
					return
				}
				user.AccountStatus.Condition = "banned"
				user.AccountStatus.Since = time.Now()
				user.AccountStatus.Bancount++
				user.AccountStatus.Until = time.Now().Add(48 * time.Hour * time.Duration(user.AccountStatus.Bancount))
				if err := cuser.UpdateId(user.Id, user); err != nil {
					alerts = append(alerts, Alert{"danger", "Could not ban user"})
					data["alerts"] = alerts
					ctx.Status(500)
					ctx.JSON(data)
					return
				}
				alerts = append(alerts, Alert{"success", "User was banned"})
				data["alerts"] = alerts
				ctx.JSON(data)
			} else {
				ctx.Status(400)
				return
			}
		} else {
			ctx.Status(401)
		}
	} else {
		ctx.Status(500)
	}
}

func (ctx *AccountContext) UserStatus() {
	data := make(map[string]interface{})
	data["user"] = userFromToken(ctx.Request)
	ctx.JSON(data)
}

func (ctx *AccountRegisterContext) Create() {
	type MailContext struct {
		Name, Surname, Host, Password string
	}
	data := make(map[string]interface{})
	e := make([]Alert, 0)

	ctx.Data.Name = strings.Trim(ctx.Data.Name, " ")
	ctx.Data.Surname = strings.Trim(ctx.Data.Surname, " ")
	ctx.Data.Email = strings.Trim(ctx.Data.Email, " ")
	ctx.Data.Email2 = strings.Trim(ctx.Data.Email2, " ")

	if ctx.Data.Name == "" {
		e = append(e, Alert{"danger", "Please fill out your name"})
	}
	if ctx.Data.Surname == "" {
		e = append(e, Alert{"danger", "Please fill out your last name"})
	}
	if ctx.Data.Email == "" {
		e = append(e, Alert{"danger", "Please enter a valid email-address"})
	}
	if ctx.Data.Email != ctx.Data.Email2 {
		e = append(e, Alert{"danger", "The two email addresses don't match"})
	}

	c, err := cuser.Find(bson.M{"email": ctx.Data.Email}).Count()
	if err != nil {
		ctx.Status(500)
		return
	}
	if c > 0 {
		e = append(e, Alert{"danger", "The user with this email address already exists."})
	}

	if len(e) == 0 {
		user := new(User)
		password := uniuri.New()
		if devmode {
			log.Printf("New User: %s, %s\n", ctx.Data.Email, password)
		}
		b := []byte(password)
		b, _ = bcrypt.GenerateFromPassword(b, 12)
		user.Id = bson.NewObjectId()
		user.Created = time.Now()
		user.AccountStatus.Condition = "new"
		user.Email = ctx.Data.Email
		user.Password = string(b)
		user.Role = "user"
		profile := new(Profile)
		profile.Id = bson.NewObjectId()
		profile.User = user.Id.Hex()
		profile.Name = ctx.Data.Name
		profile.Surname = ctx.Data.Surname

		mailtemplate, merr := template.New("newaccount.txt").ParseFiles("mail/user/newaccount.txt")
		if merr != nil {
			log.Println("AccountRegisterContext.Create: mail/user/newaccount.txt not found on medium")
			ctx.Status(500)
			return
		}
		mc := MailContext{
			Name:     ctx.Data.Name,
			Surname:  ctx.Data.Surname,
			Host:     domain,
			Password: password,
		}
		body := bytes.NewBuffer(nil)
		merr = mailtemplate.Execute(body, mc)
		if merr != nil {
			log.Println(merr)
		}

		if !devmode {
			m := mail.NewMail(mailfrom, []string{ctx.Data.Email}, "Welcome to "+domain, body.String())
			if err := m.Send(); err != nil {
				cuser.RemoveId(user.Id)
				e = append(e, Alert{"danger", "Sending mail failed. Please try again later."})
				data["alerts"] = e
				ctx.Status(500)
				ctx.JSON(data)
				return
			}
		} else {
			fmt.Println(body) //TODO
			// TODO
			err := cuser.Insert(user)
			if err != nil {
				ctx.Status(500)
				return
			}
			err = cprofile.Insert(profile)
			if err != nil {
				ctx.Status(500)
				return
			}
		}
	} else {
		data["alerts"] = e
		ctx.Status(400)
		ctx.JSON(data)
		return
	}

}

func (ctx *AccountLoginContext) Authenticate() {
	valid := false
	data := make(map[string]interface{})
	alerts := make([]Alert, 0)
	ctx.Data.Email = strings.Trim(ctx.Data.Email, " ")
	user := new(User)
	if err := cuser.Find(bson.M{"email": ctx.Data.Email}).One(user); err != nil {
		valid = false
	} else {
		// check if login allowed
		if user.LoginAllowed() {
			if valid = user.VerifyCredentials(ctx.Data.Email, ctx.Data.Password); !valid {
				if err := user.FailLogin(); err != nil {
					log.Println("DANGER! Could not FailLogin()", err, user.Id)
				}
			}
		} else {
			// login not allowed
			alerts = append(alerts, Alert{"warning", "You have failed 3 login attempts in the last 15 Minutes. Please wait 15 Minutes from now on and try again."})
			data["alerts"] = alerts
			ctx.Status(400)
			ctx.JSON(data)
			return
		}
	}
	if valid {
		rip := ctx.Request.Header.Get("X-Real-Ip")
		if rip != "" {
			uerr := user.Login(ctx.Request.UserAgent(), rip)
			if uerr != nil {
				ctx.Status(500)
				return
			}
		} else {
			uerr := user.Login(ctx.Request.UserAgent(), ctx.Request.RemoteAddr)
			if uerr != nil {
				ctx.Status(500)
				return
			}
		}
		// create a signer for rsa 256
		t := jwt.New(jwt.GetSigningMethod("RS512"))
		t.Claims["user"] = UserClaim{
			Id:   user.Id.Hex(),
			Role: user.Role,
		}
		if ctx.Data.Rememberme {
			t.Claims["exp"] = time.Now().Add(time.Hour * 24 * 30 * 12).Unix()
			data["remember"] = true
		} else {
			t.Claims["exp"] = time.Now().Add(time.Hour * 8).Unix()
			data["remember"] = false
		}
		tokenString, terr := t.SignedString(signKey)
		if terr != nil {
			alerts = append(alerts, Alert{"danger", "Token signing error."})
			data["alerts"] = alerts
			ctx.Status(500)
			ctx.JSON(data)
			log.Println("Could not retrieve tokenString:", terr)
			return
		}
		data["token"] = tokenString
	} else {
		alerts = append(alerts, Alert{"danger", "Login not successful."})
		data["alerts"] = alerts
		ctx.Status(400)
	}

	ctx.JSON(data)
}

func (ctx *AccountContext) Logout() {
	data := make(map[string]interface{})
	alerts := make([]Alert, 0)
	alerts = append(alerts, Alert{"success", "You have been logged out."})
	data["alerts"] = alerts
	data["token"] = ""
	ctx.JSON(data)
}

func (ctx *AccountResetContext) ResetRequest() {
	type MailContext struct {
		Name, Surname, Host, Token string
	}
	data := make(map[string]interface{})
	alerts := make([]Alert, 0)

	ctx.Data.Email = strings.Trim(ctx.Data.Email, " ")
	user := new(User)
	if err := cuser.Find(bson.M{"email": ctx.Data.Email}).One(user); err != nil {
		alerts = append(alerts, Alert{"danger", "There is no account with this email address."})
		data["alerts"] = alerts
		ctx.Status(500)
		ctx.JSON(data)
		return
	}
	profile := new(Profile)
	if err := cprofile.Find(bson.M{"user_id": user.Id.Hex()}).One(profile); err != nil {
		alerts = append(alerts, Alert{"danger", "There is no profile associated with this account."})
		data["alerts"] = alerts
		ctx.Status(500)
		ctx.JSON(data)
		return
	}

	user.ResetToken = authcookie.NewSinceNow(user.Email, 24*time.Hour, secret)
	mailtemplate, merr := template.New("sendpasswordtoken.txt").ParseFiles("mail/user/sendpasswordtoken.txt")
	if merr != nil {
		log.Println("AccountResetContext.ResetRequest: mail/user/sendpasswordtoken.txt not found on medium")
		ctx.Status(500)
		return
	}

	mc := MailContext{
		Name:    profile.Name,
		Surname: profile.Surname,
		Host:    domain,
		Token:   user.ResetToken,
	}
	body := bytes.NewBuffer(nil)
	merr = mailtemplate.Execute(body, mc)
	if merr != nil {
		log.Println("Error while executing mailtemplate:", merr)
		ctx.Status(500)
		return
	}

	if !devmode {
		// start
		m := mail.NewMail(mailfrom, []string{user.Email}, "Password Reset", body.String())
		if err := m.Send(); err != nil {
			log.Println("Could not send Password Reset mail")
			alerts = append(alerts, Alert{"danger", "Sending mail failed. Please try again later."})
			user.ResetToken = ""
			merr := cuser.UpdateId(user.Id, user)
			if merr != nil {
				log.Println("DANGER!! Could not reset the token to database and sending mail failed")
				alerts = append(alerts, Alert{"danger", "Internal Server Error. Please try again later."})
			}
			data["alerts"] = alerts
			ctx.JSON(data)
			return
		} else {
			merr := cuser.UpdateId(user.Id, user)
			if merr != nil {
				log.Println("DANGER!! Could not reset the token to database")
				alerts = append(alerts, Alert{"danger", "Internal Server Error. Please try again later."})
				data["alerts"] = alerts
				ctx.JSON(data)
				return
			}
			alerts = append(alerts, Alert{"success", "An email has been sent to " + ctx.Data.Email + ". Please check your email account for further instructions."})
			data["alerts"] = alerts
			ctx.JSON(data)
			return
		}
		// end
	} else {
		// delete later
		fmt.Println(body)
		merr := cuser.UpdateId(user.Id, user)
		if merr != nil {
			log.Println("DANGER!! Could not reset the token to database:", merr)
			alerts = append(alerts, Alert{"danger", "Internal Server Error. Please try again later."})
			data["alerts"] = alerts
			ctx.JSON(data)
			return
		}
		alerts = append(alerts, Alert{"success", "An email has been sent to " + ctx.Data.Email + ". Please check your email account for further instructions."})
		data["alerts"] = alerts
		ctx.JSON(data)
		return
	}

}

func (ctx *AccountResetContext) ResetPassword() {
	type MailContext struct {
		Name, Surname, Host, Password string
	}
	data := make(map[string]interface{})
	alerts := make([]Alert, 0)

	if len(ctx.Data.Token) >= 1024+authcookie.MinLength {
		ctx.Status(500)
		return
	}

	password := uniuri.New()
	b := []byte(password)
	b, _ = bcrypt.GenerateFromPassword(b, 12)

	user := new(User)
	err := cuser.Find(bson.M{"resettoken": ctx.Data.Token}).One(user)
	if err != nil {
		alerts = append(alerts, Alert{"danger", "Invalid Token"})
		data["alerts"] = alerts
		ctx.Status(500)
		ctx.JSON(data)
		return
	}
	profile := new(Profile)
	if err := cprofile.Find(bson.M{"user_id": user.Id.Hex()}).One(profile); err != nil {
		alerts = append(alerts, Alert{"danger", "There is no profile associated with this account."})
		data["alerts"] = alerts
		ctx.Status(500)
		ctx.JSON(data)
		return
	}

	mailtemplate, merr := template.New("newpassword.txt").ParseFiles("mail/user/newpassword.txt")
	if merr != nil {
		log.Println("AccountResetContext.ResetPassword: mail/user/newpassword.txt not found on medium")
		ctx.Status(500)
		return
	}
	mc := MailContext{
		Name:     profile.Name,
		Surname:  profile.Surname,
		Host:     domain,
		Password: password,
	}
	body := bytes.NewBuffer(nil)
	merr = mailtemplate.Execute(body, mc)
	if merr != nil {
		log.Println("Error while executing mailtemplate:", merr)
		ctx.Status(500)
		return
	}

	token := user.ResetToken
	login, expires, aerr := authcookie.Parse(token, secret)
	if aerr != nil {
		alerts = append(alerts, Alert{"danger", "Invalid Token"})
		data["alerts"] = alerts
		ctx.Status(500)
		ctx.JSON(data)
		return
	}
	if time.Since(expires) >= 24*time.Hour {
		alerts = append(alerts, Alert{"danger", "Token expired"})
		data["alerts"] = alerts
		ctx.Status(500)
		ctx.JSON(data)
		return
	}

	if login != user.Email {
		alerts = append(alerts, Alert{"danger", "Invalid Token"})
		data["alerts"] = alerts
		ctx.Status(500)
		ctx.JSON(data)
		return
	}

	user.Password = strings.Trim(string(b), "\x00")
	if devmode {
		fmt.Printf("User Password Changed: %s, %s\n", user.Email, password)
	}
	if !devmode {
		m := mail.NewMail(mailfrom, []string{user.Email}, "Your new password for "+domain, body.String())
		if err := m.Send(); err != nil {
			log.Println("Could not send New Password mail for user ", user.Email)
			alerts = append(alerts, Alert{"danger", "Sending mail failed. Please try again later."})
			data["alerts"] = alerts
			ctx.JSON(data)
			return
		} else {
			user.ResetToken = ""
			dberr := cuser.UpdateId(user.Id, user)
			if dberr != nil {
				alerts = append(alerts, Alert{"danger", "Could not update your password. Please try again later."})
				data["alerts"] = alerts
				ctx.JSON(data)
				return
			}
			alerts = append(alerts, Alert{"success", "Your new password has been sent. Please check your email."})
			data["alerts"] = alerts
			ctx.JSON(data)
		}
	} else {
		//deleteme
		fmt.Println(body)
		user.ResetToken = ""
		dberr := cuser.UpdateId(user.Id, user)
		if dberr != nil {
			alerts = append(alerts, Alert{"danger", "Could not update your password. Please try again later."})
			data["alerts"] = alerts
			ctx.JSON(data)
			return
		}
		alerts = append(alerts, Alert{"success", "Your new password has been sent. Please check your email."})
		data["alerts"] = alerts
		ctx.JSON(data)
	}
}

// Model Funcs

func (u *User) GenerateToken(l int) string {
	b := make([]byte, l)
	_, _ = io.ReadFull(rand.Reader, b)
	return fmt.Sprintf("%02x", b)
}

func (u *User) VerifyCredentials(email, password string) bool {
	if email != u.Email {
		return false
	}
	if bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)) != nil {
		return false
	}
	return true
}

func (u *User) Login(ua, ip string) error {
	u.LoginStatus.LastSuccessful = time.Now()
	u.AccountStatus.Condition = "active"
	l := LoginEntry{
		u.LoginStatus.LastSuccessful,
		ua,
		strings.Split(ip, ":")[0],
	}
	if len(u.LoginStatus.LoginHistory) > 10 {
		u.LoginStatus.LoginHistory = u.LoginStatus.LoginHistory[len(u.LoginStatus.LoginHistory)-10 : len(u.LoginStatus.LoginHistory)]
	}
	u.LoginStatus.LoginHistory = append(u.LoginStatus.LoginHistory, l)
	u.LoginStatus.FailedAttempts = 0
	return cuser.UpdateId(u.Id, u)
}

func (u *User) LoginAllowed() bool {
	if u.AccountStatus.Condition == "permbanned" {
		return false
	}
	if u.AccountStatus.Condition == "banned" {
		if time.Now().Before(u.AccountStatus.Until) {
			return false
		} else {
			u.AccountStatus.Condition = "active"
			u.AccountStatus.Since = time.Time{}
			u.AccountStatus.Until = time.Time{}
		}
	}
	if u.LoginStatus.FailedAttempts >= 3 {
		if time.Since(u.LoginStatus.LastFailed) >= time.Minute*15 {
			u.LoginStatus.FailedAttempts = 0
			return true
		} else {
			return false
		}
	} else {
		return true
	}
}

func (u *User) FailLogin() error {
	u.LoginStatus.FailedAttempts++
	if u.LoginStatus.FailedAttempts >= 3 {
		u.LoginStatus.LastFailed = time.Now()
	}
	return cuser.UpdateId(u.Id, u)
}
