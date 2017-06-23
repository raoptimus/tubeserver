package v1

import (
	"encoding/json"
	"errors"
	"ts/data"
	"ts/detect"
)

type SortDirect int

const (
	SortDirectDesc SortDirect = -1
	SortDirectAsc  SortDirect = 1
)

type (
	Query    map[string]interface{}
	Object   interface{}
	SortInfo struct {
		Field  string
		Direct SortDirect
	}
	Page struct {
		Skip  int
		Limit int
	}
	Request struct {
		Ip      string        `json:"Ip"`
		Token   string        `json:"Token"`
		Lang    data.Language `json:"Lang"`
		Query   *Query        `json:"Query,omitempty"`
		Object  Object        `json:"Object,omitempty"`
		Sort    *SortInfo     `json:"Sort,omitempty"`
		Page    *Page         `json:"Page,omitempty"`
		Ver     string        `json:"Ver,omitempty"`
		Project string        `json:"Project,omitempty"`

		Device *data.Device      `json:"-"`
		Geo    *detect.GeoRecord `json:"-"`
	}
)

//todo use json.RawMessage
func (s *Request) UnmarshalObject(obj interface{}) error {
	if s.Object == nil {
		return errors.New("Object is empty")
	}
	b, err := json.Marshal(s.Object)
	if err != nil {
		return errors.New("Object is not marshal")
	}
	if err := json.Unmarshal(b, &obj); err != nil {
		return errors.New("Object is not unmarshal")
	}
	return nil
}

//todo use json.RawMessage
func (s *Request) UnmarshalQuery(obj interface{}) error {
	if s.Query == nil {
		return errors.New("Query is empty")
	}
	b, err := json.Marshal(s.Query)
	if err != nil {
		return errors.New("Query is not marshal")
	}
	if err := json.Unmarshal(b, &obj); err != nil {
		return errors.New("Query is not unmarshal")
	}
	return nil
}

func (s *SortInfo) LessString(i, j string) bool {
	switch s.Direct {
	case SortDirectDesc:
		return i > j
	default:
		return i < j
	}
}

func (s *SortInfo) LessInt(i, j int) bool {
	switch s.Direct {
	case SortDirectDesc:
		return i > j
	default:
		return i < j
	}
}

func (s *SortInfo) LessFloat64(i, j float64) bool {
	switch s.Direct {
	case SortDirectDesc:
		return i > j
	default:
		return i < j
	}
}
func (s *SortInfo) LessFloat32(i, j float32) bool {
	switch s.Direct {
	case SortDirectDesc:
		return i > j
	default:
		return i < j
	}
}
