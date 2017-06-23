package data

import (
	"gopkg.in/mgo.v2/bson"
	"strconv"
	"time"
	"ts/mongodb"
)

type VideoHistoryType int

const (
	VideoHistoryTypeLike VideoHistoryType = iota
	VideoHistoryTypeView
	VideoHistoryTypeDownload
)

type (
	VideoHistory struct {
		Id        bson.ObjectId    `bson:"_id"`
		UserId    int              `bson:"UserId"`
		VideoId   int              `bson:"VideoId"`
		Type      VideoHistoryType `bson:"Type"`
		AddedDate time.Time        `bson:"AddedDate"`
	}
)

func (s *VideoHistory) Exists() (bool, error) {
	n, err := Context.VideoHistory.FindId(s.Id).Count()
	return n > 0, err
}

func NewVideoHistory(userId, videoId int, t VideoHistoryType) *VideoHistory {
	id := mongodb.GenerateObjectId(strconv.Itoa(userId), strconv.Itoa(videoId), strconv.Itoa(int(t)))
	return &VideoHistory{
		Id:        id,
		UserId:    userId,
		VideoId:   videoId,
		Type:      t,
		AddedDate: time.Now().UTC(),
	}
}

func (s *VideoHistory) Insert(d *Device) error {
	if err := Context.VideoHistory.Insert(s); err != nil {
		return err
	}
	stat := NewDailyStat(&d.Source.DeviceSource)
	switch s.Type {
	case VideoHistoryTypeLike:
		stat.LikeVideoCount++
	case VideoHistoryTypeView:
		stat.ViewVideoCount++
	case VideoHistoryTypeDownload:
		stat.DownloadVideoCount++
	}
	return stat.UpsertInc()
}
