package v1

import (
	"errors"
	"sort"
)

type VideoAction string

const (
	VideoActionAdded   VideoAction = "added"
	VideoActionRemoved VideoAction = "removed"
)

type (
	Channel struct {
		Id    int    `json:"Id"`
		Title string `json:"Title"`
	}
	Actor struct {
		Name string `json:"Name"`
	}
	Video struct {
		Id            int            `json:"Id"`
		Title         string         `json:"Title"`
		Slug          string         `json:"Slug"`
		Desc          string         `json:"Desc"`
		Duration      int64          `json:"Duration"`
		CategoryId    int            `json:"CategoryId"`
		PublishedDate int64          `json:"PublishedDate"`
		UpdateDate    int64          `json:"UpdateDate"`
		IsPremium     bool           `json:"IsPremium"`
		ViewCount     int            `json:"ViewCount"`
		LikeCount     int            `json:"LikeCount"`
		DownloadCount int            `json:"DownloadCount"`
		CommentCount  int            `json:"CommentCount"`
		Images        VideoMediaList `json:"Images"`
		IsLiked       bool           `json:"IsLiked"`
		Featured      bool           `json:"Featured"`
		//advanced fields
		AddedDate int64            `json:"AddedDate,omitempty"`
		Category  Category         `json:"Category,omitempty"`
		Files     VideoMediaList   `json:"Files,omitempty"`
		Comments  VideoCommentList `json:"Comments,omitempty"`
		Related   []int            `json:"Related,omitempty"`
		Tags      []string         `json:"Tags,omitempty"`
		Actors    ActorList        `json:"Actors,omitempty"`
		Channel   Channel          `json:"Channel,omitempty"`
	}
	VideoMedia struct {
		Format string `json:"Format"`
		Url    string `json:"Url"`
		AUrl   string `json:"AUrl,omitempty"` //alternate url
		Size   int    `json:"Size,omitempty"`
	}
	VideoComment struct {
		Id       int    `json:"Id,omitempty"`
		VideoId  int    `json:"VideoId"`
		Author   string `json:"Author,omitempty"`
		Avatar   string `json:"Avatar,omitempty"`
		Body     string `json:"Body"`
		PostDate int64  `json:"PostDate,omitempty"`
	}
	VideoActionResponse struct {
		Action     VideoAction `json:"Action,omitempty"`
		TotalCount int         `json:"TotalCount"`
	}

	VideoList        []*Video
	VideoMediaList   []*VideoMedia
	VideoCommentList []*VideoComment
	ActorList        []*Actor
)

func (s *VideoComment) Validate() error {
	if s.VideoId == 0 {
		return errors.New("Video id is empty")
	}
	if len(s.Body) < 10 {
		return errors.New("Body is small, len < 10")
	}
	return nil
}

func (s VideoCommentList) Sort() VideoCommentList {
	sort.Sort(s)
	return s
}

func (s VideoCommentList) Len() int {
	return len(s)
}

func (s VideoCommentList) Less(i, j int) bool {
	return s[i].PostDate > s[j].PostDate
}

func (s VideoCommentList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s VideoMediaList) Sort() VideoMediaList {
	sort.Sort(s)
	return s
}

func (s VideoMediaList) Len() int {
	return len(s)
}

func (s VideoMediaList) Less(i, j int) bool {
	return s[i].Format < s[j].Format
}

func (s VideoMediaList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
