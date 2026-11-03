package service

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/mbrysin/PingTower/services/checker/model"
)

type siteSource interface {
	GetEnabled(ctx context.Context) ([]model.Site, error)
}

type statWriter interface {
	CreateStat(ctx context.Context, stat *model.Stat) (*model.Stat, error)
}

type statCache interface {
	Add(ctx context.Context, stat *model.Stat) error
}

type Checker struct {
	sites    siteSource
	stats    statWriter
	cache    statCache
	client   *http.Client
	limiter  *rate.Limiter
	log      *slog.Logger
	interval time.Duration
	workers  int
}

func New(
	sites siteSource,
	stats statWriter,
	cache statCache,
	log *slog.Logger,
	interval time.Duration,
	workers int,
) (*Checker, error) {
	if interval <= 0 {
		return nil, fmt.Errorf("interval must be positive")
	}
	if workers < 1 {
		return nil, fmt.Errorf("workers must be at least 1")
	}

	return &Checker{
		sites:    sites,
		stats:    stats,
		cache:    cache,
		client:   &http.Client{Timeout: 10 * time.Second},
		limiter:  rate.NewLimiter(rate.Limit(2), 1),
		log:      log.With("service", "checker"),
		interval: interval,
		workers:  workers,
	}, nil
}

func (c *Checker) Run(ctx context.Context) error {
	for {
		c.runCycle(ctx)

		timer := time.NewTimer(c.interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

func (c *Checker) runCycle(ctx context.Context) {
	sites, err := c.sites.GetEnabled(ctx)
	if err != nil {
		c.log.ErrorContext(ctx, "load enabled sites failed", "err", err)
		return
	}

	c.log.InfoContext(ctx, "starting check cycle", "sites", len(sites))
	if len(sites) == 0 {
		return
	}

	jobs := make(chan model.Site)
	var wg sync.WaitGroup

	for i := 0; i < c.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for site := range jobs {
				if ctx.Err() != nil {
					return
				}
				c.checkOne(ctx, site)
			}
		}()
	}

dispatch:
	for _, site := range sites {
		select {
		case <-ctx.Done():
			break dispatch
		case jobs <- site:
		}
	}

	close(jobs)
	wg.Wait()

	c.log.InfoContext(ctx, "check cycle finished")
}

func (c *Checker) checkOne(ctx context.Context, site model.Site) {
	if err := c.limiter.Wait(ctx); err != nil {
		return
	}

	stat := &model.Stat{
		SiteID:    site.ID,
		CheckedAt: time.Now().UTC(),
	}

	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, site.URL, nil)
	if err != nil {
		c.saveStat(ctx, stat, err)
		return
	}
	req.Header.Set("User-Agent", "PingTower/0.0.1-pre-alpha")

	resp, err := c.client.Do(req)
	stat.ResponseTimeMs = int32Ptr(int32(time.Since(start).Milliseconds()))

	if err != nil {
		c.saveStat(ctx, stat, err)
		return
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	statusCode := int32(resp.StatusCode)
	stat.StatusCode = &statusCode
	stat.IsUp = resp.StatusCode < http.StatusInternalServerError

	c.saveStat(ctx, stat, nil)
}

func (c *Checker) saveStat(ctx context.Context, stat *model.Stat, checkErr error) {
	if checkErr != nil {
		message := checkErr.Error()
		stat.Error = &message
	}

	created, err := c.stats.CreateStat(ctx, stat)
	if err != nil {
		c.log.ErrorContext(ctx, "save stat to postgres failed",
			"site_id", stat.SiteID,
			"err", err,
		)
		return
	}

	if c.cache == nil {
		return
	}
	if err := c.cache.Add(ctx, created); err != nil {
		c.log.ErrorContext(ctx, "save stat to redis failed",
			"site_id", stat.SiteID,
			"err", err,
		)
	}
}

func int32Ptr(n int32) *int32 {
	return &n
}
