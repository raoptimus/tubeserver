package main

import (
	"errors"
	"ts/data"
	api "ts/protocol/v1"
)

type CategoryController int

func (s *CategoryController) List(req *api.Request, list *api.Categories) error {
	sort, page, lang := req.Sort, req.Page, req.Lang
	dbList := Mem.VideoCategory.FindAll()

	if len(dbList) == 0 {
		return errors.New(string(api.StatusNotFound))
	}

	*list = make(api.Categories, len(dbList))

	for i, c := range dbList {
		(*list)[i] = &api.Category{
			Id:        c.Id,
			Title:     c.Title.Get(lang, false),
			Slug:      c.Title.Get(lang, false),
			ShortDesc: c.ShortDesc,
			LongDesc:  c.LongDesc,
		}
	}

	*list = list.Sort(sort)
	{ // fake categories
		top := &api.Category{
			Id:    -1,
			Title: "TOP",
			Slug:  "top",
		}
		if req.Lang == data.LanguageRussian {
			top.Title = "ТОП"
		}
		*list = append(api.Categories{top}, *list...)
	}
	*list = list.Page(page)
	return nil

}
