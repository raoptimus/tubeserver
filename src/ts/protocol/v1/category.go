package v1

import (
	"sort"
)

type (
	Category struct {
		Id        int    `json:"Id"`
		Title     string `json:"Title"`
		Slug      string `json:"Slug"`
		ShortDesc string `json:"ShortDesc"`
		LongDesc  string `json:"LongDesc"`
	}
	Categories []*Category

	categorySorter struct {
		Categories
		sort *SortInfo
	}
)

func (s Categories) Page(p *Page) Categories {
	if p == nil {
		return s
	}

	limit, skip := p.Limit, p.Skip

	if limit > 1000 {
		limit = 1000
	}

	if limit > len(s) {
		limit = len(s)
	}

	return s[skip:limit]
}

//todo see http://godoc.org/github.com/bradfitz/slice#SortInterface

func (s Categories) Sort(sortInfo *SortInfo) Categories {
	if sortInfo == nil {
		return s
	}
	sort.Sort(categorySorter{Categories: s, sort: sortInfo})
	return s
}

func (s categorySorter) Len() int {
	return len(s.Categories)
}

func (s categorySorter) Less(i, j int) bool {
	switch s.sort.Field {
	case "Title":
		return s.sort.LessString(s.Categories[i].Title, s.Categories[j].Title)
	default:
		return s.sort.LessInt(s.Categories[i].Id, s.Categories[j].Id)
	}
}

func (s categorySorter) Swap(i, j int) {
	s.Categories[i], s.Categories[j] = s.Categories[j], s.Categories[i]
}
