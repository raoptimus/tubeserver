package data

import (
	"gopkg.in/mgo.v2/bson"
	"strings"
	"time"
)

type (
	SearchQuery struct {
		Id          string    `bson:"_id"`
		SearchCount int       `bson:"SearchCount"`
		ResultCount int       `bson:"ResultCount"`
		UpdateDate  time.Time `bson:"UpdateDate"`
		AddedDate   time.Time `bson:"AddedDate"`
	}
)

func InsertSearchQuery(keyword string, resultCount int) error {
	id := strings.ToLower(keyword)
	update := bson.M{
		"$setOnInsert": bson.M{
			"AddedDate": time.Now().UTC(),
			"_id":       id,
		},
		"$inc": bson.M{
			"SearchCount": 1,
		},
		"$set": bson.M{
			"UpdateDate":  time.Now().UTC(),
			"ResultCount": resultCount,
		},
	}
	_, err := Context.SearchQueries.UpsertId(id, update)
	return err
}
