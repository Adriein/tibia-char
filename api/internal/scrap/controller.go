package scrap

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	service *Service
}

func NewController(service *Service) *Controller {
	return &Controller{
		service: service,
	}
}

func (c *Controller) Get() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c.service.ScrapBazaar()

		ctx.JSON(http.StatusOK, gin.H{"ok": true, "data": "pong"})
	}
}
