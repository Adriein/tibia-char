package auction

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/adriein/tibia-char/internal/auction/model"
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
	GetOrCreate(world string) (*model.World, error)
}

type PgWorldRepository struct {
	connection *sql.DB
}

func NewPgWorldRepository(c *sql.DB) *PgWorldRepository {
	return &PgWorldRepository{connection: c}
}

func (wr *PgWorldRepository) GetOrCreate(world string) (*model.World, error) {
	query := `
		WITH ins AS (
			INSERT INTO tc_world (tw_name)
			VALUES ($1)
			ON CONFLICT (tw_name) DO NOTHING
			RETURNING tw_id, tw_name
		)
		SELECT tw_id, tw_name FROM ins
		UNION ALL
		SELECT tw_id, tw_name FROM tc_world WHERE tw_name = $1
		LIMIT 1;
	`

	var dto model.World

	err := wr.connection.QueryRow(query, world).Scan(&dto.Id, &dto.Name)

	if err != nil {
		return nil, eris.New(err.Error())
	}

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
	var b strings.Builder

	b.WriteString("INSERT INTO tc_auction (")
	b.WriteString("ta_id, ta_tibia_auction_link, ta_img, ta_char_name, ta_char_level, ta_char_vocation, ta_char_gender, ta_char_world, ")
	b.WriteString("ta_current_bid, ta_auction_start, ta_auction_end, ta_is_active, ta_date_add, ta_date_upd")
	b.WriteString(") VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)")

	var query = b.String()

	_, err := r.connection.Exec(
		query,
		auction.Id,
		auction.TibiaAuctionLink,
		auction.Img,
		auction.CharName,
		auction.CharLevel,
		auction.CharVocation.Id,
		auction.CharGender.Id,
		auction.CharWorld.Id,
		auction.Bid,
		auction.AuctionStart.Format(time.DateTime),
		auction.AuctionEnd.Format(time.DateTime),
		auction.IsActive,
		auction.DateAdd.Format(time.DateTime),
		auction.DateUpd.Format(time.DateTime),
	)

	if err != nil {
		return eris.New(err.Error())
	}

	return nil
}

func (r *PgAuctionRepository) GetActiveAuctions(ctx context.Context) ([]*model.Auction, error) {
	query := `
		SELECT
			a.ta_id,
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
			a.ta_is_active,
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
		WHERE
			a.ta_is_active = 1
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

		err := rows.Scan(
			&auction.Id,
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
			&auction.IsActive,
			&auction.DateAdd,
			&auction.DateUpd,
		)
		if err != nil {
			return nil, eris.Wrap(err, "Failed to scan auction")
		}

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
