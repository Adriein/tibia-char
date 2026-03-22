package auction

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"sort"
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
	tibiaAPI             *vendor.TibiaApi
	auctionRepository    AuctionRepository
	worldRepository      WorldRepository
	currencyRepository   currency.CurrencyRepository
	aggAuctionRepository AggAuctionStatsRepository
	mapper               *Mapper
	parserFactory        ParserFactory
	scrapperFactory      *CollyFactory
	logger               *slog.Logger
}

const AuctionDetailMaxConcurrency = 5
const AuctionLinkMaxConcurrency = 5

func NewService(
	tibiaAPI *vendor.TibiaApi,
	auctionRepo AuctionRepository,
	worldRepo WorldRepository,
	currencyRepo currency.CurrencyRepository,
	aggAuctionRepository AggAuctionStatsRepository,
	mapper *Mapper,
	parserFactory ParserFactory,
	scrapperFactory *CollyFactory,
	logger *slog.Logger,
) *Service {
	return &Service{
		tibiaAPI:             tibiaAPI,
		auctionRepository:    auctionRepo,
		worldRepository:      worldRepo,
		currencyRepository:   currencyRepo,
		aggAuctionRepository: aggAuctionRepository,
		mapper:               mapper,
		parserFactory:        parserFactory,
		scrapperFactory:      scrapperFactory,
		logger:               logger,
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
		return eris.Wrap(err, "Failed to load location Europe/Berlin")
	}

	now := time.Now().In(loc)

	if now.Hour() >= 10 && now.Hour() <= 11 {
		if err := s.ScrapBazaar(ctx); err != nil {
			return err
		}
	}

	if err := s.WatchActiveAuctions(ctx); err != nil {
		return err
	}

	if err := s.ConsolidateAuctions(ctx); err != nil {
		return err
	}

	return nil
}

/*
================================================================================
Stats Agg Logic
================================================================================
*/

func (s *Service) AggregateAuctionStatsPrecompute(ctx context.Context) error {
	s.logger.Info("Start auction stats aggregation")

	now := time.Now()

	historicAucPrices, err := s.auctionRepository.GetHistoricAuctionPrices(ctx)

	if err != nil {
		return err
	}

	priceSubsets := s.subsetPricesMap(historicAucPrices)

	stats := statistics.New()

	for key, prices := range priceSubsets {
		if len(prices) == 0 {
			continue
		}

		sort.Ints(prices)

		minPrice := prices[0]
		maxPrice := prices[len(prices)-1]

		median := stats.Median(prices)
		mean := stats.Mean(prices)

		stdDeviation := stats.StdDeviation(prices)
		mode := stats.Mode(prices)

		sampleSize := len(prices)

		agg := &AggAuctionStats{
			SubsetKey:    key,
			MinPrice:     minPrice,
			MaxPrice:     maxPrice,
			Median:       median,
			Mean:         mean,
			StdDeviation: stdDeviation,
			Mode:         mode,
			SampleSize:   sampleSize,
		}

		if err := s.aggAuctionRepository.Save(agg); err != nil {
			return err
		}
	}

	s.logger.Info("Finish auction stats aggregation", "duration", time.Since(now))

	return nil
}

/*
================================================================================
Scrapper Logic
================================================================================
*/

func (s *Service) ScrapBazaar(ctx context.Context) error {
	s.logger.Info("Start", "phase", constants.ScrapPhase)

	ctx = context.WithValue(ctx, constants.Phase, constants.ScrapPhase)

	now := time.Now()

	scrapperInstance := s.scrapperFactory.CreateScrapper("N")
	auctionNumberParser := s.parserFactory.CreateAuctionNumberParser(scrapperInstance)

	currentAuctions, err := auctionNumberParser.Scrap()

	s.logger.Info("Obtained active auctions", "phase", constants.ScrapPhase, "auctions", currentAuctions)

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

	s.logger.Info("Obtained worlds from tibiaAPI", "worlds", len(worlds))

	auctionLinkSet, err := s.scrapAuctionLinks(worlds)

	if err != nil {
		s.logger.Error("Finish with error", "phase", constants.ScrapPhase, "duration", time.Since(now))

		return err
	}

	if err := s.scrapAuctionDetails(ctx, auctionLinkSet); err != nil {
		s.logger.Error("Finish with error", "phase", constants.ScrapPhase, "duration", time.Since(now))

		return err
	}

	s.logger.Info("Finish", "phase", constants.ScrapPhase, "duration", time.Since(now), "auctions", currentAuctions, "links", len(auctionLinkSet.Data))

	return nil
}

