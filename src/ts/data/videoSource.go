package data

import (
	"gopkg.in/mgo.v2/bson"
	"strconv"
	"ts/mongodb"
)

type Id int

func (s *Id) SetBSON(raw bson.Raw) error {
	var i int
	if raw.Kind == 0x02 {
		var ss string
		err := raw.Unmarshal(&ss)
		if err != nil {
			return err
		}
		i, err = strconv.Atoi(ss)
		if err != nil {
			return err
		}
	} else {
		err := raw.Unmarshal(&i)
		if err != nil {
			return err
		}
	}
	*s = Id(i)
	return nil
}

type (
	VideoSource struct {
		Id                    bson.ObjectId `bson:"_id"`
		SourceId              int           `bson:"SourceId"`
		Domain                string        `bson:"Domain"`
		ScreenshotCount       int           `bson:"ScreenshotCount"`
		ScreenshotSelectIndex Id            `bson:"ScreenshotSelectIndex"`
	}
)

func NewVideoSource(srcId int, domain string, screenshotCount, screenshotIndexSelect int) *VideoSource {
	id := mongodb.GenerateObjectId(strconv.Itoa(srcId), domain)

	return &VideoSource{
		Id:                    id,
		SourceId:              srcId,
		Domain:                domain,
		ScreenshotCount:       screenshotCount,
		ScreenshotSelectIndex: Id(screenshotIndexSelect),
	}
}
