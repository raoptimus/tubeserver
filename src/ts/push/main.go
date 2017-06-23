package main

import (
	"fmt"
	"github.com/raoptimus/gserv/config"
	"github.com/raoptimus/gserv/service"
	"github.com/raoptimus/rlog"
	"os"
	"time"
	"ts/data"
	"ts/mem/targeting/country"
)

var log *rlog.Logger
var MemCountryList *country.Country

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

	m := NewManager()
	service.Init(&service.BaseService{
		Start:  m.Start,
		Stop:   m.Stop,
		Logger: log,
	})

	data.Init(false)
	countries, err := country.NewCountry()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	MemCountryList = countries

	service.Go(m.ForceUpdater)
	service.Start(true)
}

func now() time.Time {
	return time.Now().UTC()
}
