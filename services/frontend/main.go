package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"

	frontendHandler "github.com/mbrysin/PingTower/services/frontend/handler"
)

func main() {
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	frontend := frontendHandler.New()
	frontend.RegisterRoutes(router)

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8083"
	}

	log.Printf("frontend listening on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("run frontend: %v", err)
	}
}
