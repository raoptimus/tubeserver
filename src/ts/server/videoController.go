package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/raoptimus/gserv/service"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"strconv"
	"time"
	"ts/data"
	api "ts/protocol/v1"
)

type VideoController struct{}

//тест работы сервера с видео для мониторилки
func (s *VideoController) Test(unused *int, result *bool) error {
	n, err := data.Context.Videos.Count()
	*result = n > 0
	return err
}

func (s *VideoController) HistoryClear(req *api.Request, result *bool) error {
	defer service.DontPanic()
	token, query := req.Token, req.Query
	userId, err := data.GetUserId(token)
	if err != nil {
		return err
	}

	criteria := bson.M{"UserId": userId}
	if query != nil {
		t, ok := (*query)["Type"]
		if ok {
			criteria["Type"] = data.VideoHistoryType(t.(float64))
		}
	}

	_, err = data.Context.VideoHistory.RemoveAll(criteria)
	if err != nil {
		return err
	}
	*result = true
	return nil
}

func (s *VideoController) HistoryList(req *api.Request, list *api.VideoList) error {
	defer service.DontPanic()
	token, query, page, lang := req.Token, req.Query, req.Page, req.Lang
	userId, err := data.GetUserId(token)
	if err != nil {
		return errors.New(fmt.Sprintf("Cant get user by token %s, error: %v", token, err))
	}

	criteria := bson.M{"UserId": userId}
	if query != nil {
		t, ok := (*query)["Type"]
		if ok {
			criteria["Type"] = data.VideoHistoryType(t.(float64))
		}
	}

	var histList []*data.VideoHistory
	videoIdList := make([]int, 0)
	err = data.Context.VideoHistory.
		Find(criteria).
		Select(bson.M{"VideoId": 1, "AddedDate": 1, "Type": 1}).
		Sort("-AddedDate").
		Skip(page.Skip).Limit(page.Limit).
		All(&histList)
	if err != nil {
		return errors.New(fmt.Sprintf("Cant get videos by query %v, error: %v", criteria, err))
	}
	for _, h := range histList {
		videoIdList = append(videoIdList, h.VideoId)
	}

	var dbList []data.Video
	//	for _, vid := range videoIdList {
	//		var vv data.Video
	//		err = data.Context.Videos.FindId(vid).Select(bson.M{"Keywords": 0}).One(&vv)
	//		if err != nil {
	//			return errors.New(fmt.Sprintf("Cant get video by id %d, appended %d, error %v", vid, len(dbList), err))
	//		}
	//		dbList = append(dbList, vv)
	//	}
	err = data.Context.Videos.
		Find(bson.M{"_id": bson.M{"$in": videoIdList}}).
		Select(bson.M{"Keywords": 0}).
		All(&dbList)
	if err != nil {
		return errors.New(fmt.Sprintf("Cant get videos $in %v, error: %v", videoIdList, err))
	}

	*list = make(api.VideoList, len(dbList))
	for i, v := range dbList {
		(*list)[i] = s.convertVideo(&v, lang, false, req.Ip)
		(*list)[i].IsLiked = true

		for _, h := range histList {
			if h.VideoId == v.Id {
				(*list)[i].AddedDate = h.AddedDate.Unix()
				(*list)[i].IsLiked = h.Type == data.VideoHistoryTypeLike
				break
			}
		}
	}

	return nil
}

func (s *VideoController) Find(req *api.Request, list *api.VideoList) error {
	defer service.DontPanic()
	q := req.Query
	if q != nil {
		if _, ok := (*q)["$text"]; ok {
			return s.Search(req, list)
		}
	}
	return s.List(req, list)
}

