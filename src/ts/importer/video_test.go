package main

import (
	"gopkg.in/mgo.v2/bson"
	"testing"
	"ts/data"
)

func TestVideoListIn(t *testing.T) {
	var v data.Video
	err := data.Context.Videos.Find(bson.M{"Featured": true}).Select(bson.M{"_id": 1, "Tags": 1}).One(&v)
	if err != nil {
		t.Fatal(err.Error())
	}

	vm := &VideoMakeRelates{}
	err = vm.Update(v.Id, v.Tags)
	if err != nil {
		t.Fatal(err.Error())
	}
}
