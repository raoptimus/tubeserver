package category

import (
	"errors"
	"sync"
	"ts/data"
)

type (
	videoCategoryCache struct {
		mutex sync.RWMutex
		items map[int]*data.Category
	}
)

var ErrNotFound = errors.New("Not found")
var instance *videoCategoryCache

func NewVideoCategoryCache() *videoCategoryCache {
	if instance != nil {
		return instance
	}
	instance = &videoCategoryCache{
		items: make(map[int]*data.Category),
	}
	return instance
}

func (s *videoCategoryCache) FindId(id int) (*data.Category, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	v, ok := s.items[id]
	if ok {
		return v, nil
	}
	return nil, ErrNotFound
}

func (s *videoCategoryCache) FindAll() []*data.Category {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	result := make([]*data.Category, len(s.items))
	i := 0
	for _, it := range s.items {
		result[i] = it
		i++
	}
	return result
}

func (s *videoCategoryCache) UpsertId(id int, value data.Category) *data.Category {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	it, ok := s.items[id]
	if ok {
		*it = value
		return it
	} else {
		s.items[id] = &value
		return &value
	}
}

func (s *videoCategoryCache) ExistsId(id int) bool {
	s.mutex.RLock()
	_, ok := s.items[id]
	s.mutex.RUnlock()
	return ok
}

func (s *videoCategoryCache) DeleteId(id int) {
	s.mutex.Lock()
	delete(s.items, id)
	s.mutex.Unlock()
}
