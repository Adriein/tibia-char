package cron

import (
	"context"
	"log"
	"os"

	"github.com/adriein/tibia-char/internal"
	"github.com/adriein/tibia-char/internal/currency"
	"github.com/adriein/tibia-char/pkg/helper"
	"github.com/adriein/tibia-char/pkg/middleware"
	"github.com/adriein/tibia-char/pkg/vendor"
	_ "github.com/lib/pq"
	"github.com/rotisserie/eris"
)

func Currency(app *internal.App) {
	logger := log.New(os.Stderr, "[Currency Cron] ", log.LstdFlags|log.LUTC)

	currencyAPI := vendor.NewOpenCurrencyAPI()

	repository := currency.NewPgCurrencyRepository(app.Database)

	cron := currency.NewService(repository, currencyAPI, logger)

	ctx := context.WithValue(context.Background(), middleware.TraceIDKey, helper.TraceID())

	err := cron.StoreDailyRates(ctx)

	if err != nil {
		log.Fatal(eris.ToString(err, true))
	}
}
