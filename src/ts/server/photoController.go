package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	//"math/rand"
	"time"
	"ts/data"
	"ts/mongodb"
	api "ts/protocol/v1"
)

type PhotoController struct{}

func (s *PhotoController) List(req *api.Request, list *api.PhotoAlbumList) error {
	token, query, sorter, page, lang := req.Token, req.Query, req.Sort, req.Page, req.Lang
	criteria := bson.M{"Featured": true}
	err := s.convertQueryToCriteria(query, lang, &criteria)
	if err != nil {
		return err
	}
	q := data.Context.PhotoAlbums.Find(criteria)

	if sorter != nil {
		f := sorter.Field
		switch f {
		case "Id":
			{
				f = "_id"
			}
		case "PublishedDate":
			{
				//nothing
			}
		default:
			return errors.New("Sort deny for the field " + f)
		}

		if sorter.Direct == api.SortDirectDesc {
			f = "-" + f
		}
		q = q.Sort(f)
	}

	var dbList []data.PhotoAlbum
	err = q.Skip(page.Skip).Limit(page.Limit).All(&dbList)
	if err != nil {
		return err
	}

	l := len(dbList)
	if l == 0 {
		return nil
	}

	userId, _ := data.GetUserId(token)
	*list = make(api.PhotoAlbumList, l)
	pIdList := make([]bson.ObjectId, l)

	for i, v := range dbList {
		(*list)[i] = s.convertPhotoAlbum(&v, lang)
		(*list)[i].Photos = make([]*api.PhotoMain, 3)

		for j := 0; j < 3; j++ {
			(*list)[i].Photos[j] = &api.PhotoMain{
				Url: v.Photos[j].Url(),
			}
			//fmt.Println(rand.Intn(20))
		}

		if userId > 0 {
			h := data.NewPhotoAlbumHistory(userId, v.Id, data.PhotoAlbumHistoryTypeLike)
			pIdList[i] = h.Id
		}
	}

	if userId == 0 {
		return nil
	}

	var historyList []data.PhotoAlbumHistory
	err = data.Context.PhotoAlbumHistory.
		Find(bson.M{"_id": bson.M{"$in": pIdList}}).
		Select(bson.M{"PhotoAlbumId": 1}).
		All(&historyList)
	if err != nil {
		return err
	}

	hl := len(historyList)
	if hl == 0 {
		return nil
	}

	for i := 0; i < l; i++ {
		for k := 0; i < hl; i++ {
			(*list)[i].IsLiked = ((*list)[i].Id == historyList[k].PhotoAlbumId)
		}
	}
	return nil
}

func (s *PhotoController) One(req *api.Request, one *api.PhotoAlbum) error {
	query, lang := req.Query, req.Lang
	if query == nil {
		return errors.New("Query is empty")
	}

	id, ok := (*query)["Id"]
	if !ok {
		return errors.New("Id not found in query")
	}
	photoAlbumId := int(id.(float64))
	pa := &data.PhotoAlbum{}
	err := pa.LoadById(photoAlbumId)
	if err != nil {
		return err
	}

	comments := make(api.PhotoAlbumComments, len(pa.Comments))
	for i, comm := range pa.Comments {
		comments[i] = &api.PhotoAlbumComment{
			Id:           comm.Id,
			PhotoAlbumId: pa.Id,
			Author:       comm.User.Author(),
			Avatar:       string(comm.User.Avatar.Data),
			Body:         comm.Body,
			PostDate:     comm.PostDate.Unix(),
		}
	}
	comments = comments.Sort()

	*one = *s.convertPhotoAlbum(pa, lang)
	one.Comments = comments
	/*
		userId, _ := data.GetUserId(token)

		if userId == 0 {
			return nil
		}

		lc, err := data.Context.PhotoAlbumHistory.Find(bson.M{"UserId": userId, "PhotoAlbumId": one.Id, "Type": data.PhotoAlbumHistoryTypeLike}).Count()

		if err != nil {
			return err
		}

		one.IsLiked = lc > 0
	*/
	return nil
}

