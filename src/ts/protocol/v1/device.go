package v1

import (
	"crypto/md5"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"ts/data"
)

type DeviceActionType int

const (
	DeviceActionTypeDownload DeviceActionType = iota
	DeviceActionTypeStart
	DeviceActionTypeExit
)

func (s DeviceActionType) String() string {
	switch s {
	case DeviceActionTypeDownload:
		return "Download"
	case DeviceActionTypeStart:
		return "Start"
	case DeviceActionTypeExit:
		return "Exit"
	default:
		return ""
	}
}

type (
	Device struct {
		Id            string
		Os            string
		Type          string
		VerOs         string
		Serial        string
		WifiMac       string
		AdvertisingId string
		Manufacture   string
		Model         string
		SerialGsm     string
		FileName      string
		Tracking      string
	}
	ReturnToken struct {
		Alert  string
		Locked bool
		Token  string
		Exists bool
		BannerInfo
	}
	DeviceEvent struct {
		Details string
	}
	DeviceInfo struct {
		Loc      string
		GoogleId string
		Language data.Language
	}
	DeviceNet struct {
		ISP     string
		Carrier string
	}
	DeviceTracking struct {
		Apk     string
		Site    string
		Ad      string
		Landing string
		Partner string
		TransId string
		OffId   string
		AffId   string
		Uid     string
	}
)

func (s *Device) ValidateNewToken() error {
	switch {
	case s.AdvertisingId == "":
		return errors.New("AdvertisingId is empty")
	}
	return nil
}

func (s *Device) ValidateOldToken() error {
	switch {
	case s.Serial == "":
		return errors.New("Serial is empty")
	case s.SerialGsm == "":
		return errors.New("SerialGsm is empty")
	case s.Id == "":
		return errors.New("DeviceId is empty")
	}
	return nil
}

func (s *Device) GenNewToken() (string, error) {
	if err := s.ValidateNewToken(); err != nil {
		return "", err
	}
	return s.genToken(s.AdvertisingId), nil
}

func (s *Device) GenOldToken() (string, error) {
	if err := s.ValidateOldToken(); err != nil {
		return "", err
	}
	return s.genToken(s.Serial, s.SerialGsm, s.Id), nil
}

func (s *Device) GenTmpToken() (string, error) {
	getField := func(v reflect.Value, i int) reflect.Value {
		val := v.Field(i)
		if val.Kind() == reflect.Interface && !val.IsNil() {
			val = val.Elem()
		}
		return val
	}
	v := reflect.ValueOf(s).Elem()
	a := make([]string, 0)

	for i := 0; i < v.NumField(); i++ {
		vv := getField(v, i).String()
		if vv != "" {
			a = append(a, vv)
		}
	}
	if len(a) == 0 {
		return "", errors.New("All Device fields are empty")
	}
	return s.genToken(a...), nil
}

func (s *Device) genToken(items ...string) string {
	data := "apk|" + strings.Join(items, "|")
	return fmt.Sprintf("%x", md5.Sum([]byte(data)))
}

func (s *Device) ParseTracking() (t *DeviceTracking) {
	if s.Tracking == "" {
		return s.ParseFileName()
	}
	t = &DeviceTracking{
		Apk: "1",
	}
	b, err := base64.URLEncoding.DecodeString(s.Tracking)
	if err != nil {
		return t
	}
	v, err := url.ParseQuery(string(b))
	if err != nil {
		return t
	}
	t.Site = strings.Replace(v.Get("s"), ".", "_", 0)
	t.Ad = strings.Replace(v.Get("a"), ".", "_", 0)
	t.Landing = strings.Replace(v.Get("l"), ".", "_", 0)
	t.Partner = strings.Replace(v.Get("u"), ".", "_", 0)
	t.TransId = strings.Replace(v.Get("_t"), ".", "_", 0)
	t.OffId = strings.Replace(v.Get("_o"), ".", "_", 0)
	t.AffId = strings.Replace(v.Get("_a"), ".", "_", 0)
	t.Uid = strings.Replace(v.Get("_uid"), ".", "_", 0)

	if strings.Contains("apk2", t.Site) {
		t.Apk = "2"
	}
	return t
}

//deprecated
func (s *Device) ParseFileName() (t *DeviceTracking) {
	t = &DeviceTracking{
		Apk: "1",
	}
	if s.FileName == "" || s.FileName == "....apk" || s.FileName == "..." {
		return
	}
	reg := regexp.MustCompile(`^...(([A-Za-z0-9+/-]{4})*([A-Za-z0-9+/-]{2}==|[A-Za-z0-9+/-]{3}=)?).*`)
	name := reg.ReplaceAllString(s.FileName, "$1")
	b, err := base64.URLEncoding.DecodeString(name)
	if err != nil {
		return t
	}
	v, err := url.ParseQuery(string(b))
	if err != nil {
		return t
	}

	t.Site = strings.Replace(v.Get("s"), ".", "_", 0)
	t.Ad = strings.Replace(v.Get("a"), ".", "_", 0)
	t.Landing = strings.Replace(v.Get("l"), ".", "_", 0)
	if strings.Contains("apk2", t.Site) {
		t.Apk = "2"
	}
	return
}
