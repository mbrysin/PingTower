package checker

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/mbrysin/PingTower/services/checker/model"
)

type SitesRepository interface {
	GetEnabled(ctx context.Context) ([]model.Site, error)
}

type ResultsRepository interface {
	Create(ctx context.Context, result model.CheckResult) error
}

type DailyStatsRepository interface {
	RecordCheck(ctx context.Context, result model.CheckResult) error
}

type LatestCache interface {
	SetLatest(
		ctx context.Context,
		siteID int,
		result model.CheckResult,
	) error
}

type Worker struct {
	sitesRepo      SitesRepository
	resultsRepo    ResultsRepository
	dailyStatsRepo DailyStatsRepository
	cache          LatestCache
	log            *slog.Logger
}

func NewWorker(
	sitesRepo SitesRepository,
	resultsRepo ResultsRepository,
	dailyStatsRepo DailyStatsRepository,
	latestCache LatestCache,
	logger *slog.Logger,
) *Worker {
	return &Worker{
		sitesRepo:      sitesRepo,
		resultsRepo:    resultsRepo,
		dailyStatsRepo: dailyStatsRepo,
		cache:          latestCache,
		log:            logger.With("service", "checker"),
	}
}

func (w *Worker) checkAndSave(ctx context.Context, site model.Site) {
	result := Check(ctx, site.URL)
	result.SiteID = site.ID

	if err := w.resultsRepo.Create(ctx, result); err != nil {
		w.log.ErrorContext(ctx, "save check result",
			"siteID", site.ID, "err", err)
		return
	}

	if err := w.dailyStatsRepo.RecordCheck(ctx, result); err != nil {
		w.log.ErrorContext(ctx, "update daily stats",
			"siteID", site.ID, "err", err)
	}

	if err := w.cache.SetLatest(ctx, site.ID, result); err != nil {
		w.log.ErrorContext(ctx, "update latest status cache",
			"siteID", site.ID, "err", err)
	}
}

func (w *Worker) runChecks(ctx context.Context, sites []model.Site, workers int) {
	if workers < 1 {
		workers = 1
	}

	jobs := make(chan model.Site)
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for site := range jobs {
				if ctx.Err() != nil {
					return
				}
				w.checkAndSave(ctx, site)
			}
		}()
	}

	for _, site := range sites {
		select {
		case jobs <- site:
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			return
		}
	}

	close(jobs)
	wg.Wait()
}

func (w *Worker) Run(ctx context.Context,
	checkInterval time.Duration,
	reloadInterval time.Duration,
	workers int,
) error {
	if checkInterval <= 0 {
		return fmt.Errorf("check interval must be positive")
	}
	if reloadInterval <= 0 {
		return fmt.Errorf("reload interval must be positive")
	}

	sites, err := w.sitesRepo.GetEnabled(ctx)
	if err != nil {
		return fmt.Errorf("load enabled sites: %w", err)
	}

	w.runChecks(ctx, sites, workers)

	checkTicker := time.NewTicker(checkInterval)
	defer checkTicker.Stop()

	reloadTicker := time.NewTicker(reloadInterval)
	defer reloadTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-checkTicker.C:
			w.runChecks(ctx, sites, workers)

		case <-reloadTicker.C:
			updatedSites, err := w.sitesRepo.GetEnabled(ctx)
			if err != nil {
				w.log.ErrorContext(ctx, "reload enabled sites failed", "err", err)
				continue
			}
			sites = updatedSites
		}
	}
}