func (s *PhotoController) GetPhotos(req *api.Request, list *api.PhotoList) error {
	token, query, page := req.Token, req.Query, req.Page

	photoAlbumId, ok := (*query)["PhotoAlbumId"]
	if !ok {
		return errors.New("Id not found in query")
	}
	q := data.Context.Photos.Find(bson.M{"PhotoAlbumId": photoAlbumId})

	var dbList []data.Photo
	err := q.Skip(page.Skip).Limit(page.Limit).All(&dbList)
	if err != nil {
		return err
	}

	l := len(dbList)
	if l == 0 {
		return nil
	}

	userId, _ := data.GetUserId(token)
	*list = make(api.PhotoList, l)
	pIdList := make([]bson.ObjectId, l)

	for i, v := range dbList {
		(*list)[i] = s.convertPhoto(&v)

		if userId > 0 {
			h := data.NewPhotoHistory(userId, v.Id, data.PhotoHistoryTypeLike)
			pIdList[i] = h.Id
		}
	}

	if userId == 0 {
		return nil
	}

	var historyList []data.PhotoHistory
	err = data.Context.PhotoHistory.
		Find(bson.M{"_id": bson.M{"$in": pIdList}}).
		Select(bson.M{"PhotoId": 1}).
		All(&historyList)
	if err != nil {
		return err
	}

	hl := len(historyList)
	if hl == 0 {
		return nil
	}

	for i := 0; i < l; i++ {
		for k := 0; i < hl; i++ {
			(*list)[i].IsLiked = ((*list)[i].Id == historyList[k].PhotoId)
		}
	}
	return nil
}

func (s *PhotoController) WriteComment(req *api.Request, comm *api.PhotoAlbumComment) error {
	token, obj := req.Token, req.Object
	if obj == nil {
		return errors.New("Object is empty")
	}
	b, err := json.Marshal(obj)
	if err != nil {
		return errors.New("Object is not a comment")
	}
	if err := json.Unmarshal(b, &comm); err != nil {
		return errors.New("Object is not a comment")
	}

	if err := comm.Validate(); err != nil {
		return err
	}

	pc, err := data.NewPhotoAlbumComment(comm.PhotoAlbumId)
	if err != nil {
		return err
	}

	exists, err := data.PhotoAlbumExists(comm.PhotoAlbumId)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("Photo album is not exists")
	}

	u := &data.User{}
	err = u.GetOrCreate(token)
	if err != nil {
		return err
	}

	pc.PostDate = time.Now().UTC()
	pc.Language = u.Language
	pc.Body = comm.Body
	pc.Status = data.CommentStatusApproved
	pc.UserId = u.Id

	if err := data.Context.PhotoAlbumComments.Insert(pc); err != nil {
		return err
	}

	up := bson.M{"$push": bson.M{"Comments": pc}, "$inc": bson.M{"CommentCount": 1}}
	err = data.Context.PhotoAlbums.UpdateId(pc.PhotoAlbumId, up)
	if err != nil {
		return err
	}

	comm.Id = pc.Id
	comm.Author = u.Author()
	comm.Avatar = string(u.Avatar.Data)
	comm.PostDate = pc.PostDate.Unix()

	return nil
}

func (s *PhotoController) ToggleLike(req *api.Request, res *api.PhotoActionResponse) error {
	return s.incAction(data.PhotoAlbumHistoryTypeLike, req, res)
}

func (s *PhotoController) IncView(req *api.Request, res *api.PhotoActionResponse) error {
	return s.incAction(data.PhotoAlbumHistoryTypeView, req, res)
}

