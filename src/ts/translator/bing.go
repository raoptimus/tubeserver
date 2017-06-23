package main

import (
	"flag"
	btr "github.com/theplant/bingtranslator/translator"
)

type bingTranslator struct {
	clientId     string
	clientSecret string
}

func (t *bingTranslator) InitFlags() {
	// verybusymanager@yandex.ru
	// Microsoft-shit
	flag.StringVar(
		&t.clientId,
		"bing-client-id",
		"microsoft_is_fucking_awful",
		"bing client it",
	)
	flag.StringVar(
		&t.clientSecret,
		"bing-client-secret",
		"8yAWNbMh0HJ+fToKwCR/7xDvKswiycELwEoIohk5O7I=",
		"bing client secret",
	)
}

func (t *bingTranslator) Init() {
	btr.SetCredentials(t.clientId, t.clientSecret)
}

func (t *bingTranslator) Name() string {
	return "bing"
}

func (t *bingTranslator) Translate(text, from, to string) string {
	translations, err := btr.Translate(from, to, text, btr.INPUT_TEXT)
	if err != nil {
		panic("Bing translate error: " + err.Error())
	}
	return translations[0].Text
}
