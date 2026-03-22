package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/adriein/tibia-char/internal"
	"github.com/adriein/tibia-char/internal/auction"
	"github.com/adriein/tibia-char/internal/currency"
	"github.com/adriein/tibia-char/pkg/constants"
	"github.com/adriein/tibia-char/pkg/helper"
	"github.com/adriein/tibia-char/pkg/middleware"
	"github.com/adriein/tibia-char/pkg/vendor"
	_ "github.com/lib/pq"
	"github.com/rotisserie/eris"
)

func main() {
	app := internal.NewApp()

	auctionRepository := auction.NewPgAuctionRepository(app.Database)
	worldRepository := auction.NewPgWorldRepository(app.Database)
	currencyRepository := currency.NewPgCurrencyRepository(app.Database)
	aggAuctionRepository := auction.NewPgAggAuctionRepository(app.Database)

	scrapperFactory := &auction.CollyFactory{}
	parserFactory := &auction.HtmlParserFactory{}

	tibiaAPI := vendor.NewTibiaApi()

	mapper := auction.NewMapper(worldRepository)

	for true {
		traceID := helper.TraceID()

		ctx := context.WithValue(context.Background(), middleware.TraceIDKey, traceID)

		opts := &slog.HandlerOptions{
			ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
				if attr.Key == slog.TimeKey {
					formatted := attr.Value.Time().UTC().Format(time.DateTime)

					return slog.String(slog.TimeKey, formatted)
				}

				return attr
			},
		}

		var logger *slog.Logger

		if os.Getenv(constants.Env) == constants.DEV {
			logger = slog.New(slog.NewTextHandler(os.Stdout, opts)).With("source", "SCRAPPER", "trace_id", traceID)
		} else {
			logger = slog.New(slog.NewJSONHandler(os.Stdout, opts)).With("source", "SCRAPPER", "trace_id", traceID)
		}

		service := auction.NewService(tibiaAPI, auctionRepository, worldRepository, currencyRepository, aggAuctionRepository, mapper, parserFactory, scrapperFactory, logger)

		err := service.ScrapperOrchestrator(ctx)

		if err != nil {
			log.Println(eris.ToString(err, true))
		}

		time.Sleep(5 * time.Minute)
	}
}
