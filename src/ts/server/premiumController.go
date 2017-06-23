package main

import (
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"time"
	"ts/data"
	api "ts/protocol/v1"
)

type PremiumController struct{}

func (s *PremiumController) TariffList(req *api.Request, list *api.TariffList) error {
	user := &data.User{}
	if err := user.GetOrCreate(req.Token); err != nil {
		return err
	}
	// rpc will signal an error if list will be nil
	*list = make(api.TariffList, 0, 4)
	const durationCoef = uint64(time.Hour / time.Millisecond)
	var tariff data.Tariff
	iter := data.Context.Tariff.Find(bson.M{"Enabled": true}).Sort("-AproxPrice").Iter()
	for iter.Next(&tariff) {
		t := &api.Tariff{
			Id:       tariff.Id,
			Title:    tariff.Title.Get(req.Lang, true),
			Price:    fmt.Sprintf("%s%.2g", tariff.Currency, tariff.Price),
			Duration: tariff.Time * durationCoef,
			PayUrl:   tariff.PayUrl,
		}
		t.SetMetaPayUrl(req.Token, user.Id)
		*list = append(*list, t)
		tariff = data.Tariff{}
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("Cant find tariffs: %v", err)
	}
	return nil
}
