package main

import (
	"context"
	"log"

	"github.com/adriein/tibia-char/internal"
	"github.com/adriein/tibia-char/pkg/helper"
	"github.com/adriein/tibia-char/pkg/middleware"
	"github.com/rotisserie/eris"
)

func main() {
	app := internal.NewApp()

	service := app.Modules.Auction

	ctx := context.WithValue(context.Background(), middleware.TraceIDKey, helper.TraceID())

	err := service.AggregateAuctionStatsPrecompute(ctx)

	if err != nil {
		log.Println(eris.ToString(err, true))
	}
}
