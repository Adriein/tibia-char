package server

import (
	"log"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/rotisserie/eris"

	"github.com/adriein/tibia-char/internal"
	"github.com/adriein/tibia-char/internal/auction"
	"github.com/adriein/tibia-char/internal/health"
	"github.com/adriein/tibia-char/pkg/middleware"
)

type TibiaChar struct {
	app       *internal.App
	gin       *gin.Engine
	validator *validator.Validate
}

func New(port string) *TibiaChar {
	engine := gin.Default()

	ginHtmlRenderer := engine.HTMLRender

	engine.HTMLRender = &HTMLTemplRenderer{FallbackHtmlRenderer: ginHtmlRenderer}

	// Disable trusted proxy warning.
	engine.SetTrustedProxies(nil)

	engine.Use(middleware.Error())

	tibiaChar := &TibiaChar{
		app:       internal.NewApp(),
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

	//AUCTIONS
	t.gin.GET("/index", auction.NewController().Get())
}
