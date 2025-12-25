package auction

import (
	"context"
	"database/sql"
	"time"

	"github.com/adriein/tibia-char/internal/auction/model"
	"github.com/adriein/tibia-char/pkg/constants"
	"github.com/rotisserie/eris"
)

type VocationRepository interface {
	GetOrCreate(vocation string) (*model.Vocation, error)
}

type PgVocationReository struct {
	connection *sql.DB
}

func NewPgVocationRepository(c *sql.DB) *PgVocationReository {
	return &PgVocationReository{connection: c}
}

func (vr *PgVocationReository) GetOrCreate(vocation string) (*model.Vocation, error) {
	query := `
		WITH ins AS (
			INSERT INTO tc_vocation (tv_name)
			VALUES ($1)
			ON CONFLICT (tv_name) DO NOTHING
			RETURNING tv_id, tv_name
		)
		SELECT tv_id, tv_name FROM ins
		UNION ALL
		SELECT tv_id, tv_name FROM tc_vocation WHERE tv_name = $1
		LIMIT 1;
	`

	var dto model.Vocation

	err := vr.connection.QueryRow(query, vocation).Scan(&dto.Id, &dto.Name)

	if err != nil {
		return nil, eris.New(err.Error())
	}

	return &dto, nil
}

type GenderRepository interface {
	GetOrCreate(gender string) (*model.Gender, error)
}

type PgGenderRepository struct {
	connection *sql.DB
}

func NewPgGenderRepository(c *sql.DB) *PgGenderRepository {
	return &PgGenderRepository{connection: c}
}

func (gr *PgGenderRepository) GetOrCreate(gender string) (*model.Gender, error) {
	query := `
		WITH ins AS (
			INSERT INTO tc_gender (tg_name)
			VALUES ($1)
			ON CONFLICT (tg_name) DO NOTHING
			RETURNING tg_id, tg_name
		)
		SELECT tg_id, tg_name FROM ins
		UNION ALL
		SELECT tg_id, tg_name FROM tc_gender WHERE tg_name = $1
		LIMIT 1;
	`

	var dto model.Gender

	err := gr.connection.QueryRow(query, gender).Scan(&dto.Id, &dto.Name)

	if err != nil {
		return nil, eris.New(err.Error())
	}

	return &dto, nil
}

type WorldRepository interface {
	GetOrCreate(world *model.World) (*model.World, error)
}

type PgWorldRepository struct {
	connection *sql.DB
}

func NewPgWorldRepository(c *sql.DB) *PgWorldRepository {
	return &PgWorldRepository{connection: c}
}

func (wr *PgWorldRepository) GetOrCreate(world *model.World) (*model.World, error) {
	query := `
		WITH ins AS (
			INSERT INTO tc_world (tw_name, tw_location, tw_battle_eye)
			VALUES ($1, $2, $3)
			ON CONFLICT (tw_name) DO NOTHING
			RETURNING tw_id, tw_name, tw_location, tw_battle_eye
		)
		SELECT * FROM ins
		UNION ALL
		SELECT * FROM tc_world WHERE tw_name = $1
		LIMIT 1;
	`

	var dto model.World

	var battleEye string

	err := wr.connection.QueryRow(
		query,
		world.Name,
		world.Location,
		world.BattleEye.String(),
	).Scan(
		&dto.Id,
		&dto.Name,
		&dto.Location,
		&battleEye,
	)

	if err != nil {
		return nil, eris.New(err.Error())
	}

	dto.BattleEye, err = constants.GetBattleEyeFromString(battleEye)

	if err != nil {
		return nil, err
	}

	return &dto, nil
}

type WorldTransferRepository interface {
	Get(transfer string) (*model.WorldTransfer, error)
}

type PgWorldTransferRepository struct {
	connection *sql.DB
}

func NewPgWorldTransferRepository(c *sql.DB) *PgWorldTransferRepository {
	return &PgWorldTransferRepository{connection: c}
}

func (wtr *PgWorldTransferRepository) Get(transfer string) (*model.WorldTransfer, error) {
	query := `
		SELECT * FROM tc_world_transfer WHERE twt_name = $1;
	`

	var dto model.WorldTransfer
	var worldTransferAllowance string

	err := wtr.connection.QueryRow(query, transfer).Scan(&dto.Id, &worldTransferAllowance)

	if err != nil {
		return nil, eris.New(err.Error())
	}

	wta, err := constants.GetWorldTransferAllowanceFromString(worldTransferAllowance)

	if err != nil {
		return nil, err
	}

	dto.Name = wta

	return &dto, nil
}

type AuctionRepository interface {
	Save(auction *model.Auction) error
	GetActiveAuctions(ctx context.Context) ([]*model.Auction, error)
}

type PgAuctionRepository struct {
	connection *sql.DB
}

