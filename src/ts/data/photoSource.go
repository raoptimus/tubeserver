package data

import (
	"gopkg.in/mgo.v2/bson"
	"strconv"
	"ts/mongodb"
)

type (
	PhotoAlbumSource struct {
		Id               bson.ObjectId `bson:"_id"`
		SourceId         int           `bson:"SourceId"`
		Domain           string        `bson:"Domain"`
		PhotoCount       int           `bson:"PhotoCount"`
		PhotoSelectIndex int           `bson:"PhotoSelectIndex"`
	}
)

func NewPhotoAlbumSource(srcId int, domain string, photoCount int, photoIndexSelect int) *PhotoAlbumSource {
	id := mongodb.GenerateObjectId(strconv.Itoa(srcId), domain)

	return &PhotoAlbumSource{
		Id:               id,
		SourceId:         srcId,
		Domain:           domain,
		PhotoCount:       photoCount,
		PhotoSelectIndex: photoIndexSelect,
	}
}
