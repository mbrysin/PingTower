package handler

import "github.com/gin-gonic/gin"

func (h *Handler) RegisterRoutes(router *gin.Engine) {
	router.GET("/sites/latest-stats", func(c *gin.Context) {
		h.GetLatestForSites(c.Writer, c.Request)
	})

	router.GET("/sites/:siteID/stats", func(c *gin.Context) {
		c.Request.SetPathValue("siteID", c.Param("siteID"))
		h.GetStatsBetween(c.Writer, c.Request)
	})

	router.GET("/sites/:siteID/daily-stats", func(c *gin.Context) {
		c.Request.SetPathValue("siteID", c.Param("siteID"))
		h.GetDailyStatsBetween(c.Writer, c.Request)
	})

	router.GET("/sites/:siteID/latest-stat", func(c *gin.Context) {
		c.Request.SetPathValue("siteID", c.Param("siteID"))
		h.GetLatest(c.Writer, c.Request)
	})
}
