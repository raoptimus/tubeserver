package main

import (
	"errors"
	"fmt"
	"github.com/raoptimus/gserv/service"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"strconv"
	"strings"
	"time"
	"ts/data"
	api "ts/protocol/v1"
)

type (
	DeviceController struct {
	}
)

func (s *DeviceController) Reg2(req *api.Request, ret *api.ReturnToken) error {
	defer service.DontPanic()
	dev := new(api.Device)
	if err := req.UnmarshalObject(dev); err != nil {
		return err
	}

	ver := req.Ver
	ret.BannerInfo.Init(ver)

	tmpToken, err := dev.GenTmpToken()
	if err != nil {
		return errors.New("Device is not valid")
	}
	tokens := make([]string, 0)
	devices := make([]*data.Device, 0)
	newToken, _ := dev.GenNewToken()
	if newToken != "" {
		tokens = append(tokens, newToken)
	}
	if req.Token != "" {
		tokens = append(tokens, req.Token)
	}
	oldToken, _ := dev.GenOldToken()
	if oldToken != "" {
		tokens = append(tokens, oldToken)
	}
	if len(tokens) == 0 {
		tokens = append(tokens, tmpToken)
	}

	for _, token := range tokens {
		d, err := data.FindDevice(token)
		if err != nil {
			return err
		}
		if d != nil {
			devices = append(devices, d)
		}
	}

	token := tokens[0]

	//REGISTER
	if len(devices) == 0 {
		if err := s.register(dev, ver, token); err != nil {
			return err
		}
		ret.Token = token
		return nil
	}

	update := bson.M{}
	device := devices[0]
	src := s.src(dev, ver)

	if !src.Empty() && device.Source != src {
		device.Source = src // why?
		update["Source"] = src
	} else if src.Uid != "" && device.Source.Uid != src.Uid {
		device.Source.Uid = src.Uid
		update["Source.Uid"] = src.Uid
	}
	if dev.Serial != "" {
		device.Serial = dev.Serial
		update["Serial"] = dev.Serial
	}
	if dev.SerialGsm != "" {
		device.SerialGsm = dev.SerialGsm
		update["SerialGsm"] = dev.SerialGsm
	}

	if dev.Id != "" {
		device.DeviceId = dev.Id
		update["DeviceId"] = dev.Id
	}

	if dev.WifiMac != "" {
		device.WifiMac = dev.WifiMac
		update["WifiMac"] = dev.WifiMac
	}
	if dev.AdvertisingId != "" {
		device.AdvertisingId = dev.AdvertisingId
		update["AdvertisingId"] = dev.AdvertisingId
	}

	ret.Token = token
	ret.Exists = true

	//EXISTS & REPLACE
	if token != device.Id {
		if err := s.replace(device, token); err != nil {
			return err
		}
		return nil
	}

	//EXISTS
	if len(update) > 0 {
		if err := data.Context.Devices.UpdateId(token, bson.M{"$set": update}); err != nil {
			return err
		}
	}
	return nil
}

func (s *DeviceController) UpdateInfo(req *api.Request, result *bool) error {
	defer service.DontPanic()
	var info api.DeviceInfo
	if err := req.UnmarshalObject(&info); err != nil {
		return err
	}
	l, err := time.LoadLocation(info.Loc)
	if err != nil {
		return err
	}

	h := time.Now().In(l).Hour() - time.Now().UTC().Hour()
	deviceLoc := data.DeviceLocation{
		Loc: info.Loc,
		Gmt: h,
	}
	err = data.Context.Devices.UpdateId(req.Token,
		bson.M{"$set": bson.M{
			"Loc":            deviceLoc,
			"GoogleId":       info.GoogleId,
			"HasGoogleId":    info.GoogleId != "",
			"UpdateGoogleId": time.Now().UTC(),
			"Language":       info.Language,
		}})
	if err != nil {
		return err
	}
	*result = true
	return nil
}

func (s *DeviceController) ClickPush(req *api.Request, result *bool) error {
	defer service.DontPanic()
	var objId struct {
		Id string
	}
	if err := req.UnmarshalObject(&objId); err != nil {
		return err
	}
	token := req.Token

	var d data.Device
	err := data.Context.Devices.FindId(token).One(&d)
	if err != nil {
		return errors.New("Token error: " + err.Error())
	}

	update := bson.M{"$inc": bson.M{"PushClickCount": 1}}
	err = data.Context.Devices.UpdateId(token, update)
	if err != nil {
		return err
	}

	id, err := strconv.Atoi(objId.Id)
	if err != nil {
		return errors.New("Push id not valid")
	}
	if id <= 0 {
		return errors.New("Push id can't be blank")
	}

	err = data.Context.PushTask.UpdateId(id, update)
	if err != nil {
		return errors.New("PushTask $inc PushClickCount error: " + err.Error())
	}

	stat := data.NewDailyStat(&d.Source.DeviceSource)
	stat.PushClickCount++
	err = stat.UpsertInc()
	if err != nil {
		return err
	}

	*result = true
	return nil
}

