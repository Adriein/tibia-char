package main

import (
	"log"

	"github.com/adriein/tibia-char/internal"
	"github.com/adriein/tibia-char/internal/scrap"
)

func main() {
	internal.NewApp()

	cron := scrap.NewService()

	err := cron.ScrapBazaar()

	if err != nil {
		log.Fatal(err.Error())
	}
}
