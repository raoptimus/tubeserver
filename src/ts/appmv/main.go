package main

import (
	"fmt"
	"github.com/raoptimus/gserv/config"
	"github.com/raoptimus/gserv/service"
	"github.com/raoptimus/rlog"
	"os"
	"ts/data"
)

var log *rlog.Logger

func main() {
	if service.Exists() {
		os.Exit(0)
	}
	var err error
	log, err = rlog.NewLoggerDial(rlog.LoggerTypeMongoDb, "", config.String("MongoLogServer", config.String("MongoAllServer", "localhost/Logs")), "")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	service.Init(&service.BaseService{
		Start:  start,
		Logger: log,
	})
	service.Start(false)
}

func start() {
	data.Init(false)
	if err := new(Worker).Start(); err != nil {
		log.Println(err)
		fmt.Println(err)
	} else {
		fmt.Println("Worker is done")
	}
}
