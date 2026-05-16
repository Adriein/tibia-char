package middleware

import (
	"github.com/adriein/tibia-char/pkg/constants"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/posthog/posthog-go"
)

func Analytics(client posthog.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		/*sessionId := getOrSetSessionID(ctx)

		var protocol string

		if ctx.Request.Proto == "HTTP/1.1" {
			protocol = "http"
		} else {
			protocol = "https"
		}

		currentURL := fmt.Sprintf("%s://%s%s", protocol, ctx.Request.Host, ctx.Request.URL.Path)
		os := ctx.Request.Header.Get("Sec-Ch-Ua-Platform")
		browser := ctx.Request.Header.Get("Sec-Ch-Ua")
		device := getDeviceTypeFromHints(ctx)
		referrer := ctx.Request.Referer()

		//TODO: if I want the IP remember that if i have a reverse proxy i need to know which header contains the real ip
		client.Enqueue(posthog.Capture{
			DistinctId: sessionId,
			Event:      "$pageview",
			Properties: posthog.NewProperties().
				Set("$current_url", currentURL).
				Set("$os", os).
				Set("$browser", browser).
				Set("$device_type", device).
				Set("$referrer", referrer),
		})*/

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

func getDeviceTypeFromHints(ctx *gin.Context) string {
	isMobile := ctx.Request.Header.Get("Sec-Ch-Ua-Mobile")

	if isMobile == "?1" {
		return "Mobile"
	}

	if isMobile == "?0" {
		return "Desktop"
	}

	return "Unknown"
}
