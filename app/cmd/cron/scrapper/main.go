package main

import (
	"context"
	"log"
	"os"

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

	auctionRepository := auction.NewPgAuctionRepository(app.Databse)
	worldRepository := auction.NewPgWorldRepository(app.Databse)
	currencyRepository := currency.NewPgCurrencyRepository(app.Databse)

	scrapperFactory := &auction.CollyFactory{}
	parserFactory := &auction.HtmlParserFactory{}

	tibiaAPI := vendor.NewTibiaApi()

	mapper := auction.NewMapper(worldRepository)

	cron := auction.NewService(tibiaAPI, auctionRepository, worldRepository, currencyRepository, mapper, parserFactory, scrapperFactory, logger)

	ctx := context.WithValue(context.Background(), middleware.TraceIDKey, helper.TraceID())

	err := cron.ScrapBazaar(ctx)

	if err != nil {
		log.Fatal(eris.ToString(err, true))
	}
}
