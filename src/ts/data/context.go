package data

import (
	"database/sql"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/raoptimus/gserv/config"
	mgo "gopkg.in/mgo.v2"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type (
	context struct {
		//db1
		Dictionary        *mgo.Collection
		VideoCategory     *mgo.Collection
		VideoComments     *mgo.Collection
		Videos            *mgo.Collection
		SearchQueries     *mgo.Collection
		PhotoAlbumHistory *mgo.Collection
		PhotoHistory      *mgo.Collection

		//db2
		Events             *mgo.Collection
		VideoHistory       *mgo.Collection
		DailyStat          *mgo.Collection
		Devices            *mgo.Collection
		DeviceEvents       *mgo.Collection
		PushLog            *mgo.Collection
		PushTask           *mgo.Collection
		Users              *mgo.Collection
		RpcStat            *mgo.Collection
		Channel            *mgo.Collection
		Ads                *mgo.Collection
		Country            *mgo.Collection
		Tariff             *mgo.Collection
		PremiumTransaction *mgo.Collection
		PhotoAlbumComments *mgo.Collection
		PhotoAlbums        *mgo.Collection
		Photos             *mgo.Collection

		//db3
		Applications  *mgo.Collection
		Applications2 *mgo.Collection

		DbTV *sql.DB
	}
)

var Context *context

func initMongoDb(urlRaw string) (*mgo.Database, error) {
	u, err := url.Parse("mongodb://" + urlRaw)
	if err != nil {
		return nil, errors.New("Connection string of mongodb is not correct")
	}
	options := u.Query()
	w := options.Get("w")
	ww := 1
	options.Del("w")
	rs := options.Get("replicaSet")
	rpr := options.Get("readPreference")
	options.Del("readPreference")

	urlRaw = u.Host + u.Path + "?" + options.Encode()
	info, err := mgo.ParseURL(urlRaw)
	if err != nil {
		return nil, errors.New("Connection string of mongodb is not correct: " + err.Error())
	}
	info.Timeout = 1 * time.Minute

	session, err := mgo.DialWithInfo(info)
	if err != nil {
		return nil, errors.New("Can't connection to mongodb (" + urlRaw + "): " + err.Error())
	}
	go func() {
		for {
			session.Refresh()
			time.Sleep(5 * time.Minute)
		}
	}()

	if rs == "" {
		session.SetMode(mgo.Monotonic, true) //Strong no refresh session
	} else {
		if rpr == "primary" {
			session.SetMode(mgo.Monotonic, true)
		} else {
			session.SetMode(mgo.Eventual, true)
		}
	}
	if w != "" {
		ww, err = strconv.Atoi(w)
		if err != nil {
			ww = 1
		}
	}
	session.SetSafe(&mgo.Safe{
		W: ww,
	})
	db := session.DB("")
	return db, nil
}

func initMysqlDb(urlRaw string) (*sql.DB, error) {
	db, err := sql.Open("mysql", urlRaw)
	if err != nil {
		return nil, errors.New("Can't connection to mysqldb (" + urlRaw + "): " + err.Error())
	}
	return db, nil
}

func Init(createIndexes bool) {
	if Context != nil {
		return
	}
	if err := connect(createIndexes); err != nil {
		log.Fatalln(err)
	}
	config.OnAfterLoad("data.Context.reconnect", reconnect)
	log.Println("data.Context initied")
}

func reconnect() {
	err := connect(false)
	if err != nil {
		log.Println(err)
	}
}

func connect(createIndexes bool) error {
	context := &context{}
	log.Println("data.Context db connection...")

	DbTVDbUrl := config.String("MySqlDbTVServer", "")
	if DbTVDbUrl != "" {
		DbTVDb, err := initMysqlDb(DbTVDbUrl)
		if err != nil {
			return err
		}
		context.DbTV = DbTVDb
	}

	//db1
	mongoVideoUrl := config.String("MongoVideoServer", "")
	if mongoVideoUrl != "" {
		mongoDbVideo, err := initMongoDb(mongoVideoUrl)
		if err != nil {
			return err
		}
		context.Dictionary = mongoDbVideo.C("Dictionary")
		//	Context.Events = mongoDbVideo.C("Event")
		//	Context.PhotoAlbumComments = mongoDbVideo.C("PhotoAlbumComment")
		//	Context.PhotoAlbumHistory = mongoDbVideo.C("PhotoAlbumHistory")
		//	Context.PhotoAlbums = mongoDbVideo.C("PhotoAlbum")
		//	Context.PhotoHistory = mongoDbVideo.C("PhotoHistory")
		//	Context.Photos = mongoDbVideo.C("Photo")
		context.Videos = mongoDbVideo.C("Video")
		context.VideoCategory = mongoDbVideo.C("VideoCategory")
		context.VideoComments = mongoDbVideo.C("VideoComment")
		context.SearchQueries = mongoDbVideo.C("SearchQuery")
	}
	//db2
	mongoAllUrl := config.String("MongoAllServer", "")
	if mongoAllUrl != "" {
		mongoDbAll, err := initMongoDb(mongoAllUrl)
		if err != nil {
			return err
		}
		context.DailyStat = mongoDbAll.C("DailyStat")
		context.Devices = mongoDbAll.C("Device")
		context.DeviceEvents = mongoDbAll.C("DeviceEvent")
		context.PushLog = mongoDbAll.C("PushLog")
		context.PushTask = mongoDbAll.C("PushTask")
		context.Users = mongoDbAll.C("User")
		context.VideoHistory = mongoDbAll.C("VideoHistory")
		context.RpcStat = mongoDbAll.C("RpcStat")
		context.Channel = mongoDbAll.C("Channel")
		context.Ads = mongoDbAll.C("Ads")
		context.Country = mongoDbAll.C("Country")
		context.Tariff = mongoDbAll.C("Tariff")
		context.PremiumTransaction = mongoDbAll.C("PremiumTransaction")
	}
	//db3
	mongoUpdate := config.String("MongoUpdateServer", "")
	if mongoUpdate != "" {
		mongoDbUpdate, err := initMongoDb(mongoUpdate)
		if err != nil {
			return err
		}
		context.Applications = mongoDbUpdate.C("Application")
	}
	//db3.2
	mongoUpdate2 := config.String("MongoUpdateServer2", "")
	if mongoUpdate2 != "" {
		mongoDbUpdate2, err := initMongoDb(mongoUpdate2)
		if err != nil {
			return err
		}
		context.Applications2 = mongoDbUpdate2.C("Application")
	}

	Context = context

	if createIndexes {
		mongoDbCreateIndexes()
	}
	return nil
}

func mongoDbCreateIndexes() {
	log.Println("data.Context db indexing...")
	var (
		ix  mgo.Index
		err error
	)
	//Create indexes
	//===========================================
	//===========================================

	//RpcStat
	//===========================================
	ix = mgo.Index{
		Key:        []string{"-Data"},
		Background: true,
	}
	err = Context.RpcStat.EnsureIndex(ix)
	if err != nil {
		log.Fatalln(err)
	}

	//SearchQuery
	//===========================================
	ix = mgo.Index{
		Key:        []string{"_id", "-ResultCount", "-SearchCount"},
		Background: true,
	}
	err = Context.SearchQueries.EnsureIndex(ix)
	if err != nil {
		log.Fatalln(err)
	}

	//PushLog
	//===========================================
	ix = mgo.Index{
		Key:        []string{"Token", "TaskId", "-SendedDate"},
		Background: true,
	}
	err = Context.PushLog.EnsureIndex(ix)
	if err != nil {
		log.Fatalln(err)
	}

	//DailyStat
	//===========================================
	ix = mgo.Index{
		Key:        []string{"-Date"},
		Background: true,
	}
	err = Context.DailyStat.EnsureIndex(ix)
	if err != nil {
		log.Fatalln(err)
	}
	//===========================================

	//Videos
	//===========================================
	ix = mgo.Index{
		Key:        []string{"Source._id"},
		Unique:     true,
		DropDups:   false,
		Background: true,
		Sparse:     true,
	}
	err = Context.Videos.EnsureIndex(ix)
	if err != nil {
		log.Fatalln(err)
	}
	//drop old full text indexes
	Context.Videos.DropIndex("$text:Title.Quote", "$text:Desc.Quote", "$text:Tags")
	Context.Videos.DropIndex("$text:Title.Quote", "$text:Desc.Quote", "$text:Tags", "Filters")
	//
	ix = mgo.Index{
		Key:        []string{"$text:Title.Quote", "$text:Desc.Quote", "$text:Keywords", "Filters"},
		Weights:    map[string]int{"Title.Quote": 16, "Desc.Quote": 8, "Keywords": 4},
		Background: true,
		Name:       "TextIndex",
	}
	err = Context.Videos.EnsureIndex(ix)
	if err != nil {
		if strings.Index(err.Error(), "exists") == -1 {
			log.Fatalln(err)
		}
	}

	ix = mgo.Index{
		Key:        []string{"Filters", "-PublishedDate"},
		Background: true,
	}
	err = Context.Videos.EnsureIndex(ix)
	if err != nil {
		log.Fatalln(err)
	}

	ix = mgo.Index{
		Key:        []string{"Filters", "-UpdateDate"},
		Background: true,
	}
	err = Context.Videos.EnsureIndex(ix)
	if err != nil {
		log.Fatalln(err)
	}
	ix = mgo.Index{
		Key:        []string{"Filters", "-Rank"},
		Background: true,
	}
	err = Context.Videos.EnsureIndex(ix)
	if err != nil {
		log.Fatalln(err)
	}
	//===========================================

	//VideoHistory
	//===========================================
	ix = mgo.Index{
		Key:        []string{"UserId", "Type", "-AddedDate"},
		Background: true,
	}
	err = Context.VideoHistory.EnsureIndex(ix)
	if err != nil {
		log.Fatalln(err)
	}
	//===========================================

	//Applications
	//===========================================
	ix = mgo.Index{
		Key:        []string{"-Ver", "-BuildVer"},
		Background: true,
	}
	err = Context.Applications.EnsureIndex(ix)
	if err != nil {
		log.Fatalln(err)
	}
	if Context.Applications2 != nil {
		err = Context.Applications2.EnsureIndex(ix)
		if err != nil {
			log.Fatalln(err)
		}
	}
	//===========================================

	//Device Events
	//===========================================
	ix = mgo.Index{
		Key:        []string{"Action", "DeviceId", "-AddedDate"},
		Background: true,
	}
	err = Context.DeviceEvents.EnsureIndex(ix)
	if err != nil {
		log.Fatalln(err)
	}
	ix = mgo.Index{
		Key:         []string{"AddedDate"},
		Background:  true,
		ExpireAfter: 24 * 30 * time.Hour, //1 month
	}
	err = Context.DeviceEvents.EnsureIndex(ix)
	if err != nil {
		log.Fatalln(err)
	}
	//===========================================

	//User
	//===========================================
	ix = mgo.Index{
		Key:        []string{"Tokens"},
		Background: true,
		Unique:     true,
	}
	err = Context.Users.EnsureIndex(ix)
	if err != nil {
		log.Fatalln(err)
	}
	//===========================================
	err = CreateDefaultGuestUser()
	if err != nil {
		log.Fatalln(err)
	}
}
