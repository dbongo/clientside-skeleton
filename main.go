package main

import (
	"github.com/dalu/jwt"
	"github.com/gomango/pux"
	"io/ioutil"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
	"net/http"
)

const (
	mongodbHost     string = "localhost"
	mongodbDatabase string = "cside"
	domain          string = "skeleton.dev"     // used for mails
	privKeyPath            = "keys/app.rsa"     // openssl genrsa -out app.rsa keysize
	pubKeyPath             = "keys/app.rsa.pub" // openssl rsa -in app.rsa -pubout > app.rsa.pub
)

var (
	router             *pux.Router
	ms                 *mgo.Session
	mdb                *mgo.Database
	devmode            bool   = true
	mailfrom           string = "root@localhost"
	secret             []byte = []byte("some.very.secret.string.that.is.so.secret.that.you.cant.guess.it")
	verifyKey, signKey []byte
)

type Alert struct {
	Type string `json:"type"`
	Msg  string `json:"msg"`
}

func main() {
	var e error
	signKey, e = ioutil.ReadFile(privKeyPath)
	if e != nil {
		log.Fatal("Error reading private key")
		return
	}

	verifyKey, e = ioutil.ReadFile(pubKeyPath)
	if e != nil {
		log.Fatal("Error reading private key")
		return
	}

	ms, e = mgo.Dial(mongodbHost)
	if e != nil {
		panic(e)
	}
	defer ms.Close()
	mdb = ms.DB(mongodbDatabase)
	accountCollections()
	profileCollections()

	router = pux.New()
	accountRoutes()
	profileRoutes()

	http.ListenAndServe(":8080", Log(router))
}

func Log(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s %s", r.RemoteAddr, r.Method, r.Host, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func userFromToken(t *http.Request) *UserClaim {
	token, err := jwt.Parse(t.Header.Get("Authorization"), func(token *jwt.Token) ([]byte, error) {
		return verifyKey, nil
	})
	if err == nil && token.Valid {
		if u, ok := token.Claims["user"].(map[string]interface{}); ok {
			uc := new(UserClaim)
			if id, ok := u["Id"].(string); ok {
				uc.Id = id
			} else {
				return nil
			}
			if role, ok := u["Role"].(string); ok {
				uc.Role = role
			} else {
				return nil
			}
			return uc
		} else {
			return nil
		}
	} else {
		return nil
	}
}

func isBanned(r *http.Request) bool {
	if u := userFromToken(r); u != nil {
		user := new(User)
		if err := cuser.FindId(bson.ObjectIdHex(u.Id)).One(user); err != nil {
			return true
		}
		if user.AccountStatus.Condition == "banned" || user.AccountStatus.Condition == "permbanned" {
			return true
		}
		return false
	} else {
		return true
	}
}

/*
func CheckToken(tokenValue string) (bool, string) {
	token, err := jwt.Parse(tokenValue, func(token *jwt.Token) ([]byte, error) {
		// since we only use the one private key to sign the tokens,
		// we also only use its public counter part to verify
		return verifyKey, nil
	})
	// branch out into the possible error from signing
	switch err.(type) {
	case nil: // no error
		if !token.Valid { // but may still be invalid
			return false, "Token invalid"
		}
		fmt.Println(token)
		return true, ""
	case *jwt.ValidationError: // something was wrong during the validation
		vErr := err.(*jwt.ValidationError)

		switch vErr.Errors {
		case jwt.ValidationErrorExpired:
			return false, "Token expired"

		default:
			return false, fmt.Sprintf("ValidationError error: %+v\n", vErr.Errors)
		}

	default: // something else went wrong
		return false, fmt.Sprintf("Token parse error: %v\n", err)
	}
}
*/
