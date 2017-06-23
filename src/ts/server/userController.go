package main

import (
	"encoding/json"
	"time"
	"ts/data"
	api "ts/protocol/v1"

	"gopkg.in/mgo.v2/bson"
)

type UserController struct {
}

func (s *UserController) GetInfo(req *api.Request, u *api.User) error {
	du := &data.User{}
	err := du.GetOrCreate(req.Token)
	if err != nil {
		return err
	}

	*u = api.User{
		Id:           du.Id,
		UserName:     du.Author(),
		Tokens:       du.Tokens,
		Avatar:       string(du.Avatar.Data),
		Tel:          du.Tel,
		Email:        du.Email,
		Lang:         du.Language,
		CreationDate: du.CreationDate.Unix(),
		Premium:      s.premium(du),
	}

	return nil
}

func (s *UserController) UpdateInfo(req *api.Request, unused *string) error {
	token, obj := req.Token, req.Object
	b, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	var u api.User
	if err := json.Unmarshal(b, &u); err != nil {
		return err
	}

	us := &data.User{}
	if err := us.GetOrCreate(token); err != nil {
		return err
	}

	update := bson.M{}

	set := func(field, from string) {
		if from != "" {
			update[field] = from
		}
	}

	set("UserName", u.UserName)
	set("Tel", u.Tel)
	set("Email", u.Email)

	if u.Avatar != "" {
		update["Avatar"] = []byte(u.Avatar)
	}

	if u.Lang != "" {
		update["Language"] = u.Lang
	}

	if len(update) == 0 {
		return nil
	}
	return data.Context.Users.UpdateId(us.Id, bson.M{"$set": update})
}

func (s *UserController) GetPremiumStatus(req *api.Request, premium *api.Premium) error {
	user := &data.User{}
	if err := user.GetOrCreate(req.Token); err != nil {
		return err
	}
	*premium = s.premium(user)
	return nil
}

func (s *UserController) premium(user *data.User) api.Premium {
	var duration time.Duration
	if user.Premium.Expired() {
		duration = user.Premium.Duration()
	}
	if duration == 0 {
		user.Premium.Type = data.PremiumTypeNone
	}
	return api.Premium{
		Duration: uint64(duration / time.Millisecond),
		Type:     user.Premium.Type,
	}
}
