package auction

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/adriein/tibia-char/internal/currency"
	"github.com/adriein/tibia-char/pkg/constants"
	"github.com/adriein/tibia-char/pkg/enums"
	"github.com/adriein/tibia-char/pkg/helper/collections"
	"github.com/adriein/tibia-char/pkg/helper/statistics"
	"github.com/adriein/tibia-char/pkg/middleware"
	"github.com/adriein/tibia-char/pkg/vendor"
	"github.com/rotisserie/eris"
	"golang.org/x/sync/errgroup"
)

type Service struct {
	tibiaAPI           *vendor.TibiaApi
	auctionRepository  AuctionRepository
	worldRepository    WorldRepository
	currencyRepository currency.CurrencyRepository
	mapper             *Mapper
	parserFactory      ParserFactory
	scrapperFactory    *CollyFactory
	logger             *log.Logger
}

const AuctionDetailMaxConcurrency = 5
const AuctionLinkMaxConcurrency = 5

func NewService(
	tibiaAPI *vendor.TibiaApi,
	auctionRepo AuctionRepository,
	worldRepo WorldRepository,
	currencyRepo currency.CurrencyRepository,
	mapper *Mapper,
	parserFactory ParserFactory,
	scrapperFactory *CollyFactory,
	logger *log.Logger,
) *Service {
	return &Service{
		tibiaAPI:           tibiaAPI,
		auctionRepository:  auctionRepo,
		worldRepository:    worldRepo,
		currencyRepository: currencyRepo,
		mapper:             mapper,
		parserFactory:      parserFactory,
		scrapperFactory:    scrapperFactory,
		logger:             logger,
	}
}

/*
================================================================================
Scrapper Orchestrator
================================================================================
*/

func (s *Service) ScrapperOrchestrator(ctx context.Context) error {
	loc, err := time.LoadLocation("Europe/Berlin")

	if err != nil {
		s.logger.Printf("Failed to load location Europe/Berlin: %v", err)

		return err
	}

	now := time.Now().In(loc)

	if now.Hour() >= 10 && now.Hour() <= 11 {
		if err := s.ScrapBazaar(ctx); err != nil {
			return err
		}
	}

	if err := s.RefreshActiveAuctions(ctx); err != nil {
		return err
	}

	if err := s.ConsolidateAuctions(ctx); err != nil {
		return err
	}

	return nil
}

/*
================================================================================
Scrapper Logic
================================================================================
*/

func (s *Service) ScrapBazaar(ctx context.Context) error {
	traceID := ctx.Value(middleware.TraceIDKey)

	s.logger.Printf("TraceID: %s Start Scrap Phase\n", traceID)

	now := time.Now()

	scrapperInstance := s.scrapperFactory.CreateScrapper("N")
	auctionNumberParser := s.parserFactory.CreateAuctionNumberParser(scrapperInstance)

	currentAuctions, err := auctionNumberParser.Scrap()

	if err != nil {
		return err
	}

	worldDTO, err := s.tibiaAPI.GetWorlds()

	if err != nil {
		return eris.Wrap(err, "Failed to fetch worlds from Tibia API")
	}

	var worlds []*World

	for i, dto := range worldDTO.Worlds.RegularWorlds {
		var battleEye enums.BattleEye

		if dto.BattleEyeDate != "release" {
			battleEye = enums.BattleEyeYellow
		} else {
			battleEye = enums.BattleEyeGreen
		}

		worlds = append(worlds, &World{
			Id:        i + 1,
			Name:      dto.Name,
			Location:  dto.Location,
			BattleEye: battleEye,
			Pvp:       dto.PvpType,
		})
	}

	//worlds = []*World{{Id: 1, Name: "Calmera", Location: "North America", BattleEye: enums.BattleEyeYellow, Pvp: "Optional Pvp"}}

	for _, world := range worlds {
		_, err := s.worldRepository.GetOrCreate(world)

		if err != nil {
			return err
		}
	}

	auctionLinkSet, err := s.scrapAuctionLinks(ctx, worlds)

	if err != nil {
		s.logger.Printf("TraceID: %s Finished Scrap Phase with error time: %s\n", traceID, time.Since(now))

		return err
	}

	if err := s.scrapAuctionDetails(auctionLinkSet); err != nil {
		s.logger.Printf("TraceID: %s Finished Scrap Phase with error time: %s\n", traceID, time.Since(now))

		return err
	}

	s.logger.Printf("TraceID: %s Finish Scrap Phase - Links: %d/%d - Time: %s\n", traceID, currentAuctions, len(auctionLinkSet.Data), time.Since(now))

	return nil
}

