package main

import (
	"fmt"
	"github.com/raoptimus/gserv/config"
	"sync"
	api "ts/protocol/v1"
)

type (
	IpTable struct {
		sync.RWMutex
		items api.AddrList
	}
)

var table *IpTable

func NewIpTable() *IpTable {
	if table != nil {
		return table
	}
	table = &IpTable{}
	if err := table.load(); err != nil {
		log.Fatalln(err)
		return nil
	}
	config.OnAfterLoad("ServerIp.Reload", table.reload)
	return table
}

func (s *IpTable) Add(ip string, port int, server string) {
	s.Lock()
	defer s.Unlock()
	s.items = append(s.items,
		&api.Addr{
			Ip:     ip,
			Port:   port,
			BaseIp: server,
		})
}

func (s *IpTable) List() api.AddrList {
	s.RLock()
	defer s.RUnlock()
	return s.items[:len(s.items)]
}

//PRIVATE METHODS

func (s *IpTable) load() error {
	var list api.AddrList
	err := config.Object("ServerIp", &list)
	if err != nil {
		return err
	}
	s.Lock()
	s.items = list
	s.Unlock()
	return nil
}

func (s *IpTable) reload() {
	err := s.load()
	if err != nil {
		fmt.Println(err)
		log.Println(err)
	}
}
