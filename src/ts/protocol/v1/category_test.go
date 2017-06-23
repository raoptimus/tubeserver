package v1

import (
	"testing"
	"ts/data"
)

func TestCategoryList(t *testing.T) {
	req := Request{
		Lang:  data.LanguageRussian,
		Token: TEST_TOKEN,
		Sort: &SortInfo{
			Field:  "Title",
			Direct: SortDirectAsc,
		},
		Page: &Page{
			Skip:  0,
			Limit: 100,
		},
	}
	var list Categories

	if err := getRpcClient().Call("CategoryController.List", &req, &list); err != nil {
		t.Fatal(err.Error())
	}

	if len(list) == 0 {
		t.Fatal("Return data is empty")
	}

	topCategoryFound := false
	for i, cat := range list {
		switch {
		case cat.Id == -1 && i == 0:
			topCategoryFound = true
		case cat.Id <= 0:
			t.Fatal("Data not valid, Id is zero")
		case cat.Title == "":
			t.Fatal("Data not valid, Title is empty")
		}
	}
	if topCategoryFound == false {
		t.Fatal("Category `top' not found")
	}
}
