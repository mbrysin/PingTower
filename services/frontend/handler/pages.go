package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct{}

func New() *Handler {
	return &Handler{}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.GET("/healthz", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	r.GET("/readyz", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	r.LoadHTMLGlob("services/frontend/templates/*")
	r.Static("/static", "services/frontend/static")

	r.GET("/", h.Dashboard)
}

func (h *Handler) Dashboard(c *gin.Context) {
	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"title": "PingTower — Dashboard",
	})
}
