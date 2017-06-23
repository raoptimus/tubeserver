package data

type (
	Country struct {
		Id        int      `bson:"_id"`
		Name      string   `bson:"Name"`
		Code      string   `bson:"Code"`
	}
)

const CountryUnknown = "UN"
