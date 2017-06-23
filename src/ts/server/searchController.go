package main

import (
	//"errors"
	//"gopkg.in/mgo.v2/bson"
	//"strings"
	"ts/data"
	api "ts/protocol/v1"
)

type (
	SearchController struct {
	}
)

func (s *SearchController) Find(req *api.Request, list *api.SearchQueryList) error {
	/*
	if req.Query == nil {
		return errors.New("Query can't be nil")
	}
	id, ok := (*req.Query)["Id"].(string)
	if !ok || len(id) < 3 {
		return errors.New("Query can't be blank")
	}

	where := bson.M{
		"_id": bson.RegEx{
			Pattern: strings.ToLower(id) + ".*",
		},
	}

	var qList []data.SearchQuery
	err := data.Context.SearchQueries.Find(where).Sort("-ResultCount", "-SearchCount").Limit(10).All(&qList)
	if err != nil {
		return err
	}
	*list = make(api.SearchQueryList, len(qList))
	for i, q := range qList {
		(*list)[i] = s.convertObject(&q)
	}
	*/
	*list = make(api.SearchQueryList, 0)
	return nil
}

func (s *SearchController) convertObject(in *data.SearchQuery) *api.SearchQuery {
	return &api.SearchQuery{
		Text:        in.Id,
		ResultCount: in.ResultCount,
	}
}