func (s *Service) scrapAuctionLinks(ctx context.Context, worlds []*World) (*AuctionLinkSet, error) {
	traceID := ctx.Value(middleware.TraceIDKey)

	now := time.Now()

	semaphore := make(chan struct{}, AuctionLinkMaxConcurrency)

	worldsChunk := collections.Chunk(worlds, AuctionLinkMaxConcurrency)

	auctionLinkSet := NewAuctionLinkSet()
	storedAuctionLinkSet := NewAuctionLinkSet()

	auctions, err := s.auctionRepository.GetActiveAuctions(context.Background())

	if err != nil {
		return nil, eris.Wrap(err, "Error getting active auctions")
	}

	for _, auction := range auctions {
		storedAuctionLinkSet.Set(auction.AuctionID, auction.TibiaAuctionLink)
	}

	for _, chunk := range worldsChunk {
		goroutineCounter := int32(0)

		g, ctx := errgroup.WithContext(context.Background())

		for _, world := range chunk {
			g.Go(func() error {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case semaphore <- struct{}{}:
					defer func() { <-semaphore }()
				}

				goroutineNum := atomic.AddInt32(&goroutineCounter, 1)
				goroutineID := fmt.Sprintf("L-routine%d", goroutineNum)

				scrapperInstance := s.scrapperFactory.CreateScrapper(goroutineID)

				linkParser := s.parserFactory.CreateAuctionListParser(scrapperInstance)

				if err := linkParser.Scrap(world.Name, auctionLinkSet, storedAuctionLinkSet); err != nil {
					return eris.Wrapf(err, "Failed to collect link in L-routine%d", goroutineNum)
				}
				return nil
			})
		}

		err := g.Wait()

		if err != nil {
			return nil, err
		}
	}

	s.logger.Printf("TraceID: %s Finished Scrapping with error time: %s\n", traceID, time.Since(now))

	return auctionLinkSet, nil
}

func (s *Service) scrapAuctionDetails(auctionLinkSet *AuctionLinkSet) error {
	pm := NewProxyManager()

	workload := pm.BalanceLoad(auctionLinkSet.Data)

	failed := NewAuctionLinkSet()

	g, gCtx := errgroup.WithContext(context.Background())

	for proxyAddr, work := range workload {
		ctx := context.WithValue(gCtx, constants.ProxyAddr, proxyAddr)

		g.Go(func() error {
			s.scrapAuctionDetail(ctx, g, failed, work)

			return nil
		})
	}

	err := g.Wait()

	if err != nil {
		return err
	}

	s.logger.Printf("Total to scrap: %d, Failed: %d", len(auctionLinkSet.Data), len(failed.Data))

	return nil
}

func (s *Service) scrapAuctionDetail(ctx context.Context, g *errgroup.Group, failed *AuctionLinkSet, workGroup []collections.KeyValue[int, string]) {
	semaphore := make(chan struct{}, AuctionDetailMaxConcurrency)

	chunks := collections.Chunk(workGroup, AuctionDetailMaxConcurrency)

	workDoneCounter := int32(0)
	totalWorkload := len(workGroup)

	for i, chunk := range chunks {
		if i != 0 {
			randDelay := time.Duration(2+rand.Intn(5)) * time.Second

			time.Sleep(randDelay)
		}

		goroutineCounter := int32(0)

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

				goroutineNum := atomic.AddInt32(&goroutineCounter, 1)
				goroutineID := fmt.Sprintf("D-routine%d", goroutineNum)

				scrapperInstance := s.scrapperFactory.CreateScrapper(goroutineID)
				auctionParser := s.parserFactory.CreateAuctionParser(scrapperInstance)

				dto, err := auctionParser.Parse(ctx, auctionId, linkURL)

				if err != nil {
					if eris.Is(err, RateLimitError) {
						return eris.Wrapf(err, "Rate limit reached parsing auctionId %d", auctionId)
					}

					s.logger.Println(err.Error())

					failed.Set(auctionId, linkURL)

					s.notifyStatus(totalWorkload, &workDoneCounter)

					return nil
				}

				auction, err := s.mapper.FromDTO(dto)

				if err != nil {
					s.notifyStatus(totalWorkload, &workDoneCounter)

					return eris.Wrapf(err, "Error mapping from DTO auctionId %d", auctionId)
				}

				if err := s.auctionRepository.Save(auction); err != nil {
					s.notifyStatus(totalWorkload, &workDoneCounter)
					//TODO: all the return errors inside this go routine are being ignored
					return eris.Wrapf(err, "Error saving to DB auctionId %d", auctionId)
				}

				s.notifyStatus(totalWorkload, &workDoneCounter)

				return nil
			})
		}
	}
}

func (s *Service) notifyStatus(totalWork int, workDoneCounter *int32) {
	newValue := atomic.AddInt32(workDoneCounter, 1)
	s.logger.Printf("%d/%d", newValue, totalWork)
}