func (s *PhotoController) incAction(t data.PhotoAlbumHistoryType, req *api.Request, res *api.PhotoActionResponse) error {
	query := req.Query
	if query == nil {
		return errors.New("Query is empty")
	}

	id, ok := (*query)["Id"]
	if !ok {
		return errors.New("Id not found in query")
	}

	fid, ok := id.(float64)
	if !ok {
		return errors.New("Id is not integer")
	}
	pId := int(fid)
	if pId <= 0 {
		return errors.New("Video doesn't exists")
	}
	exists, err := data.VideoExists(pId)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("Video doesn't exists")
	}

	u := &data.User{}
	err = u.GetOrCreate(req.Token)
	if err != nil {
		return err
	}

	exists, err = data.PhotoAlbumExists(pId)
	if err != nil {
		return err
	}
	if !exists {
		errors.New("Video not found")
	}

	h := data.NewPhotoAlbumHistory(u.Id, pId, t)
	exists, err = h.Exists()
	if err != nil {
		return err
	}

	var field string

	switch t {
	case data.PhotoAlbumHistoryTypeView:
		field = "ViewCount"
	case data.PhotoAlbumHistoryTypeLike:
		field = "LikeCount"
	default:
		return errors.New("Action not found")
	}

	if exists {
		switch t {
		case data.PhotoAlbumHistoryTypeLike:
			{
				err := data.Context.PhotoAlbumHistory.RemoveId(&h.Id)
				if err != nil {
					return err
				}

				res.Action = api.PhotoActionRemoved
			}
		}

		var ret map[string]interface{}
		err := data.Context.PhotoAlbums.FindId(pId).Select(bson.M{field: 1}).One(&ret)
		if err != nil {
			return err
		}
		res.TotalCount = ret[field].(int)
	} else {
		err = data.Context.PhotoAlbumHistory.Insert(&h)
		if err != nil {
			return err
		}

		err = mongodb.IncAndGet(data.Context.PhotoAlbums.FindId(pId), field, 1, &res.TotalCount)

		if err != nil {
			return err
		}

		res.Action = api.PhotoActionAdded
	}

	return nil
}

func (s *PhotoController) convertQueryToCriteria(query *api.Query, lang data.Language, where *bson.M) error {
	if query == nil {
		return nil
	}
	for k, v := range *query {
		switch k {
		case "IsPremium":
			{
				(*where)[k] = v.(bool)
			}
		case "CategoryId":
			{
				(*where)[k] = int(v.(float64))
			}
		case "$text":
			{
				(*where)["$text"] = bson.M{
					"$search": v.(string),
					//todo not work with ru
					//					"$language": lang,
				}
			}
		case "Id":
			{
				vmap, ok := v.(map[string]interface{})
				if !ok {
					return errors.New("$in value is not a map")
				}
				for mk, mv := range vmap {
					switch mk {
					case "$in":
						{
							iids, ok := mv.([]interface{})
							if !ok {
								return errors.New("Id.$in value is not a slice")
							}
							ids := make([]int, len(iids))
							for i, id := range iids {
								vid, ok := id.(float64)
								if !ok {
									return errors.New("Id.$in value is not a int[] slice - " + fmt.Sprintf("%#v", id))
								}
								ids[i] = int(vid)
							}

							(*where)["_id"] = bson.M{"$in": ids}
						}
					default:
						{
							return errors.New("Field '$in." + mk + "' not accept for the query")
						}
					}
				}
			}
		default:
			{
				return errors.New("Field '" + k + "' not accept for the query")
			}
		}

	}

	return nil
}

func (s *PhotoController) convertPhotoAlbum(v *data.PhotoAlbum, lang data.Language) *api.PhotoAlbum {
	pa := &api.PhotoAlbum{
		Id:            v.Id,
		Title:         v.Title.Get(lang, false),
		Desc:          v.Desc.Get(lang, false),
		CategoryId:    v.CategoryId,
		PublishedDate: v.PublishedDate.Unix(),
		IsPremium:     v.IsPremium,
		LikeCount:     v.LikeCount,
		ViewCount:     v.ViewCount,
		CommentCount:  v.CommentCount,
		PhotoCount:    v.PhotoCount,
	}

	return pa
}

func (s *PhotoController) convertPhoto(v *data.Photo) *api.Photo {
	p := &api.Photo{
		Id:        v.Id,
		AlbumId:   v.AlbumId,
		Url:       v.Url(),
		LikeCount: v.LikeCount,
		IsLiked:   false,
	}

	return p
}
