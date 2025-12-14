package auction

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	tcerrors "github.com/adriein/tibia-char/pkg/errors"
	"github.com/rotisserie/eris"
)

type VocationRepository interface {
	GetByName(vocation string) (*Vocation, error)
}

type PgVocationReository struct {
	connection *sql.DB
}

func (r *PgVocationReository) NewPgVocationRepository(c *sql.DB) *PgVocationReository {
	return &PgVocationReository{connection: c}
}

func (v *PgVocationReository) GetByName(vocation string) (*Vocation, error) {
	statement, err := v.connection.Prepare("SELECT * FROM tc_vocation WHERE name = $1;")

	if err != nil {
		return nil, eris.New(err.Error())
	}

	var (
		id   string
		name string
	)

	if scanErr := statement.QueryRow(vocation).Scan(
		&id,
		&name,
	); scanErr != nil {
		if errors.Is(scanErr, sql.ErrNoRows) {
			return nil, eris.Wrapf(tcerrors.NotFoundError, "Vocation %s not found", vocation)
		}

		return nil, eris.New(scanErr.Error())
	}

	return &Vocation{Id: id, Name: name}, nil
}

type GenderRepository interface {
	GetByName(gender string) (*Gender, error)
}

type PgGenderRepository struct {
	connection *sql.DB
}

func (r *PgGenderRepository) NewPgGenderRepository(c *sql.DB) *PgGenderRepository {
	return &PgGenderRepository{connection: c}
}

func (v *PgGenderRepository) GetByName(gender string) (*Gender, error) {
	statement, err := v.connection.Prepare("SELECT * FROM tc_gender WHERE name = $1;")

	if err != nil {
		return nil, eris.New(err.Error())
	}

	var (
		id   string
		name string
	)

	if scanErr := statement.QueryRow(gender).Scan(
		&id,
		&name,
	); scanErr != nil {
		if errors.Is(scanErr, sql.ErrNoRows) {
			return nil, eris.Wrapf(tcerrors.NotFoundError, "Gender %s not found", gender)
		}

		return nil, eris.New(scanErr.Error())
	}

	return &Gender{Id: id, Name: name}, nil
}

type WorldRepository interface {
	GetByName(world string) (*World, error)
}

type PgWorldRepository struct {
	connection *sql.DB
}

func (r *PgGenderRepository) NewPgWorldRepository(c *sql.DB) *PgWorldRepository {
	return &PgWorldRepository{connection: c}
}

func (v *PgWorldRepository) GetByName(world string) (*World, error) {
	statement, err := v.connection.Prepare("SELECT * FROM tc_world WHERE name = $1;")

	if err != nil {
		return nil, eris.New(err.Error())
	}

	var (
		id   string
		name string
	)

	if scanErr := statement.QueryRow(world).Scan(
		&id,
		&name,
	); scanErr != nil {
		if errors.Is(scanErr, sql.ErrNoRows) {
			return nil, eris.Wrapf(tcerrors.NotFoundError, "World %s not found", world)
		}

		return nil, eris.New(scanErr.Error())
	}

	return &World{Id: id, Name: name}, nil
}

type AuctionRepository interface {
	Save(auction *Auction) error
}

type PgAuctionRepository struct {
	connection *sql.DB
}

func NewPgAuctionRepository(connection *sql.DB) *PgAuctionRepository {
	return &PgAuctionRepository{
		connection: connection,
	}
}

func (r *PgAuctionRepository) Save(auction *Auction) error {
	var b strings.Builder

	b.WriteString("INSERT INTO tc_auction (")
	b.WriteString("ta_tibia_auction_id, ta_tibia_auction_link, ta_img, ta_char_name, ta_char_level, ta_char_vocation, ta_char_gender, ta_char_world, ")
	b.WriteString("ta_current_bid, ta_auction_start, ta_auction_end, ta_is_active, ta_date_add, ta_date_upd")
	b.WriteString(") VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)")

	var query = b.String()

	_, err := r.connection.Exec(
		query,
		auction.TibiaAuctionId,
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
