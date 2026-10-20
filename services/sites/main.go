package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mbrysin/PingTower/services/sites/handler"
	"github.com/mbrysin/PingTower/services/sites/repo"
	"github.com/mbrysin/PingTower/services/sites/service"
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

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	siteRepo := repo.NewSites(db, logger)
	siteService := service.NewSiteService(siteRepo)
	siteHandler := handler.New(siteService)

	router := gin.Default()
	siteHandler.RegisterRoutes(router)

	if err := http.ListenAndServe(":8081", router); err != nil {
		log.Fatalf("start server: %v", err)
	}
}
