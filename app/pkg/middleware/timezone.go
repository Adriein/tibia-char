package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

const TimezoneKey = "ReqLocation"

func TimeZone() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		loc := time.UTC

		if cookie, err := ctx.Cookie("User_Tz"); err == nil {
			if userLoc, err := time.LoadLocation(cookie); err == nil {
				ctx.Set(TimezoneKey, userLoc)

				ctx.Next()
			}
		}

		ctx.Set(TimezoneKey, loc)

		ctx.Next()
	}
}
