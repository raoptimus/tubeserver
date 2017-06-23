package v1

import (
	"errors"
	"gopkg.in/mgo.v2/bson"
	"math"
	"strconv"
	//	"strings"
	"testing"
	"time"
	"ts/data"
)

//func TestVideoRelates(t *testing.T) {
//	var videos []data.Video
//	q := bson.M{"_id": bson.M{"$in": []int{62204, 62210, 62175, 59292, 57755, 54935}}}
//	q = bson.M{"Filters": "published"}
//	err := data.Context.Videos.
//		Find(q).
//		Limit(4).
//		Sort("-PublishedDate").
//		All(&videos)
//	if err != nil {
//		t.Fatal(err.Error())
//	}
//	for _, v := range videos {
//		text := strings.Join(v.Keywords, " ")
//		var relVideos []data.Video
//		err = data.Context.Videos.
//			Find(bson.M{"$text": bson.M{"$search": text}, "Filters": "published"}).
//			Select(bson.M{"score": bson.M{"$meta": "textScore"}}).
//			Sort("$textScore:score").
//			Limit(11).
//			All(&relVideos)
//		if err != nil {
//			t.Fatal(err.Error())
//		}
//		t.Logf("\n||%v||%v||%v||%v|\n",
//			v.Id,
//			v.Title.Get(data.LanguageRussian, false),
//			v.Keywords,
//			v.Related[:10],
//		)
//		for _, v2 := range relVideos {
//			if v.Id == v2.Id {
//				continue
//			}
//			t.Logf("|%v|%v|%v| |\n",
//				v2.Id,
//				v2.Title.Get(data.LanguageRussian, false),
//				v2.Keywords,
//			)
//		}
//	}
//}

func TestVideoListIn(t *testing.T) {
	ids := []int{21586, 43579, 59724, 34201, 58707, 59389, 25517, 20882, 56300, 56892}

	req := Request{
		Lang:  data.LanguageRussian,
		Token: TEST_TOKEN,
		Query: &Query{
			"Id": Query{
				"$in": ids,
			},
		},
		Sort: &SortInfo{
			Field:  "PublishedDate",
			Direct: SortDirectAsc,
		},
		Page: &Page{
			Skip:  0,
			Limit: 100,
		},
	}
	var list VideoList
	if err := getRpcClient().Call("VideoController.List", &req, &list); err != nil {
		t.Fatal(err.Error())
	}

	switch {
	case len(list) == 0:
		t.Fatal("Return data is empty")
	case len(list) != len(ids):
		t.Fatalf("Got %d videos, want %d", len(list), len(ids))
	}

	valid := func(want int) bool {
		for _, got := range ids {
			if got == want {
				return true
			}
		}
		return false
	}
	for _, v := range list {
		if !valid(v.Id) {
			t.Fatalf("Data not valid, want id in %v got  (%v)", ids, v.Id)
		}
		if v.Title == "" {
			t.Fatal("Data not vaild, Title is empty")
		}
	}
}

func TestVideoList(t *testing.T) {
	req := Request{
		Lang:  data.LanguageRussian,
		Token: TEST_TOKEN,
		Query: &Query{
			"IsPremium": true,
		},
		Sort: &SortInfo{
			Field:  "PublishedDate",
			Direct: SortDirectAsc,
		},
		Page: &Page{
			Skip:  0,
			Limit: 100,
		},
	}
	var list VideoList
	if err := getRpcClient().Call("VideoController.List", &req, &list); err != nil {
		t.Fatal(err.Error())
	}

	if len(list) == 0 {
		t.Fatal("Return data is empty")
	}

	for _, v := range list {
		if v.Id <= 0 {
			t.Fatal("Data not vaild, Id is zero")
		}
		if v.Title == "" {
			t.Fatal("Data not vaild, Title is empty")
		}
	}
}

