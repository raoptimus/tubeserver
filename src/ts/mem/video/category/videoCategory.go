package category

import (
	"errors"
	"fmt"
	"strconv"
	"time"
	"ts/data"
)

type (
	VideoCategory struct {
		*videoCategoryCache
	}
)

func NewVideoCategory() (*VideoCategory, error) {
	c := &VideoCategory{}
	c.videoCategoryCache = NewVideoCategoryCache()
	if err := c.load(); err != nil {
		return nil, err
	}
	go c.reload()
	return c, nil
}

func (s *VideoCategory) GetSourceIds() []string {
	list := s.videoCategoryCache.FindAll()
	ids := make([]string, 0)

	for _, cat := range list {
		for _, id := range cat.SourceId {
			ids = append(ids, strconv.Itoa(id))
		}
	}
	return ids
}

func (s *VideoCategory) FindSource(srcId int) (*data.Category, error) {
	list := s.videoCategoryCache.FindAll()

	for _, cat := range list {
		for _, id := range cat.SourceId {
			if id == srcId {
				return cat, nil
			}
		}
	}

	return nil, errors.New("Not found category " + strconv.Itoa(srcId))
}

func (s *VideoCategory) load() error {
	fmt.Println("loading video categories")
	var list []data.Category
	err := data.Context.VideoCategory.Find(nil).All(&list)
	if err != nil {
		return err
	}
	for _, c := range list {
		s.videoCategoryCache.UpsertId(c.Id, c)
	}
	fmt.Printf("loaded %d video categories\n", len(list))
	return nil
}

func (s *VideoCategory) reload() {
	for {
		time.Sleep(time.Duration(3 * time.Hour))
		s.load()
	}
}
