package country

import (
	"sync"
	"ts/data"
)

type (
	countryCache struct {
		mutex sync.RWMutex
		items map[string]*data.Country
	}
)

var instance *countryCache

func NewCountryCache() *countryCache {
	if instance != nil {
		return instance
	}
	instance = &countryCache{
		items: make(map[string]*data.Country),
	}
	return instance
}

func (s *countryCache) UpsertCountry(value data.Country) *data.Country {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	it, ok := s.items[value.Code]
	if ok {
		*it = value
		return it
	} else {
		s.items[value.Code] = &value
		return &value
	}
}

func (s *countryCache) IsExists(checkedCountry string) bool {
	s.mutex.RLock()
	_, ok := s.items[checkedCountry]
	s.mutex.RUnlock()
	return ok
}
