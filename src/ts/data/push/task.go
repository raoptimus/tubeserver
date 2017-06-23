//go:generate stringer -type=State
package push

import (
	"encoding/json"
	"time"
	"ts/data"
)

type (
	Task struct {
		Id      int         `bson:"_id"`
		Header  []data.Text `bson:"Header"`
		Message []data.Text `bson:"Message"`
		IconUrl string      `bson:"IconUrl"`
		GoUrl   string      `bson:"GoUrl"`
		//		PushClickCount int                    `bson:"PushClickCount"`
		DaysOfWeek     []int       `bson:"DaysOfWeek"`
		Hour           int         `bson:"Hour"`
		MaxHour        int         `bson:"MaxHour"`
		Repeat         Repeat      `bson:"Repeat"`
		FrequencyHours int         `bson:"FrequencyHours"`
		Action         Action      `bson:"Action"`
		Options        TaskOptions `bson:"Options"`
		State          State       `bson:"State"`
		Enabled        bool        `bson:"Enabled"`
		LastError      string      `bson:"LastError"`
		AddedDate      time.Time   `bson:"AddedDate"`
        CarrierType    []string    `bson:"CarrierType"`
        Countries      []string    `bson:"Countries"`
	}
	TaskOptions  map[string]interface{}
	NotifyReturn struct {
		ElapseDaysLastActiveFrom int `bson:"ElapseDaysLastActiveFrom"`
		ElapseDaysLastActiveTo   int `bson:"ElapseDaysLastActiveTo"`
	}
	NotifyToken struct {
		Token string `bson:"Token"`
	}
	NotifyUpgrade struct {
		Ver string `bson:"Ver"`
	}
)

type Action int

const (
	ActionNotifyAll Action = iota
	ActionNotifyToken
	ActionNotifyReturn
	ActionNotifyUpgrade
)

type Repeat int

const (
	RepeatOnce Repeat = iota
	RepeatLoop
)

type State int

const (
	StateWait State = iota
	StateInProgress
	StateError
	StateFinish
)

func (s *Task) GetHeader(lang data.Language) string {
	q := ""
	for _, text := range s.Header {
		q = text.Quote
		if text.Language == lang {
			break
		}
	}

	return q
}

func (s *Task) GetMessage(lang data.Language) string {
	q := ""
	for _, text := range s.Message {
		q = text.Quote
		if text.Language == lang {
			break
		}
	}

	return q
}

func (s TaskOptions) Object(obj interface{}) error {
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, &obj)
}
