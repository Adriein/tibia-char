package auction

import (
	"database/sql"

	"github.com/rotisserie/eris"
)

type VocationRepository interface {
	Get(vocation string) (*Vocation, error)
}

type PgVocationRepsitory struct {
	connection *sql.DB
}

func NewPgVocationRepository(c *sql.DB) *PgVocationRepsitory {
	return &PgVocationRepsitory{connection: c}
}

func (vr *PgVocationRepsitory) Get(vocation string) (*Vocation, error) {
	query := `SELECT * FROM tc_vocation WHERE tv_name = $1;`

	var dto Vocation

	err := vr.connection.QueryRow(query, vocation).Scan(&dto.Id, &dto.Name)

	if err != nil {
		return nil, eris.New(err.Error())
	}

	return &dto, nil
}
