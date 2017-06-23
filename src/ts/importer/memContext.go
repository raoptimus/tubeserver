package main

import "ts/mem/video/category"

type (
	memContext struct {
		VideoCategory *category.VideoCategory
	}
)

var ctx *memContext

func NewMemContext() (*memContext, error) {
	if ctx == nil {
		vc, err := category.NewVideoCategory()
		if err != nil {
			return nil, err
		}
		ctx = &memContext{
			VideoCategory: vc,
		}
	}

	return ctx, nil
}
