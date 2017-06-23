package main

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"os"
	"strconv"
	"time"
	"ts/data"
)

func main() {
	printTask("...")
	if err := cat(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	printTask("...")
	if err := DbTV(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	printTask("sync...")
	if err := synccat(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	printTask("reindex")
	if err := reindex(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("All finish")
}

func printTask(t string) {
	fmt.Println("=============")
	fmt.Println(t)
}

func synccat() error {
	srcCatId := 23
	catId := 13
	DbTV := data.Context.DbTV
	videos := data.Context.Videos
	srcDomain := "..."
	var (
		vid   int
		done  int
		total int
	)
	rows, err := DbTV.Query(`SELECT VideoId FROM tbh_videos WHERE CategoryId = ? ORDER BY VideoId`, srcCatId)
	if err != nil {
		return err
	}

	for rows.Next() {
		err = rows.Scan(&vid)
		if err != nil {
			return err
		}

		src := data.NewVideoSource(vid, srcDomain, 0, 0)
		err = videos.Update(bson.M{"Source._id": src.Id}, bson.M{"$set": bson.M{"CategoryId": catId}})
		if err == nil {
			done++
		} else {
			if err != mgo.ErrNotFound {
				return err
			}
		}
		total++
	}

	fmt.Println("Done", done)
	fmt.Println("NotFound", total-done)
	fmt.Println("Total", total)
	return nil
}

func reindex() error {
	db := data.Context

	it := db.Videos.Find(nil).Sort("_id").Iter()
	var v struct {
		Id         int       `bson:"_id"`
		IsPremium  bool      `bson:"IsPremium"`
		Filters    []string  `bson:"Filters"`
		Featured   bool      `bson:"Featured"`
		CategoryId int       `bson:"CategoryId"`
		ImportDate time.Time `bson:"ImportDate"`
	}
	for it.Next(&v) {
		v.Filters = make([]string, 0)
		v.Filters = append(v.Filters, c(v.IsPremium)+"premium")
		switch v.CategoryId {
		case 10, 13:
			{
				v.Featured = false
			}
		}
		v.Filters = append(v.Filters, c(v.Featured)+"featured")
		v.Filters = append(v.Filters, "c"+strconv.Itoa(v.CategoryId))
		if v.Featured {
			v.Filters = append(v.Filters, "published")
		} else {
			v.Filters = append(v.Filters, "deleted")
		}
		err := db.Videos.UpdateId(v.Id,
			bson.M{"$set": bson.M{
				"Filters":   v.Filters,
				"AddedDate": v.ImportDate,
			}})
		if err != nil {
			it.Close()
			return err
		}
	}
	err := it.Close()
	if err != nil {
		return err
	}

	db.Videos.DropIndex("Featured", "CategoryId", "IsPremium", "-PublishedDate")
	db.Videos.DropIndex("$text:Title.Quote", "$text:Desc.Quote", "$text:Tags")
	db.Videos.DropIndex("-PublishedDate")
	db.Videos.UpdateAll(nil, bson.M{"$unset": bson.M{"Featured": "", "IsPremium": "", "ImportDate": ""}})
	return nil
}

func DbTV() error {
	return data.Context.VideoCategory.UpdateId(4, bson.M{"$pull": bson.M{"SourceId": 23}})
}

func cat() error {
	c, _ := data.Context.VideoCategory.FindId(13).Count()
	if c > 0 {
		return nil
	}
	cat := data.Category{
		Id: 13,
		Title: []*data.Text{
			&data.Text{
				Quote:    "...",
				Language: data.LanguageRussian,
			},
			&data.Text{
				Quote:    "...",
				Language: data.LanguageEnglish,
			}},
		SourceId: []int{23},
	}
	return data.Context.VideoCategory.Insert(cat)
}

func c(v bool) string {
	if v {
		return ""
	} else {
		return "!"
	}
}
