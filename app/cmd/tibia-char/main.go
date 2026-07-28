package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/adriein/tibia-char/cmd/cron"
	"github.com/adriein/tibia-char/internal"
	"github.com/adriein/tibia-char/internal/server"
	"github.com/adriein/tibia-char/pkg/constants"
	_ "github.com/lib/pq"
)

func main() {
	app := internal.NewApp()

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := app.Shutdown(ctx); err != nil {
			fmt.Printf("Failed to perform graceful shutdown %v", err)
			os.Exit(1)
		}
	}()

	if len(os.Args) < 2 {
		server.New(os.Getenv(constants.ServerPort), app)

		return
	}

	switch os.Args[1] {
	case constants.CronBackfill:
		cron.Backfill(app)
	case constants.CronCurrency:
		cron.Currency(app)
	case constants.CronScrapper:
		cron.Scrapper(app, false)
	case constants.CronIngestion:
		cron.Scrapper(app, true)
	case constants.CronStats:
		cron.Stats(app)
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
