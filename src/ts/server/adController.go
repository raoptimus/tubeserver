package main

import (
	"github.com/raoptimus/gserv/config"
	"gopkg.in/mgo.v2/bson"
	"strings"
	"ts/data"
	api "ts/protocol/v1"
)

type AdController struct{}

func (s *AdController) BannerInfo(req *api.Request, bannerInfo *api.BannerInfo) error {
	token := req.Token
	if err := data.ValidateToken(token); err != nil {
		return err
	}
	var dev data.Device
	if err := dev.FindId(token); err != nil {
		return data.ErrNotFound(token)
	}
	*bannerInfo = *new(api.BannerInfo).Init(dev.Source.Ver)

	return nil
}

func (s *AdController) List(req *api.Request, list *api.AdList) error {
	carrierType := data.AdCarrierTypeUnknown

	if req.Device.LastISP != "" {
		carrierType = data.AdCarrierTypeWifi
	} else if req.Device.LastCarrier != "" {
		carrierType = data.AdCarrierTypeMobile
	}

	criteria := bson.M{
		"Status":      "Running",
		"Countries":   data.CountryUnknown,
		"CarrierType": carrierType.String(),
	}

	if req.Geo != nil && Mem.Country.IsExists(req.Geo.CountryCode) {
		criteria["Countries"] = strings.ToUpper(req.Geo.CountryCode)
	}

	q := data.Context.Ads.Find(criteria)

	var dbList []data.Ad
	err := q.Skip(req.Page.Skip).Limit(req.Page.Limit).All(&dbList)
	if err != nil {
		return err
	}

	ads := make(api.AdList, 0, len(dbList))
	for _, ad := range dbList {
		ads = append(ads, s.convertAd(&ad, req.Lang))
	}
	*list = ads.Sort()

	return nil
}

func (s *AdController) convertAd(ad *data.Ad, lang data.Language) *api.Ad {
	url := config.String("CdnAppUrl", "./") + "/ad"
	result := &api.Ad{
		Id:     ad.Id,
		Title:  ad.Title.Get(lang, false),
		Name:   ad.Name.Get(lang, true),
		Desc:   ad.Desc.Get(lang, false),
		Age:    ad.Age,
		Rating: int(ad.Rating * 10),
		Url:    ad.Link,
		Icon:   url + "/icon/" + ad.Icon,
		Images: make([]string, 0, len(ad.Screenshots)),
		Sort:   ad.Sort,
	}
	for _, hash := range ad.Screenshots {
		result.Images = append(result.Images, url+"/image/"+hash)
	}
	return result
}
