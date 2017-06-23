package v1

import (
	"errors"
	"sort"
)

type PhotoAction string

const (
	PhotoActionAdded   PhotoAction = "added"
	PhotoActionRemoved PhotoAction = "removed"
)

type (
	PhotoAlbum struct {
		Id            int                `json:"Id"`
		Title         string             `json:"Title"`
		Desc          string             `json:"Desc"`
		CategoryId    int                `json:"CategoryId"`
		PublishedDate int64              `json:"PublishedDate"`
		IsPremium     bool               `json:"IsPremium"`
		LikeCount     int                `json:"LikeCount"`
		ViewCount     int                `json:"ViewCount"`
		CommentCount  int                `json:"CommentCount"`
		PhotoCount    int                `json:"PhotoCount"`
		IsLiked       bool               `json:"IsLiked"`
		Photos        PhotoMainList      `json:"Photos,omitempty"`
		Comments      PhotoAlbumComments `json:"Comments,omitempty"`
	}

	PhotoMain struct {
		Url string `json:"Url"`
	}

	Photo struct {
		Id        int    `json:"Id"`
		AlbumId   int    `json:"AlbumId"`
		Url       string `json:"Url"`
		IsLiked   bool   `json:"IsLiked"`
		LikeCount int    `json:"LikeCount"`
	}

	PhotoAlbumComment struct {
		Id           int    `json:"Id,omitempty"`
		PhotoAlbumId int    `json:"PhotoAlbumId"`
		Author       string `json:"Author,omitempty"`
		Avatar       string `json:"Avatar,omitempty"`
		Body         string `json:"Body"`
		PostDate     int64  `json:"PostDate,omitempty"`
	}

	PhotoActionResponse struct {
		Action     PhotoAction `json:"Action,omitempty"`
		TotalCount int         `json:"TotalCount"`
	}

	PhotoAlbumList     []*PhotoAlbum
	PhotoList          []*Photo
	PhotoMainList      []*PhotoMain
	PhotoAlbumComments []*PhotoAlbumComment
)

func (s *PhotoAlbumComment) Validate() error {
	if s.PhotoAlbumId == 0 {
		return errors.New("Video id is empty")
	}
	if len(s.Body) < 10 {
		return errors.New("Body is small, len < 10")
	}
	return nil
}

func (s PhotoAlbumComments) Sort() PhotoAlbumComments {
	sort.Sort(s)
	return s
}

func (s PhotoAlbumComments) Len() int {
	return len(s)
}

func (s PhotoAlbumComments) Less(i, j int) bool {
	return s[i].PostDate > s[j].PostDate
}

func (s PhotoAlbumComments) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
