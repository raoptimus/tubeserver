package data

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type AppStatus int

const (
	AppStatusNoChecked AppStatus = 0
	AppStatusChecked   AppStatus = 1
)

type Application struct {
	Id          bson.ObjectId `bson:"_id"`
	Name        string        `bson:"Name"`
	Ver         string        `bson:"Ver"`
	BuildVer    string        `bson:"BuildVer"`
	Description string        `bson:"Description"`
	Status      AppStatus     `bson:"Status"`
	AddedDate   time.Time     `bson:"AddedDate"`
}
