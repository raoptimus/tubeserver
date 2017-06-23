package main

import (
	"fmt"
	"github.com/raoptimus/gserv/config"
	"github.com/raoptimus/gserv/service"
	"github.com/raoptimus/rlog"
	"gopkg.in/mgo.v2/bson"
	"os"
	"time"
	"ts/data"
)

var log *rlog.Logger

func main() {
	if service.Exists() {
		os.Exit(0)
	}
	var err error
	log, err = rlog.NewLoggerDial(rlog.LoggerTypeMongoDb, "", config.String("MongoLogServer", config.String("MongoAllServer", "localhost/Logs")), "")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	service.Init(&service.BaseService{
		Start:  start,
		Logger: log,
	})
	service.Start(true)
}

func start() {
	data.Init(false)
	log.Println("Video publisher is running")
	limit := 100
	var (
		err          error
		totalSuccess int
		totalErr     int
	)

	for {
		totalSuccess = 0
		totalErr = 0
		q := bson.M{"Filters": "approved", "PublishedDate": bson.M{"$lte": time.Now().UTC()}}

		for skip := 0; true; skip += limit {
			var videoList []data.Video
			if err := data.Context.Videos.Find(q).Skip(skip).Limit(limit).All(&videoList); err != nil {
				log.Printf("Dont take the video list for publishing (%v, %d, %d): %v", q, skip, limit, err.Error())
				break
			}
			if len(videoList) == 0 {
				break
			}
			for _, v := range videoList {
				for i, f := range v.Filters {
					if f == "approved" {
						v.Filters[i] = "published"
						break
					}
				}
				err = data.Context.Videos.UpdateId(v.Id, bson.M{"$set": bson.M{"Filters": v.Filters}})
				if err != nil {
					log.Printf("Push published in filters (%v) for video (%d) is error: %v", v.Filters, v.Id, err.Error())
					totalErr++
					continue
				}
				totalSuccess++
			}
		}
		log.Printf("Video publisher proccesed, success: %v, error: %v", totalSuccess, totalErr)
		time.Sleep(5 * time.Minute)
	}
}
