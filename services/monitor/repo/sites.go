package repo

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mbrysin/PingTower/services/monitor/model"
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
	r.log.DebugContext(ctx, "create site", "url", s.Url)

	const q = `
		INSERT INTO sites (name, url)
		VALUES ($1, $2)
		RETURNING id, name, url, created_at
	`

	err := r.db.QueryRow(ctx, q, s.Name, s.Url).Scan(
		&s.Id,
		&s.Name,
		&s.Url,
		&s.CreatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrSiteExists
		}
		r.log.ErrorContext(ctx, "create site failed", "url", s.Url, "err", err)
		return nil, fmt.Errorf("create site: %w", err)
	}

	return s, nil
}

func (r *Sites) Get(ctx context.Context, id int) (*model.Site, error) {
	r.log.DebugContext(ctx, "get site", "id", id)

	const q = `
		SELECT id, name, url, created_at
		FROM sites
		WHERE id = $1
	`
	var s model.Site

	err := r.db.QueryRow(ctx, q, id).Scan(
		&s.Id,
		&s.Name,
		&s.Url,
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

func (r *Sites) UpdateName(ctx context.Context, id int, name string) (*model.Site, error) {
	r.log.DebugContext(ctx, "update name", "id", id, "name", name)

	const q = `
		UPDATE sites
		SET name = $2
		WHERE id = $1
		RETURNING id, name, url, created_at
	`
	var s model.Site

	err := r.db.QueryRow(ctx, q, id, name).Scan(
		&s.Id,
		&s.Name,
		&s.Url,
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
