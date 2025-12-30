package controller

import (
	"log"
	"net/http"
	"os"

	"github.com/adriein/tibia-char/internal"
	"github.com/adriein/tibia-char/internal/auction"
	"github.com/adriein/tibia-char/internal/auction/view"
	"github.com/adriein/tibia-char/pkg/middleware"
	"github.com/adriein/tibia-char/pkg/vendor"
	"github.com/gin-gonic/gin"
	"github.com/rotisserie/eris"
)

type Controller struct {
	service *auction.Service
	logger  *log.Logger
}

func NewController(app *internal.App) *Controller {
	logger := log.New(os.Stderr, "[Auction] ", log.LstdFlags|log.LUTC)

	auctionRepository := auction.NewPgAuctionRepository(app.Databse)
	worldRepository := auction.NewPgWorldRepository(app.Databse)

	linkScrapper := auction.NewScrapper("CollectAuctionLinks")

	detailScrapper := auction.NewScrapper("CollectAuctionDetails")

	tibiaAPI := vendor.NewTibiaApi()

	auctionListHtmlParser := auction.NewAuctionListHtmlParser(linkScrapper.Collector)
	auctionHtmlParser := auction.NewAuctionHtmlParser(detailScrapper.Collector)

	mapper := auction.NewMapper(worldRepository)

	service := auction.NewService(tibiaAPI, auctionListHtmlParser, auctionHtmlParser, auctionRepository, worldRepository, mapper, logger)

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
			//Temporal logging until I decide what to do
			c.logger.Printf("TraceID %s Error getting auctions: %s\n", traceID, eris.ToString(err, true))
		}

		renderer := vendor.NewTemplRenderer(ctx, http.StatusOK, view.Index(auctions))

		ctx.Render(http.StatusOK, renderer)
	}
}
