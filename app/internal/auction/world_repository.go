package auction

import (
	"database/sql"
	"errors"

	"github.com/adriein/tibia-char/pkg/enums"
	"github.com/rotisserie/eris"
)

var ErrAggAuctionStatsNotFound = errors.New("Aggregated auction stats not found")


type WorldRepository interface {
	GetOrCreate(world *World) (*World, error)
}

type PgWorldRepository struct {
	connection *sql.DB
}

func NewPgWorldRepository(c *sql.DB) *PgWorldRepository {
	return &PgWorldRepository{connection: c}
}

func (wr *PgWorldRepository) GetOrCreate(world *World) (*World, error) {
	query := `
		WITH existing AS (
    		SELECT * FROM tc_world WHERE tw_name = $1
		),
		inserted AS (
    		INSERT INTO tc_world (tw_name, tw_location, tw_battle_eye, tw_pvp)
    		SELECT $1, $2, $3, $4
    		WHERE NOT EXISTS (SELECT 1 FROM existing)
    		ON CONFLICT (tw_name) DO NOTHING
    		RETURNING *
		)
		SELECT * FROM inserted
		UNION ALL
		SELECT * FROM existing
		LIMIT 1;
	`

	var dto World

	var battleEye string

	err := wr.connection.QueryRow(
		query,
		world.Name,
		world.Location,
		world.BattleEye.String(),
		world.Pvp,
	).Scan(
		&dto.Id,
		&dto.Name,
		&dto.Location,
		&battleEye,
		&dto.Pvp,
	)

	if err != nil {
		return nil, eris.Wrap(err, world.Name)
	}

	dto.BattleEye, err = enums.GetBattleEyeFromString(battleEye)

	if err != nil {
		return nil, err
	}

	return &dto, nil
}
