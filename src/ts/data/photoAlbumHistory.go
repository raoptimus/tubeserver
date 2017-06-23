package data

import (
	"gopkg.in/mgo.v2/bson"
	"strconv"
	"time"
	"ts/mongodb"
)

type PhotoAlbumHistoryType int

const (
	PhotoAlbumHistoryTypeLike PhotoAlbumHistoryType = iota
	PhotoAlbumHistoryTypeView
)

type (
	PhotoAlbumHistory struct {
		Id           bson.ObjectId         `bson:"_id"`
		UserId       int                   `bson:"UserId"`
		PhotoAlbumId int                   `bson:"PhotoAlbumId"`
		Type         PhotoAlbumHistoryType `bson:"Type"`
		AddedDate    time.Time             `bson:"AddedDate"`
	}
)

func (s *PhotoAlbumHistory) Exists() (bool, error) {
	n, err := Context.PhotoAlbumHistory.FindId(s.Id).Count()
	return n > 0, err
}

func NewPhotoAlbumHistory(userId, photoAlbumId int, t PhotoAlbumHistoryType) *PhotoAlbumHistory {
	id := mongodb.GenerateObjectId(strconv.Itoa(userId), strconv.Itoa(photoAlbumId), strconv.Itoa(int(t)))
	return &PhotoAlbumHistory{
		Id:           id,
		UserId:       userId,
		PhotoAlbumId: photoAlbumId,
		Type:         t,
		AddedDate:    time.Now().UTC(),
	}
}
