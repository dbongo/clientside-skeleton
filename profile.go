package main

import (
	"github.com/gomango/multicontext"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"time"
)

const cProfile string = "profiles"

var cprofile *mgo.Collection

type (
	Profile struct {
		Id                 bson.ObjectId `bson:"_id,omitempty" json:"id"`
		User               string        `bson:"user_id" json:"user_id"`
		Name               string        `bson:"name" json:"name"`
		Surname            string        `bson:"surname" json:"surname"`
		About              string        `bson:",omitempty" json:"about"`
		Sex                string        `bson:",omitempty" json:"sex"`
		SexInterest        string        `bson:",omitempty" json:"sexinterest"`
		RelationshipStatus string        `bson:",omitempty" json:"relationshipstatus"`
		Signature          string        `bson:",omitempty" json:"signature"`
		Birthday           time.Time     `bson:",omitempty" json:"birthday"`
		City               string        `bson:",omitempty" json:"city"`
		Country            string        `bson:",omitempty" json:"country"`
	}

	ProfileContext struct {
		multicontext.Context
		Data *Profile
	}
)

func profileCollections() {
	cprofile = mdb.C(cProfile)
}

func profileRoutes() {
	router.Get("/user/{id:bsonid}/profile", (*ProfileContext).FindByUserId)
	router.Get("/profile", (*ProfileContext).Me)
	router.Post("/profile", (*ProfileContext).Update)
}

func (c *ProfileContext) DecodeJSON() {
	c.Data = new(Profile)
	c.Context.Data = c.Data
	c.Context.DecodeJSON()
}

func (ctx *ProfileContext) Me() {
	if u := userFromToken(ctx.Request); u != nil {
		ctx.FindByUserId(u.Id)
	} else {
		ctx.Status(401)
	}
}

func (ctx *ProfileContext) FindByUserId(id string) {
	if u := userFromToken(ctx.Request); u != nil {
		result := new(Profile)
		err := cprofile.Find(bson.M{"user_id": id}).One(result)
		if err != nil {
			ctx.Status(400)
			return
		}
		if u.Id != result.Id.Hex() && u.Role != "admin" {
			ctx.Status(403)
			return
		}
		ctx.JSON(result)
	} else {
		ctx.Status(401)
		return
	}
}

func (ctx *ProfileContext) Update() {
	if u := userFromToken(ctx.Request); u != nil {
		data := make(map[string]interface{})
		alerts := make([]Alert, 0)
		if u.Id != ctx.Data.User && u.Role != "admin" {
			ctx.Status(403)
			return
		}

		count, err := cuser.FindId(bson.ObjectIdHex(ctx.Data.User)).Count()
		if err != nil {
			alerts = append(alerts, Alert{"danger", "Database unavailable. Please try again soon."})
			data["alerts"] = alerts
			ctx.JSON(data)
			return
		}

		if count != 1 {
			alerts = append(alerts, Alert{"danger", "Profile not found."})
			data["alerts"] = alerts
			ctx.JSON(data)
			return
		}
		if err := cprofile.UpdateId(ctx.Data.Id, ctx.Data); err != nil {
			alerts = append(alerts, Alert{"danger", "Could not update your profile. Please try again soon."})
			data["alerts"] = alerts
			ctx.JSON(data)
			return
		}
		alerts = append(alerts, Alert{"success", "Your profile has been updated."})
		data["alerts"] = alerts
		ctx.JSON(data)
		return
	} else {
		ctx.Status(401)
		return
	}
}
