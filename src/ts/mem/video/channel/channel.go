package channel

import (
	"fmt"
	"time"
	"ts/data"
)

type (
	Channel struct {
		*channelCache
	}
)

func NewChannel() (*Channel, error) {
	c := &Channel{}
	c.channelCache = NewChannelCache()
	if err := c.load(); err != nil {
		return nil, err
	}
	go c.reload()
	return c, nil
}

func (s *Channel) load() error {
	fmt.Println("loading video channels")
	var list []data.Channel
	err := data.Context.Channel.Find(nil).All(&list)
	if err != nil {
		return err
	}
	for _, c := range list {
		s.channelCache.UpsertId(c.Id, c)
	}
	fmt.Printf("loaded %d video channels\n", len(list))
	return nil
}

func (s *Channel) reload() {
	for {
		time.Sleep(time.Duration(3 * time.Hour))
		s.load()
	}
}
