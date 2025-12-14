package main

import (
	"log"
	"os"

	"github.com/adriein/tibia-char/internal"
	"github.com/adriein/tibia-char/internal/auction"
	_ "github.com/lib/pq"
)

func main() {
	app := internal.NewApp()

	logger := log.New(os.Stderr, "[Scrapper Cron] ", log.LstdFlags|log.LUTC)

	auctionRepository := auction.NewPgAuctionRepository(app.Databse)
	vocationRepository := auction.NewPgVocationRepository(app.Databse)
	genderRepository := auction.NewPgGenderRepository(app.Databse)
	wolrdRepository := auction.NewPgWorldRepository(app.Databse)

	cron := auction.NewService(logger, auctionRepository, vocationRepository, genderRepository, wolrdRepository)

	err := cron.ScrapBazaar()

	if err != nil {
		log.Fatal(err.Error())
	}
}