func TestVideoSearch(t *testing.T) {
	req := Request{
		Lang:  data.LanguageRussian,
		Token: TEST_TOKEN,
		Query: &Query{
			"$text": "юююю",
		},
		Page: &Page{
			Skip:  0,
			Limit: 100,
		},
	}
	var list VideoList
	if err := getRpcClient().Call("VideoController.Search", &req, &list); err != nil {
		t.Fatal(err.Error())
	}

	if len(list) == 0 {
		t.Fatal("Return data is empty")
	}

	for _, v := range list {
		if v.Id <= 0 {
			t.Fatal("Data not vaild, Id is zero")
		}
		if v.Title == "" {
			t.Fatal("Data not vaild, Title is empty")
		}
	}
}

func ExampleVideoOne() {
	dv, err := getModel()
	if err != nil {
		panic(err)
	}
	videoId := strconv.Itoa(dv.Id)
	runExample(`{"method": "VideoController.One", "params": [{"Token": "` + TEST_TOKEN + `", "Query": {"Id": ` + videoId + `}}], "id": "1"}`)
	// Output:
}

func TestVideoOne(t *testing.T) {
	dv, err := getModel()
	if err != nil {
		t.Fatal(err.Error())
	}
	videoId := dv.Id

	req := Request{
		Ip:    "206.54.164.72",
		Lang:  data.LanguageRussian,
		Token: TEST_TOKEN,
		Query: &Query{
			"Id": videoId,
		},
	}
	var v Video
	if err := getRpcClient().Call("VideoController.One", &req, &v); err != nil {
		t.Fatalf("Call VideoController.One (Query: %v, Token: %v) return error: %v", *req.Query, req.Token, err)
	}

	if v.Id != videoId {
		t.Fatal("Data is not valid; id not equals")
	}

	if v.Title == "" {
		t.Fatal("Data is not valid; Title is empty")
	}
	if v.Id <= 0 {
		t.Fatal("Data is not valid")
	}

	// we cannot get image size so test only video files
	for _, file := range v.Files {
		if file.Size == 0 {
			t.Fatalf("video id = %v file has empty size: %+v", v.Id, file)
		}
		if err := checkVideoFileUrl(file.Url); err != nil {
			t.Fatalf("Do not getting the file %v: %v", file.Url, err)
		}
		if err := checkVideoFileUrl(file.AUrl); err != nil {
			t.Fatalf("Do not getting the file %v: %v", file.AUrl, err)
		}
	}

	for _, c := range v.Comments {
		if c.VideoId != videoId {
			t.Fatal("Data is not valid. Comment.VideoId != Request.Query.Id")
		}
	}
}

func TestVideoToggleLike(t *testing.T) {
	dv, err := getModel()
	if err != nil {
		t.Fatal(err.Error())
	}

	req := Request{
		Token: TEST_TOKEN,
		Query: &Query{
			"Id": dv.Id,
		},
	}
	var res VideoActionResponse
	if err := getRpcClient().Call("VideoController.ToggleLike", &req, &res); err != nil {
		t.Fatal(err.Error())
	}

	// fmt.Printf("%#v", res)
}

func TestVideoIncView(t *testing.T) {
	dv, err := getModel()
	if err != nil {
		t.Fatal(err.Error())
	}

	req := Request{
		Token: TEST_TOKEN,
		Query: &Query{
			"Id": dv.Id,
		},
	}
	var res VideoActionResponse
	if err := getRpcClient().Call("VideoController.IncView", &req, &res); err != nil {
		t.Fatal(err.Error())
	}
}

func TestWriteVideoComment(t *testing.T) {
	req := Request{
		Token: TEST_TOKEN,
		Object: VideoComment{
			VideoId: 2,
			Body:    "мама мыла раму",
		},
	}

	var comm VideoComment
	if err := getRpcClient().Call("VideoController.WriteComment", &req, &comm); err != nil {
		t.Fatal(err.Error())
	}

	if comm.Id <= 0 {
		t.Fatal("Comment id is empty")
	}

	if comm.VideoId <= 0 {
		t.Fatal("Comment videoId is empty")
	}

	if comm.Author == "" {
		t.Fatal("Comment author is empty")
	}
}

