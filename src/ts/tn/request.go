package tn

import (
	"net/url"
	"strings"
)

type (
	Request struct {
		Action  string `json:"action"`
		Charset string `json:"charset"`
		Width   int    `json:"width"`
		MiscId  int    `json:"misc_id"`
		BlockId int    `json:"block_id"`
		SiteId  int    `json:"site_id"`
		Site    *Site  `json:"site"`
		User    *User  `json:"user"`
		Ad      *Ad    `json:"ad"`
		Https   bool   `json:"https"`
	}
	Site struct {
		Referer string `json:"referer"`
		Page    string `json:"page"`
	}
	User struct {
		Ua   string `json:"ua"`
		Ip   string `json:"ip"`
		Lang string `json:"lang"`
	}
	Ad struct {
		Show      []string `json:"show,int,omitempty"`
		Amount    int      `json:"amount"`
		MinBid    float32  `json:"min_bid"`
		NoContent bool     `json:"no_content"`
	}
	Filter struct {
		AllowHosts   []string
		AllowFormats []string
		SiteId       int
		BlockId      int
		WithWap      bool
	}
	FilterList []*Filter
)

const (
	bidReqAction           = "getTeasers"
	bidReqCharset          = "utf-8"
	minBid         float32 = 0.01
	defaultSiteId          = 271840
	defaultBlockId         = 680430
)

var filters FilterList

func NewRequest(page, ref, ua, ip, lang string, w, count, subId int, shown []string, format string, withWap bool) *Request {
	req := &Request{
		Action:  bidReqAction,
		Charset: bidReqCharset,
		Width:   w,
		MiscId:  subId,
		Site: &Site{
			Page:    page,
			Referer: ref,
		},
		User: &User{
			Ua:   ua,
			Ip:   ip,
			Lang: lang,
		},
		Ad: &Ad{
			NoContent: true,
			Amount:    count,
			Show:      shown,
			MinBid:    minBid,
		},
	}

	host := ""
	link := page
	if link == "" {
		link = ref
	}
	if link != "" {
		if u, err := url.Parse(link); err == nil {
			req.Https = u.Scheme == "https"
			host = u.Host
		}
	}

	req.SiteId, req.BlockId = req.filter(host, format, withWap)
	return req
}

func (s *Request) filter(host, format string, withWap bool) (siteId, blockId int) {
	for _, f := range filters {
		if withWap != f.WithWap {
			continue
		}
		if f.hostIsAllow(host) && f.formatIsAllow(format) {
			siteId = f.SiteId
			blockId = f.BlockId
			break
		}
	}

	if siteId == 0 {
		siteId = defaultSiteId
		blockId = defaultBlockId
	}

	return
}

func (s *Filter) hostIsAllow(host string) bool {
	for _, h := range s.AllowHosts {
		if h == "*" || h == host {
			return true
		}
		if string(h[0]) == "*" {
			if h[1:] == host || strings.HasSuffix(host, h[1:]) {
				return true
			}
		}
	}
	return false
}

func (s *Filter) formatIsAllow(format string) bool {
	for _, f := range s.AllowFormats {
		if f == "*" || f == format {
			return true
		}
	}
	return false
}

func init() {
	filters = []*Filter{

		&Filter{
			AllowHosts:   []string{"*......"},
			AllowFormats: []string{"*"},
			WithWap:      true,
			SiteId:       0,
			BlockId:      0,
		},
		&Filter{
			AllowHosts:   []string{"*......"},
			AllowFormats: []string{"pauseroll", "postroll", "seekroll"},
			WithWap:      false,
			SiteId:       0,
			BlockId:      0,
		},
		&Filter{
			AllowHosts:   []string{"*......"},
			AllowFormats: []string{"*"},
			WithWap:      false,
			SiteId:       0,
			BlockId:      0,
		},

		&Filter{
			AllowHosts:   []string{"*......."},
			AllowFormats: []string{"*"},
			WithWap:      true,
			SiteId:       0,
			BlockId:      0,
		},
		&Filter{
			AllowHosts:   []string{"*....tv"},
			AllowFormats: []string{"pauseroll", "postroll", "seekroll"},
			WithWap:      false,
			SiteId:       0,
			BlockId:      0,
		},
		&Filter{
			AllowHosts:   []string{"*....tv"},
			AllowFormats: []string{"*"},
			WithWap:      false,
			SiteId:       0,
			BlockId:      0,
		},
		
		&Filter{
			AllowHosts:   []string{"*....tv"},
			AllowFormats: []string{"*"},
			WithWap:      true,
			SiteId:       0,
			BlockId:      0,
		},
		&Filter{
			AllowHosts:   []string{"*....tv"},
			AllowFormats: []string{"pauseroll", "postroll", "seekroll"},
			WithWap:      false,
			SiteId:       0,
			BlockId:      0,
		},
		&Filter{
			AllowHosts:   []string{"*....tv"},
			AllowFormats: []string{"*"},
			WithWap:      false,
			SiteId:       0,
			BlockId:      0,
		},

		&Filter{
			AllowHosts:   []string{"*.....tv"},
			AllowFormats: []string{"*"},
			WithWap:      true,
			SiteId:       0,
			BlockId:      0,
		},
		&Filter{
			AllowHosts:   []string{"*.....tv"},
			AllowFormats: []string{"pauseroll", "postroll", "seekroll"},
			WithWap:      false,
			SiteId:       0,
			BlockId:      0,
		},
		&Filter{
			AllowHosts:   []string{"*.....tv"},
			AllowFormats: []string{"*"},
			WithWap:      false,
			SiteId:       0,
			BlockId:      0,
		},
		
		&Filter{
			AllowHosts:   []string{"*....com"},
			AllowFormats: []string{"*"},
			WithWap:      true,
			SiteId:       0,
			BlockId:      0,
		},
		&Filter{
			AllowHosts:   []string{"*.....com"},
			AllowFormats: []string{"pauseroll", "postroll", "seekroll"},
			WithWap:      false,
			SiteId:       0,
			BlockId:      0,
		},
		&Filter{
			AllowHosts:   []string{"*....com"},
			AllowFormats: []string{"*"},
			WithWap:      false,
			SiteId:       0,
			BlockId:      0,
		},

		&Filter{
			AllowHosts:   []string{"*.....com"},
			AllowFormats: []string{"*"},
			WithWap:      true,
			SiteId:       0,
			BlockId:      0,
		},
		&Filter{
			AllowHosts:   []string{"*.....com"},
			AllowFormats: []string{"pauseroll", "postroll", "seekroll"},
			WithWap:      false,
			SiteId:       0,
			BlockId:      0,
		},
		&Filter{
			AllowHosts:   []string{"*.....com"},
			AllowFormats: []string{"*"},
			WithWap:      false,
			SiteId:       0,
			BlockId:      0,
		},

		&Filter{
			AllowHosts:   []string{"*"},
			AllowFormats: []string{"*"},
			WithWap:      true,
			SiteId:       0,
			BlockId:      0,
		},
		&Filter{
			AllowHosts:   []string{"*"},
			AllowFormats: []string{"pauseroll", "postroll", "seekroll"},
			WithWap:      false,
			SiteId:       0,
			BlockId:      0,
		},
		&Filter{
			AllowHosts:   []string{"*"},
			AllowFormats: []string{"*"},
			WithWap:      false,
			SiteId:       0,
			BlockId:      0,
		},
	}
}
