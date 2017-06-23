package data

import (
	"time"
)

type PremiumType string

const (
	PremiumTypeNone   = "none"
	PremiumTypeTrial  = "trial"
	PremiumTypeSignup = "signup"
)

type Premium struct {
	Expires time.Time   `bson:"Expires"`
	Type    PremiumType `bson:"Type"`
}

func (p *Premium) Expired() bool {
	return p.Expires.UTC().After(time.Now().UTC())
}

func (p *Premium) Duration() time.Duration {
	return p.Expires.UTC().Sub(time.Now())
}

func (p *Premium) setTrial(hours int) {
	if hours > 0 {
		dur := time.Duration(hours) * time.Hour
		p.Expires = time.Now().Add(dur)
		p.Type = PremiumTypeTrial
	} else {
		p.Type = PremiumTypeNone
	}
}

type Tariff struct {
	Id       int      `bson:"_id"`
	Price    float64  `bson:"AproxPrice"`
	Currency string   `bson:"Currency"`
	Title    TextList `bson:"Title"`
	Time     uint64   `bson:"Time"` // duration in hours
	PayUrl   string   `bson:"PayUrl"`
}

type Transaction struct {
	Token     string      `bson:"Token"`
	Price     float64     `bson:"Price"`
	Duration  int         `bson:"Duration"` //hours
	UserId    int         `bson:"UserId"`
	TariffId  int         `bson:"TariffId"`
	Type      PremiumType `bson:"Type"`
	AddedDate time.Time   `bson:"AddedDate"`
	// TransactionData interface{} `bson:"TransactionData"`
	// BillingId       interface{} `bson:"BillingId"`
}
