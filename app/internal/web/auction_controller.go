package web

import (
	"log/slog"
	"net/http"

	"github.com/adriein/tibia-char/internal"
	"github.com/adriein/tibia-char/internal/auction"
	"github.com/adriein/tibia-char/pkg/middleware"
	"github.com/adriein/tibia-char/pkg/vendor"
	"github.com/adriein/tibia-char/ui/html"
	"github.com/gin-gonic/gin"
	"github.com/rotisserie/eris"
)

type AuctionController struct {
	service auction.AuctionService
	logger  *slog.Logger
}

func NewAuctionController(app *internal.App) *AuctionController {
	return &AuctionController{
		service: app.Modules.Auction,
		logger:  app.Logger,
	}
}

func (c *AuctionController) Get() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		traceID := ctx.Value(middleware.TraceIDKey)

		filter, err := auction.FilterFromQueryParams(ctx.QueryMap(auction.QueryFilter))

		if err != nil {
			//TODO: Temporal logging until I decide what to do
			c.logger.Error("Error getting auctions", "trace_id", traceID, "error", eris.ToString(err, true))
		}

		page, err := c.service.GetAuctions(ctx, filter)

		if err != nil {
			//TODO: Temporal logging until I decide what to do
			c.logger.Error("Error getting auctions", "trace_id", traceID, "error", eris.ToString(err, true))
		}

		renderer := vendor.NewTemplRenderer(ctx, http.StatusOK, html.AuctionsView(page))

		ctx.Render(http.StatusOK, renderer)
	}
}
