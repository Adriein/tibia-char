package internal

import (
	"database/sql"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/adriein/tibia-char/database"
	"github.com/adriein/tibia-char/internal/auction"
	"github.com/adriein/tibia-char/internal/currency"
	"github.com/adriein/tibia-char/pkg/constants"
	"github.com/adriein/tibia-char/pkg/helper"
	"github.com/adriein/tibia-char/pkg/vendor"
	"github.com/joho/godotenv"
)

type Modules struct {
	Auction auction.AuctionService
}

type App struct {
	Database *sql.DB
	Modules  *Modules
	Logger   *slog.Logger
}

func NewApp() *App {
	if os.Getenv(constants.Env) != constants.Prod {
		dotenvErr := godotenv.Load()

		if dotenvErr != nil {
			log.Fatal("Error loading .env file")
		}
	}

	checker := helper.NewEnvVarChecker(
		constants.DatabaseUser,
		constants.DatabasePassword,
		constants.DatabaseName,
		constants.ServerPort,
		constants.Env,
	)

	if envCheckerErr := checker.Check(); envCheckerErr != nil {
		log.Fatal(envCheckerErr.Error())
	}

	logger := initLogger()

	db := database.New()
	modules := initModules(db, logger)

	return &App{
		Database: db,
		Modules:  modules,
		Logger:   logger,
	}
}

func initLogger() *slog.Logger {
	opts := &slog.HandlerOptions{
		ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
			if attr.Key == slog.TimeKey {
				formatted := attr.Value.Time().UTC().Format(time.DateTime)

				return slog.String(slog.TimeKey, formatted)
			}

			return attr
		},
	}

	if os.Getenv(constants.Env) == constants.Dev {
		return slog.New(slog.NewTextHandler(os.Stdout, opts))
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, opts))

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
