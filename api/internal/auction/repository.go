package auction

import (
	"database/sql"
	"strings"
	"time"

	"github.com/rotisserie/eris"
)

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
	b.WriteString("ta_tibia_auction_id, ta_img, ta_char_name, ta_char_level, ta_char_vocation, ta_char_gender, ta_char_world, ")
	b.WriteString("ta_current_bid, ta_auction_start, ta_auction_end, ta_is_active, ta_date_add, ta_date_upd")
	b.WriteString(") VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)")

	var query = b.String()

	_, err := r.connection.Exec(
		query,
		auction.TibiaAuctionId,
		auction.Img,
		auction.Name,
		auction.Level,
		auction.Vocation,
		auction.Gender,
		auction.World,
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
