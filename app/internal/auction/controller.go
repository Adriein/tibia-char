package auction

import (
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/adriein/tibia-char/internal"
	"github.com/adriein/tibia-char/internal/currency"
	"github.com/adriein/tibia-char/pkg/constants"
	"github.com/adriein/tibia-char/pkg/vendor"
	"github.com/gin-gonic/gin"
	"github.com/rotisserie/eris"
)

type Controller struct {
	service *Service
	logger  *slog.Logger
}

func NewController(app *internal.App) *Controller {
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

	auctionRepository := NewPgAuctionRepository(app.Database)
	worldRepository := NewPgWorldRepository(app.Database)
	currencyRepository := currency.NewPgCurrencyRepository(app.Database)
	aggAuctionRepository := NewPgAggAuctionRepository(app.Database)

	tibiaAPI := vendor.NewTibiaApi()

	scrapperFactory := &CollyFactory{}
	parserFactory := &HtmlParserFactory{}

	mapper := NewMapper(worldRepository)

	service := NewService(tibiaAPI, auctionRepository, worldRepository, currencyRepository, aggAuctionRepository, mapper, parserFactory, scrapperFactory, logger)

	return &Controller{
		service: service,
		logger:  logger,
	}
}

func (c *Controller) Get() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		qty, err := strconv.Atoi(ctx.DefaultQuery("qty", "20"))

		if err != nil {
			c.logger.Error("Error getting auctions", "error", eris.ToString(err, true))
		}

		pageNum, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))

		if err != nil {
			c.logger.Error("Error getting auctions", "error", eris.ToString(err, true))
		}

		filter := &AuctionFilter{
			Pagination: &FilterPagination{
				Limit:     qty,
				Page:      pageNum - 1,
				SortBy:    SortByEndTime,
				SortOrder: SortOrderAsc,
			},
		}

		page, err := c.service.GetAuctions(ctx, filter)

		if err != nil {
			//TODO: Temporal logging until I decide what to do
			c.logger.Error("Error getting auctions", "error", eris.ToString(err, true))
		}

		renderer := vendor.NewTemplRenderer(ctx, http.StatusOK, AuctionsView(page))

		ctx.Render(http.StatusOK, renderer)
	}
}
