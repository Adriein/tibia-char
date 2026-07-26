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
	"github.com/adriein/tibia-char/pkg/helper/beautify"
	"github.com/adriein/tibia-char/pkg/helper/collections"
	"github.com/adriein/tibia-char/pkg/helper/statistics"
	"github.com/adriein/tibia-char/pkg/middleware"
	"github.com/adriein/tibia-char/pkg/vendor"
	"github.com/rotisserie/eris"
	"golang.org/x/sync/errgroup"
)

type AuctionService interface {
	ScrapperOrchestrator(ctx context.Context) error
	GetAuctions(ctx context.Context, filter *AuctionFilter) (*PaginatedAuctions, error)
	AggregateAuctionStatsPrecompute(ctx context.Context) error
	WatchActiveAuctions(ctx context.Context) error
	ConsolidateAuctions(ctx context.Context) error
	BackfillAuctionFlags(ctx context.Context) error
}

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

func (s *Service) BackfillAuctionFlags(ctx context.Context) error {
	var auctions []*Auction

	historical, err := s.auctionRepository.GetHistoricAuctionPrices(ctx)

	if err != nil {
		return err
	}

	auctions = append(auctions, historical...)

	active, err := s.auctionRepository.GetActiveAuctions(ctx)

	if err != nil {
		return err
	}

	auctions = append(auctions, active...)

	traceID := ctx.Value(middleware.TraceIDKey)
	total := len(auctions)

	s.logger.Info("Backfilling", "trace_id", traceID, "n_auc", total)

	for i, auction := range auctions {

		stats, err := s.aggAuctionRepository.GetByKey(auction.SubsetKey())

		if err != nil {
			if errors.Is(err, ErrAggAuctionStatsNotFound) {
				continue
			}

			return err
		}

		auction.CalculateFlags(stats)

		err = s.auctionRepository.UpdateFlags(ctx, auction)

		if err != nil {
			return err
		}

		s.logger.Info(fmt.Sprintf("Backfilling %d/%d", i+1, total), "trace_id", traceID)
	}

	return nil
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

	if now.Hour() == 10 && now.Minute() <= 59 {
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
	source := ctx.Value(constants.SourceKey)
	traceID := ctx.Value(middleware.TraceIDKey)

	s.logger.Info("Start auction stats aggregation", "trace_id", traceID, "source", source)

	start := time.Now()

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

	s.logger.Info("Finish auction stats aggregation", "trace_id", traceID, "source", source, "duration", time.Since(start))

	return nil
}

/*
================================================================================
Scrapper Logic
================================================================================
*/

func (s *Service) ScrapBazaar(ctx context.Context) error {
	source := ctx.Value(constants.SourceKey)
	traceID := ctx.Value(middleware.TraceIDKey)

	s.logger.Info("Start", "trace_id", traceID, "source", source, "phase", constants.ScrapPhase)

	ctx = context.WithValue(ctx, constants.Phase, constants.ScrapPhase)

	start := time.Now()

	scrapperInstance := s.scrapperFactory.CreateScrapper("N")
	auctionNumberParser := s.parserFactory.CreateAuctionNumberParser(scrapperInstance)

	currentAuctions, err := auctionNumberParser.Scrap()

	s.logger.Info("Obtained active auctions", "trace_id", traceID, "source", source, "phase", constants.ScrapPhase, "auctions", currentAuctions)

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

	s.logger.Info("Obtained worlds from tibiaAPI", "trace_id", traceID, "source", source, "worlds", len(worlds))

	maxRetries := 5

	for i := range maxRetries {
		if i > 0 {
			backoff := time.Duration(i) * time.Minute

			s.logger.Warn("Rate limited, backing off", "trace_id", traceID, "source", source, "phase", constants.ScrapPhase, "retry", i, "backoff", backoff)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}

		auctionLinkSet, err := s.scrapAuctionLinks(ctx, worlds)

		if err != nil {
			if errors.Is(err, RateLimitError) {
				s.logger.Error("Rate limit error, retrying", "trace_id", traceID, "source", source, "phase", constants.ScrapPhase, "duration", time.Since(start))

				continue
			}

			s.logger.Error("Finish with error", "trace_id", traceID, "source", source, "phase", constants.ScrapPhase, "duration", time.Since(start))

			return eris.Wrap(err, "Error scrapping auction link")
		}

		if err := s.scrapAuctionDetails(ctx, auctionLinkSet); err != nil {
			if errors.Is(err, RateLimitError) {
				s.logger.Error("Rate limit error, retrying", "trace_id", traceID, "source", source, "phase", constants.ScrapPhase, "duration", time.Since(start))

				continue
			}

			s.logger.Error("Finish with error", "trace_id", traceID, "source", source, "phase", constants.ScrapPhase, "duration", time.Since(start))

			return err
		}

		s.logger.Info("Finish", "trace_id", traceID, "source", source, "phase", constants.ScrapPhase, "duration", time.Since(start), "auctions", currentAuctions, "links", len(auctionLinkSet.Data))

		return nil
	}

	return eris.New("Failed to scrap auctions: rate limit retries exhausted")
}

func (s *Service) scrapAuctionLinks(ctx context.Context, worlds []*World) (*AuctionLinkSet, error) {
	source := ctx.Value(constants.SourceKey)
	traceID := ctx.Value(middleware.TraceIDKey)

	start := time.Now()

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

				start := time.Now()

				goroutineNum := atomic.AddInt32(&goroutineCounter, 1)
				goroutineID := fmt.Sprintf("L-routine%d", goroutineNum)

				scrapperInstance := s.scrapperFactory.CreateScrapper(goroutineID)

				linkParser := s.parserFactory.CreateAuctionListParser(scrapperInstance)

				s.logger.Info("Start world link extraction", "trace_id", traceID, "source", source, "phase", constants.ScrapPhase, "world", world.Name, "routine_id", goroutineID)

				if err := linkParser.Scrap(world.Name, auctionLinkSet, storedAuctionLinkSet); err != nil {
					return eris.Wrapf(err, "Failed to collect link in L-routine%d", goroutineNum)
				}

				s.logger.Info("Finish world link extraction", "trace_id", traceID, "source", source, "phase", constants.ScrapPhase, "world", world.Name, "routine_id", goroutineID, "duration", time.Since(start))

				return nil
			})
		}

		err := g.Wait()

		if err != nil {
			return nil, eris.Wrap(err, "Failed scrapping")
		}
	}

	s.logger.Info("Finish links extraction", "trace_id", traceID, "source", source, "phase", constants.ScrapPhase, "duration", time.Since(start))

	return auctionLinkSet, nil
}

