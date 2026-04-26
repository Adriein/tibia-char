package beautify

import (
	"fmt"
	"time"
)

func FormatTimeAgo(duration time.Duration) string {
	if duration < time.Minute {
		return "just now"
	}

	minutes := int(duration.Minutes())
	if minutes < 60 {
		return fmt.Sprintf("%dm ago", minutes)
	}

	hours := int(duration.Hours())
	if hours < 24 {
		return fmt.Sprintf("%dh ago", hours)
	}

	days := hours / 24
	if days < 30 {
		return fmt.Sprintf("%dd ago", days)
	}

	months := days / 30
	if months < 12 {
		return fmt.Sprintf("%dmo ago", months)
	}

	years := months / 12
	return fmt.Sprintf("%dy ago", years)
}

func TimeLeft(endAt time.Time) string {
	var timeLeft string

	duration := time.Until(endAt)

	d := int(duration.Hours()) / 24
	h := int(duration.Hours()) % 24
	m := int(duration.Minutes()) % 60

	switch {
	case duration <= 0:
		timeLeft = "Finished"
	case d > 0:
		timeLeft = fmt.Sprintf("%dd %dh", d, h)
	case h > 0:
		timeLeft = fmt.Sprintf("%dh %dm", h, m)
	default:
		timeLeft = fmt.Sprintf("%dm", m)
	}

	return timeLeft
}