func NewPgAuctionRepository(connection *sql.DB) *PgAuctionRepository {
	return &PgAuctionRepository{
		connection: connection,
	}
}

func (r *PgAuctionRepository) Save(auction *model.Auction) error {
	tx, err := r.connection.Begin()

	if err != nil {
		return eris.Wrap(err, "Error creating transaction")
	}

	auctionQuery := `
		INSERT INTO tc_auction (
			ta_auction_id,
			ta_tibia_auction_link,
			ta_img,
			ta_char_name,
			ta_char_level,
			ta_char_vocation,
			ta_char_gender,
			ta_char_world,
			ta_world_transfer,
			ta_current_bid,
			ta_auction_start,
			ta_auction_end,
			ta_date_add,
			ta_date_upd
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING ta_id;
	`

	var generatedId int64

	err = tx.QueryRow(
		auctionQuery,
		auction.AuctionID,
		auction.TibiaAuctionLink,
		auction.Img,
		auction.CharName,
		auction.CharLevel,
		auction.CharVocation.Id,
		auction.CharGender.Id,
		auction.CharWorld.Id,
		auction.WorldTransfer.Id,
		auction.Bid,
		auction.AuctionStart,
		auction.AuctionEnd,
		auction.DateAdd,
		auction.DateUpd,
	).Scan(&generatedId)

	if err != nil {
		tx.Rollback()

		return eris.Wrap(err, "Error on insert on tc_auction")
	}

	recordingQuery := `
		INSERT INTO tc_auction_recording (
			tar_auction_id,
			tar_recordable_id,
			tar_status,
			tar_date_add,
			tar_date_upd
		)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (tar_auction_id) DO UPDATE SET
			tar_status = EXCLUDED.tar_status,
			tar_date_upd = EXCLUDED.tar_date_upd;
	`

	_, err = tx.Exec(
		recordingQuery,
		auction.AuctionID,
		generatedId,
		auction.Status.String(),
		auction.DateAdd,
		auction.DateUpd,
	)

	if err != nil {
		tx.Rollback()

		return eris.Wrap(err, "Error in upsert on tc_auction_recording")
	}

	return tx.Commit()
}

func (r *PgAuctionRepository) GetActiveAuctions(ctx context.Context) ([]*model.Auction, error) {
	query := `
		SELECT
			a.ta_id,
			a.ta_auction_id,
			a.ta_tibia_auction_link,
			a.ta_img,
			a.ta_char_name,
			a.ta_char_level,
			v.tv_id,
			v.tv_name,
			g.tg_id,
			g.tg_name,
			w.tw_id,
			w.tw_name,
			a.ta_current_bid,
			a.ta_auction_start,
			a.ta_auction_end,
			tar.tar_status,
			a.ta_date_add,
			a.ta_date_upd
		FROM
			tc_auction a
		INNER JOIN
			tc_vocation v ON a.ta_char_vocation = v.tv_id
		INNER JOIN
			tc_gender g ON a.ta_char_gender = g.tg_id
		INNER JOIN
			tc_world w ON a.ta_char_world = w.tw_id
		INNER JOIN
			tc_auction_recording tar ON a.ta_id = tar.tar_recordable_id
		WHERE
			tar.tar_status = 'active'
		ORDER BY a.ta_auction_end ASC;
	`

	ctxTimeout, cancel := context.WithTimeout(ctx, time.Second*10)

	defer cancel()

	rows, err := r.connection.QueryContext(ctxTimeout, query)

	if err != nil {
		return nil, eris.Wrap(err, "Failed to query active auctions")
	}

	defer rows.Close()

	var auctions []*model.Auction

	for rows.Next() {
		var auction model.Auction
		var vocation model.Vocation
		var gender model.Gender
		var world model.World
		var statusString string

		err := rows.Scan(
			&auction.ID,
			&auction.AuctionID,
			&auction.TibiaAuctionLink,
			&auction.Img,
			&auction.CharName,
			&auction.CharLevel,
			&vocation.Id,
			&vocation.Name,
			&gender.Id,
			&gender.Name,
			&world.Id,
			&world.Name,
			&auction.Bid,
			&auction.AuctionStart,
			&auction.AuctionEnd,
			&statusString,
			&auction.DateAdd,
			&auction.DateUpd,
		)

		if err != nil {
			return nil, eris.Wrap(err, "Failed to scan auction")
		}

		status, err := constants.GetAuctionRecordableStatusFromString(statusString)

		if err != nil {
			return nil, eris.Wrap(err, "Failed to parse auction status")
		}

		auction.Status = status
		auction.CharVocation = &vocation
		auction.CharGender = &gender
		auction.CharWorld = &world

		auctions = append(auctions, &auction)
	}

	if err = rows.Err(); err != nil {
		return nil, eris.Wrap(err, "Failed iterating rows")
	}

	return auctions, nil
}
