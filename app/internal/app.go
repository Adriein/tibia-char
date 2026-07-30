package internal

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/adriein/tibia-char/database"
	"github.com/adriein/tibia-char/internal/auction"
	"github.com/adriein/tibia-char/internal/currency"
	"github.com/adriein/tibia-char/pkg/constants"
	"github.com/adriein/tibia-char/pkg/helper"
	"github.com/adriein/tibia-char/pkg/logger"
	"github.com/adriein/tibia-char/pkg/vendor"
	"github.com/joho/godotenv"
)

type ShutdownFunc func(context.Context) error
type Modules struct {
	Auction auction.AuctionService
}

type App struct {
	Database *sql.DB
	Modules  *Modules
	Logger   *slog.Logger
	Shutdown ShutdownFunc
}

func NewApp() *App {
	if os.Getenv(constants.Env) != constants.Prod {
		dotenvErr := godotenv.Load()

		if dotenvErr != nil {
			log.Fatal("Error loading .env file")
		}
	}

	checker := helper.NewEnvVarChecker(
		constants.DatabaseUrl,
		constants.ServerPort,
		constants.PosthogSdkApiKey,
		constants.Env,
		constants.ImgVersion,
	)

	if envCheckerErr := checker.Check(); envCheckerErr != nil {
		log.Fatal(envCheckerErr.Error())
	}

	logger, loggerShutdown := logger.Create()

	db := database.New(logger)
	modules := initModules(db, logger)

	shudownFn := gracefulShutdown(loggerShutdown)

	return &App{
		Database: db,
		Modules:  modules,
		Logger:   logger,
		Shutdown: shudownFn,
	}
}

func initModules(db *sql.DB, logger *slog.Logger) *Modules {
	auctionRepository := auction.NewPgAuctionRepository(db)
	worldRepository := auction.NewPgWorldRepository(db)
	currencyRepository := currency.NewPgCurrencyRepository(db)
	aggAuctionRepository := auction.NewPgAggAuctionRepository(db)

	scrapperFactory := &auction.CollyFactory{}
	parserFactory := &auction.HtmlParserFactory{}

	tibiaAPI := vendor.NewTibiaApi()

	mapper := auction.NewMapper(worldRepository)
	service := auction.NewService(tibiaAPI, auctionRepository, worldRepository, currencyRepository, aggAuctionRepository, mapper, parserFactory, scrapperFactory, logger)

	return &Modules{
		Auction: service,
	}
}

func gracefulShutdown(cleanups ...ShutdownFunc) ShutdownFunc {
	return func(ctx context.Context) error {
		var combinedErr error

		for i, cleanup := range cleanups {
			if cleanup == nil {
				continue
			}

			if err := cleanup(ctx); err != nil {
				slog.Error("Cleanup task failed", "task_index", i, "error", err)
				if combinedErr == nil {
					combinedErr = err
				} else {
					combinedErr = fmt.Errorf("%v; %w", combinedErr, err)
				}
			}
		}

		return combinedErr
	}
}
