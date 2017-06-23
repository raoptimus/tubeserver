package data

import (
	"errors"
	"time"
	"ts/mongodb"
)

type (
	PhotoAlbumComment struct {
		Id           int           `bson:"_id"`
		PhotoAlbumId int           `bson:"PhotoAlbumId"`
		UserId       int           `bson:"UserId"`
		Body         string        `bson:"Body"`
		Status       CommentStatus `bson:"Status"`
		PostDate     time.Time     `bson:"PostDate"`
		Language     Language      `bson:"Language"`
		User         *User         `-`
	}
)

func NewPhotoAlbumComment(photoAlbumId int) (pc *PhotoAlbumComment, err error) {
	id, err := mongodb.GetNewIncId(Context.PhotoAlbumComments)

	if err != nil {
		return nil, errors.New("Sequence PhotoAlbumComment.Id error: " + err.Error())
	}

	return &PhotoAlbumComment{
		Id:           id,
		PhotoAlbumId: photoAlbumId,
	}, nil
}
