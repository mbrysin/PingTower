package repo

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mbrysin/PingTower/services/stats/model"
)

var (
	ErrStatNotFound = errors.New("stat not found")
)

type Stats struct {
	db  *pgxpool.Pool
	log *slog.Logger
}

func NewStats(db *pgxpool.Pool, log *slog.Logger) *Stats {
	return &Stats{
		db:  db,
		log: log.With("repo", "stats"),
	}
}

func (r *Stats) GetStat(ctx context.Context, stat_id int) (*model.Stat, error) {
	r.log.With("repo", "get", "id", stat_id)

	const q = `
		SELECT id, site_id, checked_at, is_up, status_code, response_time_ms, error from stat
		WHERE id = $1
	`

	var s model.Stat
	err := r.db.QueryRow(ctx, q, stat_id).Scan(
		&s.ID,
		&s.SiteID,
		&s.CheckedAt,
		&s.IsUp,
		&s.StatusCode,
		&s.ResponseTimeMs,
		&s.Error,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrStatNotFound
		}
		r.log.ErrorContext(ctx, "get stat failed", "id", stat_id, "err", err)
		return nil, fmt.Errorf("get stat %d: %w", stat_id, err)
	}

	return &s, nil
}

func (r *Stats) GetStatsByHour(
	ctx context.Context,
	siteID int,
	at time.Time,
) ([]model.Stat, error) {
	r.log.With("repo", "get_stats_by_hour", "siteID", siteID)

	start := at.Truncate(time.Hour)
	end := start.Add(time.Hour)

	const q = `
		SELECT id, site_id, checked_at, is_up, status_code,
		       response_time_ms, error
		FROM stat
		WHERE site_id = $1
		  AND checked_at >= $2
		  AND checked_at < $3
		ORDER BY checked_at
	`

	rows, err := r.db.Query(ctx, q, siteID, start, end)
	if err != nil {
		r.log.ErrorContext(ctx, "get hourly stats failed",
			"site_id", siteID, "start", start, "err", err)
		return nil, fmt.Errorf("get hourly stats for site %d: %w", siteID, err)
	}
	defer rows.Close()

	stats := make([]model.Stat, 0)
	for rows.Next() {
		var s model.Stat
		if err := rows.Scan(
			&s.ID,
			&s.SiteID,
			&s.CheckedAt,
			&s.IsUp,
			&s.StatusCode,
			&s.ResponseTimeMs,
			&s.Error,
		); err != nil {
			return nil, fmt.Errorf("scan hourly stat: %w", err)
		}
		stats = append(stats, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate hourly stats: %w", err)
	}

	return stats, nil
}

func (r *Stats) GetStatsBetween(
	ctx context.Context,
	siteID int,
	from time.Time,
	to time.Time,
) ([]model.Stat, error) {
	r.log.With("repo", "get_stats_between", "siteID", siteID)
	if !from.Before(to) {
		return nil, fmt.Errorf("invalid time range: from must be before to")
	}

	const q = `
		SELECT id, site_id, checked_at, is_up, status_code,
		       response_time_ms, error
		FROM stat
		WHERE site_id = $1
		  AND checked_at >= $2
		  AND checked_at < $3
		ORDER BY checked_at
	`

	rows, err := r.db.Query(ctx, q, siteID, from, to)
	if err != nil {
		r.log.ErrorContext(ctx, "get stats between dates failed",
			"site_id", siteID, "from", from, "to", to, "err", err)
		return nil, fmt.Errorf("get stats for site %d between dates: %w", siteID, err)
	}
	defer rows.Close()

	stats := make([]model.Stat, 0)
	for rows.Next() {
		var s model.Stat
		if err := rows.Scan(
			&s.ID,
			&s.SiteID,
			&s.CheckedAt,
			&s.IsUp,
			&s.StatusCode,
			&s.ResponseTimeMs,
			&s.Error,
		); err != nil {
			return nil, fmt.Errorf("scan stat: %w", err)
		}
		stats = append(stats, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate stats: %w", err)
	}

	return stats, nil
}

func (r *Stats) GetLatest(
	ctx context.Context,
	siteID int,
) (*model.Stat, error) {
	const query = `
		SELECT id, site_id, checked_at, is_up, status_code,
		       response_time_ms, error
		FROM stat
		WHERE site_id = $1
		ORDER BY checked_at DESC, id DESC
		LIMIT 1
	`

	var stat model.Stat
	err := r.db.QueryRow(ctx, query, siteID).Scan(
		&stat.ID,
		&stat.SiteID,
		&stat.CheckedAt,
		&stat.IsUp,
		&stat.StatusCode,
		&stat.ResponseTimeMs,
		&stat.Error,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query latest stat for site %d: %w", siteID, err)
	}

	return &stat, nil
}
