package data

import (
	"errors"
	"github.com/raoptimus/gserv/config"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"strconv"
	"time"
	"ts/mongodb"
)

const USER_ID_GUEST int = 1
const UNIX_TIME_1970 int64 = -10800
const USER_PREF_GUEST = "Guest"
const USER_ANON = "Anonymous"

type (
	User struct {
		Id           int         `bson:"_id"`
		UserName     string      `bson:"UserName"`
		Tokens       []*string   `bson:"Tokens"`
		Avatar       bson.Binary `bson:"Avatar"`
		Tel          string      `bson:"Tel"`
		Email        string      `bson:"Email"`
		Language     Language    `bson:"Language"`
		CreationDate time.Time   `bson:"CreationDate"`
		Premium      Premium     `bson:"Premium"`
	}
)

func (s *User) Author() string {
	if s.UserName != "" {
		return s.UserName
	}

	if s.Id == 1 {
		return USER_ANON
	}

	return USER_PREF_GUEST + strconv.Itoa(s.Id)
}

func GetUserByToken(token string) (u *User, err error) {
	err = Context.Users.Find(bson.M{"Tokens": token}).One(&u)
	return
}

func (s *User) GetOrCreate(token string) (err error) {
	err = Context.Users.Find(bson.M{"Tokens": token}).One(&s)

	if err == mgo.ErrNotFound {
		s.Tokens = []*string{&token}
		err = s.Insert()
	}

	return
}

func (s *User) Insert() error {
	if len(s.Tokens) == 0 {
		return errors.New("Tokens is empty")
	}

	if s.Id <= 0 {
		var err error
		s.Id, err = mongodb.GetNewIncId(Context.Users)
		if err != nil {
			return err
		}
	}

	s.CreationDate = time.Now().UTC()
	trialHours := config.Int("TrialDays", 0) * 24
	s.Premium.setTrial(trialHours)
	if err := Context.Users.Insert(s); err != nil {
		return err
	}
	if s.Premium.Type == PremiumTypeTrial {
		s.insertTrialTransaction(trialHours)
	}
	return nil
}

func (s *User) insertTrialTransaction(hours int) {
	tr := Transaction{
		UserId:    s.Id,
		Price:     0,
		Duration:  hours,
		Type:      PremiumTypeTrial,
		AddedDate: time.Now(),
	}
	if err := Context.PremiumTransaction.Insert(tr); err != nil {
		// well mongo sucks
	}
}

func UserExists(id int) (bool, error) {
	n, err := Context.Users.FindId(id).Count()
	return n > 0, err
}

func GetUserId(token string) (id int, err error) {
	var u User
	err = Context.Users.Find(bson.M{"Tokens": token}).Select(bson.M{"_id": 1}).One(&u)

	if err == nil {
		id = u.Id
	}

	return
}

func CreateDefaultGuestUser() error {
	if exists, _ := UserExists(1); !exists {
		_, err := mongodb.GetNewIncId(Context.Users)
		if err != nil {
			return err
		}

		token := ""
		u := &User{Id: 1, Tokens: []*string{&token}}
		return u.Insert()
	}

	return nil
}
