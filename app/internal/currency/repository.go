package currency

import (
	"context"
	"database/sql"
	"time"
)

type CurrencyRepository interface {
	Save(ctx context.Context, currency *CurrencyRate) error
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

	return err
}
