package data

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type (
	event struct {
		Id       bson.ObjectId `bson:"_id"`
		Token    string        `bson:"Token"`
		Category string        `bson:"Category"`
		Action   string        `bson:"Action"`
		Details  string        `bson:"Details"`
		Date     time.Time     `bson:"Date"`
	}
)

func EventTrack(token, category, action, details string) error {
	e := event{
		Id:       bson.NewObjectId(),
		Token:    token,
		Category: category,
		Action:   action,
		Details:  details,
		Date:     time.Now().UTC(),
	}

	return Context.Events.Insert(e)
}
