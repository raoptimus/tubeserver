package main

import (
	"database/sql"
	"errors"
	"fmt"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"strconv"
	"strings"
	"time"
	"ts/data"
)

type (
	photoImporter struct {
		lastSrcId        int
		srcDb            *sql.DB
		domain           string
		sourceCategories string
	}
)

func New() (*photoImporter, error) {
	s := &photoImporter{}
	s.srcDb = data.Context.DbTV
	s.domain = DOMAIN
	lid, err := s.getLastSrcId()
	if err != nil {
		return nil, err
	}
	s.lastSrcId = lid
	s.sourceCategories = strings.Join(Mem.VideoCategory.GetSourceIds(), ",")
	return s, nil
}

func (s *photoImporter) Start() {
	for {
		fmt.Println("-> move data begining")
		s.moveData()
		fmt.Println("-> move data finish")
		time.Sleep(1 * time.Hour)
	}
}

func (s *photoImporter) moveData() {
	fmt.Println("domain=", s.domain)
	fmt.Println("lastSrcId=", s.lastSrcId)

	if err := s.srcDb.Ping(); err != nil {
		log.Err(err.Error())
		return
	}

	rows, err := s.srcDb.Query(`SELECT
				PhotoAlbumId, Title, Description,
	 			UNIX_TIMESTAMP(PublishedDate),
				CategoryId, IsPremium,
				MainPhotoId, ApprovedPhotoCount, FavoriteCount, ViewCount, IFNULL(StringTags, ''), IFNULL(StringModels, '')
			FROM tbh_photo_albums
			WHERE Approved = 1 AND IsPublic = 1 AND IsIndividual = 0 AND CategoryId IN(`+s.sourceCategories+`)
			AND PublishedDate >= (SELECT PublishedDate FROM tbh_photo_albums WHERE PhotoAlbumId >= ? LIMIT 1)
			AND PublishedDate <= NOW()
			ORDER BY PublishedDate`, s.lastSrcId)

	if err != nil {
		log.Err(err.Error())
		return
	}

	for rows.Next() {
		var srcId int
		var title string
		var desc string
		var p int64
		var catId int
		var isPremium int
		var mainPhotoId int
		var photoCount int
		var favoriteCount int
		var viewCount int
		var strTags string
		var strModels string

		err = rows.Scan(&srcId, &title, &desc, &p, &catId, &isPremium,
			&mainPhotoId, &photoCount, &favoriteCount, &viewCount, &strTags, &strModels)

		if err != nil {
			log.Err("Scan row photo albums error:" + err.Error())
			return
		}

		published := time.Unix(p, 0)
		src := data.NewPhotoAlbumSource(srcId, s.domain, photoCount, mainPhotoId)

		if s.existsPhotoAlbum(src) {
			s.lastSrcId = srcId
			continue
		}

		pa, err := data.NewPhotoAlbum()

		if err != nil {
			log.Err("New photo album error:" + err.Error())
			return
		}

		pa.Source = src

		if err != nil {
			log.Err("New photo album error:" + err.Error())
			return
		}

		cat, err := Mem.VideoCategory.FindSource(catId)
		if err != nil {
			log.Err(err.Error())
			return
		}

		tags := ""

		for _, c := range cat.Title {
			tags += strings.ToLower(c.Quote) + ","
		}

		tags += strTags + "," + strModels

		pa.Tags = uniqLower(tags)
		pa.Title = data.TextList{
			&data.Text{
				Quote:    title,
				Language: data.LanguageRussian,
			}}
		pa.Desc = data.TextList{
			&data.Text{
				Quote:    desc,
				Language: data.LanguageRussian,
			}}
		pa.CategoryId = cat.Id
		pa.IsPremium = isPremium == 1
		pa.LikeCount = favoriteCount
		pa.ViewCount = viewCount
		pa.Featured = published.Year() >= 13 || favoriteCount > 83
		pa.PublishedDate = published

		pa.Comments, err = s.getPhotoAlbumComments(srcId, pa.Id)

		if err != nil {
			log.Err("Get photo album (" + strconv.Itoa(srcId) + ") comment list error:" + err.Error())
			return
		}

		pa.CommentCount = len(pa.Comments)

		pa.Photos, pa.MainPhotoId, err = s.getPhotos(srcId, pa.Id, mainPhotoId)

		if err != nil {
			log.Err("Get photo album (" + strconv.Itoa(srcId) + ") relates list error:" + err.Error())
			return
		}

		pa.PhotoCount = len(pa.Photos)

		err = data.Context.PhotoAlbums.Insert(pa)
		if err != nil {
			log.Err("Insert photo album  error: " + err.Error())
			return
		}

		for _, ph := range pa.Photos {
			if err := data.Context.Photos.Insert(ph); err != nil {
				log.Err("Insert photo error:" + err.Error())
				return
			}
		}

		for _, comm := range pa.Comments {
			if err := data.Context.PhotoAlbumComments.Insert(comm); err != nil {
				log.Err("Insert photo comment error:" + err.Error())
				return
			}
		}

		s.lastSrcId = srcId
	}
}