func (s *VideoController) List(req *api.Request, list *api.VideoList) error {
	defer service.DontPanic()
	token, query, sorter, page, lang := req.Token, req.Query, req.Sort, req.Page, req.Lang
	// top category

	if query != nil {
		if cat, ok := (*query)["CategoryId"]; ok && cat == float64(-1) {
			query = nil
			sorter = &api.SortInfo{
				Field:  "Rank",
				Direct: api.SortDirectDesc,
			}
		}
	}

	criteria, err := s.convertQueryToCriteria(query, lang)
	if err != nil {
		return err
	}

	q := data.Context.Videos.Find(criteria)

	if sorter != nil {
		f := sorter.Field
		switch f {
		case "Id":
			{
				f = "_id"
			}
		case "PublishedDate", "UpdateDate", "Rank":
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

	var dbList []data.Video
	err = q.Skip(page.Skip).Limit(page.Limit).Select(bson.M{"Keywords": 0}).All(&dbList)
	if err != nil {
		return err
	}

	l := len(dbList)
	*list = make(api.VideoList, l)
	if l == 0 {
		return nil
	}

	userId, _ := data.GetUserId(token)
	videoIdList := make([]bson.ObjectId, l)

	for i, v := range dbList {
		(*list)[i] = s.convertVideo(&v, lang, false, req.Ip)

		if userId > 0 {
			h := data.NewVideoHistory(userId, v.Id, data.VideoHistoryTypeLike)
			videoIdList[i] = h.Id
		}
	}

	if userId == 0 {
		return nil
	}

	var historyList []data.VideoHistory
	err = data.Context.VideoHistory.
		Find(bson.M{"_id": bson.M{"$in": videoIdList}}).
		Select(bson.M{"VideoId": 1}).
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
			(*list)[i].IsLiked = ((*list)[i].Id == historyList[k].VideoId)
		}
	}
	return nil
}

func (s *VideoController) Count(req *api.Request, count *int) error {
	defer service.DontPanic()
	query := req.Query
	criteria, err := s.convertQueryToCriteria(query, data.LanguageEnglish)
	if err != nil {
		return err
	}

	*count, err = data.Context.Videos.Find(criteria).Count()
	if err != nil {
		return err
	}

	return nil
}

func (s *VideoController) Search(req *api.Request, list *api.VideoList) error {
	defer service.DontPanic()
	query, page, lang := req.Query, req.Page, req.Lang
	if query == nil {
		return errors.New("Query can't be null")
	}
	text, ok := (*query)["$text"].(string)
	if !ok {
		return errors.New("Query is not valid")
	}
	if len(text) < 3 {
		return errors.New("Query must be 3 symbols or more")
	}
	criteria, err := s.convertQueryToCriteria(query, lang)
	if err != nil {
		return err
	}

	var dbList []data.Video
	q := data.Context.Videos.
		Find(criteria).
		Select(bson.M{"score": bson.M{"$meta": "textScore"}}).
		Sort("$textScore:score").
		Skip(page.Skip).
		Limit(page.Limit)
	err = q.All(&dbList)
	if err != nil {
		return err
	}

	data.InsertSearchQuery(text, len(dbList))

	*list = make(api.VideoList, len(dbList))
	for i, v := range dbList {
		(*list)[i] = s.convertVideo(&v, lang, false, req.Ip)
	}

	return nil
}

func (s *VideoController) One(req *api.Request, one *api.Video) error {
	defer service.DontPanic()
	query, lang := req.Query, req.Lang
	if query == nil {
		return errors.New("Query is empty")
	}

	id, ok := (*query)["Id"].(float64)
	if !ok {
		return errors.New("Id not found in query or wrong value")
	}

	videoId := int(id)
	v := &data.Video{}
	err := v.LoadById(videoId)
	if err != nil {
		return err
	}
	*one = *s.convertVideo(v, lang, true, req.Ip)
	return nil
}

func (s *VideoController) WriteComment(req *api.Request, comm *api.VideoComment) error {
	defer service.DontPanic()
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

	device := &data.Device{}
	if err := device.FindId(token); err != nil {
		return err
	}

	if err := comm.Validate(); err != nil {
		return err
	}

	vc, err := data.NewVideoComment(comm.VideoId)
	if err != nil {
		return err
	}

	exists, err := data.VideoExists(comm.VideoId)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("Video doesn't exists")
	}

	u := &data.User{}
	err = u.GetOrCreate(token)
	if err != nil {
		return err
	}

	vc.PostDate = time.Now().UTC()
	vc.Language = req.Lang
	vc.Body = comm.Body
	//todo check spam
	vc.Status = data.CommentStatusApproved
	vc.UserId = u.Id

	if err := vc.Insert(&device.Source.DeviceSource); err != nil {
		return err
	}

	up := bson.M{"$push": bson.M{"Comments": vc}, "$inc": bson.M{"CommentCount": 1}}
	err = data.Context.Videos.UpdateId(vc.VideoId, up)
	if err != nil {
		return err
	}

	comm.Id = vc.Id
	comm.Author = u.Author()
	comm.Avatar = string(u.Avatar.Data)
	comm.PostDate = vc.PostDate.Unix()
	return nil
}

