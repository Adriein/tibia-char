package main

import (
	"context"
	"log"

	"github.com/adriein/tibia-char/internal"
	"github.com/adriein/tibia-char/pkg/constants"
	"github.com/adriein/tibia-char/pkg/helper"
	"github.com/adriein/tibia-char/pkg/middleware"
	_ "github.com/lib/pq"
	"github.com/rotisserie/eris"
)

func main() {
	app := internal.NewApp()

	traceID := helper.TraceID()

	ctx := context.WithValue(context.Background(), middleware.TraceIDKey, traceID)
	ctx = context.WithValue(ctx, constants.SourceKey, "BACKFILL_CRON")

	service := app.Modules.Auction

	if err := service.BackfillAuctionFlags(ctx); err != nil {
		log.Println(eris.ToString(err, true))
	}
}
