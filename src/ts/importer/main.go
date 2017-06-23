package main

import (
	"fmt"
	"github.com/raoptimus/gserv/config"
	"github.com/raoptimus/gserv/service"
	"github.com/raoptimus/rlog"
	"os"
	"strings"
	"ts/data"
	"unicode"
)

const DOMAIN string = "..."

var log *rlog.Logger
var Mem *memContext

func main() {
	if service.Exists() {
		os.Exit(0)
	}
	var err error
	log, err = rlog.NewLoggerDial(rlog.LoggerTypeStd, "", config.String("MongoLogServer", config.String("MongoAllServer", "localhost/Logs")), "")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	service.Init(&service.BaseService{
		Start:  start,
		Logger: log,
	})
	service.Start(true)
}

func start() {
	data.Init(false)
	var err error
	Mem, err = NewMemContext()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	//		bt.GoCatchPanic(new(PhotoImporter).Start)
	vi, err := NewVideoImporter()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	service.Go(vi.Start)
}

func uniqLower(tags string) []string {
	tagMap := make(map[string]bool)

	for _, t := range strings.Split(strings.ToLower(tags), ",") {
		tg := strings.Trim(t, " ")

		if tg == "" {
			continue
		}

		tagMap[tg] = true
	}

	res := make([]string, len(tagMap))
	ti := 0

	for t, _ := range tagMap {
		res[ti] = t
		ti++
	}

	return res
}

func generateSlug(str string) (slug string) {
	return strings.Map(func(r rune) rune {
		switch {
		case r == ' ', r == '-':
			return '-'
		case r == '_', unicode.IsLetter(r), unicode.IsDigit(r):
			return r
		default:
			return -1
		}
		return -1
	}, strings.ToLower(strings.TrimSpace(str)))
}
