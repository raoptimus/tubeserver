package main

import (
	"flag"
)

var googleApiKey = ""

type googleTranslator struct {
}

func (t *googleTranslator) InitFlags() {
	flag.StringVar(&googleApiKey, "google-api-key", googleApiKey, "google api key")
}

func (t *googleTranslator) Init() {

}

func (t *googleTranslator) Name() string {
	return "google"
}

func (t *googleTranslator) Translate(text, from, to string) string {
	return text
}
