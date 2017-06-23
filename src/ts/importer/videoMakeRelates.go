package main

import (
	"errors"
	"fmt"
	"github.com/raoptimus/gserv/config"
	"github.com/raoptimus/gserv/service"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"strings"
	"time"
	"ts/data"
)

const VIDEO_REL_LIMIT = 100

type (
	VideoMakeRelates struct {
		working bool
	}
)

func (s *VideoMakeRelates) Resume() {
	enabled := config.Bool("MakeVideoRelatesVideoDaemon", false)
	if !enabled {
		return
	}
	if s.working {
		return
	}

	s.working = true
	service.Go(s.work)
}

func (s *VideoMakeRelates) work() {
	fmt.Println("-> Start video remake relates")
	defer func() {
		//TODO: race condition. fixme
		s.working = false
	}()

	sleep := config.Duration("VideoMakeRelatesSleepDuration", 1*time.Second)
	numWorkers := config.Int("VideoMakeRelatesNumWorkers", 1)

	ch := make(chan *data.Video)
	done := make(chan int, numWorkers)

	for i := 0; i < numWorkers; i++ {
		go s.startWorker(ch, done, sleep)
	}
	s.feedWorkers(ch)
	close(ch) // stop workers
	processed := 0
	for i := 0; i < numWorkers; i++ {
		processed += <-done
	}

	fmt.Println("<- Finish video remake relates =", processed)
}

func (s *VideoMakeRelates) find() *mgo.Query {
	return data.Context.Videos.
		Find(bson.M{"Filters": "*"}).
		Select(bson.M{"_id": 1, "Keywords": 1}).
		Sort("-PublishedDate")
}

func (s *VideoMakeRelates) startWorker(ch chan *data.Video, done chan int, sleep time.Duration) {
	processed := 0
	for v := range ch {
		processed++
		if s.hasValidRelates(v) {
			continue
		}
		if err := s.Update(v.Id, v.Keywords); err != nil {
			log.Err(err.Error())
		}
		time.Sleep(sleep)
	}
	done <- processed
}

func (s *VideoMakeRelates) hasValidRelates(v *data.Video) bool {
	if len(v.Related) == 0 {
		return false
	}
	c, err := data.Context.Videos.Find(bson.M{"_id": bson.M{"$in": v.Related}, "Filters": "published"}).Count()
	// skip futher processin on error, but make log record
	if err != nil {
		log.Err(err.Error())
		return true
	}
	return c == len(v.Related)
}

func (s *VideoMakeRelates) feedWorkers(ch chan *data.Video) {
	q := s.find()
	count, _ := q.Count()
	s.debug("found: %d videos", count)
	var v = new(data.Video)
	iter := q.Iter()
	for iter.Next(v) {
		s.debug("feeding video %v to worker", v.Id)
		ch <- v
		v = new(data.Video)
	}
	if err := iter.Err(); err != nil {
		s.debug("iter error: %v", err)
		if err != mgo.ErrNotFound {
			log.Err("Error video remake relates: " + err.Error())
		}
	}
	iter.Close()
}

func (s *VideoMakeRelates) Update(videoId int, videoTags []string) (err error) {
	text := strings.Join(videoTags, " ")
	var res []data.Video
	err = data.Context.Videos.
		Find(bson.M{"$text": bson.M{"$search": text}, "Filters": "published"}).
		Select(bson.M{"score": bson.M{"$meta": "textScore"}, "_id": 1}).
		Sort("$textScore:score").
		Limit(VIDEO_REL_LIMIT + 1).
		All(&res)
	if err != nil {
		log.Err("Error video update relates:" + err.Error())
		return
	}

	ret := make([]int, 0)

	for _, id := range res {
		if id.Id == videoId {
			continue
		}
		ret = append(ret, id.Id)
	}

	if len(ret) == 0 {
		return errors.New(fmt.Sprintf("Video relateds not found for video %d, tags %v", videoId, videoTags))
	}

	return data.Context.Videos.UpdateId(videoId, bson.M{
		"$set": bson.M{
			"Related":       ret,
			"UpRelatesDate": time.Now().UTC(),
		},
	})
}

func (s *VideoMakeRelates) debug(format string, args ...interface{}) {
	if false {
		fmt.Printf("VideoMakeRelates: "+format+"\n", args...)
	}
}
