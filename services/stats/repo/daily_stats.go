package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/mbrysin/PingTower/services/stats/model"
)

func (r *Stats) GetMonthlyStats(
	ctx context.Context,
	siteID int,
	year int,
	month time.Month,
) ([]model.DailyStat, error) {
	start := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)

	const q = `
		SELECT id, site_id, stat_date, checks, failed_checks,
		       success_percent, last_check, avg_response_time_ms,
		       min_response_time_ms, max_response_time_ms
		FROM daily_stat
		WHERE site_id = $1
		  AND stat_date >= $2
		  AND stat_date < $3
		ORDER BY stat_date
	`

	rows, err := r.db.Query(ctx, q, siteID, start, end)
	if err != nil {
		r.log.ErrorContext(ctx, "get monthly stats failed",
			"site_id", siteID, "year", year, "month", month, "err", err)
		return nil, fmt.Errorf("get monthly stats for site %d: %w", siteID, err)
	}
	defer rows.Close()

	stats := make([]model.DailyStat, 0)
	for rows.Next() {
		var s model.DailyStat
		if err := rows.Scan(
			&s.ID,
			&s.SiteID,
			&s.StatDate,
			&s.Checks,
			&s.FailedChecks,
			&s.SuccessPercent,
			&s.LastCheck,
			&s.AvgResponseTimeMs,
			&s.MinResponseTimeMs,
			&s.MaxResponseTimeMs,
		); err != nil {
			return nil, fmt.Errorf("scan daily stat: %w", err)
		}
		stats = append(stats, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate monthly stats: %w", err)
	}

	return stats, nil
}

func (r *Stats) GetDailyStatsBetween(
	ctx context.Context,
	siteID int,
	from time.Time,
	to time.Time,
) ([]model.DailyStat, error) {
	fromDate := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
	toDate := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, time.UTC)

	if !fromDate.Before(toDate) {
		return nil, fmt.Errorf("invalid date range: from must be before to")
	}

	const q = `
		SELECT id, site_id, stat_date, checks, failed_checks,
		       success_percent, last_check, avg_response_time_ms,
		       min_response_time_ms, max_response_time_ms
		FROM daily_stat
		WHERE site_id = $1
		  AND stat_date >= $2
		  AND stat_date < $3
		ORDER BY stat_date
	`

	rows, err := r.db.Query(ctx, q, siteID, fromDate, toDate)
	if err != nil {
		r.log.ErrorContext(ctx, "get daily stats between dates failed",
			"site_id", siteID, "from", fromDate, "to", toDate, "err", err)
		return nil, fmt.Errorf("get daily stats for site %d between dates: %w", siteID, err)
	}
	defer rows.Close()

	stats := make([]model.DailyStat, 0)
	for rows.Next() {
		var s model.DailyStat
		if err := rows.Scan(
			&s.ID,
			&s.SiteID,
			&s.StatDate,
			&s.Checks,
			&s.FailedChecks,
			&s.SuccessPercent,
			&s.LastCheck,
			&s.AvgResponseTimeMs,
			&s.MinResponseTimeMs,
			&s.MaxResponseTimeMs,
		); err != nil {
			return nil, fmt.Errorf("scan daily stat: %w", err)
		}
		stats = append(stats, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate daily stats: %w", err)
	}

	return stats, nil
}
