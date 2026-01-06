package currency

import (
	"math"
	"strings"
	"time"

	"github.com/adriein/tibia-char/pkg/enums"
)

type CurrencyRate struct {
	USD     float64
	EUR     float64
	AUD     float64
	GBP     float64
	PLN     float64
	BRL     float64
	DateUpd time.Time
}

func (c *CurrencyRate) Exchange(priceInEUR int, target enums.Currency) int {
	var rate float64

	switch target {
	case enums.CurrencyUSD:
		rate = c.USD
	case enums.CurrencyEUR:
		rate = c.EUR
	case enums.CurrencyAUD:
		rate = c.AUD
	case enums.CurrencyGBP:
		rate = c.GBP
	case enums.CurrencyPLN:
		rate = c.PLN
	case enums.CurrencyBRL:
		rate = c.BRL
	default:
		return priceInEUR
	}

	converted := float64(priceInEUR) * rate

	return int(math.Round(converted))
}

func FromLocation(loc *time.Location) enums.Currency {
	if loc == nil {
		return enums.CurrencyUSD
	}

	name := loc.String()

	switch {
	case strings.HasPrefix(name, "Europe/"):
		if strings.Contains(name, "Warsaw") {
			return enums.CurrencyPLN
		}

		if strings.Contains(name, "London") {
			return enums.CurrencyGBP
		}

		return enums.CurrencyEUR

	case strings.HasPrefix(name, "Australia/"):
		return enums.CurrencyAUD

	case strings.HasPrefix(name, "America/Sao_Paulo"):
		return enums.CurrencyBRL

	default:
		return enums.CurrencyUSD
	}
}
