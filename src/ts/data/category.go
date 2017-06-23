package data

type (
	Category struct {
		Id        int      `bson:"_id"`
		Title     TextList `bson:"Title"`
		SourceId  []int    `bson:"SourceId"`
		Slug      TextList `bson:"Slug"`
		ShortDesc string   `bson:"ShortDesc"`
		LongDesc  string   `bson:"LongDesc"`
	}
)
