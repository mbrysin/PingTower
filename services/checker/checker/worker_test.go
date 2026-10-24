package checker

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mbrysin/PingTower/services/checker/model"
)

type fakeSitesRepo struct {
	sites []model.Site
	err   error
}

func (f *fakeSitesRepo) GetEnabled(context.Context) ([]model.Site, error) {
	return f.sites, f.err
}

type fakeResultsRepo struct {
	result model.CheckResult
	calls  int
	err    error
}

func (f *fakeResultsRepo) Create(
	_ context.Context,
	result model.CheckResult,
) error {
	f.calls++
	f.result = result
	return f.err
}

type fakeDailyStatsRepo struct {
	result model.CheckResult
	calls  int
	err    error
}

func (f *fakeDailyStatsRepo) RecordCheck(
	_ context.Context,
	result model.CheckResult,
) error {
	f.calls++
	f.result = result
	return f.err
}

type fakeLatestCache struct {
	siteID int
	result model.CheckResult
	calls  int
	err    error
}

func (f *fakeLatestCache) SetLatest(
	_ context.Context,
	siteID int,
	result model.CheckResult,
) error {
	f.calls++
	f.siteID = siteID
	f.result = result
	return f.err
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestCheckAndSave(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(
		w http.ResponseWriter,
		_ *http.Request,
	) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	results := &fakeResultsRepo{}
	dailyStats := &fakeDailyStatsRepo{}
	latestCache := &fakeLatestCache{}

	worker := NewWorker(
		&fakeSitesRepo{},
		results,
		dailyStats,
		latestCache,
		testLogger(),
	)

	site := model.Site{
		ID:  42,
		URL: server.URL,
	}

	worker.checkAndSave(context.Background(), site)

	if results.calls != 1 {
		t.Fatalf("results Create calls = %d, want 1", results.calls)
	}
	if dailyStats.calls != 1 {
		t.Fatalf("daily stats calls = %d, want 1", dailyStats.calls)
	}
	if latestCache.calls != 1 {
		t.Fatalf("cache calls = %d, want 1", latestCache.calls)
	}

	if results.result.SiteID != site.ID {
		t.Errorf("saved result SiteID = %d, want %d",
			results.result.SiteID, site.ID)
	}
	if !results.result.IsUp {
		t.Error("expected result to be up")
	}
	if latestCache.siteID != site.ID {
		t.Errorf("cached siteID = %d, want %d",
			latestCache.siteID, site.ID)
	}
}

func TestCheckAndSaveDoesNotUpdateCacheWhenSavingResultFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(
		w http.ResponseWriter,
		_ *http.Request,
	) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	results := &fakeResultsRepo{err: errors.New("database unavailable")}
	dailyStats := &fakeDailyStatsRepo{}
	latestCache := &fakeLatestCache{}

	worker := NewWorker(
		&fakeSitesRepo{},
		results,
		dailyStats,
		latestCache,
		testLogger(),
	)

	worker.checkAndSave(context.Background(), model.Site{
		ID:  42,
		URL: server.URL,
	})

	if dailyStats.calls != 0 {
		t.Errorf("daily stats calls = %d, want 0", dailyStats.calls)
	}
	if latestCache.calls != 0 {
		t.Errorf("cache calls = %d, want 0", latestCache.calls)
	}
}

func TestCheckAndSaveContinuesWhenDailyStatsFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(
		w http.ResponseWriter,
		_ *http.Request,
	) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	results := &fakeResultsRepo{}
	dailyStats := &fakeDailyStatsRepo{err: errors.New("stats unavailable")}
	latestCache := &fakeLatestCache{}

	worker := NewWorker(
		&fakeSitesRepo{},
		results,
		dailyStats,
		latestCache,
		testLogger(),
	)

	worker.checkAndSave(context.Background(), model.Site{
		ID:  42,
		URL: server.URL,
	})

	if results.calls != 1 {
		t.Errorf("results Create calls = %d, want 1", results.calls)
	}
	if latestCache.calls != 1 {
		t.Errorf("cache calls = %d, want 1", latestCache.calls)
	}
}
