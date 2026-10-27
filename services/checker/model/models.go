package model

import "time"

type CheckResult struct {
	ID             int       `json:"id"`
	SiteID         int       `json:"site_id"`
	CheckedAt      time.Time `json:"checked_at"`
	IsUp           bool      `json:"is_up"`
	StatusCode     int       `json:"status_code"`
	ResponseTimeMs int64     `json:"response_time_ms"`
	Error          string    `json:"error,omitempty"`
}

type SiteDailyStat struct {
	ID            int       `json:"id"`
	SiteID        int       `json:"site_id"`
	Date          time.Time `json:"date"`
	Status        string    `json:"status"`
	ChecksTotal   int       `json:"checks_total"`
	ChecksFailed  int       `json:"checks_failed"`
	UptimePercent float64   `json:"uptime_percent"`
	LastCheckedAt time.Time `json:"last_checked_at"`
}
