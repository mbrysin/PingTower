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
	Create(ctx context.Context, name, rawURL string) (*model.Site, error)
	Get(ctx context.Context, id int) (*model.Site, error)
	List(ctx context.Context, limit, offset int) ([]model.Site, error)
	UpdateName(ctx context.Context, id int, name string) (*model.Site, error)
	SetEnabled(ctx context.Context, id int, enabled bool) (*model.Site, error)
	Delete(ctx context.Context, id int) error
}

type Handler struct {
	sites SitesService
}

func New(sites SitesService) *Handler {
	return &Handler{
		sites: sites,
	}
}

func (h *Handler) Create(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
		URL  string `json:"url" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	site, err := h.sites.Create(c.Request.Context(), req.Name, req.URL)
	if err != nil {
		if errors.Is(err, service.ErrSiteExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "site already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, site)
}

func (h *Handler) Get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid site id"})
		return
	}

	site, err := h.sites.Get(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrSiteNotFound) {
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

func (h *Handler) UpdateName(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid site id"})
		return
	}

	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	site, err := h.sites.UpdateName(c.Request.Context(), id, req.Name)
	if err != nil {
		if errors.Is(err, service.ErrSiteNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "site not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, site)
}

func (h *Handler) SetEnabled(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid site id"})
		return
	}

	var req struct {
		Enabled *bool `json:"enabled" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	site, err := h.sites.SetEnabled(c.Request.Context(), id, *req.Enabled)
	if err != nil {
		if errors.Is(err, service.ErrSiteNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "site not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, site)
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid site id"})
		return
	}

	if err := h.sites.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, service.ErrSiteNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "site not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.Status(http.StatusNoContent)
}
