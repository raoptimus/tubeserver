package data

type (
	Channel struct {
		Id    int    `bson:"_id"`
		Title string `bson:"Title"`
	}
)
