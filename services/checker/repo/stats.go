package repo

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mbrysin/PingTower/services/checker/model"
)

type Results struct {
	db  *pgxpool.Pool
	log *slog.Logger
}

type DailyStats struct {
	db  *pgxpool.Pool
	log *slog.Logger
}

func NewResults(db *pgxpool.Pool, log *slog.Logger) *Results {
	return &Results{
		db:  db,
		log: log.With("repo", "results"),
	}
}

func NewDailyStats(db *pgxpool.Pool, log *slog.Logger) *DailyStats {
	return &DailyStats{
		db:  db,
		log: log.With("repo", "results"),
	}
}

func (r *Results) Create(ctx context.Context, result model.CheckResult) error {
	r.log.DebugContext(ctx, "create check result", "SiteID", result.SiteID)

	const q = `
    INSERT INTO check_results (
        site_id, checked_at, is_up, status_code, response_time_ms, error
    )
    VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.Exec(ctx, q,
		result.SiteID,
		result.CheckedAt,
		result.IsUp,
		result.StatusCode,
		result.ResponseTimeMs,
		result.Error,
	)

	if err != nil {
		r.log.ErrorContext(ctx, "create check failed", "siteID", result.SiteID, "err", err)
		return fmt.Errorf("create check: %w", err)
	}

	return nil
}

func (r *DailyStats) RecordCheck(
	ctx context.Context,
	result model.CheckResult,
) error {
	date := result.CheckedAt.UTC()
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)

	failed := !result.IsUp

	const q = `
		INSERT INTO site_daily_stats (
			site_id,
			date,
			status,
			checks_total,
			checks_failed,
			uptime_percent,
			last_checked_at
		)
		VALUES (
			$1,
			$2,
			CASE WHEN $3 THEN 'up' ELSE 'degraded' END,
			1,
			CASE WHEN $3 THEN 0 ELSE 1 END,
			CASE WHEN $3 THEN 100.00 ELSE 0.00 END,
			$4
		)
		ON CONFLICT (site_id, date) DO UPDATE SET
			checks_total = site_daily_stats.checks_total + 1,
			checks_failed = site_daily_stats.checks_failed
				+ CASE WHEN $3 THEN 0 ELSE 1 END,
			uptime_percent = (
				(
					site_daily_stats.checks_total + 1
					- site_daily_stats.checks_failed
					- CASE WHEN $3 THEN 0 ELSE 1 END
				)::NUMERIC
				/ (site_daily_stats.checks_total + 1)
			) * 100,
			status = CASE
				WHEN site_daily_stats.checks_failed
					+ CASE WHEN $3 THEN 0 ELSE 1 END > 0
				THEN 'degraded'
				ELSE 'up'
			END,
			last_checked_at = EXCLUDED.last_checked_at
	`

	_, err := r.db.Exec(
		ctx,
		q,
		result.SiteID,
		date,
		!failed,
		result.CheckedAt,
	)
	if err != nil {
		return fmt.Errorf("record daily site check: %w", err)
	}

	return nil
}
