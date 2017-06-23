package data

import (
	"errors"
	"fmt"
	"github.com/raoptimus/gserv/config"
	"gopkg.in/mgo.v2/bson"
	"time"
	"ts/mongodb"
)

type (
	Video struct {
		Id            int                `bson:"_id"`
		Source        *VideoSource       `bson:"Source"`
		Files         []*VideoFile       `bson:"Files"`
		Screenshots   []*VideoScreenshot `bson:"Screenshots"`
		Comments      []*VideoComment    `bson:"Comments"`
		Title         TextList           `bson:"Title"`
		Slug          TextList           `bson:"Slug"`
		Desc          TextList           `bson:"Desc"`
		Duration      int64              `bson:"Duration"`
		CategoryId    int                `bson:"CategoryId"`
		PublishedDate time.Time          `bson:"PublishedDate"`
		Related       []int              `bson:"Related"`
		Tags          TagsList           `bson:"Tags"`
		Keywords      []string           `bson:"Keywords"`
		Filters       []string           `bson:"Filters"`
		AddedDate     time.Time          `bson:"AddedDate"`
		UpdateDate    time.Time          `bson:"UpdateDate"`
		ChannelId     int                `bson:"ChannelId"`
		Actors        []string           `bson:"Actors"`
		FilesId       string             `bson:"FilesId"`
		VideoCounters `bson:"VideoCounters,inline"`
	}
	VideoScreenshot struct {
		W      int               `bson:"W"`
		H      int               `bson:"H"`
		Image  bson.Binary       `bson:"Image"`
		Thumbs []*VideoThumbnail `bson:"Thumbs"`
	}
	VideoThumbnail struct {
		W     int         `bson:"W"`
		H     int         `bson:"H"`
		Image bson.Binary `bson:"Image"`
	}
	VideoCounters struct {
		ViewCount     int `bson:"ViewCount"`
		LikeCount     int `bson:"LikeCount"`
		DownloadCount int `bson:"DownloadCount"`
		CommentCount  int `bson:"CommentCount"`
	}
)

func (s *Video) IsPremium() bool {
	for _, f := range s.Filters {
		if f == "premium" {
			return true
		}
		if f == "!premium" {
			return false
		}
	}
	return false
}
func (s *Video) IsFeatured() bool {
	for _, f := range s.Filters {
		if f == "featured" {
			return true
		}
		if f == "!featured" {
			return false
		}
	}
	return false
}

func (s *Video) GetCommentCount() int {
	return len(s.Comments)
}

func (s *Video) setAllUsersForComments() error {
	userIdMap := make(map[int]bool)
	userIds := make([]int, 0)

	for _, c := range s.Comments {
		if userIdMap[c.UserId] {
			continue
		}

		userIdMap[c.UserId] = true
		userIds = append(userIds, c.UserId)
	}

	var users []User
	err := Context.Users.Find(bson.M{"_id": bson.M{"$in": userIds}}).All(&users)

	if err != nil {
		return err
	}

	for i := 0; i < len(s.Comments); i++ {
		for _, u := range users {
			if u.Id == s.Comments[i].UserId {
				s.Comments[i].User = &u
				break
			}
		}
	}

	for i := 0; i < len(s.Comments); i++ {
		if s.Comments[i].User == nil {
			s.Comments[i].User = &User{}
		}
	}

	return nil
}

func VideoExists(id int) (bool, error) {
	n, err := Context.Videos.FindId(id).Count()
	return n > 0, err
}

func (s *Video) LoadById(id int) error {
	err := Context.Videos.FindId(id).One(&s)
	if err != nil {
		return err
	}
	return s.setAllUsersForComments()
}

func (s *Video) MainScreenshotUrl() string {
	host := config.String("ImageHost", "....com")
	d4 := fmt.Sprintf("%4d", s.Source.SourceId)
	return fmt.Sprintf("http://%s/%s/%s/%d/%dx%d/%010d.jpg",
		host, d4[0:2], d4[2:4], s.Source.SourceId, 640, 480, s.Source.ScreenshotSelectIndex)
}

func NewVideo() (v *Video, err error) {
	id, err := mongodb.GetNewIncId(Context.Videos)

	if err != nil {
		return nil, errors.New("Sequence Video.Id error: " + err.Error())
	}

	return &Video{
		Id:         id,
		Filters:    make([]string, 0),
		AddedDate:  time.Now().UTC(),
		UpdateDate: time.Now().UTC(),
	}, nil
}

func (vc *VideoCounters) Rank() float64 {
	return float64(vc.LikeCount+vc.DownloadCount) / float64(vc.ViewCount)
}
