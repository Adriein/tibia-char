package auction

import (
	"database/sql"

	"github.com/rotisserie/eris"
)

type GenderRepository interface {
	Get(gender string) (*Gender, error)
}

type PgGenderRepository struct {
	connection *sql.DB
}

func NewPgGenderRepository(c *sql.DB) *PgGenderRepository {
	return &PgGenderRepository{connection: c}
}

func (gr *PgGenderRepository) Get(gender string) (*Gender, error) {
	query := `SELECT * FROM tc_gender WHERE tg_name = $1`

	var dto Gender

	err := gr.connection.QueryRow(query, gender).Scan(&dto.Id, &dto.Name)

	if err != nil {
		return nil, eris.New(err.Error())
	}

	return &dto, nil
}
