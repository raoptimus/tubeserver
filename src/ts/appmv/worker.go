package main

import (
	"encoding/json"
	"errors"
	"github.com/raoptimus/gserv/config"
	"gopkg.in/mgo.v2/bson"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"time"
	"ts/data"
)

const FILE_APK_TYPE = "application/vnd.android.package-archive"

type (
	Worker struct {
	}
	App struct {
		Application string `json:"application"`
		Description string `json:"description"`
		VersionName string `json:"versionName"`
		VersionCode int    `json:"versionCode"`
	}
)

func (s *Worker) Start() (err error) {
	fileApkConf := config.String("ApkConfigFileName", "ApkConfigFileName")
	fileApkName := config.String("ApkFileName", "ApkFileName")

	//parse json with info
	b, err := ioutil.ReadFile(fileApkConf)
	if err != nil {
		return err
	}
	var app App
	err = json.Unmarshal(b, &app)
	if err != nil {
		return err
	}
	//check exists
	build := strconv.Itoa(app.VersionCode)
	n, err := data.Context.Applications.Find(bson.M{"Ver": app.VersionName, "BuildVer": build}).Count()
	if err != nil {
		return err
	}
	if n > 0 {
		return errors.New("Application already exists")
	}
	//copy apk to grid fs
	f, err := os.Open(fileApkName)
	if err != nil {
		return err
	}
	defer f.Close()

	fDb, err := data.Context.Applications.Database.GridFS("fs").Create(fileApkName)
	if err != nil {
		return err
	}
	fDb.SetContentType(FILE_APK_TYPE)
	_, err = io.Copy(fDb, f)
	if err != nil {
		return err
	}
	err = fDb.Close()
	if err != nil {
		return err
	}
	//insert new app record
	appDb := data.Application{
		Id:          fDb.Id().(bson.ObjectId),
		Name:        app.Application,
		Ver:         app.VersionName,
		BuildVer:    build,
		Description: app.Description,
		Status:      data.AppStatusNoChecked,
		AddedDate:   time.Now().UTC(),
	}
	err = data.Context.Applications.Insert(appDb)
	if err != nil {
		return err
	}

	return os.Remove(fileApkName)
}
