package repo

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mbrysin/PingTower/services/sites/model"
)

var (
	ErrSiteNotFound = errors.New("site not found")
	ErrSiteExists   = errors.New("site exists")
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

func (r *Sites) Create(ctx context.Context, s *model.Site) (*model.Site, error) {
	r.log.DebugContext(ctx, "create site", "url", s.URL)

	const q = `
		INSERT INTO sites (name, url)
		VALUES ($1, $2)
		RETURNING id, name, url, enabled, created_at
	`

	err := r.db.QueryRow(ctx, q, s.Name, s.URL).Scan(
		&s.ID,
		&s.Name,
		&s.URL,
		&s.Enabled,
		&s.CreatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrSiteExists
		}
		r.log.ErrorContext(ctx, "create site failed", "url", s.URL, "err", err)
		return nil, fmt.Errorf("create site: %w", err)
	}

	return s, nil
}

func (r *Sites) Get(ctx context.Context, id int) (*model.Site, error) {
	r.log.DebugContext(ctx, "get site", "id", id)

	const q = `
		SELECT id, name, url, enabled, created_at
		FROM sites
		WHERE id = $1
	`
	var s model.Site

	err := r.db.QueryRow(ctx, q, id).Scan(
		&s.ID,
		&s.Name,
		&s.URL,
		&s.Enabled,
		&s.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSiteNotFound
		}
		r.log.ErrorContext(ctx, "get site failed", "id", id, "err", err)
		return nil, fmt.Errorf("get site %d: %w", id, err)
	}

	return &s, nil
}

func (r *Sites) List(
	ctx context.Context,
	limit int,
	offset int,
) ([]model.Site, error) {
	r.log.DebugContext(ctx, "list sites", "limit", limit, "offset", offset)

	if limit <= 0 {
		return nil, fmt.Errorf("list sites: limit must be greater than 0")
	}
	if offset < 0 {
		return nil, fmt.Errorf("list sites: offset must not be negative")
	}

	const q = `
		SELECT id, name, url, enabled, created_at
		FROM sites
		ORDER BY id
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, q, limit, offset)
	if err != nil {
		r.log.ErrorContext(ctx, "list sites query failed", "err", err)
		return nil, fmt.Errorf("list sites: query: %w", err)
	}
	defer rows.Close()

	sites := make([]model.Site, 0)
	for rows.Next() {
		var s model.Site
		if err := rows.Scan(
			&s.ID,
			&s.Name,
			&s.URL,
			&s.Enabled,
			&s.CreatedAt,
		); err != nil {
			r.log.ErrorContext(ctx, "list sites scan failed", "err", err)
			return nil, fmt.Errorf("list sites: scan: %w", err)
		}
		sites = append(sites, s)
	}

	if err := rows.Err(); err != nil {
		r.log.ErrorContext(ctx, "list sites rows failed", "err", err)
		return nil, fmt.Errorf("list sites: rows: %w", err)
	}

	return sites, nil
}

func (r *Sites) UpdateName(ctx context.Context, id int, name string) (*model.Site, error) {
	r.log.DebugContext(ctx, "update name", "id", id, "name", name)

	const q = `
		UPDATE sites
		SET name = $2
		WHERE id = $1
		RETURNING id, name, url, enabled, created_at
	`
	var s model.Site

	err := r.db.QueryRow(ctx, q, id, name).Scan(
		&s.ID,
		&s.Name,
		&s.URL,
		&s.Enabled,
		&s.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSiteNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrSiteExists
		}
		r.log.ErrorContext(ctx, "update site failed", "name", s.Name, "err", err)
		return nil, fmt.Errorf("update site: %w", err)
	}

	return &s, err
}

func (r *Sites) SetEnabled(ctx context.Context, id int, enabled bool) (*model.Site, error) {
	r.log.DebugContext(ctx, "set site enabled", "id", id, "enabled", enabled)

	const q = `
		UPDATE sites
		SET enabled = $2
		WHERE id = $1
		RETURNING id, name, url, enabled, created_at
	`

	var s model.Site
	err := r.db.QueryRow(ctx, q, id, enabled).Scan(
		&s.ID,
		&s.Name,
		&s.URL,
		&s.Enabled,
		&s.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSiteNotFound
		}
		r.log.ErrorContext(ctx, "set site enabled failed",
			"id", id,
			"enabled", enabled,
			"err", err,
		)
		return nil, fmt.Errorf("set site enabled: %w", err)
	}

	return &s, nil
}

func (r *Sites) Delete(ctx context.Context, id int) error {
	r.log.DebugContext(ctx, "delete site", "id", id)

	const q = `DELETE FROM sites WHERE id = $1`

	tag, err := r.db.Exec(ctx, q, id)
	if err != nil {
		r.log.ErrorContext(ctx, "delete site failed", "id", id, "err", err)
		return fmt.Errorf("delete site %d: %w", id, err)
	}

	if tag.RowsAffected() == 0 {
		return ErrSiteNotFound
	}

	return nil
}
