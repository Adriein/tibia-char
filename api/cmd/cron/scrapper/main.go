package main

import (
	"log"
	"os"

	"github.com/adriein/tibia-char/internal"
	"github.com/adriein/tibia-char/internal/scrap"
)

func main() {
	internal.NewApp()

	logger := log.New(os.Stderr, "[Scrapper Cron] ", log.LstdFlags|log.LUTC)

	cron := scrap.NewService(logger)

	err := cron.ScrapBazaar()

	if err != nil {
		log.Fatal(err.Error())
	}
}
