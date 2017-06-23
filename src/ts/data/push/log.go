package push

import (
	"gopkg.in/mgo.v2/bson"
	"time"
	"ts/data"
	"ts/mongodb"
)

type (
	Log struct {
		Id         int       `bson:"_id"`
		Token      string    `bson:"Token"`
		Status     Status    `bson:"Status"`
		TaskId     int       `bson:"TaskId"`
		Error      string    `bson:"Error"`
		SendedDate time.Time `bson:"SendedDate"`
	}
)

type Status int

const (
	StatusSuccess Status = iota
	StatusError
)

func WriteLog(d *data.Device, st Status, taskId int, lerr string) error {
	//todo inc stats and device.stats
	id, err := mongodb.GetNewIncId(data.Context.PushLog)
	if err != nil {
		return err
	}
	l := Log{
		Id:         id,
		Token:      d.Id,
		Status:     st,
		TaskId:     taskId,
		Error:      lerr,
		SendedDate: time.Now().UTC(),
	}
	err = data.Context.PushLog.Insert(l)
	if err != nil {
		return err
	}

	if st == StatusSuccess {
		stat := data.NewDailyStat(&d.Source.DeviceSource)
		stat.PushSendedCount++
		err := stat.UpsertInc()
		if err != nil {
			return err
		}

		update := bson.M{"$inc": bson.M{"PushSendedCount": 1}}
		err = data.Context.Devices.UpdateId(d.Id, update)
		if err == nil {
			//todo inc in manager after send, inc after finish task
			err = data.Context.PushTask.UpdateId(taskId, update)

		}
	}
	return nil
}
