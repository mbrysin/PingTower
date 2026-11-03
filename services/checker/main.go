package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
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

	sitesRepo := repo.NewSites(pgPool, logger)
	statsRepo := repo.NewStats(pgPool, logger)
	redisStatsRepo := repo.NewRedisStats(redisClient, retention)

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

	if err := checker.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("checker service stopped: %v", err)
	}

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