func (s *DeviceController) AddAction(req *api.Request, result *bool) error {
	defer service.DontPanic()

	var inEvent api.DeviceEvent
	if err := req.UnmarshalObject(&inEvent); err != nil {
		return err
	}
	dl := strings.ToLower(inEvent.Details)
	var act data.DeviceAction
	stat := data.NewDailyStat(&req.Device.Source.DeviceSource)
	stat.TLaunchCount++
	now := time.Now().UTC()

	isUniq := now.Sub(req.Device.LastActiveTime.UTC()) > 24*time.Hour
	if isUniq {
		stat.ULaunchCount++
	}

	switch {
	case strings.Contains(dl, "user") && strings.Contains(dl, "updated"):
		{
			act = data.DeviceActionUpdate
			stat.UpgradeCount++
		}
	case strings.Contains(dl, "first") && strings.Contains(dl, "launch"):
		{
			act = data.DeviceActionFLaunch
			stat.FLaunchCount++
		}
	case !strings.Contains(dl, "first") && strings.Contains(dl, "launch"):
		{
			act = data.DeviceActionLaunch
		}
	case strings.Contains(dl, "reinstall"):
		{
			act = data.DeviceActionReInstall
			stat.ReinstallCount++
		}
	default:
		{
			return errors.New("The unknown details. Unable to define action")
		}
	}

	ev := &data.DeviceEvent{
		Id:        bson.NewObjectId(),
		DeviceId:  req.Device.Id,
		Action:    act,
		Details:   dl,
		Ip:        req.Ip,
		Ver:       req.Ver,
		AddedDate: now,
	}

	set := bson.M{
		"LastActiveTime": time.Now().UTC(),
	}
	if req.Ip != req.Device.LastIp {
		set["LastIp"] = req.Ip
		set["LastGeo"] = req.Geo
	}
	if req.Ver != req.Device.Source.Ver {
		set["Source.Ver"] = req.Ver
	}

	update := bson.M{
		"$set": set,
		"$inc": bson.M{"LaunchCount": 1},
	}
	if err := data.Context.Devices.UpdateId(req.Device.Id, update); err != nil {
		return fmt.Errorf("Device can't be update: %v", err)
	}
	if err := stat.UpsertInc(); err != nil {
		return fmt.Errorf("Daily stat can't be upsert: %v", err)
	}
	if err := data.Context.DeviceEvents.Insert(ev); err != nil {
		return fmt.Errorf("Device event can't be insert: %v", err)
	}

	*result = true
	return nil
}

func (s *DeviceController) UpdateNet(req *api.Request, res *bool) error {
	defer service.DontPanic()
	var net api.DeviceNet
	if err := req.UnmarshalObject(&net); err != nil {
		return err
	}
	up := bson.M{"$set": bson.M{
		"LastISP":     net.ISP,
		"LastCarrier": net.Carrier,
	}}
	if err := data.Context.Devices.UpdateId(req.Token, up); err != nil {
		return err
	}
	*res = true
	return nil
}

// ===============
// PRIVATE METHODS

func (s *DeviceController) register(dev *api.Device, ver, token string) error {
	device := data.Device{
		Id:            token,
		DeviceId:      dev.Id,
		Os:            dev.Os,
		Type:          dev.Type,
		VerOs:         dev.VerOs,
		Serial:        dev.Serial,
		WifiMac:       dev.WifiMac,
		AdvertisingId: dev.AdvertisingId,
		Manufacture:   dev.Manufacture,
		Model:         dev.Model,
		SerialGsm:     dev.SerialGsm,
		Source:        s.src(dev, ver),
	}
	device.AddedDate = time.Now().UTC()
	device.LastActiveTime = device.AddedDate
	if err := data.Context.Devices.Insert(device); err != nil {
		return err
	}
	stat := data.NewDailyStat(&device.Source.DeviceSource)
	stat.DeviceRegCount++
	if err := stat.UpsertInc(); err != nil {
		return err
	}
	PostBack.GoSend(&device) //post back to partners
	return nil
}

func (s *DeviceController) replace(device *data.Device, newToken string) error {
	id := device.Id
	device.Id = newToken
	device.OldToken = id

	user, err := data.GetUserByToken(id)
	if err != nil && err != mgo.ErrNotFound {
		return errors.New("Can't take user for device")
	}
	if user != nil {
		userTokens := []string{device.Id}
		for _, t := range user.Tokens {
			if *t == device.Id || *t == id {
				continue
			}
			userTokens = append(userTokens, *t)
		}
		err = data.Context.Users.UpdateId(user.Id, bson.M{"$set": bson.M{"Tokens": userTokens}})
		if err != nil {
			return err
		}
	}

	if err := data.Context.Devices.Insert(device); err != nil {
		return err
	}
	if err := data.Context.Devices.RemoveId(id); err != nil {
		return err
	}

	return nil
}

func (s *DeviceController) src(dev *api.Device, ver string) data.DeviceSourceExt {
	t := dev.ParseTracking()
	return data.DeviceSourceExt{
		DeviceSource: data.DeviceSource{
			Apk:     t.Apk,
			Ad:      t.Ad,
			Landing: t.Landing,
			Partner: t.Partner,
			Site:    t.Site,
			Ver:     ver,
		},
		TransId: t.TransId,
		AffId:   t.AffId,
		OffId:   t.OffId,
		Uid:     t.Uid,
	}
}
