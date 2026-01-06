package currency

import (
	"context"
	"database/sql"
	"time"

	"github.com/rotisserie/eris"
)

type CurrencyRepository interface {
	Save(ctx context.Context, currency *CurrencyRate) error
	GetLatest(ctx context.Context) (*CurrencyRate, error)
}

type PgCurrencyRepository struct {
	connection *sql.DB
}

func NewPgCurrencyRepository(db *sql.DB) *PgCurrencyRepository {
	return &PgCurrencyRepository{connection: db}
}

func (r *PgCurrencyRepository) Save(ctx context.Context, currency *CurrencyRate) error {
	query := `
		INSERT INTO tc_currency_rates (
			tcr_usd,
			tcr_eur,
			tcr_aud,
			tcr_gbp,
			tcr_pln,
			tcr_brl,
			tcr_date_upd
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7);
	`

	_, err := r.connection.ExecContext(
		ctx,
		query,
		currency.USD,
		currency.EUR,
		currency.AUD,
		currency.GBP,
		currency.PLN,
		currency.BRL,
		currency.DateUpd.Format(time.DateTime),
	)

	if err != nil {
		return eris.Wrap(err, "Error saving currency rates")
	}

	return nil
}

func (r *PgCurrencyRepository) GetLatest(ctx context.Context) (*CurrencyRate, error) {
	query := `
		SELECT
			tcr_usd,
      tcr_eur,
      tcr_aud,
      tcr_gbp,
      tcr_pln,
      tcr_brl,
      tcr_date_upd
		FROM
			tc_currency_rates
		ORDER BY
			tcr_date_upd DESC
		LIMIT 1
		;
	`

	row := r.connection.QueryRowContext(ctx, query)

	var currency CurrencyRate

	err := row.Scan(
		&currency.USD,
		&currency.EUR,
		&currency.AUD,
		&currency.GBP,
		&currency.PLN,
		&currency.BRL,
		&currency.DateUpd,
	)

	currency.DateUpd = currency.DateUpd.In(time.UTC)

	if err != nil {
		return nil, eris.Wrap(err, "Failed to query currency rates")
	}

	return &currency, nil
}