func (s *Service) scrapAuctionDetails(ctx context.Context, auctionLinkSet *AuctionLinkSet) error {
	source := ctx.Value(constants.SourceKey)
	traceID := ctx.Value(middleware.TraceIDKey)

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

		s.logger.Info("Balance workload with proxy", "trace_id", traceID, "source", source, "phase", phase, "routine_id", goroutineID, "proxy_addr", proxyAddr)

		g.Go(func() error {
			s.scrapAuctionDetail(ctx, g, failed, work)

			return nil
		})
	}

	err := g.Wait()

	if err != nil {
		return err
	}

	s.logger.Info("Finish scrapping auction details", "trace_id", traceID, "source", source, "phase", phase, "total", len(auctionLinkSet.Data), "failed", len(failed.Data), "duration", time.Since(start))

	return nil
}

func (s *Service) scrapAuctionDetail(ctx context.Context, g *errgroup.Group, failed *AuctionLinkSet, workGroup []collections.KeyValue[int, string]) {
	source := ctx.Value(constants.SourceKey)
	traceID := ctx.Value(middleware.TraceIDKey)

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

					s.logger.Error(fmt.Sprintf("Failed auction detail scrapping: %s", eris.ToString(err, true)), "trace_id", traceID, "source", source, "phase", phase, "routine_id", goroutineID, "auction_id", auctionId, "duration", time.Since(start))

					failed.Set(auctionId, linkURL)

					s.notifyStatus(ctx, totalWorkload, &workDoneCounter)

					return nil
				}

				auction, err := s.mapper.FromDTO(dto)

				if err != nil {
					s.notifyStatus(ctx, totalWorkload, &workDoneCounter)

					s.logger.Error(fmt.Sprintf("Failed mapping DTO to auction: %s", eris.ToString(err, true)), "trace_id", traceID, "source", source, "phase", phase, "routine_id", goroutineID, "auction_id", auctionId, "duration", time.Since(start))

					return nil
				}

				existingAuction, err := s.auctionRepository.GetAuctionByAuctionID(ctx, auction.AuctionID)

				if err != nil {
					s.notifyStatus(ctx, totalWorkload, &workDoneCounter)

					s.logger.Error(fmt.Sprintf("Failed finding auction by auction ID: %s", eris.ToString(err, true)), "trace_id", traceID, "source", source, "phase", phase, "routine_id", goroutineID, "auction_id", auctionId, "duration", time.Since(start))

					return nil
				}

				if existingAuction != nil {
					auction.BidRegistry = existingAuction.BidRegistry
				}

				if auction.IsEqual(existingAuction) {
					if auction.ShouldBeArchived(existingAuction) {

						err := s.auctionRepository.DeactivateAuctionRecord(ctx, auction.AuctionID)

						if err != nil {
							s.logger.Error(fmt.Sprintf("Failed deactivating auction: %s", eris.ToString(err, true)), "trace_id", traceID, "source", source, "phase", phase, "routine_id", goroutineID, "auction_id", auctionId, "duration", time.Since(start))

							s.notifyStatus(ctx, totalWorkload, &workDoneCounter)

							return nil
						}

						s.notifyStatus(ctx, totalWorkload, &workDoneCounter)

						s.logger.Info("Auction deactivated", "trace_id", traceID, "source", source, "phase", phase, "routine_id", goroutineID, "auction_id", auctionId, "duration", time.Since(start))

						return nil
					}

					err := s.auctionRepository.MarkAuctionAsUpdated(ctx, auction.AuctionID)

					if err != nil {
						s.logger.Error(fmt.Sprintf("Failed marking auction as updated: %s", eris.ToString(err, true)), "trace_id", traceID, "source", source, "phase", phase, "routine_id", goroutineID, "auction_id", auctionId, "duration", time.Since(start))

						s.notifyStatus(ctx, totalWorkload, &workDoneCounter)

						return nil
					}

					s.notifyStatus(ctx, totalWorkload, &workDoneCounter)

					s.logger.Info("Skipping auction save, no significant difference detected", "trace_id", traceID, "source", source, "phase", phase, "routine_id", goroutineID, "auction_id", auctionId, "duration", time.Since(start))

					return nil
				}

				stats, err := s.aggAuctionRepository.GetByKey(auction.SubsetKey())

				if err != nil {
					if errors.Is(err, ErrAggAuctionStatsNotFound) {
						err := s.auctionRepository.Save(auction)

						if err != nil {
							s.logger.Error(fmt.Sprintf("Failed saving auction: %s", eris.ToString(err, true)), "trace_id", traceID, "source", source, "phase", phase, "routine_id", goroutineID, "auction_id", auctionId, "duration", time.Since(start))

							s.notifyStatus(ctx, totalWorkload, &workDoneCounter)

							return nil
						}

						s.notifyStatus(ctx, totalWorkload, &workDoneCounter)

						return nil
					}

					s.logger.Error(fmt.Sprintf("Failed getting stats: %s", eris.ToString(err, true)), "trace_id", traceID, "source", source, "phase", phase, "routine_id", goroutineID, "auction_id", auctionId, "duration", time.Since(start))

					s.notifyStatus(ctx, totalWorkload, &workDoneCounter)

					return nil
				}

				auction.CalculateFlags(stats)

				err = s.auctionRepository.Save(auction)

				if err != nil {
					s.logger.Error(fmt.Sprintf("Failed saving auction: %s", eris.ToString(err, true)), "trace_id", traceID, "source", source, "phase", phase, "routine_id", goroutineID, "auction_id", auctionId, "duration", time.Since(start))

					s.notifyStatus(ctx, totalWorkload, &workDoneCounter)

					return nil
				}

				s.notifyStatus(ctx, totalWorkload, &workDoneCounter)

				s.logger.Info("Finish auction detail scrapping", "trace_id", traceID, "source", source, "phase", phase, "routine_id", goroutineID, "auction_id", auctionId, "duration", time.Since(start))

				return nil
			})
		}
	}
}

