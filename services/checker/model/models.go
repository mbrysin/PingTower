package model

import "time"

type Site struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	URL     string `json:"url"`
	Enabled bool   `json:"enabled"`
}

type Stat struct {
	ID             int       `db:"id"               json:"id"`
	SiteID         int       `db:"site_id"          json:"site_id"`
	CheckedAt      time.Time `db:"checked_at"       json:"checked_at"`
	IsUp           bool      `db:"is_up"            json:"is_up"`
	StatusCode     *int32    `db:"status_code"      json:"status_code,omitempty"`
	ResponseTimeMs *int32    `db:"response_time_ms" json:"response_time_ms,omitempty"`
	Error          *string   `db:"error"            json:"error,omitempty"`
}
