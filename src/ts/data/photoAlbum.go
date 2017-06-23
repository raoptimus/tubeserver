package data

import (
	"errors"
	"gopkg.in/mgo.v2/bson"
	"time"
	"ts/mongodb"
)

type (
	PhotoAlbum struct {
		Id            int                  `bson:"_id"`
		Source        *PhotoAlbumSource    `bson:"Source"`
		Photos        []*Photo             `bson:"Photos"`
		Comments      []*PhotoAlbumComment `bson:"Comments"`
		Title         TextList             `bson:"Title"`
		Desc          TextList             `bson:"Desc"`
		CategoryId    int                  `bson:"CategoryId"`
		PublishedDate time.Time            `bson:"PublishedDate"`
		IsPremium     bool                 `bson:"IsPremium"`
		LikeCount     int                  `bson:"LikeCount"`
		ViewCount     int                  `bson:"ViewCount"`
		CommentCount  int                  `bson:"CommentCount"`
		PhotoCount    int                  `bson:"PhotoCount"`
		MainPhotoId   int                  `bson:"MainPhotoId"`
		ImportDate    time.Time            `bson:"ImportDate"`
		Featured      bool                 `bson:"Featured"`
		Tags          []string             `bson:"Tags"`
	}
)

func (s *PhotoAlbum) LoadById(id int) error {
	err := Context.PhotoAlbums.FindId(id).One(&s)
	if err != nil {
		return err
	}
	return s.setAllUsersForComments()
}

func (s *PhotoAlbum) setAllUsersForComments() error {
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

	return nil
}

func NewPhotoAlbum() (v *PhotoAlbum, err error) {
	id, err := mongodb.GetNewIncId(Context.PhotoAlbums)

	if err != nil {
		return nil, errors.New("Sequence PhotoAlbum.Id error: " + err.Error())
	}

	return &PhotoAlbum{
		Id:         id,
		ImportDate: time.Now().UTC(),
	}, nil
}

func PhotoAlbumExists(id int) (bool, error) {
	n, err := Context.PhotoAlbums.FindId(id).Count()
	return n > 0, err
}
