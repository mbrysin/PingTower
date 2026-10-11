package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mbrysin/PingTower/services/auth/middleware"
)

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	auth := r.Group("/auth")
	auth.POST("/register", h.Register)
	auth.POST("/login", h.Login)

	protected := r.Group("/api")
	protected.Use(middleware.Auth(h.tokens))
	protected.GET("/me", h.Me)

}
