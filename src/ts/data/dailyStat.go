package data

import (
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type (
	DailyStat struct {
		Id         DailyStatId `bson:"_id"`
		DailyCount `bson:"DailyCount,inline"`
	}
	DailyStatId struct {
		Date   time.Time    `bson:"Date"`
		Source DeviceSource `bson:"Source,inline"`
	}
	DailyCount struct {
		//Кол-во регистраций новых устройств
		DeviceRegCount int `bson:"RegCount"`
		//Кол-во первых запусков (Первых установок)
		FLaunchCount int `bson:"FLaunchCount"`
		//Кол-во пере-установок
		ReinstallCount int `bson:"ReinstallCount"`
		//Кол-во обновлений
		UpgradeCount int `bson:"UpgradeCount"`
		//Кол-во уникальных запусков за день
		ULaunchCount int `bson:"ULaunchCount"`
		//Total Launch кол-во всех запусков
		TLaunchCount int `bson:"TLaunchCount"`
		//Кол-во лайков
		LikeVideoCount     int `bson:"LikeVideoCount"`
		ViewVideoCount     int `bson:"ViewVideoCount"`
		DownloadVideoCount int `bson:"DownloadVideoCount"`
		PushSendedCount    int `bson:"PushSendedCount"`
		PushClickCount     int `bson:"PushClickCount"`
		VideoCommentCount  int `bson:"VideoCommentCount"`
		PhotoCommentCount  int `bson:"PhotoCommentCount"`
	}
)

//new or from queue
func NewDailyStat(src *DeviceSource) *DailyStat {
	s := &DailyStat{}
	s.Id = DailyStatId{
		Date: time.Now().UTC().Truncate(24 * time.Hour),
	}
	if src != nil {
		s.Id.Source = *src
	} else {
		s.Id.Source = DeviceSource{}
	}

	s.DailyCount = DailyCount{}
	return s
}

//todo queue with channel and merging
func (s *DailyStat) UpsertInc() error {
	_, err := Context.DailyStat.UpsertId(
		fmt.Sprintf("%v", s.Id),
		bson.M{"$inc": s.DailyCount, "$setOnInsert": s.Id})
	return err
}
