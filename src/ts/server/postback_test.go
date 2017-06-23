package main

import (
	"testing"
	"ts/data"
)

func Test(t *testing.T) {
	var d data.Device
	err := data.Context.Devices.Find(nil).One(&d)
	if err != nil {
		t.Fatal(err.Error())
	}
	d.Source.AffId = "0"
	d.Source.TransId = "00000000000000000000"
	d.Source.OffId = "0"

	d.Source.Partner = "mobionetwork"
	err = PostBack.Send(&d)
	if err != nil {
		t.Fatal(err.Error())
	}

	d.Source.Partner = "cpaplanet"
	err = PostBack.Send(&d)
	if err != nil {
		t.Fatal(err.Error())
	}
}
