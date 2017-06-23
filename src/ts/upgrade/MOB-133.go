package main

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"ts/data"
)

func main() {
	srcCatId := 26
	catId := 12
	DbTV := data.Context.DbTV
	videos := data.Context.Videos
	srcDomain := "..."
	var (
		vid   int
		done  int
		total int
	)
	rows, err := DbTV.Query(`SELECT VideoId FROM tbh_videos WHERE CategoryId = ? ORDER BY VideoId`, srcCatId)
	if err != nil {
		fmt.Println("Select:", err)
		return
	}

	for rows.Next() {
		err = rows.Scan(&vid)
		if err != nil {
			fmt.Println("Scan:", err)
			return
		}

		src := data.NewVideoSource(vid, srcDomain, 0, 0)
		err = videos.Update(bson.M{"Source._id": src.Id}, bson.M{"$set": bson.M{"CategoryId": catId}})
		if err == nil {
			done++
		} else {
			if err != mgo.ErrNotFound {
				fmt.Println("Update: ", err)
				return
			}
		}
		total++
	}

	fmt.Println("All finish")
	fmt.Println("Done", done)
	fmt.Println("NotFound", total-done)
	fmt.Println("Total", total)
}
