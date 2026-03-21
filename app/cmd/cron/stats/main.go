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
	"github.com/rotisserie/eris"
)

func main() {
	app := internal.NewApp()

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
		logger = slog.New(slog.NewTextHandler(os.Stdout, opts))
	} else {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, opts))
	}

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

	err := service.AggregateAuctionStatsPrecompute(ctx)

	if err != nil {
		log.Println(eris.ToString(err, true))
	}
}