func (s *Service) scrapAuctionLinks(worlds []*World) (*AuctionLinkSet, error) {
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

				now := time.Now()

				goroutineNum := atomic.AddInt32(&goroutineCounter, 1)
				goroutineID := fmt.Sprintf("L-routine%d", goroutineNum)

				scrapperInstance := s.scrapperFactory.CreateScrapper(goroutineID)

				linkParser := s.parserFactory.CreateAuctionListParser(scrapperInstance)

				s.logger.Info("Start world link extraction", "phase", constants.ScrapPhase, "world", world.Name, "routine_id", goroutineID)

				if err := linkParser.Scrap(world.Name, auctionLinkSet, storedAuctionLinkSet); err != nil {
					return eris.Wrapf(err, "Failed to collect link in L-routine%d", goroutineNum)
				}

				s.logger.Info("Finish world link extraction", "phase", constants.ScrapPhase, "world", world.Name, "routine_id", goroutineID, "duration", time.Since(now))

				return nil
			})
		}

		err := g.Wait()

		if err != nil {
			return nil, err
		}
	}

	s.logger.Info("Finish links extraction", "phase", constants.ScrapPhase, "duration", time.Since(now))

	return auctionLinkSet, nil
}

func (s *Service) scrapAuctionDetails(ctx context.Context, auctionLinkSet *AuctionLinkSet) error {
	pm := NewProxyManager()

	workload := pm.BalanceLoad(auctionLinkSet.Data)

	failed := NewAuctionLinkSet()

	g, gCtx := errgroup.WithContext(ctx)

	goroutineCounter := int32(0)

	phase := ctx.Value(constants.Phase)

	start := time.Now()

	for proxyAddr, work := range workload {
		goroutineNum := atomic.AddInt32(&goroutineCounter, 1)
		goroutineID := fmt.Sprintf("P-routine%d", goroutineNum)

		ctx := context.WithValue(gCtx, constants.ProxyAddr, proxyAddr)

		s.logger.Info("Balance workload with proxy", "phase", phase, "routine_id", goroutineID, "proxy_addr", proxyAddr)

		g.Go(func() error {
			s.scrapAuctionDetail(ctx, g, failed, work)

			return nil
		})
	}

	err := g.Wait()

	if err != nil {
		return err
	}

	s.logger.Info("Finish scrapping auction details", "phase", phase, "total", len(auctionLinkSet.Data), "failed", len(failed.Data), "duration", time.Since(start))

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

				start := time.Now()
				phase := ctx.Value(constants.Phase)

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

					s.logger.Error(err.Error())

					failed.Set(auctionId, linkURL)

					s.notifyStatus(ctx, totalWorkload, &workDoneCounter)

					return nil
				}

				auction, err := s.mapper.FromDTO(dto)

				if err != nil {
					s.notifyStatus(ctx, totalWorkload, &workDoneCounter)

					s.logger.Error("Failed auction detail scrapping", "phase", phase, "routine_id", goroutineID, "duration", time.Since(start))

					return eris.Wrapf(err, "Error mapping from DTO auctionId %d", auctionId)
				}

				existingAuction, err := s.auctionRepository.GetAuctionByAuctionID(ctx, auction.AuctionID)

				if err != nil {
					s.notifyStatus(ctx, totalWorkload, &workDoneCounter)

					s.logger.Error("Failed auction detail scrapping", "phase", phase, "routine_id", goroutineID, "duration", time.Since(start))

					return eris.Wrapf(err, "Error getting existing auction for auctionId %d", auctionId)
				}

				if existingAuction != nil && auction.IsEqual(existingAuction) {
					if auction.ShouldBeArchived() {
						s.auctionRepository.DeactivateAuctionRecord(ctx, auction.AuctionID)

						s.notifyStatus(ctx, totalWorkload, &workDoneCounter)

						s.logger.Info("Auction deactivated", "phase", phase, "routine_id", goroutineID, "duration", time.Since(start))

						return nil
					}

					s.auctionRepository.MarkAuctionAsUpdated(ctx, auction.AuctionID)

					s.notifyStatus(ctx, totalWorkload, &workDoneCounter)

					s.logger.Info("Skipping auction save, no significant difference detected", "phase", phase, "routine_id", goroutineID, "duration", time.Since(start))

					return nil
				}

				stats, err := s.aggAuctionRepository.GetByKey(auction.SubsetKey())

				if err != nil {
					if errors.Is(err, ErrAggAuctionStatsNotFound) {
						if err := s.auctionRepository.Save(auction); err != nil {
							s.notifyStatus(ctx, totalWorkload, &workDoneCounter)

							s.logger.Error("Failed auction detail scrapping", "phase", phase, "routine_id", goroutineID, "duration", time.Since(start))

							//TODO: all the return errors inside this go routine are being ignored
							return eris.Wrapf(err, "Error saving to DB auctionId %d", auctionId)
						}

						s.notifyStatus(ctx, totalWorkload, &workDoneCounter)

						s.logger.Error("Failed auction detail scrapping", "phase", phase, "routine_id", goroutineID, "duration", time.Since(start))

						return nil
					}

					return eris.Wrap(err, "Failed getting stats")
				}

				auction.CalculateFlags(stats)

				if err := s.auctionRepository.Save(auction); err != nil {
					s.notifyStatus(ctx, totalWorkload, &workDoneCounter)

					s.logger.Error("Failed auction detail scrapping", "phase", phase, "routine_id", goroutineID, "duration", time.Since(start))

					//TODO: all the return errors inside this go routine are being ignored
					return eris.Wrapf(err, "Error saving to DB auctionId %d", auctionId)
				}

				s.notifyStatus(ctx, totalWorkload, &workDoneCounter)

				s.logger.Info("Finish auction detail scrapping", "phase", phase, "routine_id", goroutineID, "duration", time.Since(start))

				return nil
			})
		}
	}
}

