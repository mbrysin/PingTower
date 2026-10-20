package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mbrysin/PingTower/services/sites/model"
	"github.com/mbrysin/PingTower/services/sites/service"
)

type SitesService interface {
	Get(ctx context.Context, id int) (*model.Site, error)
	List(ctx context.Context, limit, offset int) ([]model.Site, error)
}

type Handler struct {
	sites SitesService
}

func New(sites SitesService) *Handler {
	return &Handler{
		sites: sites,
	}
}

func (h *Handler) Get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid site id"})
		return
	}

	site, err := h.sites.Get(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrInvalidSite) {
			c.JSON(http.StatusNotFound, gin.H{"error": "site not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, site)
}

func (h *Handler) List(c *gin.Context) {
	limit, offset := 20, 0

	if value := c.Query("limit"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit"})
			return
		}
		limit = parsed
	}

	if value := c.Query("offset"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid offset"})
			return
		}
		offset = parsed
	}

	sites, err := h.sites.List(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pagination"})
		return
	}

	c.JSON(http.StatusOK, sites)
}
