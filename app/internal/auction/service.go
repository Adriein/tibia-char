package auction

import (
	"context"
	"log"
	"math/rand"
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

const AuctionDetailMaxConcurrency = 5
const AuctionLinkMaxConcurrency = 10

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

	auctionLinkSet, err := s.scrapAuctionLinks(worlds)

	s.logger.Printf("TraceID: %s Total auction links %d - Auction links obtained %d\n", traceID, currentAuctions, len(auctionLinkSet.Data))

	if err != nil {
		s.logger.Printf("TraceID: %s Finished Scrapping with error in %s\n", traceID, time.Since(now))

		return err
	}

	if err := s.scrapAuctionDetail(auctionLinkSet); err != nil {
		s.logger.Printf("TraceID: %s Finished Scrapping with error in %s\n", traceID, time.Since(now))

		return err
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

func (s *Service) scrapAuctionLinks(worlds []string) (*model.AuctionLinkSet, error) {
	semaphore := make(chan struct{}, AuctionLinkMaxConcurrency)

	worldsChunk := array.Chunk(worlds, AuctionLinkMaxConcurrency)

	auctionLinkSet := model.NewAuctionLinkSet()

	for _, chunk := range worldsChunk {
		g, ctx := errgroup.WithContext(context.Background())

		for _, world := range chunk {
			g.Go(func() error {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case semaphore <- struct{}{}:
					defer func() { <-semaphore }()
				}

				if err := s.linkParser.ScrapeWorld(world, auctionLinkSet); err != nil {
					return err
				}
				return nil
			})
		}

		err := g.Wait()

		if err != nil {
			return nil, err
		}
	}

	return auctionLinkSet, nil
}

func (s *Service) scrapAuctionDetail(auctionLinkSet *model.AuctionLinkSet) error {
	links := array.ChunkMap(auctionLinkSet.Data, AuctionDetailMaxConcurrency)

	semaphore := make(chan struct{}, AuctionDetailMaxConcurrency)

	for i, chunk := range links {
		g, ctx := errgroup.WithContext(context.Background())

		if i != 0 {
			randDelay := time.Duration(1+rand.Intn(5)) * time.Second

			time.Sleep(randDelay)
		}

		for _, kv := range chunk {
			g.Go(func() error {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case semaphore <- struct{}{}:
					defer func() { <-semaphore }()
				}

				auctionId := kv.Key
				linkURL := kv.Value

				dto, err := s.auctionParser.Parse(auctionId, linkURL)

				if err != nil {
					return err
				}

				auction, err := s.mapper.FromDTO(dto)

				if err != nil {
					return err
				}

				if err := s.auctionRepository.Save(auction); err != nil {
					return err
				}

				return nil
			})
		}

		g.Wait()
	}

	return nil
}
