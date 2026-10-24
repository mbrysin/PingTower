package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	goredis "github.com/redis/go-redis/v9"

	"github.com/mbrysin/PingTower/services/checker/cache"
	"github.com/mbrysin/PingTower/services/checker/checker"
	"github.com/mbrysin/PingTower/services/checker/repo"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	if err := run(ctx, logger); err != nil {
		logger.Error("checker stopped", "err", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, logger *slog.Logger) error {
	db, err := pgxpool.New(ctx, requiredEnv("DATABASE_URL"))
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		return err
	}

	redisClient := goredis.NewClient(&goredis.Options{
		Addr: envOr("REDIS_ADDR", "redis:6379"),
	})
	defer redisClient.Close()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		return err
	}

	sitesRepo := repo.NewSites(db, logger)
	resultsRepo := repo.NewResults(db, logger)
	dailyStatsRepo := repo.NewDailyStats(db, logger)
	latestCache := cache.NewRedis(redisClient)

	worker := checker.NewWorker(
		sitesRepo,
		resultsRepo,
		dailyStatsRepo,
		latestCache,
		logger,
	)

	return worker.Run(
		ctx,
		durationEnv("CHECK_INTERVAL", time.Minute),
		durationEnv("SITES_RELOAD_INTERVAL", 5*time.Minute),
		intEnv("CHECKER_WORKERS", 10),
	)
}

func requiredEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("required environment variable %s is not set", key)
	}
	return value
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func durationEnv(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	d, err := time.ParseDuration(value)
	if err != nil || d <= 0 {
		log.Fatalf("%s must be a positive duration: %q", key, value)
	}
	return d
}

func intEnv(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	n, err := strconv.Atoi(value)
	if err != nil || n < 1 {
		log.Fatalf("%s must be a positive integer: %q", key, value)
	}
	return n
}
