package auction

import (
	"context"
	"log"
	"math/rand"
	"time"

	"github.com/adriein/tibia-char/internal/currency"
	"github.com/adriein/tibia-char/pkg/enums"
	"github.com/adriein/tibia-char/pkg/helper/collections"
	"github.com/adriein/tibia-char/pkg/middleware"
	"github.com/adriein/tibia-char/pkg/vendor"
	"github.com/rotisserie/eris"
	"golang.org/x/sync/errgroup"
)

type Service struct {
	tibiaAPI           *vendor.TibiaApi
	linkParser         *AuctionListHtmlParser
	auctionParser      *AuctionHtmlParser
	auctionRepository  AuctionRepository
	worldRepository    WorldRepository
	currencyRepository currency.CurrencyRepository
	mapper             *Mapper
	logger             *log.Logger
}

const AuctionDetailMaxConcurrency = 5
const AuctionLinkMaxConcurrency = 5

func NewService(
	ta *vendor.TibiaApi,
	lp *AuctionListHtmlParser,
	ap *AuctionHtmlParser,
	ar AuctionRepository,
	wr WorldRepository,
	cr currency.CurrencyRepository,
	m *Mapper,
	logger *log.Logger,
) *Service {
	return &Service{
		tibiaAPI:           ta,
		linkParser:         lp,
		auctionParser:      ap,
		auctionRepository:  ar,
		worldRepository:    wr,
		currencyRepository: cr,
		mapper:             m,
		logger:             logger,
	}
}

func (s *Service) ScrapBazaar(ctx context.Context) error {
	traceID := ctx.Value(middleware.TraceIDKey)

	s.logger.Printf("TraceID: %s Start Scrap Bazaar\n", traceID)

	now := time.Now()

	currentAuctions, err := s.linkParser.GetTotalCurrentAuctions()

	if err != nil {
		return err
	}

	/*worlds, err := s.tibiaAPI.GetWorlds()

	if err != nil {
		return eris.Wrap(err, "Failed to fetch worlds from Tibia API")
	}*/

	worlds := []*World{{Id: 1, Name: "Calmera", Location: "North America", BattleEye: enums.BattleEyeYellow, Pvp: "Optional Pvp"}}

	for _, world := range worlds {
		_, err := s.worldRepository.GetOrCreate(world)

		if err != nil {
			return err
		}
	}

	auctionLinkSet, err := s.scrapAuctionLinks(worlds)

	if err != nil {
		s.logger.Printf("TraceID: %s Finished Scrapping with error in %s\n", traceID, time.Since(now))

		return err
	}

	s.logger.Printf("TraceID: %s Total auction links %d - Auction links obtained %d\n", traceID, currentAuctions, len(auctionLinkSet.Data))

	if err := s.scrapAuctionDetails(auctionLinkSet); err != nil {
		s.logger.Printf("TraceID: %s Finished Scrapping with error: %s in %s\n", traceID, err.Error(), time.Since(now))

		return err
	}

	s.logger.Printf("TraceID: %s Finished Scrapping in %s\n", traceID, time.Since(now))

	return nil
}

func (s *Service) GetAuctions(ctx context.Context) ([]*Auction, error) {
	auctions, err := s.auctionRepository.GetActiveAuctions(ctx)

	if err != nil {
		return nil, err
	}

	loc, ok := ctx.Value(middleware.TimezoneKey).(*time.Location)

	if !ok {
		return nil, eris.New("Error, location not stored in ctx")
	}

	targetCurrency := currency.FromLocation(loc)

	conRate, err := s.currencyRepository.GetLatest(ctx)

	if err != nil {
		return nil, err
	}

	for _, auction := range auctions {
		auction.BidFiat = conRate.Exchange(auction.BidFiat, targetCurrency)
		auction.BidCurrency = targetCurrency
	}

	return auctions, nil
}

func (s *Service) scrapAuctionLinks(worlds []*World) (*AuctionLinkSet, error) {
	semaphore := make(chan struct{}, AuctionLinkMaxConcurrency)

	worldsChunk := collections.Chunk(worlds, AuctionLinkMaxConcurrency)

	auctionLinkSet := NewAuctionLinkSet()

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

				if err := s.linkParser.ScrapeWorld(world.Name, auctionLinkSet); err != nil {
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

func (s *Service) scrapAuctionDetails(auctionLinkSet *AuctionLinkSet) error {
	pm := NewProxyManager()

	workload := pm.BalanceLoad(auctionLinkSet.Data)

	failed := NewAuctionLinkSet()

	g, gCtx := errgroup.WithContext(context.Background())

	for proxyAddr, work := range workload {
		ctx := context.WithValue(gCtx, "Addr", proxyAddr)

		g.Go(func() error {
			s.scrapAuctionDetail(ctx, g, failed, work)

			return nil
		})
	}

	err := g.Wait()

	if err != nil {
		return err
	}

	/*links := collections.ChunkMap(auctionLinkSet.Data, AuctionDetailMaxConcurrency)

	semaphore := make(chan struct{}, AuctionDetailMaxConcurrency)

	failed := NewAuctionLinkSet()
	scrapped := 0

	for i, chunk := range links {
		g, ctx := errgroup.WithContext(context.Background())

		if i != 0 {
			randDelay := time.Duration(2+rand.Intn(5)) * time.Second

			time.Sleep(randDelay)
		}

		for _, kv := range chunk {
			scrapped++
			g.Go(func() error {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case semaphore <- struct{}{}:
					defer func() { <-semaphore }()
				}

				auctionId := kv.Key
				linkURL := kv.Value

				dto, err := s.auctionParser.Parse(ctx, auctionId, linkURL)

				if err != nil {
					if eris.Is(err, RateLimitError) {
						return eris.Wrapf(err, "Rate limit reached parsing auctionId %d", auctionId)
					}

					failed.Set(auctionId, linkURL)

					return nil
				}

				auction, err := s.mapper.FromDTO(dto)

				if err != nil {
					return eris.Wrapf(err, "Error mapping from DTO auctionId %d", auctionId)
				}

				if err := s.auctionRepository.Save(auction); err != nil {
					return eris.Wrapf(err, "Error saving to DB auctionId %d", auctionId)
				}

				return nil
			})
		}

		err := g.Wait()

		if err != nil {
			return err
		}
	}*/

	s.logger.Printf("Total to scrap: %d, Failed: %d", len(auctionLinkSet.Data), len(failed.Data))

	return nil
}

func (s *Service) scrapAuctionDetail(ctx context.Context, g *errgroup.Group, failed *AuctionLinkSet, workGroup []collections.KeyValue[int, string]) {
	semaphore := make(chan struct{}, AuctionDetailMaxConcurrency)

	chunks := collections.Chunk(workGroup, AuctionDetailMaxConcurrency)

	for i, chunk := range chunks {
		if i != 0 {
			randDelay := time.Duration(2+rand.Intn(5)) * time.Second

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

				dto, err := s.auctionParser.Parse(ctx, auctionId, linkURL)

				if err != nil {
					if eris.Is(err, RateLimitError) {
						return eris.Wrapf(err, "Rate limit reached parsing auctionId %d", auctionId)
					}

					failed.Set(auctionId, linkURL)

					return nil
				}

				auction, err := s.mapper.FromDTO(dto)

				if err != nil {
					return eris.Wrapf(err, "Error mapping from DTO auctionId %d", auctionId)
				}

				if err := s.auctionRepository.Save(auction); err != nil {
					return eris.Wrapf(err, "Error saving to DB auctionId %d", auctionId)
				}

				return nil
			})
		}
	}

}
