package channel

import (
	"errors"
	"sync"
	"ts/data"
)

type (
	channelCache struct {
		mutex sync.RWMutex
		items map[int]*data.Channel
	}
)

var ErrNotFound = errors.New("Not found")
var instance *channelCache

func NewChannelCache() *channelCache {
	if instance != nil {
		return instance
	}
	instance = &channelCache{
		items: make(map[int]*data.Channel),
	}
	return instance
}

func (s *channelCache) FindId(id int) (*data.Channel, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	v, ok := s.items[id]
	if ok {
		return v, nil
	}
	return nil, ErrNotFound
}

func (s *channelCache) FindAll() []*data.Channel {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	result := make([]*data.Channel, len(s.items))
	i := 0
	for _, it := range s.items {
		result[i] = it
		i++
	}
	return result
}

func (s *channelCache) UpsertId(id int, value data.Channel) *data.Channel {
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

func (s *channelCache) ExistsId(id int) bool {
	s.mutex.RLock()
	_, ok := s.items[id]
	s.mutex.RUnlock()
	return ok
}

func (s *channelCache) DeleteId(id int) {
	s.mutex.Lock()
	delete(s.items, id)
	s.mutex.Unlock()
}
