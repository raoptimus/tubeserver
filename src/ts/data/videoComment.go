package data

import (
	"errors"
	"time"
	"ts/mongodb"
)

type (
	VideoComment struct {
		Id       int           `bson:"_id"`
		VideoId  int           `bson:"VideoId"`
		UserId   int           `bson:"UserId"`
		Body     string        `bson:"Body"`
		Status   CommentStatus `bson:"Status"`
		PostDate time.Time     `bson:"PostDate"`
		Language Language      `bson:"Language"`
		User     *User         `-`
	}
)

func NewVideoComment(videoId int) (vc *VideoComment, err error) {
	id, err := mongodb.GetNewIncId(Context.VideoComments)

	if err != nil {
		return nil, errors.New("Sequence VideoComment.Id error: " + err.Error())
	}

	return &VideoComment{
		Id:      id,
		VideoId: videoId,
	}, nil
}

func (s *VideoComment) Insert(src *DeviceSource) error {
	err := Context.VideoComments.Insert(s)
	if err != nil {
		return err
	}

	stat := NewDailyStat(src)
	stat.VideoCommentCount++
	return stat.UpsertInc()
}
