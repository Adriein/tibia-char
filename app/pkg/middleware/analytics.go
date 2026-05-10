package middleware

import (
	"github.com/adriein/tibia-char/pkg/constants"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/posthog/posthog-go"
)

func Analytics(client posthog.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		sessionId := getOrSetSessionID(ctx)

		client.Enqueue(posthog.Capture{
			DistinctId: sessionId,
			Event:      "$pageview",
			Properties: posthog.NewProperties().
				Set("$current_url", "https://example.com"),
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
		86400,
		"/",
		"",
		true,
		true,
	)

	return sessionID
}
