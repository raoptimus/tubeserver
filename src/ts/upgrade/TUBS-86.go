package main

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"ts/data"
)

func main() {
	DbTV := data.Context.DbTV
	videos := data.Context.Videos
	srcDomain := "..."
	var (
		vid      int
		fileHash string
		done     int
		total    int
	)
	rows, err := DbTV.Query(`SELECT VideoId, FileHash FROM tbh_videos ORDER BY VideoId`)
	if err != nil {
		fmt.Println("Select:", err)
		return
	}

	for rows.Next() {
		err = rows.Scan(&vid, &fileHash)
		if err != nil {
			fmt.Println("Scan:", err)
			return
		}

		src := data.NewVideoSource(vid, srcDomain, 0, 0)
		err = videos.Update(bson.M{"Source._id": src.Id},
			bson.M{"$set": bson.M{"FilesId": fileHash}})
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

	fmt.Println("Done", done)
	fmt.Println("NotFound", total-done)
	fmt.Println("Total", total)
	fmt.Println("----------------------")

	fmt.Println("Search in trash...")
	total = 0
	done = 0

	rows, err = DbTV.Query(`SELECT VideoId, FileHash FROM tbh_videostrash ORDER BY VideoId`)
	if err != nil {
		fmt.Println("Select:", err)
		return
	}

	for rows.Next() {
		err = rows.Scan(&vid, &fileHash)
		if err != nil {
			fmt.Println("Scan:", err)
			return
		}

		src := data.NewVideoSource(vid, srcDomain, 0, 0)
		err = videos.Update(bson.M{"Source._id": src.Id},
			bson.M{"$set": bson.M{"FilesId": fileHash}})
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

	fmt.Println("Done", done)
	fmt.Println("NotFound", total-done)
	fmt.Println("Total", total)

	fmt.Println("All finish")
}
