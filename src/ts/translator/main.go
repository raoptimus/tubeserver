package main

import (
	"flag"
	"fmt"
	"github.com/fiam/gounidecode/unidecode"
	"github.com/raoptimus/gserv/config"
	"github.com/raoptimus/gserv/service"
	"github.com/raoptimus/rlog"
	"gopkg.in/mgo.v2/bson"
	"os"
	"regexp"
	"strings"
	"time"
	"ts/data"
	"ts/translator/blacklist"
)

var langs = []data.Language{
	"en", // английский
	// now we want only english
	// "de", // немецкий
	// "es", // испанский
	// "it", // итальянский
	// "fr", // французский
}

var baseLang = "ru"

type Translator interface {
	Translate(text, from, to string) string
	InitFlags()
	Init() // we must init translators AFTER all InitFlags called
	Name() string
}

// used in that order
var translators = []Translator{
	&yandexTranslator{},
	&bingTranslator{},
	// &googleTranslator{}, // not implemented: no free version available
}

var log *rlog.Logger
var ErrNoBaseText = fmt.Errorf("Base(%s) text not found", baseLang)
var traceEnabled = true
var videosLimit = 0

func main() {
	if service.Exists() {
		os.Exit(0)
	}
	initFlags()
	initLogger()
	data.Init(false)
	blacklist.Init()
	initTranslators()
	trace("Base lang is %s", baseLang)

	service.Init(&service.BaseService{
		Start:  start,
		Logger: log,
	})
	service.Start(true)
}

func start() {
	for range time.Tick(5 * time.Minute) {
		translator()
	}
}

func initFlags() {
	flag.IntVar(&videosLimit, "limit", videosLimit, "limit query size; 0 means no imit")
	flag.BoolVar(&traceEnabled, "v", traceEnabled, "Be very verbose (enable tracing)")
	for _, translator := range translators {
		translator.InitFlags()
	}

	// stupid bootstrap/config calls flag.Parse so our flag.* calls are useless
	// flag.Parse()
}

func initLogger() {
	var err error
	if traceEnabled {
		log, err = rlog.NewLogger(rlog.LoggerTypeStd, "")
	} else {
		cs := config.String("MongoLogServer", config.String("MongoAllServer", "localhost/Logs"))
		log, err = rlog.NewLoggerDial(rlog.LoggerTypeMongoDb, "", cs, "")
	}
	if err != nil {
		log.Fatalln(err)
	}
}

func initTranslators() {
	for _, translator := range translators {
		translator.Init()
	}
}

func translator() {
	// find videos having any translation missing
	// db.Video.findOne({Title: {$not: {$elemMatch: {Language: {$in: ["en", "de", "es", "it", "fr"]}}}}}, {Title: 1})
	q := bson.M{"Title": bson.M{"$not": bson.M{"$elemMatch": bson.M{"Language": bson.M{"$in": langs}}}}}
	selector := bson.M{"Title": true}
	query := data.Context.Videos.Find(q).Select(selector).Sort("-PublishedDate")
	if videosLimit > 0 {
		query = query.Limit(videosLimit)
	}
	stats := struct {
		done, found, errors int
	}{}
	stats.found, _ = query.Count()
	trace("Found %d videos", stats.found)

	iter := query.Iter()
	var video *data.Video
	for iter.Next(&video) {
		if err := handle(video); err != nil {
			log.Err(err.Error())
			stats.errors++
		}
		stats.done++
		trace("Done [%d/%d] | Errors [%d]", stats.done, stats.found, stats.errors)
		video = nil //clear object
	}
	if err := iter.Close(); err != nil {
		log.Err("Translator: iteration error: " + err.Error())
	}

}

func handle(video *data.Video) error {
	title := video.Title
	// translate title to missing languages
	trace("Translating video id(%d); title = %s", video.Id, title)
	title, err := translate(title)
	if err != nil {
		return fmt.Errorf("Translator: translation error: %s", err.Error())
	}
	video.Title = title
	blacklist.CensorTextList(title)
	// update db
	if err := save(video); err != nil {
		return fmt.Errorf("Translator: saving error: %s", err.Error())
	}
	return nil
}

func translate(t data.TextList) (data.TextList, error) {
	t, en, err := translateToEn(t)
	if err != nil {
		return nil, err
	}

	// find missing translations to title
	missing := getMissingLangs(t)
	trace("\tMissing langs: %+v", missing)

	// translate via remote server
	for _, lang := range missing {
		result := smartTranslate(en, "en", string(lang))
		trace("\t[%s] => %q", lang, result)
		t = append(t, &data.Text{
			Quote:    result,
			Language: lang,
		})
	}
	return t, nil
}

func translateToEn(t data.TextList) (data.TextList, string, error) {
	// get text to translate from
	base := baseText(t)
	if base == nil {
		return nil, "", ErrNoBaseText
	}

	trace("\tTranslating %q", base.Quote)
	result := smartTranslate(base.Quote, baseLang, "en")
	trace("\t[en] => %q", result)

	return append(t, &data.Text{
		Quote:    result,
		Language: "en",
	}), result, nil
}

func baseText(t data.TextList) *data.Text {
	for _, text := range t {
		if string(text.Language) == baseLang {
			return text
		}
	}
	return nil
}

func getMissingLangs(t data.TextList) []data.Language {
	var missing []data.Language
outer:
	for _, lang := range langs {
		for _, text := range t {
			if text.Language == lang {
				continue outer
			}
		}
		missing = append(missing, lang)
	}
	return missing
}

var cyrillicRegExp = regexp.MustCompile("\\p{Cyrillic}")

func smartTranslate(text, from, to string) string {
	for _, translator := range translators {
		result := translator.Translate(text, from, to)
		trace("\t\tTrying %s: %s", translator.Name(), result)
		if goodEnough(result) {
			trace("\t\tAccepted")
			return result
		}
		if from == baseLang {
			saveUntranslatedWords(result)
		}
	}
	trace("\t\tfallback to transliteration")
	return transliterate(text)
}

func goodEnough(translation string) bool {
	// for now we accept translation if it has no cyrillic symbols
	return len(translation) > 0 && !cyrillicRegExp.MatchString(translation)
}

func transliterate(text string) string {
	return unidecode.Unidecode(text)
}

func saveUntranslatedWords(text string) {
	for _, word := range strings.Fields(text) {
		if cyrillicRegExp.MatchString(word) {
			saveUntranslatedWord(word)
		}
	}
}

var savedWords = make(map[string]struct{})

func saveUntranslatedWord(word string) {
	word = strings.ToLower(word)
	if _, ok := savedWords[word]; ok {
		return
	}
	trace("\t\tSaving untranslated word %q", word)
	savedWords[word] = struct{}{}
	dict := data.Dictionary{
		Id: word,
	}
	data.Context.Dictionary.Insert(dict)
}

func save(v *data.Video) error {
	update := bson.M{"$set": bson.M{"Title": v.Title}}
	return data.Context.Videos.UpdateId(v.Id, update)
}

func trace(format string, args ...interface{}) {
	if traceEnabled {
		log.Printf(format, args...)
	}
}
