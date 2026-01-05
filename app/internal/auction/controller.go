package auction

import (
	"log"
	"net/http"
	"os"

	"github.com/adriein/tibia-char/internal"
	"github.com/adriein/tibia-char/pkg/middleware"
	"github.com/adriein/tibia-char/pkg/vendor"
	"github.com/gin-gonic/gin"
	"github.com/rotisserie/eris"
)

type Controller struct {
	service *Service
	logger  *log.Logger
}

func NewController(app *internal.App) *Controller {
	logger := log.New(os.Stderr, "[Auction] ", log.LstdFlags|log.LUTC)

	auctionRepository := NewPgAuctionRepository(app.Databse)
	worldRepository := NewPgWorldRepository(app.Databse)

	linkScrapper := NewScrapper("CollectAuctionLinks")

	detailScrapper := NewScrapper("CollectAuctionDetails")

	tibiaAPI := vendor.NewTibiaApi()

	auctionListHtmlParser := NewAuctionListHtmlParser(linkScrapper.Collector)
	auctionHtmlParser := NewAuctionHtmlParser(detailScrapper.Collector)

	mapper := NewMapper(worldRepository)

	service := NewService(tibiaAPI, auctionListHtmlParser, auctionHtmlParser, auctionRepository, worldRepository, mapper, logger)

	return &Controller{
		service: service,
		logger:  logger,
	}
}

func (c *Controller) Get() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		traceID, _ := ctx.Get(middleware.TraceIDKey)

		auctions, err := c.service.GetAuctions(ctx)

		if err != nil {
			//TODO: Temporal logging until I decide what to do
			c.logger.Printf("TraceID %s Error getting auctions: %s\n", traceID, eris.ToString(err, true))
		}

		renderer := vendor.NewTemplRenderer(ctx, http.StatusOK, AuctionsView(auctions))

		ctx.Render(http.StatusOK, renderer)
	}
}