func (s *Service) notifyStatus(ctx context.Context, totalWork int, workDoneCounter *int32) {
	newValue := atomic.AddInt32(workDoneCounter, 1)

	phase := ctx.Value(constants.Phase)

	s.logger.Info(fmt.Sprintf("%d/%d", newValue, totalWork), "phase", phase)
}

/*
================================================================================
Refresh Auctions Phase
================================================================================
*/

func (s *Service) WatchActiveAuctions(ctx context.Context) error {
	const (
		FiveMin   = 5 * time.Minute
		TenMin    = 10 * time.Minute
		TwentyMin = 20 * time.Minute
		OneHour   = 60 * time.Minute
		FourHours = 4 * time.Hour
		SixHours  = 6 * time.Hour
		OneDay    = 24 * time.Hour
		OneMonth  = 30 * 24 * time.Hour
	)

	ctx = context.WithValue(ctx, constants.Phase, constants.WatchPhase)

	now := time.Now()

	s.logger.Info("Start", "phase", constants.WatchPhase)

	type refreshPolicy struct {
		finishingIn    time.Duration
		updateInterval time.Duration
	}

	policies := []refreshPolicy{
		{finishingIn: TenMin, updateInterval: FiveMin},
		{finishingIn: OneHour, updateInterval: TwentyMin},
		{finishingIn: OneDay, updateInterval: FourHours},
		{finishingIn: OneMonth, updateInterval: SixHours},
	}

	auctionsToUpdate := NewAuctionLinkSet()

	for _, policy := range policies {
		auctionAddedCounter := 0

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
			auctionAddedCounter++
		}

		s.logger.Info("Added active auctions for policy", "phase", constants.WatchPhase, "policy", policy.finishingIn.String(), "total", auctionAddedCounter)
	}

	if auctionsToUpdate.AllowLongTail() {
		auctionAddedCounter := 0

		longTailPolicy := []refreshPolicy{
			{finishingIn: OneDay, updateInterval: OneHour},
			{finishingIn: OneMonth, updateInterval: OneHour},
		}

		for _, policy := range longTailPolicy {
			auctions, err := s.auctionRepository.GetActiveAuctionsFinishingIn(ctx, policy.finishingIn)

			if err != nil {
				return err
			}

			for _, auction := range auctions {
				if !auctionsToUpdate.AllowLongTail() {
					break
				}

				if _, exists := auctionsToUpdate.Get(auction.AuctionID); exists {
					continue
				}

				if auction.DateUpd.After(time.Now().Add(-policy.updateInterval)) {
					continue
				}

				auctionsToUpdate.Set(auction.AuctionID, auction.TibiaAuctionLink)
				auctionAddedCounter++
			}
		}

		s.logger.Info("Added active auctions for long tail", "phase", constants.WatchPhase, "total", auctionAddedCounter)
	}

	if auctionsToUpdate.IsEmpty() {
		s.logger.Info("Finish phase", "phase", constants.WatchPhase, "duration", time.Since(now))

		return nil
	}

	if err := s.scrapAuctionDetails(ctx, auctionsToUpdate); err != nil {
		s.logger.Info("Finish phase", "phase", constants.WatchPhase, "duration", time.Since(now))

		return eris.Wrap(err, "Failed to refresh auctions")
	}

	s.logger.Info("Finish phase", "phase", constants.WatchPhase, "duration", time.Since(now))

	return nil
}

