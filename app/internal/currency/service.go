package currency

import (
	"context"
	"log"

	"github.com/adriein/tibia-char/pkg/middleware"
	"github.com/adriein/tibia-char/pkg/vendor"
)

type Service struct {
	repository CurrencyRepository
	api        *vendor.OpenCurrencyAPI
	logger     *log.Logger
}

func NewService(repo CurrencyRepository, api *vendor.OpenCurrencyAPI, logger *log.Logger) *Service {
	return &Service{repository: repo, api: api, logger: logger}
}

func (s *Service) StoreDailyRates(ctx context.Context) error {
	traceID := ctx.Value(middleware.TraceIDKey)

	s.logger.Printf("TraceID: %s Start Currency Cron\n", traceID)

	conRatesDTO, err := s.api.GetConversionRates(ctx, "EUR")

	if err != nil {
		s.logger.Printf("TraceID: %s Finish Currency Cron with error\n", traceID)

		return err
	}

	currencyRate := CurrencyRate{
		USD: conRatesDTO.Rates.USD,
		EUR: conRatesDTO.Rates.EUR,
		AUD: conRatesDTO.Rates.AUD,
		GBP: conRatesDTO.Rates.GBP,
		PLN: conRatesDTO.Rates.PLN,
		BRL: conRatesDTO.Rates.BRL,
	}

	if err := s.repository.Save(&currencyRate); err != nil {
		s.logger.Printf("TraceID: %s Finish Currency Cron with error\n", traceID)

		return err
	}

	return nil
}
