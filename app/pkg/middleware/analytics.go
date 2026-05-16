package middleware

import (
	"fmt"

	"github.com/adriein/tibia-char/pkg/constants"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/posthog/posthog-go"
)

func Analytics(client posthog.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		sessionId := getOrSetSessionID(ctx)

		fmt.Println(sessionId)

		var protocol string

		if ctx.Request.Proto == "HTTP/1.1" {
			protocol = "http"
		} else {
			protocol = "https"
		}

		currentURL := fmt.Sprintf("%s://%s%s", protocol, ctx.Request.Host, ctx.Request.URL.Path)

		client.Enqueue(posthog.Capture{
			DistinctId: sessionId,
			Event:      "$pageview",
			Properties: posthog.NewProperties().
				Set("$current_url", currentURL),
		})

		ctx.Next()
	}
}

func getOrSetSessionID(ctx *gin.Context) string {
	cookie, err := ctx.Cookie(constants.SessionCookie)

	if err == nil {
		return cookie
	}

	sessionID := uuid.NewString()

	ctx.SetCookie(
		constants.SessionCookie,
		sessionID,
		constants.DayInSeconds,
		"/",
		"",
		true,
		true,
	)

	return sessionID
}
