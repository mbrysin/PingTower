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

//
// type DailyStat struct {
// 	ID                int        `db:"id"                  json:"id"`
// 	SiteID            int        `db:"site_id"             json:"site_id"`
// 	StatDate          time.Time  `db:"stat_date"           json:"stat_date"`
// 	Checks            int        `db:"checks"              json:"checks"`
// 	FailedChecks      int        `db:"failed_checks"       json:"failed_checks"`
// 	SuccessPercent    *float64   `db:"success_percent"     json:"success_percent"`
// 	LastCheck         *time.Time `db:"last_check"          json:"last_check"`
// 	AvgResponseTimeMs *float64   `db:"avg_response_time_ms" json:"avg_response_time_ms"`
// 	MinResponseTimeMs *int64     `db:"min_response_time_ms" json:"min_response_time_ms"`
// 	MaxResponseTimeMs *int64     `db:"max_response_time_ms" json:"max_response_time_ms"`
// }
