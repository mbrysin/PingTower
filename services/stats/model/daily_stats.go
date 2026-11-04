package model

import "time"

type DailyStat struct {
	ID                int        `db:"id"                  json:"id"`
	SiteID            int        `db:"site_id"             json:"site_id"`
	StatDate          time.Time  `db:"stat_date"           json:"stat_date"`
	Checks            int        `db:"checks"              json:"checks"`
	FailedChecks      int        `db:"failed_checks"       json:"failed_checks"`
	SuccessPercent    *float64   `db:"success_percent"     json:"success_percent"`
	LastCheck         *time.Time `db:"last_check"          json:"last_check"`
	AvgResponseTimeMs *float64   `db:"avg_response_time_ms" json:"avg_response_time_ms"`
	MinResponseTimeMs *int64     `db:"min_response_time_ms" json:"min_response_time_ms"`
	MaxResponseTimeMs *int64     `db:"max_response_time_ms" json:"max_response_time_ms"`
}
