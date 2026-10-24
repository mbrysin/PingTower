package repo

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mbrysin/PingTower/services/checker/model"
)

type Sites struct {
	db  *pgxpool.Pool
	log *slog.Logger
}

func NewSites(db *pgxpool.Pool, log *slog.Logger) *Sites {
	return &Sites{
		db:  db,
		log: log.With("repo", "sites"),
	}
}

func (r *Sites) GetEnabled(ctx context.Context) ([]model.Site, error) {
	r.log.DebugContext(ctx, "get enabled sites")

	const q = `
		SELECT id, name, url, enabled
		FROM sites
		WHERE enabled = TRUE
		ORDER BY id
	`

	rows, err := r.db.Query(ctx, q)
	if err != nil {
		r.log.ErrorContext(ctx, "get sites enabled failed", "err", err)
		return nil, fmt.Errorf("get sites enabled: %w", err)
	}
	defer rows.Close()

	sites := make([]model.Site, 0)

	for rows.Next() {
		var site model.Site

		if err := rows.Scan(
			&site.ID,
			&site.Name,
			&site.URL,
			&site.Enabled,
		); err != nil {
			return nil, fmt.Errorf("scan site: %w", err)
		}

		sites = append(sites, site)
	}

	if err := rows.Err(); err != nil {
		r.log.ErrorContext(ctx, "iterate sites failed", "err", err)
		return nil, fmt.Errorf("iterate sites: %w", err)
	}

	return sites, nil
}
