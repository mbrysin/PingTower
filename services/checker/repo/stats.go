package repo

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mbrysin/PingTower/services/checker/model"
)

var (
	ErrStatExists = errors.New("stat exists")
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

func (r *Stats) CreateStat(ctx context.Context, s *model.Stat) (*model.Stat, error) {
	r.log.DebugContext(ctx, "create stat", "site_id", s.SiteID)

	const q = `
		INSERT INTO stat (site_id, checked_at, is_up, status_code, reponse_time_ms, error)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	err := r.db.QueryRow(
		ctx,
		q,
		s.SiteID,
		s.CheckedAt,
		s.IsUp,
		s.StatusCode,
		s.ResponseTimeMs,
		s.Error,
	).Scan(&s.ID)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrStatExists
		}
		r.log.ErrorContext(ctx, "create stat failed", "site_id", s.SiteID, "err", err)
		return nil, fmt.Errorf("create stat: %w", err)
	}

	return s, nil
}
