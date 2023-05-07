package main

import (
	"log"

	"github.com/e-pas/yandex-praktikum-shortener/internal/app"
)

func main() {

	app, err := app.New()
	if err != nil {
		log.Fatal(err)
	}

	err = app.Run()
	if err != nil {
		log.Fatal(err)
	}

}
