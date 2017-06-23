package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/raoptimus/gserv/config"
	"gopkg.in/mgo.v2/bson"
	"strconv"
	"strings"
	"time"
	"ts/data"
)

const LIMIT = 100

type (
	videoImporter struct {
		recordCount      int
		srcDb            *sql.DB
		domain           string
		videoMakeRelates *VideoMakeRelates
	}
)

func NewVideoImporter() (*videoImporter, error) {
	s := &videoImporter{}
	s.srcDb = data.Context.DbTV
	s.domain = DOMAIN
	s.videoMakeRelates = &VideoMakeRelates{}
	s.recordCount = s.getDefaultSkip()
	fmt.Println("Starting from skip=", s.recordCount)
	return s, nil
}

func (s *videoImporter) getDefaultSkip() int {
	n, _ := data.Context.Videos.Count()
	return n / 2
}

func (s *videoImporter) Start() {
	log.Info("Video importer working now")
	s.videoMakeRelates.Resume()
	for {
		fmt.Println("-> move data begining")
		before := s.recordCount
		s.moveData()

		if s.recordCount > before {
			s.videoMakeRelates.Resume()
		}

		fmt.Println("-> move data finish")

		if s.recordCount == before {
			time.Sleep(1 * time.Hour)
			s.recordCount = s.getDefaultSkip()
		}
	}
}

func (s *videoImporter) moveData() {
	if err := s.srcDb.Ping(); err != nil {
		log.Err(fmt.Sprintf("MysqlDb ping error: %v", err))
		return
	}
	srcCatList := strings.Join(Mem.VideoCategory.GetSourceIds(), ",")
	q := `SELECT
				VideoId, Title, Description, Duration,
	 			UNIX_TIMESTAMP(PublishedDate),
				CategoryId, IsPremium,
				MP4Low, IFNULL(MP4Medium, ''), IFNULL(MP4High, ''),
				LowSpace, MediumSpace, HighSpace,
				ThumbSelect, ThumbCount, FavoriteCount, ViewCount,
				IFNULL(StringTags, ''), IFNULL(StringModels, ''),
				FileHash
			FROM tbh_videos
			WHERE Approved = 1
			    AND ContentCopied = 1
			    AND Hidden = 0
			    AND CategoryId IN(` + srcCatList + `)
			    AND MP4Low NOT LIKE '%.flv'
			    AND PublishedDate <= NOW()
			ORDER BY VideoId
			LIMIT ?, ?`
	rows, err := s.srcDb.Query(q, s.recordCount, LIMIT)
	if err != nil {
		log.Err(fmt.Sprintf("Query `%s` error: %v", q, err))
		return
	}
	stats := struct {
		found    int
		inserted int
	}{}

	for rows.Next() {
		stats.found++
		var (
			srcId     int
			title     string
			desc      string
			duration  int64
			p         int64
			catId     int
			isPremium int

			mp4Low     string
			mp4Medium  string
			mp4High    string
			sizeLow    int
			sizeMedium int
			sizeHigh   int

			thumbSelect   int
			thumbCount    int
			favoriteCount int
			viewCount     int
			strTags       string
			strModels     string

			fileHash string
		)

		err = rows.Scan(&srcId, &title, &desc, &duration, &p, &catId, &isPremium,
			&mp4Low, &mp4Medium, &mp4High, &sizeLow, &sizeMedium, &sizeHigh,
			&thumbSelect, &thumbCount, &favoriteCount, &viewCount,
			&strTags, &strModels, &fileHash)

		if err != nil {
			log.Err("Scan row video error:" + err.Error())
			return
		}
		if len(fileHash) != 32 {
			log.Err(fmt.Sprintf("FileHash '%s' by video '%d' can't be blank or length can't < 32", fileHash, srcId))
			continue
		}

		published := time.Unix(p, 0)
		src := data.NewVideoSource(srcId, s.domain, thumbCount, thumbSelect)

		if s.existsVideo(src) {
			s.recordCount++
			continue
		}

		v, err := data.NewVideo()
		v.FilesId = fileHash
		v.Source = src

		if err != nil {
			log.Err("New video error:" + err.Error())
			return
		}

		cat, err := Mem.VideoCategory.FindSource(catId)
		if err != nil {
			log.Err("Find source error:" + err.Error())
			return
		}
		//tags
		strTags += "," + cat.Title.Get(data.LanguageRussian, true)
		v.Tags = data.TagsList{&data.Tags{
			Language: data.LanguageRussian,
			Tags:     uniqLower(strTags),
		}}

		for _, c := range cat.Title {
			if c.Language == data.LanguageRussian {
				continue
			}
			v.Tags = append(v.Tags, &data.Tags{
				Language: c.Language,
				Tags:     []string{strings.ToLower(c.Quote)},
			})
		}
		//
		//actors
		actors := uniqLower(strModels)
		v.Actors = make([]string, 0)
		for _, actor := range actors {
			v.Actors = append(v.Actors, strings.Title(actor))
		}
		//
		//keywords
		excludeKeywords := map[string]bool{
			"DbTV":  true,
			"DbTVo": true,
			"порн":  true,
			"порно": true,
			"секс":  true,
		}
		keywordsMap := make(map[string]bool)
		for _, actor := range actors {
			if _, ok := excludeKeywords[actor]; !ok {
				continue
			}
			keywordsMap[actor] = true
		}
		for _, t := range v.Tags {
			for _, tag := range t.Tags {
				if _, ok := excludeKeywords[tag]; !ok {
					continue
				}
				keywordsMap[tag] = true
			}
		}
		v.Keywords = make([]string, len(keywordsMap))
		ti := 0

		for k, _ := range keywordsMap {
			v.Keywords[ti] = k
			ti++
		}

		//
		v.Title = data.TextList{}
		if config.Bool("AppendTitle", true) {
			v.Title = append(v.Title, &data.Text{
				Quote:    title,
				Language: data.LanguageRussian,
			})
		}
		v.Desc = data.TextList{}
		if config.Bool("AppendDesc", true) {
			v.Desc = append(v.Desc, &data.Text{
				Quote:    desc,
				Language: data.LanguageRussian,
			})
		}
		v.Duration = duration
		v.CategoryId = cat.Id
		v.Files = make([]*data.VideoFile, 0)
		v.LikeCount = favoriteCount
		v.ViewCount = viewCount
		featured := (published.Year() >= 13 || favoriteCount > 83) && (cat.Id != 13 && cat.Id != 10)
		v.Filters = []string{
			"*",
			s.not(config.Bool("AppendAsApproved", true), "approved"),
			"c" + strconv.Itoa(cat.Id),
			s.not(isPremium == 1, "premium"),
			s.not(featured, "featured"),
		}
		v.PublishedDate = published
		if mp4Low != "" {
			v.Files = append(v.Files, data.NewVideoFile(mp4Low, 0, 320, sizeLow))
		}
		if mp4Medium != "" {
			v.Files = append(v.Files, data.NewVideoFile(mp4Medium, 0, 480, sizeMedium))
		}
		if mp4High != "" {
			v.Files = append(v.Files, data.NewVideoFile(mp4High, 0, 720, sizeHigh))
		}

		if config.Bool("ImportComments", true) {
			v.Comments, err = s.getVideoComments(srcId, v.Id)
			if err != nil {
				log.Err("Get video (" + strconv.Itoa(srcId) + ") comment list error:" + err.Error())
				return
			}
			v.CommentCount = len(v.Comments)
		}

		err = data.Context.Videos.Insert(v)
		if err != nil {
			log.Err("Insert video error: " + err.Error())
			return
		}

		s.recordCount++
		stats.inserted++

		err = s.videoMakeRelates.Update(v.Id, v.Keywords)
		if err != nil {
			log.Err("Make relates for video error: " + err.Error())
		}

		for _, comm := range v.Comments {
			if err := data.Context.VideoComments.Insert(comm); err != nil {
				log.Err("Insert video comment error: " + err.Error())
			}
		}
	}

	fmt.Println("Videos found:\t", stats.found)
	fmt.Println("Videos inserted: ", stats.inserted)
	fmt.Println("Videos skipped:\t", stats.found-stats.inserted)
}

