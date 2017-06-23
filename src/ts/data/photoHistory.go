package data

import (
	"gopkg.in/mgo.v2/bson"
	"strconv"
	"time"
	"ts/mongodb"
)

type (
	PhotoHistory struct {
		Id        bson.ObjectId    `bson:"_id"`
		UserId    int              `bson:"UserId"`
		PhotoId   int              `bson:"PhotoAlbumId"`
		Type      PhotoHistoryType `bson:"Type"`
		AddedDate time.Time        `bson:"AddedDate"`
	}
)

func (s *PhotoHistory) Exists() (bool, error) {
	n, err := Context.PhotoHistory.FindId(s.Id).Count()
	return n > 0, err
}

func NewPhotoHistory(userId, photoId int, t PhotoHistoryType) *PhotoHistory {
	id := mongodb.GenerateObjectId(strconv.Itoa(userId), strconv.Itoa(photoId), strconv.Itoa(int(t)))
	return &PhotoHistory{
		Id:        id,
		UserId:    userId,
		PhotoId:   photoId,
		Type:      t,
		AddedDate: time.Now().UTC(),
	}
}
