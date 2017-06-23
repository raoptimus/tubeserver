package data

type AdCarrierType int

const (
    AdCarrierTypeUnknown AdCarrierType = iota
    AdCarrierTypeWifi
    AdCarrierTypeMobile
)

func (s AdCarrierType) String() string {
	switch s {
	case AdCarrierTypeWifi:
		return "wifi"
	case AdCarrierTypeMobile:
		return "mobile"
	default:
		return "unknown"
	}
}

type Ad struct {
	Id          int      `bson:"_id"`
	Title       TextList `bson:"Title"`
	Name        TextList `bson:"Name"`
	Desc        TextList `bson:"Desc"`
	Age         int      `bson:"Age"`
	Rating      float64  `bson:"Rating"`
	Link        string   `bson:"Link"`
	Sort        int      `bson:"Sort"`
	Status      string   `bson:"Status"`
	Icon        string   `bson:"Icon"`
	Screenshots []string `bson:"Screenshots"`
}
