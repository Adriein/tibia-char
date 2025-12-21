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
	app *internal.App
}

func New(port string) *TibiaChar {
	router := gin.Default()

	ginHtmlRenderer := router.HTMLRender

	router.HTMLRender = &HTMLTemplRenderer{FallbackHtmlRenderer: ginHtmlRenderer}

	// Disable trusted proxy warning.
	router.SetTrustedProxies(nil)

	router.Use(middleware.Error())

	//router.LoadHTMLGlob("views/*")

	tibiaChar := &TibiaChar{
		app: internal.NewApp(),
	}

	tibiaChar.app.SetRouter(router)
	tibiaChar.app.SetValidator(validator.New())

	tibiaChar.routeSetup()

	if ginErr := router.Run(port); ginErr != nil {
		err := eris.Wrap(ginErr, "Error starting HTTP server")

		log.Fatal(eris.ToString(err, true))
	}

	slog.Info("Starting the TibiaChar at " + port)

	return tibiaChar
}

func (t *TibiaChar) routeSetup() {
	//HEALTH CHECK
	t.app.Router.GET("/ping", health.NewController().Get())

	//AUCTIONS
	t.app.Router.GET("/index", auction.NewController().Get())
}