func (s *Service) notifyStatus(ctx context.Context, totalWork int, workDoneCounter *int32) {
	source := ctx.Value(constants.SourceKey)
	traceID := ctx.Value(middleware.TraceIDKey)

	newValue := atomic.AddInt32(workDoneCounter, 1)

	phase := ctx.Value(constants.Phase)

	s.logger.Info(fmt.Sprintf("%d/%d", newValue, totalWork), "trace_id", traceID, "source", source, "phase", phase)
}

/*
================================================================================
Refresh Auctions Phase
================================================================================
*/

func (s *Service) WatchActiveAuctions(ctx context.Context) error {
	source := ctx.Value(constants.SourceKey)
	traceID := ctx.Value(middleware.TraceIDKey)

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

	start := time.Now()

	s.logger.Info("Start", "trace_id", traceID, "source", source, "phase", constants.WatchPhase)

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

			if auction.DateUpd.After(time.Now().UTC().Add(-policy.updateInterval)) {
				continue
			}

			auctionsToUpdate.Set(auction.AuctionID, auction.TibiaAuctionLink)
			auctionAddedCounter++
		}

		s.logger.Info("Added active auctions for policy", "trace_id", traceID, "source", source, "phase", constants.WatchPhase, "policy", policy.finishingIn.String(), "total", auctionAddedCounter)
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

				if auction.DateUpd.After(time.Now().UTC().Add(-policy.updateInterval)) {
					continue
				}

				auctionsToUpdate.Set(auction.AuctionID, auction.TibiaAuctionLink)
				auctionAddedCounter++
			}
		}

		s.logger.Info("Added active auctions for long tail", "trace_id", traceID, "source", source, "phase", constants.WatchPhase, "total", auctionAddedCounter)
	}

	if auctionsToUpdate.IsEmpty() {
		s.logger.Info("Finish phase", "trace_id", traceID, "source", source, "phase", constants.WatchPhase, "duration", time.Since(start))

		return nil
	}

	if err := s.scrapAuctionDetails(ctx, auctionsToUpdate); err != nil {
		s.logger.Info("Finish phase", "trace_id", traceID, "source", source, "phase", constants.WatchPhase, "duration", time.Since(start))

		return eris.Wrap(err, "Failed to refresh auctions")
	}

	s.logger.Info("Finish phase", "trace_id", traceID, "source", source, "phase", constants.WatchPhase, "duration", time.Since(start))

	return nil
}

