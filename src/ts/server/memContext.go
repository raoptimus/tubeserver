package main

import (
	"ts/mem/video/category"
	"ts/mem/video/channel"
	"ts/mem/targeting/country"
)

type (
	memContext struct {
		VideoCategory *category.VideoCategory
		VideoChannel  *channel.Channel
        Country       *country.Country
	}
)

var ctx *memContext

func NewMemContext() (*memContext, error) {
	if ctx == nil {
		vc, err := category.NewVideoCategory()
		if err != nil {
			return nil, err
		}
		vch, err := channel.NewChannel()
		if err != nil {
			return nil, err
		}
		country, err := country.NewCountry()
		if err != nil {
			return nil, err
		}
		ctx = &memContext{
			VideoCategory: vc,
			VideoChannel:  vch,
            Country: country,
		}
	}

	return ctx, nil
}
