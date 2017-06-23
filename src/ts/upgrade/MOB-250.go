package main

import (
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"time"
	"ts/data"
)

func main() {
	it := data.Context.Devices.Find(bson.M{"HasGoogleId": true}).Iter()
	var d data.Device

	for it.Next(&d) {
		l, err := time.LoadLocation(d.Loc.Loc)
		if err == nil {
			h := time.Now().In(l).Hour() - time.Now().UTC().Hour()
			err = data.Context.Devices.UpdateId(d.Id, bson.M{"$set": bson.M{"Loc.Gmt": h}})
		}

		if err != nil {
			fmt.Println(err)
		}
	}

	err := it.Close()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Finish")
}
