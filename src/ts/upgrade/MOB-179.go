package main

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"ts/data"
)

func main() {
	if err := work(); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("All done")
}

func work() error {
	videos := data.Context.Videos
	DbTV := data.Context.DbTV

	n, _ := videos.Find(bson.M{"Files.Size": nil}).Count()
	fmt.Println("Found videos without file size", n)
	it := videos.Find(bson.M{"Files.Size": nil}).Iter()
	var (
		v          data.Video
		sizeLow    int
		sizeMedium int
		sizeHigh   int
	)

	defer func(it *mgo.Iter) {
		if err := it.Close(); err != nil {
			fmt.Println(err)
		}
	}(it)

	for it.Next(&v) {
		rows, err := DbTV.Query(`SELECT LowSpace, MediumSpace, HighSpace
            FROM tbh_videos WHERE VideoId = ?`, v.Source.SourceId)
		if err != nil {
			return err
		}
		if !rows.Next() {
			fmt.Println("Row has not been received", v.Id, v.Source.SourceId)
			continue
		}

		if err := rows.Scan(&sizeLow, &sizeMedium, &sizeHigh); err != nil {
			rows.Close()
			return err
		}
		rows.Close()

		for i := 0; i < len(v.Files); i++ {
			switch v.Files[i].H {
			case 320:
				v.Files[i].Size = sizeLow
			case 480:
				v.Files[i].Size = sizeMedium
			case 720:
				v.Files[i].Size = sizeHigh

			}
		}

		err = videos.UpdateId(v.Id, bson.M{"$set": bson.M{"Files": v.Files}})
		if err != nil {
			return err
		}
	}

	return nil
}
