package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mbrysin/PingTower/services/stats/model"
)

type StatsService interface {
	GetStatsBetween(
		ctx context.Context,
		siteID int,
		from time.Time,
		to time.Time,
	) ([]model.Stat, error)

	GetDailyStatsBetween(
		ctx context.Context,
		siteID int,
		from time.Time,
		to time.Time,
	) ([]model.DailyStat, error)

	GetLatest(
		ctx context.Context,
		siteID int,
	) (*model.Stat, error)

	GetLatestForSites(
		ctx context.Context,
		siteIDs []int,
	) (map[int]model.Stat, error)
}

type Handler struct {
	service StatsService
}

func New(service StatsService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetStatsBetween(w http.ResponseWriter, r *http.Request) {
	siteID, ok := parseSiteID(w, r)
	if !ok {
		return
	}

	from, to, ok := parseDateTimeRange(w, r)
	if !ok {
		return
	}

	stats, err := h.service.GetStatsBetween(r.Context(), siteID, from, to)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

func (h *Handler) GetDailyStatsBetween(w http.ResponseWriter, r *http.Request) {
	siteID, ok := parseSiteID(w, r)
	if !ok {
		return
	}

	from, to, ok := parseDateRange(w, r)
	if !ok {
		return
	}

	stats, err := h.service.GetDailyStatsBetween(r.Context(), siteID, from, to)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

func parseSiteID(w http.ResponseWriter, r *http.Request) (int, bool) {
	siteID, err := strconv.Atoi(r.PathValue("siteID"))
	if err != nil || siteID <= 0 {
		http.Error(w, "invalid siteID", http.StatusBadRequest)
		return 0, false
	}
	return siteID, true
}

func parseDateTimeRange(w http.ResponseWriter, r *http.Request) (time.Time, time.Time, bool) {
	from, err := time.Parse(time.RFC3339, r.URL.Query().Get("from"))
	if err != nil {
		http.Error(w, "from must be RFC3339", http.StatusBadRequest)
		return time.Time{}, time.Time{}, false
	}

	to, err := time.Parse(time.RFC3339, r.URL.Query().Get("to"))
	if err != nil {
		http.Error(w, "to must be RFC3339", http.StatusBadRequest)
		return time.Time{}, time.Time{}, false
	}

	if !from.Before(to) {
		http.Error(w, "from must be before to", http.StatusBadRequest)
		return time.Time{}, time.Time{}, false
	}

	return from, to, true
}

func parseDateRange(w http.ResponseWriter, r *http.Request) (time.Time, time.Time, bool) {
	from, err := time.Parse("2006-01-02", r.URL.Query().Get("from"))
	if err != nil {
		http.Error(w, "from must use YYYY-MM-DD format", http.StatusBadRequest)
		return time.Time{}, time.Time{}, false
	}

	to, err := time.Parse("2006-01-02", r.URL.Query().Get("to"))
	if err != nil {
		http.Error(w, "to must use YYYY-MM-DD format", http.StatusBadRequest)
		return time.Time{}, time.Time{}, false
	}

	if !from.Before(to) {
		http.Error(w, "from must be before to", http.StatusBadRequest)
		return time.Time{}, time.Time{}, false
	}

	return from, to, true
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func (h *Handler) GetLatest(w http.ResponseWriter, r *http.Request) {
	siteID, ok := parseSiteID(w, r)
	if !ok {
		return
	}

	stat, err := h.service.GetLatest(r.Context(), siteID)
	if err != nil {
		http.Error(w, "failed to get latest stat", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, stat)
}

func (h *Handler) GetLatestForSites(w http.ResponseWriter, r *http.Request) {
	rawIDs := r.URL.Query().Get("site_ids")
	if rawIDs == "" {
		http.Error(w, "site_ids is required", http.StatusBadRequest)
		return
	}

	parts := strings.Split(rawIDs, ",")
	if len(parts) > 100 {
		http.Error(w, "at most 100 site_ids are allowed", http.StatusBadRequest)
		return
	}

	siteIDs := make([]int, 0, len(parts))
	for _, part := range parts {
		id, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || id <= 0 {
			http.Error(w, "site_ids must be comma-separated positive integers", http.StatusBadRequest)
			return
		}
		siteIDs = append(siteIDs, id)
	}

	stats, err := h.service.GetLatestForSites(r.Context(), siteIDs)
	if err != nil {
		http.Error(w, "failed to get latest stats", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, stats)
}
