package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mbrysin/PingTower/services/auth/handler"
	"github.com/mbrysin/PingTower/services/auth/repo"
	"github.com/mbrysin/PingTower/services/auth/service"
	"github.com/mbrysin/PingTower/services/auth/token"
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

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	tokens, err := token.NewManager(os.Getenv("JWT_SECRET"))
	if err != nil {
		log.Fatalf("create token manager: %v", err)
	}

	userRepo := repo.NewUsers(db, logger)
	authService := service.NewAuthService(userRepo)
	authHandler := handler.New(authService, tokens)

	router := gin.Default()

	authHandler.RegisterRoutes(router)

	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("start server: %v", err)
	}
}
