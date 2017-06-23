package main

import (
	"log"
	"net/url"
)

func main() {

	icon, err := url.Parse("/icons/1.jpg?a=1")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%#v", icon)
}
