package auction

import (
	"log"
	"time"

	"github.com/adriein/tibia-char/pkg/constants"
	"github.com/gocolly/colly/v2"
)

type Service struct {
	logger             *log.Logger
	auctionRepository  AuctionRepository
	vocationRepository VocationRepository
	genderRepository   GenderRepository
	worldRepository    WorldRepository
}

func NewService(logger *log.Logger, ar AuctionRepository, vr VocationRepository, gr GenderRepository, wr WorldRepository) *Service {
	return &Service{
		logger:             logger,
		auctionRepository:  ar,
		vocationRepository: vr,
		genderRepository:   gr,
		worldRepository:    wr,
	}
}

func (s *Service) ScrapBazaar() error {
	s.logger.Println("Start Scrap Bazaar")

	now := time.Now()

	linkScrapper := NewScrapper("CollectAuctionLinks")

	linkScrapper.Collector.Limit(&colly.LimitRule{
		DomainGlob:  constants.TibiaOfficialWebsite,
		RandomDelay: 5 * time.Second,
	})

	detailScrapper := NewScrapper("CollectAuctionDetails")

	detailScrapper.Collector.Limit(&colly.LimitRule{
		DomainGlob: constants.TibiaOfficialWebsite,
	})

	auctions := Auctions{}

	auctions.Scrap(s.auctionRepository, s.vocationRepository, s.genderRepository, s.worldRepository, linkScrapper, detailScrapper, s.logger)

	s.logger.Printf("Finished Scrapping in %s", time.Since(now))

	return nil
}
