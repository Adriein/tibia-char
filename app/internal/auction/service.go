package auction

import (
	"context"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/adriein/tibia-char/internal/auction/model"
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

func (s *Service) ScrapBazaar(ctx context.Context) error {
	traceID := ctx.Value("traceID")

	s.logger.Printf("TraceID: %s Start Scrap Bazaar\n", traceID)

	now := time.Now()

	currentAuctions, err := s.linkParser.GetTotalCurrentAuctions()

	if err != nil {
		return err
	}

	auctionLinkSet, err := s.linkParser.GetLinks()

	s.logger.Printf("TraceID: %s Current auctions %d - Scrapped Auctions %d\n", traceID, currentAuctions, len(auctionLinkSet))

	if err != nil {
		return err
	}

	const MaxConcurrency = 5

	links := array.ChunkMap(auctionLinkSet, MaxConcurrency)

	var wg sync.WaitGroup

	maxWorkers := make(chan struct{}, MaxConcurrency)

	for i, chunk := range links {
		if i != 0 {
			randDelay := time.Duration(1+rand.Intn(5)) * time.Second

			time.Sleep(randDelay)
		}

		for _, kv := range chunk {
			maxWorkers <- struct{}{}

			wg.Add(1)

			id := kv.Key
			link := kv.Value

			go func(auctionId int, linkURL string) {
				defer wg.Done()

				defer func() { <-maxWorkers }()

				dto, err := s.auctionParser.Parse(auctionId, linkURL)

				if err != nil {
					s.logger.Printf("TraceID: %s Parsing of auction id: %d failed with: %s\n", traceID, auctionId, err.Error())
				}

				auction, err := s.mapper.ToDomain(dto)

				if err != nil {
					s.logger.Printf("TraceID: %s Error mapping auction dto to auction for auction id: %d: %v\n", traceID, auctionId, err)
				}

				if err := s.auctionRepository.Save(auction); err != nil {
					s.logger.Printf("TraceID: %s Error saving auction: %d: %s\n", traceID, auctionId, err.Error())
				}

			}(id, link)
		}

		wg.Wait()
	}

	s.logger.Printf("TraceID: %s Finished Scrapping in %s\n", traceID, time.Since(now))

	return nil
}

func (s *Service) GetAuctions(ctx context.Context) ([]*model.Auction, error) {
	auctions, err := s.auctionRepository.GetActiveAuctions(ctx)

	if err != nil {
		return nil, err
	}

	return auctions, nil
}
