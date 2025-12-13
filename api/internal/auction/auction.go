package auction

import "time"

type Auction struct {
	Id             int
	TibiaAuctionId int
	Img            string
	Name           string
	Level          int
	Vocation       string
	Gender         string
	World          string
	Bid            int
	AuctionStart   time.Time
	AuctionEnd     time.Time
	IsActive       bool
	DateAdd        time.Time
	DateUpd        time.Time
}
