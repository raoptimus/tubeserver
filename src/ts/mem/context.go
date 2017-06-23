package mem

import "ts/mem/video/category"

type (
	Context struct {
		VideoCategory *category.VideoCategory
	}
)

var ctx *Context

func NewContext() *Context {
	if ctx == nil {
		vc, err := category.NewVideoCategory()
		if err != nil {
			panic(err)
		}
		ctx = &Context{
			VideoCategory: vc,
		}
	}

	return ctx
}

func init() {
	NewContext()
}
