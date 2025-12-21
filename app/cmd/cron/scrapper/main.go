package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/adriein/tibia-char/internal"
	"github.com/adriein/tibia-char/internal/auction"
	"github.com/adriein/tibia-char/pkg/constants"
	"github.com/adriein/tibia-char/pkg/helper"
	"github.com/adriein/tibia-char/pkg/vendor"
	"github.com/gocolly/colly/v2"
	_ "github.com/lib/pq"
)

func main() {
	app := internal.NewApp()

	logger := log.New(os.Stderr, "[Scrapper Cron] ", log.LstdFlags|log.LUTC)

	auctionRepository := auction.NewPgAuctionRepository(app.Databse)
	worldRepository := auction.NewPgWorldRepository(app.Databse)

	linkScrapper := auction.NewScrapper("CollectAuctionLinks")

	linkScrapper.Collector.Limit(&colly.LimitRule{
		DomainGlob:  constants.TibiaOfficialWebsite,
		RandomDelay: 5 * time.Second,
	})

	detailScrapper := auction.NewScrapper("CollectAuctionDetails")

	detailScrapper.Collector.Limit(&colly.LimitRule{
		DomainGlob: constants.TibiaOfficialWebsite,
	})

	tibiaAPI := vendor.NewTibiaApi()

	auctionListHtmlParser := auction.NewAuctionListHtmlParser(tibiaAPI, worldRepository, linkScrapper.Collector)
	auctionHtmlParser := auction.NewAuctionHtmlParser(detailScrapper.Collector)

	mapper := auction.NewMapper(worldRepository)

	cron := auction.NewService(auctionListHtmlParser, auctionHtmlParser, auctionRepository, worldRepository, mapper, logger)

	ctx := context.WithValue(context.Background(), "traceID", helper.TraceID())

	err := cron.ScrapBazaar(ctx)

	if err != nil {
		log.Fatal(err.Error())
	}
}
