package service

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/time/rate"

	"github.com/mbrysin/PingTower/services/checker/model"
)

type statsStub struct {
	got *model.Stat
	err error
}

func (s *statsStub) CreateStat(_ context.Context, stat *model.Stat) (*model.Stat, error) {
	s.got = stat
	if s.err != nil {
		return nil, s.err
	}
	return stat, nil
}

type cacheStub struct {
	got *model.Stat
	err error
}

func (s *cacheStub) Add(_ context.Context, stat *model.Stat) error {
	s.got = stat
	return s.err
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestNewValidation(t *testing.T) {
	logger := testLogger()

	tests := []struct {
		name     string
		interval time.Duration
		workers  int
		wantErr  bool
	}{
		{name: "valid", interval: time.Second, workers: 1},
		{name: "zero interval", interval: 0, workers: 1, wantErr: true},
		{name: "negative interval", interval: -time.Second, workers: 1, wantErr: true},
		{name: "zero workers", interval: time.Second, workers: 0, wantErr: true},
		{name: "negative workers", interval: time.Second, workers: -1, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(nil, nil, nil, logger, tt.interval, tt.workers)
			if (err != nil) != tt.wantErr {
				t.Fatalf("New() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCheckOneHTTPStatus(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantUp     bool
	}{
		{name: "success", statusCode: http.StatusOK, wantUp: true},
		{name: "not found is considered up", statusCode: http.StatusNotFound, wantUp: true},
		{name: "server error", statusCode: http.StatusInternalServerError, wantUp: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			stats := &statsStub{}
			checker := &Checker{
				stats:   stats,
				client:  server.Client(),
				limiter: rate.NewLimiter(rate.Inf, 1),
				log:     testLogger(),
			}

			checker.checkOne(context.Background(), model.Site{
				ID:  7,
				URL: server.URL,
			})

			if stats.got == nil {
				t.Fatal("expected stat to be saved")
			}
			if stats.got.SiteID != 7 {
				t.Errorf("SiteID = %d, want 7", stats.got.SiteID)
			}
			if stats.got.IsUp != tt.wantUp {
				t.Errorf("IsUp = %v, want %v", stats.got.IsUp, tt.wantUp)
			}
			if stats.got.StatusCode == nil {
				t.Fatal("StatusCode is nil")
			}
			if *stats.got.StatusCode != int32(tt.statusCode) {
				t.Errorf("StatusCode = %d, want %d", *stats.got.StatusCode, tt.statusCode)
			}
		})
	}
}

func TestCheckOneInvalidURLSavesError(t *testing.T) {
	stats := &statsStub{}
	checker := &Checker{
		stats:   stats,
		client:  &http.Client{Timeout: time.Second},
		limiter: rate.NewLimiter(rate.Inf, 1),
		log:     testLogger(),
	}

	checker.checkOne(context.Background(), model.Site{
		ID:  3,
		URL: "://invalid-url",
	})

	if stats.got == nil {
		t.Fatal("expected stat to be saved")
	}
	if stats.got.Error == nil || *stats.got.Error == "" {
		t.Fatal("expected stat to contain an error")
	}
	if stats.got.IsUp {
		t.Error("IsUp = true, want false for invalid URL")
	}
}

func TestSaveStat(t *testing.T) {
	t.Run("saves to cache after PostgreSQL", func(t *testing.T) {
		stats := &statsStub{}
		cache := &cacheStub{}
		checker := &Checker{
			stats: stats,
			cache: cache,
			log:   testLogger(),
		}

		checker.saveStat(context.Background(), &model.Stat{SiteID: 5}, nil)

		if stats.got == nil {
			t.Fatal("expected PostgreSQL stat to be saved")
		}
		if cache.got == nil {
			t.Fatal("expected stat to be added to cache")
		}
	})

	t.Run("does not write to cache if PostgreSQL fails", func(t *testing.T) {
		stats := &statsStub{err: errors.New("database unavailable")}
		cache := &cacheStub{}
		checker := &Checker{
			stats: stats,
			cache: cache,
			log:   testLogger(),
		}

		checker.saveStat(context.Background(), &model.Stat{SiteID: 5}, nil)

		if cache.got != nil {
			t.Error("cache should not be called when PostgreSQL save fails")
		}
	})

	t.Run("stores check error in stat", func(t *testing.T) {
		stats := &statsStub{}
		checker := &Checker{
			stats: stats,
			log:   testLogger(),
		}

		checker.saveStat(
			context.Background(),
			&model.Stat{SiteID: 5},
			errors.New("connection refused"),
		)

		if stats.got == nil || stats.got.Error == nil {
			t.Fatal("expected check error to be stored in stat")
		}
		if *stats.got.Error != "connection refused" {
			t.Errorf("Error = %q, want %q", *stats.got.Error, "connection refused")
		}
	})
}
