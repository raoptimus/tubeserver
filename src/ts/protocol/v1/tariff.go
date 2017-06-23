package v1

import (
	"strconv"
	"strings"
)

type (
	Tariff struct {
		Id       int    `json:"Id"`
		Title    string `json:"Title"`
		Price    string `json:"Price"`
		Duration uint64 `json:"Duration"` //ms
		PayUrl   string `json:"PayUrl"`   //ms
	}
	TariffList []*Tariff
)

func (s *Tariff) SetMetaPayUrl(token string, userId int) {
	s.PayUrl = strings.Replace(s.PayUrl, "{TOKEN}", token, -1)
	s.PayUrl = strings.Replace(s.PayUrl, "{USER_ID}", strconv.Itoa(userId), -1)
	s.PayUrl = strings.Replace(s.PayUrl, "{TARIFF_ID}", strconv.Itoa(s.Id), -1)
}