func (s *VideoController) ToggleLike(req *api.Request, res *api.VideoActionResponse) error {
	defer service.DontPanic()
	return s.incAction(data.VideoHistoryTypeLike, req, res)
}

func (s *VideoController) IncView(req *api.Request, res *api.VideoActionResponse) error {
	defer service.DontPanic()
	return s.incAction(data.VideoHistoryTypeView, req, res)
}

//deprecated
func (s *VideoController) IncDownload(req *api.Request, res *api.VideoActionResponse) error {
	defer service.DontPanic()
	return s.incAction(data.VideoHistoryTypeDownload, req, res)
}

func (s *VideoController) ToggleDownload(req *api.Request, res *api.VideoActionResponse) error {
	defer service.DontPanic()
	return s.incAction(data.VideoHistoryTypeDownload, req, res)
}

func (s *VideoController) incAction(t data.VideoHistoryType, req *api.Request, res *api.VideoActionResponse) (err error) {
	var query api.Id
	if err := req.UnmarshalQuery(&query); err != nil {
		return err
	}
	device := &data.Device{}
	if err := device.FindId(req.Token); err != nil {
		return err
	}

	videoId := query.Id
	if videoId <= 0 {
		return errors.New("Video doesn't exists")
	}

	u := &data.User{}
	if err := u.GetOrCreate(req.Token); err != nil {
		return err
	}

	var counters data.VideoCounters
	err = data.Context.Videos.FindId(videoId).One(&counters)
	if err != nil {
		return err
	}

	h := data.NewVideoHistory(u.Id, videoId, t)
	var exists bool
	exists, err = h.Exists()
	if err != nil {
		return err
	}

	var incField string
	switch t {
	case data.VideoHistoryTypeView:
		incField = "ViewCount"
	case data.VideoHistoryTypeDownload:
		incField = "DownloadCount"
	case data.VideoHistoryTypeLike:
		incField = "LikeCount"
	default:
		return errors.New("Action not found")
	}

	var inc int
	if exists {
		switch t {
		case data.VideoHistoryTypeLike, data.VideoHistoryTypeDownload:
			{
				res.Action = api.VideoActionRemoved
				if t == data.VideoHistoryTypeDownload {
					inc = -1
				}
				if err := data.Context.VideoHistory.RemoveId(&h.Id); err != nil {
					return err
				}
			}
		}
	} else {
		res.Action = api.VideoActionAdded
		inc = 1
		if err := h.Insert(device); err != nil {
			return err
		}
	}

	//update rank
	if inc != 0 {
		change := mgo.Change{
			Update:    bson.M{"$inc": bson.M{incField: inc}},
			ReturnNew: true,
			Upsert:    true,
		}
		_, err := data.Context.Videos.FindId(videoId).Apply(change, &counters)
		if err != nil {
			return err
		}
		rank := counters.Rank()
		err = data.Context.Videos.UpdateId(videoId, bson.M{"$set": bson.M{"Rank": rank}})
		if err != nil {
			return err
		}
	}
	//
	switch t {
	case data.VideoHistoryTypeView:
		res.TotalCount = counters.ViewCount
	case data.VideoHistoryTypeDownload:
		res.TotalCount = counters.DownloadCount
	case data.VideoHistoryTypeLike:
		res.TotalCount = counters.LikeCount
	}

	return nil
}

