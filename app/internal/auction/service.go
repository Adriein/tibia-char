package auction

import (
	"context"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/adriein/tibia-char/internal/auction/model"
	"github.com/adriein/tibia-char/pkg/helper/array"
	"github.com/adriein/tibia-char/pkg/vendor"
	"github.com/rotisserie/eris"
	"golang.org/x/sync/errgroup"
)

type Service struct {
	tibiaAPI          *vendor.TibiaApi
	linkParser        *AuctionListHtmlParser
	auctionParser     *AuctionHtmlParser
	auctionRepository AuctionRepository
	worldRepository   WorldRepository
	mapper            *Mapper
	logger            *log.Logger
}

func NewService(ta *vendor.TibiaApi, lp *AuctionListHtmlParser, ap *AuctionHtmlParser, ar AuctionRepository, wr WorldRepository, m *Mapper, logger *log.Logger) *Service {
	return &Service{
		tibiaAPI:          ta,
		linkParser:        lp,
		auctionParser:     ap,
		auctionRepository: ar,
		worldRepository:   wr,
		mapper:            m,
		logger:            logger,
	}
}

func (s *Service) ScrapBazaar(ctx context.Context) error {
	const MaxConcurrency = 5
	const LinkMaxConcurrency = 5

	traceID := ctx.Value("traceID")

	s.logger.Printf("TraceID: %s Start Scrap Bazaar\n", traceID)

	now := time.Now()

	currentAuctions, err := s.linkParser.GetTotalCurrentAuctions()

	if err != nil {
		return err
	}

	worlds, err := s.tibiaAPI.GetWorlds()

	if err != nil {
		return eris.Wrap(err, "Failed to fetch worlds from Tibia API")
	}

	for _, world := range worlds {
		_, err := s.worldRepository.GetOrCreate(world)

		if err != nil {
			return err
		}
	}

	semaphore := make(chan struct{}, LinkMaxConcurrency)

	worldsChunk := array.Chunk(worlds, LinkMaxConcurrency)

	auctionLinkSet := model.NewAuctionLinkSet()

	for i, chunk := range worldsChunk {
		g, ctx := errgroup.WithContext(context.Background())

		if i != 0 {
			randDelay := time.Duration(1+rand.Intn(10)) * time.Second

			time.Sleep(randDelay)
		}

		for _, world := range chunk {
			g.Go(func() error {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case semaphore <- struct{}{}:
					defer func() { <-semaphore }()
				}

				if err := s.linkParser.ScrapeWorld(world, auctionLinkSet); err != nil {
					s.logger.Printf("Error for world %s: %v\n", world, err)

					return err
				}
				return nil
			})

		}

		err := g.Wait()

		if err != nil {
			return err
		}
	}

	s.logger.Printf("TraceID: %s Current auctions %d - Scrapped Auctions %d\n", traceID, currentAuctions, len(auctionLinkSet.Data))

	links := array.ChunkMap(auctionLinkSet.Data, MaxConcurrency)

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
					s.logger.Printf("TraceID: %s Parsing of auction id: %d failed with: %s\n", traceID, auctionId, eris.ToString(err, true))
					return
				}

				auction, err := s.mapper.FromDTO(dto)

				if err != nil {
					s.logger.Printf("TraceID: %s Error mapping auction dto to auction for auction id: %d: %s\n", traceID, auctionId, eris.ToString(err, true))
					return
				}

				if err := s.auctionRepository.Save(auction); err != nil {
					s.logger.Printf("TraceID: %s Error saving auction: %d: %s\n", traceID, auctionId, eris.ToString(err, true))
					return
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
