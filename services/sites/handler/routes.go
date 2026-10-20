package handler

import "github.com/gin-gonic/gin"

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	sites := r.Group("/sites")
	sites.GET("", h.List)
	sites.GET("/:id", h.Get)
}
