package main

import (
	"flag"
	"fmt"
	"github.com/icrowley/go-yandex-translate"
)

type yandexTranslator struct {
	tr     *yandex_translate.Translator
	apiKey string
}

func (t *yandexTranslator) InitFlags() {
	// verybusymanager@yandex.ru
	// omgplzhalp
	flag.StringVar(
		&t.apiKey,
		"yandex-api-key",
		"trnsl.1.1.20150729T162322Z.1c8b29c9d608c600.3e2e74b50dd9b0bcafff255434eab9c423b67b28",
		"yandex api key",
	)
}

func (t *yandexTranslator) Init() {
	t.tr = yandex_translate.New(t.apiKey)
}

func (t *yandexTranslator) Translate(text, from, to string) string {
	lang := fmt.Sprintf("%s-%s", from, to)
	translation, err := t.tr.Translate(lang, text)
	if err != nil {
		panic("Yandex translate error: " + err.Error())
	}
	return translation.Result()
}

func (t *yandexTranslator) Name() string {
	return "yandex"
}
