package main

import (
	"gopkg.in/mgo.v2/bson"
	"testing"
	"time"
	"ts/data"
	"ts/data/push"
)

func TestConditionAllowed(t *testing.T) {
	var device data.Device
	err := data.Context.Devices.FindId("45089afadf0f0bcc03d563d380bad6c4").One(&device) // несуществующий девайс
	if err != nil {
		t.Fatal(err.Error())
	}

	device.LastActiveTime = now().Add(time.Duration(-5) * 24 * time.Hour)

	var task push.Task
	err = data.Context.PushTask.Find(bson.M{"Options.ElapseDaysLastActiveFrom": bson.M{"$exists": true}}).One(&task)
	if err != nil {
		t.Error(err.Error())
	}
	task.Options = make(map[string]interface{})
	task.Options["ElapseDaysLastActiveFrom"] = 4
	task.Options["ElapseDaysLastActiveTo"] = 6

	fullTask := NewTask(&task)
	if !fullTask.conditionAllowed(&device) {
		t.Error("condition not allowed")
	}
}
