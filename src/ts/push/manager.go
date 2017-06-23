package main

import (
	"errors"
	"fmt"
	"github.com/googollee/go-gcm"
	"github.com/raoptimus/gserv/config"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
	"ts/data"
	"ts/data/push"
)

const PUSH_COUNT_PER_MESSAGE = 100

type (
	Manager struct {
		sync.RWMutex
		workingTaskList map[int]*Task
		stop            chan bool
		update          chan bool
		debug           struct {
			disableSending bool
		}
		sync.WaitGroup
	}
)

func NewManager() *Manager {
	return &Manager{
		workingTaskList: make(map[int]*Task),
		stop:            make(chan bool),
		update:          make(chan bool),
	}
}

func (s *Manager) Stop() {
	s.stop <- true

	for _, task := range s.workingTaskList {
		task.Disable()
	}

	s.WaitGroup.Wait()
}

func (s *Manager) Start() {
	for {
		if !s.mainLoop() {
			return
		}
		select {
		case <-s.update:
		case <-time.After(5 * time.Minute):
		case <-s.stop:
			return
		}
	}
}

func (s *Manager) ForceUpdater() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGUSR1)
	for range c {
		s.ForceUpdate()
	}
}

func (s *Manager) ForceUpdate() {
	s.update <- true
}

func (s *Manager) mainLoop() bool {
	var pushTaskList []*push.Task
	query := bson.M{"State": bson.M{"$ne": push.StateFinish}}
	err := data.Context.PushTask.Find(query).All(&pushTaskList)
	if err != nil {
		log.Err(err.Error())
	}
	for _, pushTask := range pushTaskList {
		select {
		case <-s.stop:
			return false
		default:
		}

		s.RLock()
		taskMem, ok := s.workingTaskList[pushTask.Id]
		s.RUnlock()

		if !ok {
			if !pushTask.Enabled {
				continue
			}
			taskMem = NewTask(pushTask)
			s.Lock()
			s.workingTaskList[pushTask.Id] = taskMem
			s.Unlock()
			go s.startTask(taskMem)
			continue
		}

		taskMem.UpdateFrom(pushTask)
		if !pushTask.Enabled {
			continue
		}

		switch taskMem.Push().State {
		case push.StateWait, push.StateError:
			go s.startTask(taskMem)
		}
	}
	return true
}

func (s *Manager) startTask(task *Task) {
	s.WaitGroup.Add(1)
	pushTask := task.Push()
	fmt.Printf("[%v] Start task (%d)\n", time.Now(), pushTask.Id)
	task.SetState(push.StateInProgress, nil)

	state := push.StateFinish
	err := s.jobTask(task)
	if err != nil {
		state = push.StateError
	}
	if err := task.SetState(state, err); err != nil {
		log.Err(err.Error())
	}
	pushTask = task.Push()
	if pushTask.State == push.StateFinish {
		s.Lock()
		delete(s.workingTaskList, pushTask.Id)
		s.Unlock()
	}
	fmt.Printf("[%v] Finished task (%d) with state (%s), err (%v)\n", time.Now(), pushTask.Id, state, err)
	s.WaitGroup.Done()
}

