package auction

import (
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/adriein/tibia-char/pkg/helper/array"
)

type Service struct {
	linkParser        *AuctionListHtmlParser
	auctionParser     *AuctionHtmlParser
	auctionRepository AuctionRepository
	worldRepository   WorldRepository
	mapper            *Mapper
	logger            *log.Logger
}

func NewService(lp *AuctionListHtmlParser, ap *AuctionHtmlParser, ar AuctionRepository, wr WorldRepository, m *Mapper, logger *log.Logger) *Service {
	return &Service{
		linkParser:        lp,
		auctionParser:     ap,
		auctionRepository: ar,
		worldRepository:   wr,
		mapper:            m,
		logger:            logger,
	}
}

func (s *Service) ScrapBazaar() error {
	s.logger.Println("Start Scrap Bazaar")

	now := time.Now()

	currentAuctions, err := s.linkParser.GetTotalCurrentAuctions()

	if err != nil {
		return err
	}

	auctionLinkSet, err := s.linkParser.GetLinks()

	s.logger.Printf("Current auctions %d - Scrapped Auctions %d", currentAuctions, len(auctionLinkSet))

	if err != nil {
		return err
	}

	const MaxConcurrency = 5

	links := array.Chunk(auctionLinkSet.Values(), MaxConcurrency)

	var wg sync.WaitGroup

	maxWorkers := make(chan struct{}, MaxConcurrency)

	for i, chunk := range links {
		if i != 0 {
			randDelay := time.Duration(1+rand.Intn(5)) * time.Second

			time.Sleep(randDelay)
		}

		for auctionId, link := range chunk {
			maxWorkers <- struct{}{}

			wg.Add(1)

			go func(url string) {
				defer wg.Done()

				defer func() { <-maxWorkers }()

				dto, err := s.auctionParser.Parse(auctionId, link)

				if err != nil {
					s.logger.Printf("Parsing of auction id: %d failed with: %s\n", auctionId, err.Error())
				}

				auction, err := s.mapper.ToDomain(dto)

				if err != nil {
					s.logger.Printf("Error mapping auction dto to auction for auction id: %d: %v\n", auctionId, err)
				}

				if err := s.auctionRepository.Save(auction); err != nil {
					s.logger.Printf("Error saving auction: %d: %s\n", auctionId, err.Error())
				}

			}(link)
		}

		wg.Wait()
	}

	s.logger.Printf("Finished Scrapping in %s", time.Since(now))

	return nil
}
