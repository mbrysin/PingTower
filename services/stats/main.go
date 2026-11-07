package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/mbrysin/PingTower/services/stats/handler"
	"github.com/mbrysin/PingTower/services/stats/repo"
	"github.com/mbrysin/PingTower/services/stats/service"
)

func main() {
	ctx := context.Background()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	db, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		log.Fatalf("ping database: %v", err)
	}

	redisAddr := envOrDefault("REDIS_ADDR", "localhost:6379")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDB := envNonNegativeInt("REDIS_DB", 0)

	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})
	defer redisClient.Close()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("connect to Redis: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	pgStatsRepo := repo.NewStats(db, logger)
	redisStatsRepo := repo.NewRedisStats(redisClient)

	statsService := service.NewStats(pgStatsRepo, redisStatsRepo)
	statsHandler := handler.New(statsService)

	router := gin.Default()
	statsHandler.RegisterRoutes(router)

	router.GET("/healthz", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	router.GET("/readyz", func(c *gin.Context) {
		checkCtx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		if err := db.Ping(checkCtx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready"})
			return
		}

		if err := redisClient.Ping(checkCtx).Err(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	addr := envOrDefault("STATS_ADDR", ":8082")
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("start stats server: %v", err)
	}
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

func envNonNegativeInt(name string, fallback int) int {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}

	n, err := strconv.Atoi(value)
	if err != nil || n < 0 {
		log.Fatalf("environment variable %s must be a non-negative integer", name)
	}
	return n
}
