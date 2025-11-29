package middleware

import (
	"log/slog"
	"net/http"

	"github.com/adriein/tibia-char/pkg/constants"
	"github.com/gin-gonic/gin"
	"github.com/rotisserie/eris"
)

func Error() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()

		if len(ctx.Errors) > 0 {
			err := ctx.Errors.Last().Err

			slog.Error(eris.ToString(err, true))

			ctx.JSON(http.StatusInternalServerError, gin.H{
				constants.OkResKey:   false,
				constants.DataResKey: constants.ServerGenericError,
			})
		}
	}
}
