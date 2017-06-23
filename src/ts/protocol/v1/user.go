package v1

import (
	"ts/data"
)

type (
	User struct {
		Id           int
		UserName     string
		Tokens       []*string
		Avatar       string
		Tel          string
		Email        string
		Lang         data.Language
		CreationDate int64
		Premium      Premium
	}
	Premium struct {
		Duration uint64 //ms
		Type     data.PremiumType
	}
)
