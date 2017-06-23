package main

import (
	"github.com/raoptimus/gserv/service"
	"github.com/raoptimus/rlog"
	"ts/data"
	"ts/translator/blacklist"
)

func main() {
	log, _ := rlog.NewLogger(rlog.LoggerTypeStd, "")
	service.Init(&service.BaseService{
		Start:  start,
		Logger: log,
	})
	service.Start(false)
}

func start() {
	data.Init(false)
	blacklist.Init()
	blacklist.CensorAllVideos()
}
