package blacklist

import (
	"encoding/csv"
	"fmt"
	"github.com/raoptimus/gserv/config"
	"gopkg.in/mgo.v2/bson"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"ts/data"
)

type substitution struct {
	regexp  *regexp.Regexp
	replace string
	pattern string
}

var bl []*substitution

func Init() {
	blfn := config.String("TranslatorBlackList", "blacklist.csv")
	file, err := os.Open(blfn)
	if err != nil {
		log.Fatalln("Cannot open blacklist.csv")
	}
	r := csv.NewReader(file)
	r.Comment = '#'
	unique := make(map[string]string)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalln(err)
		}
		if len(record) < 2 {
			continue
		}
		pattern := strings.ToLower(record[0])
		replace := strings.ToLower(record[1])

		if duplicate, ok := unique[pattern]; ok {
			fmt.Printf("Pattern duplicate found: %q, %q", duplicate, pattern)
		}
		unique[pattern] = replace
		bl = append(bl, &substitution{
			regexp:  regexp.MustCompile(`(?i)\b` + pattern + `\b`),
			replace: replace,
			pattern: pattern,
		})
	}
}

func CensorTextList(t data.TextList) bool {
	ii := 0
	for i := 0; i < len(t); i++ {
		if !Censor(&t[i].Quote) {
			continue
		}
		ii++
	}
	return ii > 0
}

func Censor(s *string) bool {
	o := *s
	for _, subst := range bl {
		*s = subst.regexp.ReplaceAllString(*s, subst.replace)
	}
	return o != *s
}

func CensorAllVideos() {
	// ({"Title.Language": "en", Title: {$elemMatch: {Quote: {$in: [/kid/, /child/]}}}}, {Title: 1})
	//	patterns := make([]bson.RegEx, 0, len(bl))
	//	for _, subst := range bl {
	//		regex := bson.RegEx{
	//			Pattern: `\b` + subst.pattern + `\b`,
	//			Options: "i",
	//		}
	//		patterns = append(patterns, regex)
	//	}
	//	q := bson.M{
	//		"Title": bson.M{
	//			"$elemMatch": bson.M{
	//				"Language": data.LanguageEnglish,
	//				"Quote":    bson.M{"$in": patterns},
	//			},
	//		},
	//	}
	iter := data.Context.Videos.Find(nil).Select(bson.M{"Title": 1}).Iter()
	var video data.Video

	for iter.Next(&video) {
		if !CensorTextList(video.Title) {
			continue
		}
		update := bson.M{"$set": bson.M{"Title": video.Title}}
		if err := data.Context.Videos.UpdateId(video.Id, update); err != nil {
			log.Fatalf("Translator: saving error: %v", err)
		}
	}
	if err := iter.Close(); err != nil {
		log.Fatalf("Translator: censorAll iteration error: %v", err)
	}

}