/*
================================================================================
Consolidate Auctions Phase
================================================================================
*/

func (s *Service) ConsolidateAuctions(ctx context.Context) error {
	source := ctx.Value(constants.SourceKey)
	traceID := ctx.Value(middleware.TraceIDKey)

	ctx = context.WithValue(ctx, constants.Phase, constants.ConsolidatePhase)

	s.logger.Info("Start", "trace_id", traceID, "source", source, "phase", constants.ConsolidatePhase)

	start := time.Now()

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
		s.logger.Info("Finish phase", "trace_id", traceID, "source", source, "phase", constants.ConsolidatePhase, "duration", time.Since(start))

		return nil
	}

	if err := s.scrapAuctionDetails(ctx, auctionsToUpdate); err != nil {
		s.logger.Info("Finish phase", "trace_id", traceID, "source", source, "phase", constants.ConsolidatePhase, "duration", time.Since(start))

		return eris.Wrap(err, "Failed to consolidate auctions")
	}

	s.logger.Info("Finish phase", "trace_id", traceID, "source", source, "phase", constants.ConsolidatePhase, "duration", time.Since(start))

	return nil
}

/*
================================================================================
Get Auctions Logic
================================================================================
*/

func (s *Service) GetAuctions(ctx context.Context, filter *AuctionFilter) (*PaginatedAuctions, error) {
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

		formattedAuctionEnd := auction.AuctionEnd.In(loc).Format("Jan 02 at 15:04")
		abbr, _ := auction.AuctionEnd.In(loc).Zone()

		viewModels = append(viewModels, &AuctionViewModel{
			Auction:         auction,
			LastUpdated:     beautify.FormatTimeAgo(time.Since(auction.DateUpd)),
			TimeLeft:        beautify.TimeLeft(auction.AuctionEnd),
			AucEndFormatted: fmt.Sprintf("%s %s", formattedAuctionEnd, abbr),
		})
	}

	totalCount, err := s.auctionRepository.CountActiveAuctions(ctx, filter)

	if err != nil {
		return nil, err
	}

	totalPages := totalCount / filter.Pagination.Limit

	viewFilters := FilterParams{
		Flags:  filter.Flags(),
		Status: filter.Status(),
	}

	result := &PaginatedAuctions{
		ViewModels: viewModels,
		TotalCount: totalCount,
		PageSize:   filter.Pagination.Limit,
		Page:       filter.Pagination.Page + 1,
		TotalPages: totalPages,
		Filters:    viewFilters,
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
