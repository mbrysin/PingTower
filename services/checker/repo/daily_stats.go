package repo

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DailyStats struct {
	db  *pgxpool.Pool
	log *slog.Logger
}

func NewDailyStats(db *pgxpool.Pool, log *slog.Logger) *DailyStats {
	return &DailyStats{
		db:  db,
		log: log.With("repo", "daily_stats"),
	}
}

func (r *DailyStats) AggregateDay(ctx context.Context, day time.Time) error {
	day = time.Date(day.UTC().Year(), day.UTC().Month(), day.UTC().Day(), 0, 0, 0, 0, time.UTC)

	const q = `
		INSERT INTO daily_stat (
			site_id,
			stat_date,
			checks,
			failed_checks,
			success_percent,
			last_check,
			avg_response_time_ms,
			min_response_time_ms,
			max_response_time_ms
		)
		SELECT
			site_id,
			$1::date,
			COUNT(*),
			COUNT(*) FILTER (WHERE NOT is_up),
			100.0 * COUNT(*) FILTER (WHERE is_up) / NULLIF(COUNT(*), 0),
			MAX(checked_at),
			AVG(response_time_ms),
			MIN(response_time_ms),
			MAX(response_time_ms)
		FROM stat
		WHERE checked_at >= $1
		  AND checked_at < $1 + INTERVAL '1 day'
		GROUP BY site_id
		ON CONFLICT (site_id, stat_date) DO UPDATE SET
			checks = EXCLUDED.checks,
			failed_checks = EXCLUDED.failed_checks,
			success_percent = EXCLUDED.success_percent,
			last_check = EXCLUDED.last_check,
			avg_response_time_ms = EXCLUDED.avg_response_time_ms,
			min_response_time_ms = EXCLUDED.min_response_time_ms,
			max_response_time_ms = EXCLUDED.max_response_time_ms
	`

	if _, err := r.db.Exec(ctx, q, day); err != nil {
		r.log.ErrorContext(ctx, "aggregate daily stats failed", "day", day, "err", err)
		return fmt.Errorf("aggregate daily stats for %s: %w", day.Format("2006-01-02"), err)
	}

	return nil
}
