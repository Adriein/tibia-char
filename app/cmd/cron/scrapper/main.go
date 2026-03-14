package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/adriein/tibia-char/internal"
	"github.com/adriein/tibia-char/internal/auction"
	"github.com/adriein/tibia-char/internal/currency"
	"github.com/adriein/tibia-char/pkg/helper"
	"github.com/adriein/tibia-char/pkg/middleware"
	"github.com/adriein/tibia-char/pkg/vendor"
	_ "github.com/lib/pq"
	"github.com/rotisserie/eris"
)

func main() {
	app := internal.NewApp()

	logger := log.New(os.Stderr, "[Scrapper Cron] ", log.LstdFlags|log.LUTC)

	auctionRepository := auction.NewPgAuctionRepository(app.Database)
	worldRepository := auction.NewPgWorldRepository(app.Database)
	currencyRepository := currency.NewPgCurrencyRepository(app.Database)
	aggAuctionRepository := auction.NewPgAggAuctionRepository(app.Database)

	scrapperFactory := &auction.CollyFactory{}
	parserFactory := &auction.HtmlParserFactory{}

	tibiaAPI := vendor.NewTibiaApi()

	mapper := auction.NewMapper(worldRepository)

	service := auction.NewService(tibiaAPI, auctionRepository, worldRepository, currencyRepository, aggAuctionRepository, mapper, parserFactory, scrapperFactory, logger)

	ctx := context.WithValue(context.Background(), middleware.TraceIDKey, helper.TraceID())

	for true {
		err := service.ScrapperOrchestrator(ctx)

		if err != nil {
			log.Println(eris.ToString(err, true))
		}

		time.Sleep(5 * time.Minute)
	}
}
