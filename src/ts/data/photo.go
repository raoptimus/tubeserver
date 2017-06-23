package data

import (
	"errors"
	"fmt"
	"hash/crc32"
	"ts/mongodb"
)

type PhotoHistoryType int

const (
	PhotoHistoryTypeLike PhotoHistoryType = iota
)

type (
	Photo struct {
		Id        int    `bson:"_id"`
		AlbumId   int    `bson:"PhotoAlbumId"`
		W         int    `bson:"W"`
		H         int    `bson:"H"`
		Path      string `bson:"Path"`
		Name      string `bson:"Name"`
		Hash      string `bson:"Hash"`
		Ext       string `bson:"Ext"`
		LikeCount int    `bson:"LikeCount"`
	}
)

func NewPhoto(albumId int, hash string, w, h int) (p *Photo, err error) {
	id, err := mongodb.GetNewIncId(Context.Photos)

	if err != nil {
		return nil, errors.New("Sequence Photo.Id error: " + err.Error())
	}

	t := crc32.MakeTable(crc32.IEEE)
	crc := crc32.New(t)
	crc.Write([]byte(hash))
	sh := fmt.Sprintf("%x", crc.Sum32())
	path := fmt.Sprintf("/%s/%s/%s/", sh[0:1], sh[1:3], sh)

	p = &Photo{
		Id:      id,
		AlbumId: albumId,
		Path:    "Images/Photos" + path,
		Hash:    hash,
		Name:    "640x0.jpg",
		Ext:     "jpg",
		W:       w,
		H:       h,
	}

	return p, nil
}

func (s *Photo) Url() string {
	return "http://cdn-i1..../" + s.Path + s.Name
}
