package main

import (
	"fmt"
	"ts/data"
)

//Перенести категории в базу
func main() {
	n, err := data.Context.VideoCategory.Count()
	if err != nil {
		panic(err)
	}
	if n > 0 {
		panic("VideoCategory collections exists")
	}
	list := []*data.Category{
		&data.Category{
			Id: 1,
			Title: []*data.Text{
				&data.Text{
					Quote:    "...",
					Language: data.LanguageRussian,
				},
				&data.Text{
					Quote:    "...",
					Language: data.LanguageEnglish,
				}},
			SourceId: []int{5},
		},
		&data.Category{
			Id: 2,
			Title: []*data.Text{
				&data.Text{
					Quote:    "...",
					Language: data.LanguageRussian,
				},
				&data.Text{
					Quote:    "...",
					Language: data.LanguageEnglish,
				}},
			SourceId: []int{3},
		},
		&data.Category{
			Id: 3,
			Title: []*data.Text{
				&data.Text{
					Quote:    "...",
					Language: data.LanguageRussian,
				},
				&data.Text{
					Quote:    "...",
					Language: data.LanguageEnglish,
				}},
			SourceId: []int{7},
		},
		&data.Category{
			Id: 4,
			Title: []*data.Text{
				&data.Text{
					Quote:    "...",
					Language: data.LanguageRussian,
				},
				&data.Text{
					Quote:    "...",
					Language: data.LanguageEnglish,
				}},
			SourceId: []int{9, 19, 10, 23, 40, 33, 6, 16, 29, 22, 43, 42, 24, 35, 8, 1, 11, 39, 25},
		},
		&data.Category{
			Id: 5,
			Title: []*data.Text{
				&data.Text{
					Quote:    "...",
					Language: data.LanguageRussian,
				},
				&data.Text{
					Quote:    "...",
					Language: data.LanguageEnglish,
				}},
			SourceId: []int{15, 12},
		},
		&data.Category{
			Id: 6,
			Title: []*data.Text{
				&data.Text{
					Quote:    "...",
					Language: data.LanguageRussian,
				},
				&data.Text{
					Quote:    "...",
					Language: data.LanguageEnglish,
				}},
			SourceId: []int{41, 36},
		},
		&data.Category{
			Id: 7,
			Title: []*data.Text{
				&data.Text{
					Quote:    "...",
					Language: data.LanguageRussian,
				},
				&data.Text{
					Quote:    "...",
					Language: data.LanguageEnglish,
				}},
			SourceId: []int{4},
		},
		&data.Category{
			Id: 8,
			Title: []*data.Text{
				&data.Text{
					Quote:    "...",
					Language: data.LanguageRussian,
				},
				&data.Text{
					Quote:    "...",
					Language: data.LanguageEnglish,
				}},
			SourceId: []int{2},
		},
		&data.Category{
			Id: 9,
			Title: []*data.Text{
				&data.Text{
					Quote:    "...",
					Language: data.LanguageRussian,
				},
				&data.Text{
					Quote:    "...",
					Language: data.LanguageEnglish,
				}},
			SourceId: []int{32},
		},
		&data.Category{
			Id: 10,
			Title: []*data.Text{
				&data.Text{
					Quote:    "...",
					Language: data.LanguageRussian,
				},
				&data.Text{
					Quote:    "...",
					Language: data.LanguageEnglish,
				}},
			SourceId: []int{37},
		},
		&data.Category{
			Id: 11,
			Title: []*data.Text{
				&data.Text{
					Quote:    "...",
					Language: data.LanguageRussian,
				},
				&data.Text{
					Quote:    "...",
					Language: data.LanguageEnglish,
				}},
			SourceId: []int{34},
		},
		&data.Category{
			Id: 12,
			Title: []*data.Text{
				&data.Text{
					Quote:    "...",
					Language: data.LanguageRussian,
				},
				&data.Text{
					Quote:    "...",
					Language: data.LanguageEnglish,
				}},
			SourceId: []int{26},
		},
	}
	for _, c := range list {
		err := data.Context.VideoCategory.Insert(c)
		if err != nil {
			data.Context.VideoCategory.DropCollection()
			panic(err)
		}
	}

	fmt.Println("All updates are successfully done")
}
