package mem

import (
	"errors"
	"sync"
)

type (
	//replace and create new file cache with other struct(value) and id
	CacheId   interface{}
	CacheItem interface{}
	//
	cache struct {
		mutex sync.RWMutex
		items map[CacheId]*CacheItem
	}
)

var ErrNotFound = errors.New("Not found")
var instance *cache

func NewCache() *cache {
	if instance != nil {
		return instance
	}
	instance = &cache{
		items: make(map[CacheId]*CacheItem),
	}
	return instance
}

func (s *cache) FindId(id CacheId) (*CacheItem, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	v, ok := s.items[id]
	if ok {
		return v, nil
	}
	return nil, ErrNotFound
}

func (s *cache) FindAll() []*CacheItem {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	result := make([]*CacheItem, len(s.items))
	i := 0
	for _, it := range s.items {
		result[i] = it
		i++
	}
	return result
}

func (s *cache) UpsertId(id CacheId, value CacheItem) *CacheItem {
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

func (s *cache) ExistsId(id CacheId) bool {
	s.mutex.RLock()
	_, ok := s.items[id]
	s.mutex.RUnlock()
	return ok
}

func (s *cache) DeleteId(id CacheId) {
	s.mutex.Lock()
	delete(s.items, id)
	s.mutex.Unlock()
}
