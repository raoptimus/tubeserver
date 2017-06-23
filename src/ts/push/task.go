package main

import (
	"errors"
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"sync"
	"time"
	"ts/data"
	"ts/data/push"
)

type (
	Task struct {
		sync.RWMutex
		t *push.Task
	}
)

func NewTask(t *push.Task) *Task {
	return &Task{
		t: t,
	}
}

func (s *Task) Push() push.Task {
	s.RLock()
	push := *s.t
	s.RUnlock()
	return push
}

func (s *Task) Disable() {
	s.Lock()
	defer s.Unlock()
	s.t.Enabled = false
}

func (s *Task) UpdateFrom(tsk *push.Task) {
	s.Lock()
	defer s.Unlock()

	s.t.Header = tsk.Header
	s.t.Message = tsk.Message
	s.t.IconUrl = tsk.IconUrl
	s.t.GoUrl = tsk.GoUrl
	s.t.DaysOfWeek = tsk.DaysOfWeek
	s.t.Hour = tsk.Hour
	s.t.MaxHour = tsk.MaxHour
	s.t.Repeat = tsk.Repeat
	s.t.FrequencyHours = tsk.FrequencyHours
	s.t.Options = tsk.Options

	s.t.Enabled = tsk.Enabled
}

func (s *Task) SetState(state push.State, lastErr error) error {
	p := s.Push()
	lerr := ""
	if lastErr != nil {
		state = push.StateError
		lerr = lastErr.Error()
	} else if p.Repeat == push.RepeatLoop && state == push.StateFinish {
		state = push.StateWait
	}
	//todo update lastError only on State Err
	err := data.Context.PushTask.UpdateId(p.Id,
		bson.M{"$set": bson.M{"State": state, "LastError": lerr}})
	if err == nil {
		s.Lock()
		s.t.State = state
		s.Unlock()
	}
	return err
}

func (s *Task) CanDo(d *data.Device) error {
	reason := ""
	p := s.Push()
	switch {
	case !p.Enabled:
		reason = "Task is disabled"
	case !s.googleIdValid(d):
		reason = "Google is has no valid"
	case !s.frequencyExpired(d, &p):
		reason = "Frequency not expired"
	case !s.hourAllowed(d, &p):
		reason = "Hour not allowed"
	case !s.weekdayAllowed(d, &p):
		reason = "Day of the week not allowed"
	case !s.conditionAllowed(d, &p):
		reason = "Condition of action not allowed"
	case !s.countryAllowed(d, &p):
		reason = "Country not allowed"
	case !s.netAllowed(d, &p):
		reason = "Net not allowed"
	default:
		return nil
	}

	return errors.New(fmt.Sprintf("Task %d & Device %s is denied because %+v", p.Id, d.Id, reason))
}

func (s *Task) googleIdValid(d *data.Device) bool {
	return d.HasGoogleId && d.GoogleId != ""
}

func (s *Task) conditionAllowed(d *data.Device, p *push.Task) bool {
	switch p.Action {
	case push.ActionNotifyReturn:
		{
			var obj push.NotifyReturn
			if err := p.Options.Object(&obj); err != nil {
				return false
			}
			if obj.ElapseDaysLastActiveTo > 0 {
				return d.LastActiveTime.UTC().Add(time.Duration(obj.ElapseDaysLastActiveFrom)*24*time.Hour).
					Before(now()) && d.LastActiveTime.UTC().Add(time.Duration(obj.ElapseDaysLastActiveTo)*24*time.Hour).
					After(now())
			} else {
				return d.LastActiveTime.UTC().Add(time.Duration(obj.ElapseDaysLastActiveFrom) * 24 * time.Hour).
					Before(now())
			}
		}
	case push.ActionNotifyUpgrade:
		{
			var obj push.NotifyUpgrade
			if err := p.Options.Object(&obj); err != nil {
				return false
			}
			return obj.Ver > d.Source.Ver
		}
	default:
		{
			return true
		}
	}
}

func (s *Task) weekdayAllowed(d *data.Device, p *push.Task) bool {
	if len(p.DaysOfWeek) == 0 || len(p.DaysOfWeek) == 7 {
		return true
	}
	weekday := d.CurrDayOfWeek()

	if weekday == 0 {
		weekday = 7
	}

	for _, dw := range p.DaysOfWeek {
		if dw == weekday {
			return true
		}
	}
	return false
}

func (s *Task) frequencyExpired(d *data.Device, p *push.Task) bool {
	var (
		l   push.Log
		err error
	)
	if p.FrequencyHours == 0 {
		return false
	}
	err = data.Context.PushLog.
		Find(bson.M{"Token": d.Id, "TaskId": p.Id}).
		Sort("-SendedDate").One(&l)
	if err != nil {
		if err == mgo.ErrNotFound {
			return true
		}
		return false
	}
	if l.Status == push.StatusError {
		return true
	}
	if l.SendedDate.UTC().Add(time.Duration(p.FrequencyHours) * time.Hour).After(now()) {
		return false
	}
	return true
}

func (s *Task) hourAllowed(d *data.Device, p *push.Task) bool {
	hour := d.CurrHour()
	since := p.Hour
	till := p.MaxHour

	if since == 0 && till == 0 {
		return true
	}

	if till > since {
		return hour >= since && hour < till
	} else {
		return (hour >= since && hour < 23) || (hour >= 0 && hour < till)
	}
}

func (s *Task) countryAllowed(d *data.Device, p *push.Task) bool {
	var deviceCountry = data.CountryUnknown
	if d.LastGeo != nil && MemCountryList.IsExists(d.LastGeo.CountryCode) {
		deviceCountry = d.LastGeo.CountryCode
	}

	for _, taskCountry := range p.Countries {
		if taskCountry == deviceCountry {
			return true
		}
	}
	return false
}

func (s *Task) netAllowed(d *data.Device, p *push.Task) bool {
	for _, allowedCarrierType := range p.CarrierType {
		switch allowedCarrierType {
		case data.AdCarrierTypeWifi.String():
			if d.LastISP != "" || (d.LastISP == "" && d.LastCarrier == "") { //for default without data
				return true
			}
		case data.AdCarrierTypeMobile.String():
			if d.LastCarrier != "" {
				return true
			}
		}
	}
	return false
}
