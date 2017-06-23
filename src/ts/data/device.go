package data

import (
	"errors"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
	"ts/detect"
)

type (
	Device struct {
		Id              string            `bson:"_id"`
		OldToken        string            `bson:"OldToken"` //temporary for debug
		DeviceId        string            `bson:"DeviceId"`
		WifiMac         string            `bson:"WifiMac"`
		AdvertisingId   string            `bson:"AdvertisingId"`
		Os              string            `bson:"Os"`
		Type            string            `bson:"Type"`
		VerOs           string            `bson:"VerOs"`
		Serial          string            `bson:"Serial"`
		Manufacture     string            `bson:"Manufacture"`
		Model           string            `bson:"Model"`
		SerialGsm       string            `bson:"SerialGsm"`
		StartCount      int               `bson:"StartCount"`
		PushClickCount  int               `bson:"PushClickCount"`
		PushSendedCount int               `bson:"PushSendedCount"`
		DownloadCount   int               `bson:"DownloadCount"`
		ExitCount       int               `bson:"ExitCount"`
		Source          DeviceSourceExt   `bson:"Source"`
		LastGeo         *detect.GeoRecord `bson:"LastGeo"`
		LastIp          string            `bson:"LastIp"`
		LastActiveTime  time.Time         `bson:"LastActiveTime"`
		LastISP         string            `bson:"LastISP"`
		LastCarrier     string            `bson:"LastCarrier"`
		GoogleId        string            `bson:"GoogleId"`
		HasGoogleId     bool              `bson:"HasGoogleId"`
		Loc             *DeviceLocation   `bson:"Loc"`
		Language        Language          `bson:"Language"`
		AddedDate       time.Time         `bson:"AddedDate"`
		UpdateGoogleId  time.Time         `bson:"UpdateGoogleId"`
	}
	DeviceLocation struct {
		Loc string `bson:"Loc"`
		Gmt int    `bson:"Gmt"`
	}
	//use in key for stats
	DeviceSource struct {
		//площадка
		Site string `bson:"Site"`
		//целевая страница
		Landing string `bson:"Landing"`
		//банер
		Ad string `bson:"Ad"`
		//Apk
		Apk string `bson:"Apk"`
		//apk version
		Ver string `bson:"Ver"`
		//userId
		Partner string `bson:"Partner"`
	}
	DeviceSourceExt struct {
		DeviceSource `bson:"DeviceSource,inline"`
		//transaction id
		TransId string `bson:"TransId"`
		//affiliate id
		AffId string `bson:"AffId"`
		//offer id
		OffId string `bson:"OffId"`
		//uid from web cookie
		Uid string `bson:"Uid"`
	}
)

func (s *DeviceSourceExt) Empty() bool {
	return s.Site+s.Landing+s.Ad+s.Partner == ""
}

func (s *Device) CurrHour() int {
	gmt := 0
	if s.Loc != nil {
		gmt = s.Loc.Gmt
	}
	return time.Now().UTC().Add(time.Duration(gmt) * time.Hour).Hour()
}

func (s *Device) CurrDayOfWeek() int {
	gmt := 0
	if s.Loc != nil {
		gmt = s.Loc.Gmt
	}
	return int(time.Now().UTC().Add(time.Duration(gmt) * time.Hour).Weekday())
}

func (s *Device) FindId(token string) error {
	return Context.Devices.FindId(token).One(s)
}

func (s *Device) FindUser() (user *User, err error) {
	err = Context.Users.Find(bson.M{"Token": s.Id}).One(&user)
	return
}

func ValidateToken(token string) error {
	if len(token) != 32 {
		return errors.New("Token is empty or incorrect")
	}
	return nil
}

func ErrNotFound(token string) error {
	return errors.New("Token " + token + " not found")
}

// FindDevice returns *Device or nil if one is not found
func FindDevice(token string) (*Device, error) {
	var device Device
	err := Context.Devices.FindId(token).One(&device)
	if err == mgo.ErrNotFound {
		return nil, nil
	}
	return &device, err
}
