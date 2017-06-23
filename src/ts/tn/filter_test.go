package tn

import "testing"

type (
	filterTest struct {
		host    string
		format  string
		withWap bool

		siteId  int
		blockId int
	}
)

func TestFilter(t *testing.T) {

	filterTestList := []filterTest{
		filterTest{
			host:    "....com",
			format:  "preroll",
			withWap: true,
			siteId:  0,
			blockId: 0,
		},
		//....
	}

	r := Request{}

	for _, f := range filterTestList {
		siteId, blockId := r.filter(f.host, f.format, f.withWap)
		if f.siteId != siteId {
			t.Fatalf("Site can be %d but its %d by %v", f.siteId, siteId, f)
		}
		if f.blockId != blockId {
			t.Fatalf("Block can be %d but its %d %v", f.blockId, blockId, f)
		}
	}
}
