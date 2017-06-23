package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"github.com/raoptimus/gserv/config"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"ts/data"
	api "ts/protocol/v1"
)

type (
	AppController struct {
	}
)

func (s *AppController) GetRequest(req *api.Request, res *api.Request) error {
	*res = *req
	return nil
}

func (s *AppController) LastBuild(req *api.Request, app *api.Application) error {
	token, q := req.Token, req.Query
	if q == nil {
		return errors.New("Query is empty")
	}

	ver, ok := (*q)["Ver"].(string)
	if !ok {
		return errors.New("Query.Ver is incorrect")
	}
	name, ok := (*q)["Name"].(string)
	if !ok {
		name = "apk1"
	}
	if name == "apk2" {
		return mgo.ErrNotFound
	}
	withApk, ok := (*q)["WithApk"].(bool)
	if !ok {
		withApk = true
	}

	isBeta := false
	for _, t := range BetaTokenList {
		if token == t {
			isBeta = true
			break
		}
	}

	where := bson.M{}
	if !isBeta {
		where["Status"] = data.AppStatusChecked
	}
	where["Ver"] = bson.M{"$gt": ver}

	var appDb data.Application
	c := data.Context.Applications
	downloadFile := "/download.php"

	if req.Project == "pro....." {
		c = data.Context.Applications2
		downloadFile = "/download_pro.php"
	}
	err := c.Find(where).Sort("-Ver", "-BuildVer").Limit(1).One(&appDb)
	if err != nil {
		return err
	}

	var apk string
	if withApk {
		file, err := c.Database.GridFS("fs").OpenId(appDb.Id)
		if err != nil {
			return err
		}
		var buf bytes.Buffer
		_, err = buf.ReadFrom(file)
		file.Close()
		apk = base64.StdEncoding.EncodeToString(buf.Bytes())
	}
	appUrl := config.String("AppUrl", "")
	if appUrl == "" {
		return errors.New("Cannot build download url: AppUrl is empty")
	}
	isAllow := true

	switch {
	case appUrl == "http://....tv" && ver < "2.7":
		isAllow = false
	case appUrl == "http://......" && ver < "2.1":
		isAllow = false
	}

	*app = api.Application{
		Name:         appDb.Name,
		Ver:          appDb.Ver,
		BuildVer:     appDb.BuildVer,
		Description:  appDb.Description,
		Url:          appUrl + downloadFile + "?id=" + appDb.Id.Hex() + "&v=" + appDb.Ver,
		Apk:          apk,
		CurrVerAllow: isAllow,
	}

	return nil
}

func (s *AppController) IpList(req *api.Request, list *api.AddrList) error {
	*list = Ips.List()
	return nil
}
