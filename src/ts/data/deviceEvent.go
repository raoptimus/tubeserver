package data

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type DeviceAction int

const (
	DeviceActionLaunch DeviceAction = iota
	DeviceActionFLaunch
	DeviceActionReInstall
	DeviceActionUpdate
)

type (
	DeviceEvent struct {
		Id        bson.ObjectId `bson:"_id"`
		AddedDate time.Time     `bson:"AddedDate"`
		Action    DeviceAction  `bson:"Action"`
		DeviceId  string        `bson:"DeviceId"`
		Details   string        `bson:"Details"`
		Ip        string        `bson:"Ip"`
		Ver       string        `bson:"Ver"`
	}
	DeviceEventList []*DeviceEvent
)