/*
================================================================================
Refresh Auctions Phase
================================================================================
*/

func (s *Service) RefreshActiveAuctions(ctx context.Context) error {
	traceID := ctx.Value(middleware.TraceIDKey)

	now := time.Now()

	s.logger.Printf("TraceID: %s Start refresh phase\n", traceID)

	type refreshPolicy struct {
		finishingIn    time.Duration
		updateInterval time.Duration
	}

	policies := []refreshPolicy{
		{finishingIn: 10 * time.Minute, updateInterval: 5 * time.Minute},
		{finishingIn: 60 * time.Minute, updateInterval: 20 * time.Minute},
		{finishingIn: 24 * time.Hour, updateInterval: 4 * time.Hour},
		{finishingIn: 30 * 24 * time.Hour, updateInterval: 6 * time.Hour},
	}

	auctionsToUpdate := NewAuctionLinkSet()

	for _, policy := range policies {
		auctions, err := s.auctionRepository.GetActiveAuctionsFinishingIn(ctx, policy.finishingIn)

		if err != nil {
			return err
		}

		for _, auction := range auctions {
			if _, exists := auctionsToUpdate.Get(auction.AuctionID); exists {
				continue
			}

			if auction.DateUpd.After(time.Now().Add(-policy.updateInterval)) {
				continue
			}

			auctionsToUpdate.Set(auction.AuctionID, auction.TibiaAuctionLink)
		}
	}

	if err := s.scrapAuctionDetails(auctionsToUpdate); err != nil {
		return eris.Wrap(err, "Failed to refresh auctions")
	}

	s.logger.Printf("TraceID: %s Finish refresh phase in: %s\n", traceID, time.Since(now))

	return nil
}

/*
================================================================================
Consolidate Auctions Phase
================================================================================
*/

func (s *Service) ConsolidateAuctions(ctx context.Context) error {
	traceID := ctx.Value(middleware.TraceIDKey)

	now := time.Now()

	s.logger.Printf("TraceID: %s Start consolidate phase\n", traceID)

	auctionsToUpdate := NewAuctionLinkSet()

	auctions, err := s.auctionRepository.GetAuctionsPendingToConsolidate(ctx)

	if err != nil {
		return err
	}

	for _, auction := range auctions {
		if _, exists := auctionsToUpdate.Get(auction.AuctionID); exists {
			continue
		}

		auctionsToUpdate.Set(auction.AuctionID, auction.TibiaAuctionLink)
	}

	if err := s.scrapAuctionDetails(auctionsToUpdate); err != nil {
		return eris.Wrap(err, "Failed to consolidate auctions")
	}

	s.logger.Printf("TraceID: %s Finish consolidate phase in: %s\n", traceID, time.Since(now))

	return nil
}

/*
================================================================================
Get Auctions Logic
================================================================================
*/

func (s *Service) GetAuctions(ctx context.Context, filter *AuctionFilter) (*PaginatedAuctions, error) {
	if filter == nil {
		filter = DefaultAuctionFilter()
	}

	auctions, err := s.auctionRepository.GetAuctionsWithFilter(ctx, filter)

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

	historicAucPrices, err := s.auctionRepository.GetHistoricAuctionPrices(ctx)

	stdDevSubsets := s.calculateStdDeviationForPriceSubset(historicAucPrices)

	var viewModels []*AuctionViewModel

	for _, auction := range auctions {
		auction.BidFiat = conRate.Exchange(auction.BidFiat, targetCurrency)
		auction.BidCurrency = targetCurrency

		stdDev := stdDevSubsets[auction.SubsetKey()]

		fmt.Printf("%s, %.2f", auction.SubsetKey(), stdDev)

		viewModels = append(viewModels, &AuctionViewModel{
			Auction:      auction,
			StdDeviation: stdDev,
		})
	}

	totalCount, err := s.auctionRepository.CountActiveAuctions(ctx)

	if err != nil {
		return nil, err
	}

	totalPages := totalCount / filter.Limit

	result := &PaginatedAuctions{
		ViewModels: viewModels,
		TotalCount: totalCount,
		PageSize:   filter.Limit,
		Page:       filter.Page + 1,
		TotalPages: totalPages,
	}

	return result, nil
}

func (s *Service) calculateStdDeviationForPriceSubset(auctions []*Auction) StdDeviationSubsets {
	stdDevSubset := make(StdDeviationSubsets)

	priceSubsets := make(map[string][]int)

	for _, auction := range auctions {
		key := auction.SubsetKey()

		priceSubsets[key] = append(priceSubsets[key], auction.Bid)
	}

	for key, prices := range priceSubsets {
		stdDevSubset[key] = statistics.New().StdDeviation(prices)
	}

	return stdDevSubset
}