func (s *Manager) jobTask(task *Task) error {
	googleRegKey := config.String("GoogleRegKey", "")
	if googleRegKey == "" {
		panic(errors.New("Google registration key cant be blank"))
	}
	client := gcm.New(config.String("GoogleRegKey", ""))
	var resp *gcm.Response
	deviceQuery := bson.M{"HasGoogleId": true}

	switch task.Push().Action {
	case push.ActionNotifyToken:
		{
			var obj push.NotifyToken
			if err := task.Push().Options.Object(&obj); err != nil {
				return errors.New("Options is no valid: " + err.Error())
			}
			if len(obj.Token) != 32 {
				return errors.New("Options.Token is no valid")
			}
			deviceQuery["_id"] = obj.Token
		}
	case push.ActionNotifyReturn:
		{
			var obj push.NotifyReturn
			if err := task.Push().Options.Object(&obj); err != nil {
				return errors.New("Options is no valid: " + err.Error())
			}
			if obj.ElapseDaysLastActiveTo > 0 {
				deviceQuery["LastActiveTime"] = bson.M{
					"$lte": now().Add(time.Duration(-obj.ElapseDaysLastActiveFrom) * 24 * time.Hour),
					"$gte": now().Add(time.Duration(-obj.ElapseDaysLastActiveTo) * 24 * time.Hour),
				}
			} else {
				deviceQuery["LastActiveTime"] = bson.M{
					"$lte": now().Add(time.Duration(-obj.ElapseDaysLastActiveFrom) * 24 * time.Hour),
				}
			}
		}
	case push.ActionNotifyUpgrade:
		{
			var obj push.NotifyUpgrade
			if err := task.Push().Options.Object(&obj); err != nil {
				return errors.New("Options is no valid: " + err.Error())
			}
			deviceQuery["Source.Ver"] = bson.M{"$lt": obj.Ver}
		}
	}

	switch task.Push().Action {
	case push.ActionNotifyToken, push.ActionNotifyAll, push.ActionNotifyReturn, push.ActionNotifyUpgrade:
		{
			var deviceList []data.Device

			for skip := 0; true; skip += PUSH_COUNT_PER_MESSAGE {
				if !task.Push().Enabled {
					//stop all, task is disabled
					return nil
				}
				deviceQuery["Loc.Gmt"] = task.Push().Hour - now().Hour()

				err := data.Context.Devices.Find(deviceQuery).
					Skip(skip).
					Limit(PUSH_COUNT_PER_MESSAGE).
					All(&deviceList)

				// fmt.Printf("Task (%d), found: %d devices, err: %v\n",
				// 	task.Push().Id, len(deviceList), err)
				if err != nil && err != mgo.ErrNotFound {
					return errors.New("Dont find device list: " + err.Error())
				}
				if len(deviceList) == 0 {
					break
				}

				deviceList = s.filterDeviceList(task, deviceList...)
				if len(deviceList) == 0 {
					continue
				}

				msg := s.createMessage(task, deviceList...)

				if s.debug.disableSending {
					resp = &gcm.Response{}
				} else {
					resp, err = client.Send(msg)
					if err != nil {
						return errors.New("Dont send gcm message:" + err.Error())
					}
				}

				for _, d := range deviceList {
					derr := ""
					for _, r := range resp.Results {
						if d.GoogleId != r.RegistrationID {
							continue
						}
						derr = r.Error
					}
					status := push.StatusError
					if derr == "" {
						status = push.StatusSuccess
					}
					if err := push.WriteLog(&d, status, task.Push().Id, derr); err != nil {
						log.Err("Dont write to push log: " + err.Error())
					}
				}
			}
		}
	default:
		{
			return errors.New(fmt.Sprintf("Action by Task(%d) not found", task.Push().Id))
		}
	}
	return nil
}

func (s *Manager) filterDeviceList(task *Task, list ...data.Device) (result []data.Device) {
	result = make([]data.Device, 0)
	for _, d := range list {
		if err := task.CanDo(&d); err != nil {
			continue
		}
		result = append(result, d)
	}
	return result
}

func (s *Manager) createMessage(task *Task, deviceList ...data.Device) *gcm.Message {
	msg := gcm.NewMessage()
	for _, d := range deviceList {
		msg.AddRecipient(d.GoogleId)
	}

	push := task.Push()
	iconUrl := push.IconUrl

	icon, err := url.Parse(push.IconUrl)
	if err == nil {
		if icon.Host == "" {
			appUrl := config.String("CdnAppUrl", "")
			if appUrl != "" {
				iconUrl = appUrl + icon.Path
				if icon.RawQuery != "" {
					iconUrl += "?" + icon.RawQuery
				}
			}
		}
	}

	msg.CollapseKey = strconv.Itoa(push.Id)
	msg.DelayWhileIdle = true
	msg.TimeToLive = 4 * 7 * 24 * 60 * 60                // 4 weeks
	msg.Data["H"] = push.GetHeader(data.LanguageRussian) //todo get lang from device
	msg.Data["M"] = push.GetMessage(data.LanguageRussian)
	msg.Data["U"] = push.GoUrl
	msg.Data["I"] = iconUrl
	msg.Data["Id"] = strconv.Itoa(push.Id)
	return msg
}

func (s *Manager) deleteTask(id int) {
	s.Lock()
	delete(s.workingTaskList, id)
	s.Unlock()
}