func TestVideoIncDownload(t *testing.T) {
	videoId := 2
	req := Request{
		Token: TEST_TOKEN,
		Query: &Query{
			"Id": videoId,
		},
	}
	var res VideoActionResponse
	if err := getRpcClient().Call("VideoController.IncDownload", &req, &res); err != nil {
		t.Fatal(err.Error())
	}
}

func TestVideoHistoryList(t *testing.T) {
	req := Request{
		Lang:  data.LanguageRussian,
		Token: TEST_TOKEN,
		Query: &Query{
			"Type": 0,
		},
		Page: &Page{
			Skip:  0,
			Limit: 100,
		},
	}
	var list VideoList
	if err := getRpcClient().Call("VideoController.HistoryList", &req, &list); err != nil {
		t.Fatal(err.Error())
	}
	//	if len(list) == 0 {
	//		fmt.Println("--- TestVideoHistoryList - Return data is empty")
	//	}
}

func TestCategoryTop(t *testing.T) {
	req := Request{
		Lang:  data.LanguageRussian,
		Token: TEST_TOKEN,
		Query: &Query{
			"IsPremium":  true,
			"CategoryId": -1, //top
		},
		Sort: &SortInfo{
			Field:  "PublishedDate",
			Direct: SortDirectAsc,
		},
		Page: &Page{
			Skip:  0,
			Limit: 100,
		},
	}
	var list VideoList

	if err := getRpcClient().Call("VideoController.List", &req, &list); err != nil {
		t.Fatal(err.Error())
	}

	if len(list) == 0 {
		t.Fatal("Return data is empty")
	}

	prevRank := math.Inf(+1)
	for _, v := range list {
		if v.Id <= 0 {
			t.Fatal("Data not vaild, Id is zero")
		}
		if v.Title == "" {
			t.Fatal("Data not vaild, Title is empty")
		}

		rank := float64(v.LikeCount+v.DownloadCount) / float64(v.ViewCount)
		t.Logf("id: %v, rank: %v = (%v + %v) / %v", v.Id, rank,
			v.LikeCount, v.DownloadCount, v.ViewCount)
		if rank > prevRank {
			t.Fatal("Category top: wrong sort")
		}
		prevRank = rank
	}
}

func TestVideoSorting(t *testing.T) {
	req := Request{
		Lang:  data.LanguageRussian,
		Token: TEST_TOKEN,
		Sort: &SortInfo{
			Field:  "PublishedDate",
			Direct: SortDirectDesc,
		},
		Page: &Page{
			Skip:  0,
			Limit: 100,
		},
	}
	var list1 VideoList
	if err := getRpcClient().Call("VideoController.List", &req, &list1); err != nil {
		t.Fatal(err.Error())
	}

	var list2 []data.Video
	q := bson.M{"Filters": bson.M{"$all": []string{"published", "featured"}}}
	data.Context.Videos.Find(q).Sort("-PublishedDate").Limit(100).All(&list2)
	if len(list1) != len(list2) {
		t.Fatalf("len1(%v) != len2(%v", len(list1), len(list2))
	}
	for i := 0; i < len(list1); i++ {
		v1 := list1[i]
		v2 := list2[i]
		t.Log(v1.Id, time.Unix(v1.PublishedDate, 0).Format(time.ANSIC))
		if v1.Id != v2.Id {
			t.Fatalf("%v != %v", v1.Id, v2.Id)
		}
	}
}

func getModel() (*data.Video, error) {
	var dv data.Video
	err := data.Context.Videos.Find(bson.M{"Filters": "published"}).One(&dv)
	if err != nil {
		err = errors.New("Do not find one video: " + err.Error())
	}
	return &dv, err
}