func (s *VideoController) convertQueryToCriteria(query *api.Query, lang data.Language) (where bson.M, err error) {
	haveFeatured := true
	filters := []string{"published"}
	where = bson.M{}

	if query != nil {
		for k, v := range *query {
			switch k {
			case "IsPremium":
				{
					filters = append(filters, s.not(v.(bool), "premium"))
					haveFeatured = false
				}
			case "CategoryId":
				{
					filters = append(filters, s.cat(int(v.(float64))))
					haveFeatured = false
				}
			case "$text":
				{
					where["$text"] = bson.M{
						"$search": v.(string),
						//todo not work with ru
						//"$language": lang,
					}
					haveFeatured = false
				}
			case "Id":
				{
					vmap, ok := v.(map[string]interface{})
					if !ok {
						err = errors.New("$in value is not a map")
						return
					}
					for mk, mv := range vmap {
						switch mk {
						case "$in":
							{
								haveFeatured = false
								iids, ok := mv.([]interface{})
								if !ok {
									err = errors.New("Id.$in value is not a slice")
									return
								}
								ids := make([]int, len(iids))
								for i, id := range iids {
									vid, ok := id.(float64)
									if !ok {
										err = errors.New("Id.$in value is not a int[] slice - " + fmt.Sprintf("%#v", id))
										return
									}
									ids[i] = int(vid)
								}

								where["_id"] = bson.M{"$in": ids}
							}
						default:
							{
								err = errors.New("Field '$in." + mk + "' not accept for the query")
								return
							}
						}
					}
				}
			default:
				{
					err = errors.New("Field '" + k + "' not accept for the query")
					return
				}
			}

		}
	}

	if haveFeatured {
		filters = append(filters, "featured")
	}
	where["Filters"] = bson.M{"$all": filters}
	return
}

func (s *VideoController) cat(catId int) string {
	return "c" + strconv.Itoa(catId)
}
func (s *VideoController) not(v bool, w string) string {
	if v {
		return w
	} else {
		return "!" + w
	}
}

func (s *VideoController) convertVideo(v *data.Video, lang data.Language, isFull bool, ip string) *api.Video {
	result := &api.Video{
		Id:            v.Id,
		Title:         v.Title.Get(lang, false),
		Slug:          v.Slug.Get(lang, true),
		Desc:          v.Desc.Get(lang, false),
		Duration:      v.Duration,
		CategoryId:    v.CategoryId,
		PublishedDate: v.PublishedDate.Unix(),
		UpdateDate:    v.UpdateDate.Unix(),
		IsPremium:     v.IsPremium(),
		ViewCount:     v.ViewCount,
		LikeCount:     v.LikeCount,
		CommentCount:  v.CommentCount,
		DownloadCount: v.DownloadCount,
		Featured:      v.IsFeatured(),
		Images: api.VideoMediaList{
			&api.VideoMedia{
				Format: "origin",
				Url:    v.MainScreenshotUrl(),
			},
		},
	}
	if isFull {
		result.AddedDate = v.AddedDate.Unix()
		result.Tags = v.Tags.Get(lang)
		result.Actors = api.ActorList{}
		for _, actor := range v.Actors {
			result.Actors = append(result.Actors, &api.Actor{
				Name: actor,
			})
		}
		if v.ChannelId > 0 {
			ch, err := Mem.VideoChannel.FindId(v.ChannelId)
			if err == nil {
				result.Channel = api.Channel{
					Id:    ch.Id,
					Title: ch.Title,
				}
			}
		}
		c, err := Mem.VideoCategory.FindId(v.CategoryId)
		if err == nil {
			result.Category = api.Category{
				Id:    c.Id,
				Title: c.Title.Get(lang, false),
			}
		}

		vFiles := make(api.VideoMediaList, len(v.Files))
		for i, f := range v.Files {
			vFiles[i] = &api.VideoMedia{
				Format: fmt.Sprintf("%dp", f.H),
				Url:    f.GetViewUrl(ip),
				AUrl:   f.GetDownloadUrl(ip),
				Size:   f.Size,
			}
		}
		result.Files = vFiles.Sort()

		comments := make(api.VideoCommentList, len(v.Comments))
		for i, comm := range v.Comments {
			comments[i] = &api.VideoComment{
				Id:       comm.Id,
				VideoId:  v.Id,
				Author:   comm.User.Author(),
				Avatar:   string(comm.User.Avatar.Data),
				Body:     comm.Body,
				PostDate: comm.PostDate.Unix(),
			}
		}
		result.Comments = comments.Sort()

		rels := v.Related
		if len(rels) == 0 {
			var list []data.Video
			err := data.Context.Videos.
				Find(bson.M{"Filters": "published"}).
				Select(bson.M{"_id": 1}).
				Limit(31).
				Sort("-PublishedDate").
				All(&list)
			rels = make([]int, 0)
			if err == nil {
				for _, vid := range list {
					if vid.Id == v.Id {
						continue
					}
					rels = append(rels, vid.Id)
				}
			}
		}
		if len(rels) > 30 {
			rels = rels[:30]
		}
		result.Related = rels
	}

	return result
}
