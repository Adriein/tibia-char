package server

import (
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/rotisserie/eris"

	"github.com/adriein/tibia-char/internal"
	"github.com/adriein/tibia-char/internal/health"
	"github.com/adriein/tibia-char/internal/web"
	"github.com/adriein/tibia-char/pkg/middleware"
	"github.com/adriein/tibia-char/pkg/vendor"
)

type TibiaChar struct {
	app       *internal.App
	gin       *gin.Engine
	validator *validator.Validate
}

func New(port string) *TibiaChar {
	app := internal.NewApp()

	engine := gin.New()

	ginHtmlRenderer := engine.HTMLRender

	engine.HTMLRender = &vendor.HTMLTemplRenderer{FallbackHtmlRenderer: ginHtmlRenderer}

	// Disable trusted proxy warning.
	engine.SetTrustedProxies(nil)

	engine.Use(middleware.Error(), gin.Logger(), gin.Recovery(), middleware.Tracer(), middleware.TimeZone())

	tibiaChar := &TibiaChar{
		app:       app,
		gin:       engine,
		validator: validator.New(),
	}

	tibiaChar.routeSetup()

	if ginErr := engine.Run(port); ginErr != nil {
		err := eris.Wrap(ginErr, "Error starting HTTP server")

		log.Fatal(eris.ToString(err, true))
	}

	slog.Info("Starting the TibiaChar at " + port)

	return tibiaChar
}

func (t *TibiaChar) routeSetup() {
	//HEALTH CHECK
	t.gin.GET("/ping", health.NewController().Get())

	cwd, _ := os.Getwd()

	//STATIC
	t.gin.Static("/ui/static", fmt.Sprintf("%s/ui/static", cwd))

	//AUCTIONS
	t.gin.GET("/index", web.NewAuctionController(t.app).Get())
}
