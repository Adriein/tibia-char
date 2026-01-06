package vendor

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/adriein/tibia-char/pkg/enums"
	"github.com/rotisserie/eris"
)

type CurrencyConversionRates struct {
	Result     string `json:"result"`
	ErrorType  string `json:"error-type"`
	LastUpdate string `json:"time_last_update_utc"`
	Rates      struct {
		USD float64 `json:"USD"`
		EUR float64 `json:"EUR"`
		AUD float64 `json:"AUD"`
		GBP float64 `json:"GBP"`
		PLN float64 `json:"PLN"`
		BRL float64 `json:"BRL"`
	} `json:"rates"`
}

type OpenCurrencyAPI struct{}

func NewOpenCurrencyAPI() *OpenCurrencyAPI {
	return &OpenCurrencyAPI{}
}

func (o *OpenCurrencyAPI) GetConversionRates(ctx context.Context, currency enums.Currency) (*CurrencyConversionRates, error) {
	if ok := currency.Valid(); !ok {
		return nil, eris.Errorf("Currency %s not supported", currency)
	}

	url := fmt.Sprintf("https://open.er-api.com/v6/latest/%s", currency)

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	if err != nil {
		return nil, eris.Wrapf(err, "Error creating request for currency: %s", currency)
	}

	httpRes, err := http.DefaultClient.Do(request)

	if err != nil {
		return nil, eris.Wrapf(err, "Failed to fetch currency: %s", currency)
	}

	defer httpRes.Body.Close()

	if httpRes.StatusCode != http.StatusOK {
		return nil, eris.Errorf("open.er-api.com responded with http code %d fetching currency %s", httpRes.StatusCode, currency)
	}

	var response CurrencyConversionRates

	err = json.NewDecoder(httpRes.Body).Decode(&response)

	if err != nil {
		return nil, eris.Wrapf(err, "Error fetching currency: %s", currency)
	}

	if response.Result != "success" {
		return nil, eris.Errorf("Error fetching currency %s with status %s and type %s", currency, response.Result, response.ErrorType)
	}

	return &response, nil
}
