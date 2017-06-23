package country

import (
	"fmt"
	"time"
	"ts/data"
)

type (
	Country struct {
		*countryCache
	}
)

func NewCountry() (*Country, error) {
	c := &Country{}
	c.countryCache = NewCountryCache()
	if err := c.load(); err != nil {
		return nil, err
	}
	go c.reload()
	return c, nil
}

func (s *Country) load() error {
	fmt.Println("loading countries")
	var list []data.Country
	err := data.Context.Country.Find(nil).All(&list)
	if err != nil {
		return err
	}
	for _, c := range list {
		s.countryCache.UpsertCountry(c)
	}
	fmt.Printf("loaded %d countries\n", len(list))
	return nil
}

func (s *Country) reload() {
	for {
		time.Sleep(time.Duration(3 * time.Hour))
		s.load()
	}
}

func (s *Country) IsExists(checkedCountry string) bool {
    return s.countryCache.IsExists(checkedCountry)
}
