package cron

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/adriein/tibia-char/internal"
	"github.com/adriein/tibia-char/pkg/constants"
	"github.com/adriein/tibia-char/pkg/helper"
	"github.com/adriein/tibia-char/pkg/middleware"
	_ "github.com/lib/pq"
	"github.com/rotisserie/eris"
)

func Scrapper(app *internal.App, ingestion bool) {
	logger := app.Logger

	if ingestion {
		traceID := helper.TraceID()

		ctx := context.WithValue(context.Background(), middleware.TraceIDKey, traceID)
		ctx = context.WithValue(ctx, constants.SourceKey, "SCRAP_CRON")

		service := app.Modules.Auction

		if err := service.ScrapNewAuctions(ctx); err != nil {
			logger.Error("Error scrapping new auctions", "error", eris.ToJSON(err, true))
		}

		return
	}

	if os.Getenv(constants.Env) != constants.Prod {
		for true {
			traceID := helper.TraceID()

			ctx := context.WithValue(context.Background(), middleware.TraceIDKey, traceID)
			ctx = context.WithValue(ctx, constants.SourceKey, "SCRAP_CRON")

			service := app.Modules.Auction

			if err := service.ScrapperOrchestrator(ctx); err != nil {
				log.Println(eris.ToString(err, true))
			}

			time.Sleep(5 * time.Minute)
		}

		return
	}

	traceID := helper.TraceID()

	ctx := context.WithValue(context.Background(), middleware.TraceIDKey, traceID)
	ctx = context.WithValue(ctx, constants.SourceKey, "SCRAP_CRON")

	service := app.Modules.Auction

	if err := service.ScrapperOrchestrator(ctx); err != nil {
		log.Println(eris.ToString(err, true))
	}
}
