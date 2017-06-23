package v1

import (
	"strings"
	"testing"
	"ts/data"
)

// тест автодополнения для поиска
func TestSearch(t *testing.T) {
	pattern := "ююю"
	queries := []string{
		pattern,
		"ююю",
		"ююю",
		"ююю",
		"ююю",
	}

	// put data
	for _, q := range queries {
		req := Request{
			Lang:  data.LanguageRussian,
			Token: TEST_TOKEN,
			Query: &Query{
				"$text": q,
			},
			Page: &Page{
				Skip:  0,
				Limit: 100,
			},
		}
		var list VideoList
		if err := getRpcClient().Call("VideoController.Find", &req, &list); err != nil {
			t.Fatal(err.Error())
		}
	}
	req := Request{
		Lang:  data.LanguageRussian,
		Token: TEST_TOKEN,
		Query: &Query{
			"Id": pattern,
		},
	}
	var list SearchQueryList
	if err := getRpcClient().Call("SearchController.Find", &req, &list); err != nil {
		t.Fatal(err.Error())
	}
	if len(list) == 0 {
		t.Fatal("Seach result must not be empty")
	}
	for _, s := range list {
		if !strings.Contains(s.Text, pattern) {
			t.Fatalf("Unexpected search result: %q; want `%s.*`", s.Text, pattern)
		}
	}
}