/*
================================================================================
Consolidate Auctions Phase
================================================================================
*/

func (s *Service) ConsolidateAuctions(ctx context.Context) error {
	ctx = context.WithValue(ctx, constants.Phase, constants.ConsolidatePhase)

	s.logger.Info("Start", "phase", constants.ConsolidatePhase)

	now := time.Now()

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

	if auctionsToUpdate.IsEmpty() {
		s.logger.Info("Finish phase", "phase", constants.ConsolidatePhase, "duration", time.Since(now))

		return nil
	}

	if err := s.scrapAuctionDetails(ctx, auctionsToUpdate); err != nil {
		s.logger.Info("Finish phase", "phase", constants.ConsolidatePhase, "duration", time.Since(now))

		return eris.Wrap(err, "Failed to consolidate auctions")
	}

	s.logger.Info("Finish phase", "phase", constants.ConsolidatePhase, "duration", time.Since(now))

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

	var viewModels []*AuctionViewModel

	for _, auction := range auctions {
		auction.BidFiat = conRate.Exchange(auction.BidFiat, targetCurrency)
		auction.BidCurrency = targetCurrency

		stats, err := s.aggAuctionRepository.GetByKey(auction.SubsetKey())

		if err != nil {
			if errors.Is(err, ErrAggAuctionStatsNotFound) {
				viewModels = append(viewModels, &AuctionViewModel{
					Auction: auction,
					ZScore:  0,
				})

				continue
			}
			//TODO: decide what to do because this will cause 0 auctions to the frontend maybe we can be more optimistic
			return nil, err
		}

		if auction.CharVocation.Id == constants.VocationNone {
			viewModels = append(viewModels, &AuctionViewModel{
				Auction: auction,
				ZScore:  0,
			})

			continue
		}

		ZScore := float64(auction.Bid-int(stats.Median)) / stats.StdDeviation

		viewModels = append(viewModels, &AuctionViewModel{
			Auction: auction,
			ZScore:  ZScore,
		})
	}

	totalCount, err := s.auctionRepository.CountActiveAuctions(ctx)

	if err != nil {
		return nil, err
	}

	totalPages := totalCount / filter.Pagination.Limit

	result := &PaginatedAuctions{
		ViewModels: viewModels,
		TotalCount: totalCount,
		PageSize:   filter.Pagination.Limit,
		Page:       filter.Pagination.Page + 1,
		TotalPages: totalPages,
	}

	return result, nil
}

func (s *Service) subsetPricesMap(auctions []*Auction) PriceSubsets {
	priceSubsets := make(PriceSubsets)

	for _, auction := range auctions {
		key := auction.SubsetKey()

		priceSubsets[key] = append(priceSubsets[key], auction.Bid)
	}

	return priceSubsets
}
