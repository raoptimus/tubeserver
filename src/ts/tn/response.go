package tn

import "encoding/json"

type (
	Response struct {
		Error   string   `json:"error,omitempty"`
		Hash    string   `json:"hash"`
		Teasers []Teaser `json:"teasers"`
	}
	Teaser struct {
		Id    json.Number `json:"id,Number"`
		Price float32     `json:"price,string"`
		Ctr   float32     `json:"ctr,string"`
		Title string      `json:"title"`
		Img   string      `json:"img"`
		Url   string      `json:"url"`
	}
)

const minCtr = 0.01

func (s *Teaser) CtrInPercent() float32 {
	if s.Ctr < minCtr {
		return minCtr
	} else {
		return s.Ctr * 100.0
	}
}
