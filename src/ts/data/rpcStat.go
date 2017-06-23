package data

import (
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type (
	RpcStat struct {
		Id       RpcStatId       `bson:"_id"`
		Counters RpcStatCounters `bson:"Counters,inline"`
	}
	RpcStatId struct {
		Date   time.Time     `bson:"Date"`
		Source RpcStatSource `bson:"Source,inline"`
	}
	RpcStatSource struct {
		Ver    string `bson:"Ver"`
		Method string `bson:"Method"`
	}
	RpcStatCounters struct {
		//Кол-во успешных вызовов метода
		SuccessCount int `bson:"SuccessCount"`
		//Кол-во вызовов метода с ошибкой
		ErrorCount int `bson:"ErrorCount"`
	}
)

func NewRpcStat(src *RpcStatSource) *RpcStat {
	s := &RpcStat{}
	s.Id = RpcStatId{
		Date: time.Now().UTC().Truncate(24 * time.Hour),
	}
	if src != nil {
		s.Id.Source = *src
	} else {
		s.Id.Source = RpcStatSource{}
	}

	s.Counters = RpcStatCounters{}
	return s
}

//todo queue with channel and merging
func (s *RpcStat) UpsertInc() error {
	_, err := Context.RpcStat.UpsertId(
		fmt.Sprintf("%v", s.Id),
		bson.M{"$inc": s.Counters, "$setOnInsert": s.Id})
	return err
}