func (s *videoImporter) getVideoComments(srcId, videoId int) (vcl []*data.VideoComment, err error) {
	err = s.srcDb.Ping()
	if err != nil {
		return nil, err
	}

	rows, err := s.srcDb.Query(`SELECT Comment, UNIX_TIMESTAMP(AddedDate) FROM tbh_comments WHERE Approved = 1 AND VideoId = ?`, srcId)

	if err != nil {
		return nil, err
	}

	vcl = make([]*data.VideoComment, 0)

	for rows.Next() {

		var body string
		var p int64

		err = rows.Scan(&body, &p)

		if err != nil {
			return nil, errors.New("Scan row comment error:" + err.Error())
		}

		addedDate := time.Unix(p, 0)
		vc, err := data.NewVideoComment(videoId)

		if err != nil {
			return nil, errors.New("Create video comment error:" + err.Error())
		}
		vc.PostDate = addedDate
		vc.Body = body
		vc.Language = data.LanguageRussian
		vc.Status = data.CommentStatusApproved
		vc.UserId = data.USER_ID_GUEST
		vcl = append(vcl, vc)
	}

	return vcl, nil
}

func (s *videoImporter) existsVideo(src *data.VideoSource) (exists bool) {
	n, err := data.Context.Videos.Find(bson.M{"Source._id": src.Id}).Count()

	if err != nil {
		return false
	}

	return n > 0
}

func (s *videoImporter) not(v bool, w string) string {
	if v {
		return w
	} else {
		return "!" + w
	}
}
