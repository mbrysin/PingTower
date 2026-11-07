package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/mbrysin/PingTower/services/checker/repo"
	"github.com/mbrysin/PingTower/services/checker/service"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	databaseURL := mustEnv("DATABASE_URL")
	redisAddr := envOrDefault("REDIS_ADDR", "localhost:6379")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDB := envInt("REDIS_DB", 0)
	interval := time.Duration(envInt("CHECK_INTERVAL", 5)) * time.Minute
	workers := envInt("WORKERS", 10)
	retention := time.Duration(envInt("REDIS_RETENTION_DAYS", 30)) * 24 * time.Hour

	pgPool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("create PostgreSQL pool: %v", err)
	}
	defer pgPool.Close()

	if err := pgPool.Ping(ctx); err != nil {
		log.Fatalf("connect to PostgreSQL: %v", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})
	defer redisClient.Close()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("connect to Redis: %v", err)
	}

	var ready atomic.Bool

	healthServer := &http.Server{
		Addr:              envOrDefault("HEALTH_ADDR", ":8084"),
		ReadHeaderTimeout: 2 * time.Second,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/healthz":
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("ok"))

			case "/readyz":
				if !ready.Load() {
					http.Error(w, "not ready", http.StatusServiceUnavailable)
					return
				}

				checkCtx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
				defer cancel()

				if err := pgPool.Ping(checkCtx); err != nil {
					http.Error(w, "database unavailable", http.StatusServiceUnavailable)
					return
				}

				if err := redisClient.Ping(checkCtx).Err(); err != nil {
					http.Error(w, "redis unavailable", http.StatusServiceUnavailable)
					return
				}

				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("ready"))

			default:
				http.NotFound(w, r)
			}
		}),
	}

	go func() {
		if err := healthServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("health server stopped", "err", err)
			stop()
		}
	}()

	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := healthServer.Shutdown(shutdownCtx); err != nil {
			logger.Error("health server shutdown failed", "err", err)
		}
	}()

	sitesRepo := repo.NewSites(pgPool, logger)
	statsRepo := repo.NewStats(pgPool, logger)
	redisStatsRepo := repo.NewRedisStats(redisClient, retention)

	dailyStatsRepo := repo.NewDailyStats(pgPool, logger)
	go runDailyAggregation(ctx, dailyStatsRepo, logger)

	checker, err := service.New(
		sitesRepo,
		statsRepo,
		redisStatsRepo,
		logger,
		interval,
		workers,
	)
	if err != nil {
		log.Fatalf("create checker service: %v", err)
	}

	logger.Info("checker service started",
		"interval", interval.String(),
		"workers", workers,
		"redis_retention", retention.String(),
	)

	ready.Store(true)

	if err := checker.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("checker service stopped: %v", err)
	}

	ready.Store(false)

	logger.Info("checker service stopped")
}

func mustEnv(name string) string {
	value := os.Getenv(name)
	if value == "" {
		log.Fatalf("required environment variable %s is not set", name)
	}
	return value
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func envInt(name string, fallback int) int {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}

	n, err := strconv.Atoi(value)
	if err != nil || n <= 0 {
		log.Fatalf("environment variable %s must be a positive integer", name)
	}
	return n
}

type dailyAggregator interface {
	AggregateDay(ctx context.Context, day time.Time) error
}

func runDailyAggregation(
	ctx context.Context,
	aggregator dailyAggregator,
	logger *slog.Logger,
) {
	run := func() {
		yesterday := time.Now().UTC().AddDate(0, 0, -1)
		day := time.Date(
			yesterday.Year(), yesterday.Month(), yesterday.Day(),
			0, 0, 0, 0, time.UTC,
		)

		if err := aggregator.AggregateDay(ctx, day); err != nil {
			logger.ErrorContext(ctx, "daily aggregation failed", "day", day, "err", err)
		}
	}

	run()

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			run()
		}
	}
}
