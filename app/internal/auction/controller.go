package auction

import (
	"net/http"

	"github.com/adriein/tibia-char/views"
	"github.com/gin-gonic/gin"
)

type Controller struct{}

func NewController() *Controller {

	return &Controller{}
}

func (c *Controller) Get() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		ctx.HTML(http.StatusOK, "", views.Auctions("Adri"))
	}
}
