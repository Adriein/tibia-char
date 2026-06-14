package cron

import (
	"context"
	"log"

	"github.com/adriein/tibia-char/internal"
	"github.com/adriein/tibia-char/pkg/constants"
	"github.com/adriein/tibia-char/pkg/helper"
	"github.com/adriein/tibia-char/pkg/middleware"
	"github.com/rotisserie/eris"
)

func Stats(app *internal.App) {
	service := app.Modules.Auction

	ctx := context.WithValue(context.Background(), middleware.TraceIDKey, helper.TraceID())
	ctx = context.WithValue(ctx, constants.SourceKey, "STATS_CRON")

	err := service.AggregateAuctionStatsPrecompute(ctx)

	if err != nil {
		log.Println(eris.ToString(err, true))
	}
}