func (s *photoImporter) getPhotos(srcId int, photoAlbumId int, srcMainPhotoId int) (pl []*data.Photo, mainPhotoId int, err error) {
	err = s.srcDb.Ping()

	if err != nil {
		return nil, 0, err
	}

	rows, err := s.srcDb.Query(`SELECT PhotoId, Width, Height, FileHash FROM tbh_photos WHERE PhotoAlbumId = ?`, srcId)

	if err != nil {
		return nil, 0, err
	}

	pl = make([]*data.Photo, 0)

	for rows.Next() {
		var photoId int
		var width int
		var height int
		var fileHash string

		err = rows.Scan(&photoId, &width, &height, &fileHash)

		if err != nil {
			return nil, 0, errors.New("Scan row relate error:" + err.Error())
		}

		p, err := data.NewPhoto(photoAlbumId, fileHash, width, height)

		if err != nil {
			return nil, 0, errors.New("New photo error:" + err.Error())
		}

		if photoId == srcMainPhotoId {
			mainPhotoId = p.Id
		}

		pl = append(pl, p)
	}

	if mainPhotoId == 0 {
		mainPhotoId = pl[0].Id
	}

	return pl, mainPhotoId, nil
}

func (s *photoImporter) getPhotoAlbumComments(srcId, photoAlbumId int) (pcl []*data.PhotoAlbumComment, err error) {
	err = s.srcDb.Ping()

	if err != nil {
		return nil, err
	}

	rows, err := s.srcDb.Query(`SELECT Body, UNIX_TIMESTAMP(AddedDate) FROM tbh_photo_comments WHERE Approved = 1 AND PhotoAlbumId = ?`, srcId)

	if err != nil {
		return nil, err
	}

	pcl = make([]*data.PhotoAlbumComment, 0)

	for rows.Next() {

		var body string
		var p int64

		err = rows.Scan(&body, &p)

		if err != nil {
			return nil, errors.New("Scan row comment error:" + err.Error())
		}

		addedDate := time.Unix(p, 0)
		pc, err := data.NewPhotoAlbumComment(photoAlbumId)

		if err != nil {
			return nil, errors.New("Create photo comment error:" + err.Error())
		}
		pc.PostDate = addedDate
		pc.Body = body
		pc.Language = data.LanguageRussian
		pc.Status = data.CommentStatusApproved
		pc.UserId = data.USER_ID_GUEST

		pcl = append(pcl, pc)
	}

	return pcl, nil
}

func (s *photoImporter) existsPhotoAlbum(src *data.PhotoAlbumSource) (exists bool) {
	n, err := data.Context.PhotoAlbums.Find(bson.M{"Source._id": src.Id}).Count()

	if err != nil {
		return false
	}

	return n > 0
}

func (s *photoImporter) getLastSrcId() (int, error) {
	//todo сохранять в конфиг базу посл Source и брать от туда
	pa := &data.PhotoAlbum{}
	err := data.Context.PhotoAlbums.Find(bson.M{"Source.Domain": s.domain}).Select(bson.M{"Source": 1}).Sort("-ImportDate").One(&pa)

	if err != nil {
		if err == mgo.ErrNotFound {
			return 0, nil
		}
		return 0, err
	}

	return pa.Source.SourceId, nil
}
