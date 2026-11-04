package repo

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
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

func (r *Stats) SelectStat(ctx context.Context)
