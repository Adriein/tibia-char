package currency

import "database/sql"

type CurrencyRepository interface {
	Save(currency *CurrencyRate) error
}

type PgCurrencyRepository struct {
	connection *sql.DB
}

func NewPgCurrencyRepository(db *sql.DB) *PgCurrencyRepository {
	return &PgCurrencyRepository{connection: db}
}

func (r *PgCurrencyRepository) Save(currency *CurrencyRate) error {
	return nil
}
