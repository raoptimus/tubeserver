package v1

import (
	"github.com/raoptimus/gserv/config"
	"sort"
	"strings"
)

type BannerInfo struct {
	NeedBanner      bool
	BannerFrequency int
	BannersUrl      string
}

func (b *BannerInfo) Init(version string) *BannerInfo {
	b.NeedBanner = config.Bool("NeedBanner", false)
	b.BannerFrequency = config.Int("BannerFrequency", 0)
	b.BannersUrl = config.String("BannersUrl", "")
	if strings.Contains(b.BannersUrl, "12place.com") {
		b.BannersUrl += "&p=apk"
	}
	if version != "" {
		delim := "?"
		if strings.IndexByte(b.BannersUrl, '?') != -1 {
			delim = "&"
		}
		b.BannersUrl += delim + "v=" + version
	}
	return b
}

type Ad struct {
	Id     int
	Title  string
	Name   string
	Desc   string
	Age    int
	Rating int
	Url    string
	Icon   string
	Images []string
	Sort   int `bson:"-"`
}
type AdList []*Ad

func (s AdList) Sort() AdList {
	sort.Sort(s)
	return s
}

func (s AdList) Len() int {
	return len(s)
}

func (s AdList) Less(i, j int) bool {
	return s[i].Sort > s[j].Sort
}

func (s AdList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
