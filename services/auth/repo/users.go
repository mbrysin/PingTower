package repo

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mbrysin/PingTower/services/auth/model"
)

const (
	defaultListLimit = 50
	maxListLimit     = 500
)

var (
	ErrUserNotFound  = errors.New("user not found")
	ErrUsernameTaken = errors.New("username already taken")
)

type Users struct {
	db  *pgxpool.Pool
	log *slog.Logger
}

type ListParams struct {
	Limit  int
	Offset int
}

func NewUsers(db *pgxpool.Pool, log *slog.Logger) *Users {
	return &Users{
		db:  db,
		log: log.With("repo", "users"),
	}
}

func (r *Users) Create(ctx context.Context, u *model.User) (*model.User, error) {
	r.log.DebugContext(ctx, "create user", "username", u.Username)

	const q = `
		INSERT INTO users (username, password_hash)
		VALUES ($1, $2)
		RETURNING id, username, created_at
	`

	err := r.db.QueryRow(ctx, q, u.Username, u.PasswordHash).Scan(
		&u.Id,
		&u.Username,
		&u.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrUsernameTaken
		}
		r.log.ErrorContext(ctx, "create user failed", "username", u.Username, "err", err)
		return nil, fmt.Errorf("create user: %w", err)
	}

	return u, nil
}

func (r *Users) Get(ctx context.Context, id int) (*model.User, error) {
	r.log.DebugContext(ctx, "get user", "id", id)

	const q = `
		SELECT id, username, password_hash, created_at
		FROM users
		WHERE id = $1
	`

	var u model.User
	err := r.db.QueryRow(ctx, q, id).Scan(
		&u.Id,
		&u.Username,
		&u.PasswordHash,
		&u.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		r.log.ErrorContext(ctx, "get user failed", "id", id, "err", err)
		return nil, fmt.Errorf("get user %d: %w", id, err)
	}

	return &u, nil
}

func (r *Users) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	r.log.DebugContext(ctx, "get user by username", "username", username)

	const q = `
		SELECT id, username, password_hash, created_at
		FROM users
		WHERE lower(username) = lower($1)
	`

	var u model.User
	err := r.db.QueryRow(ctx, q, username).Scan(
		&u.Id,
		&u.Username,
		&u.PasswordHash,
		&u.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		r.log.ErrorContext(ctx, "get user failed", "username", username, "err", err)
		return nil, fmt.Errorf("get user %s: %w", username, err)
	}

	return &u, nil
}

func (r *Users) List(ctx context.Context, p ListParams) ([]*model.User, error) {
	limit, offset := p.Limit, p.Offset
	r.log.DebugContext(ctx, "list users", "limit", limit, "offset", offset)

	if limit <= 0 {
		limit = defaultListLimit
	}
	if limit > maxListLimit {
		limit = maxListLimit
	}
	if offset < 0 {
		offset = 0
	}

	const q = `
		SELECT id, username, password_hash, created_at
		FROM users
		ORDER BY created_at DESC, id DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, q, limit, offset)
	if err != nil {
		r.log.ErrorContext(ctx, "list users failed", "limit", limit, "offset", offset, "err", err)
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	users := make([]*model.User, 0, limit)
	for rows.Next() {
		var u model.User
		if err := rows.Scan(
			&u.Id,
			&u.Username,
			&u.PasswordHash,
			&u.CreatedAt,
		); err != nil {
			r.log.ErrorContext(ctx, "scan user failed", "err", err)
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, &u)
	}

	if err := rows.Err(); err != nil {
		r.log.ErrorContext(ctx, "iterate users failed", "err", err)
		return nil, fmt.Errorf("iterate users: %w", err)
	}

	return users, nil
}

func (r *Users) UpdateUsername(ctx context.Context, id int, username string) (*model.User, error) {
	r.log.DebugContext(ctx, "update user username", "id", id)

	const q = `
		UPDATE users
		SET username = $2
		WHERE id = $1
		RETURNING id, username, password_hash, created_at
	`

	var u model.User
	err := r.db.QueryRow(ctx, q, id, username).Scan(
		&u.Id,
		&u.Username,
		&u.PasswordHash,
		&u.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrUsernameTaken
		}
		r.log.ErrorContext(ctx, "update user username failed", "id", id, "err", err)
		return nil, fmt.Errorf("update user username %d: %w", id, err)
	}

	return &u, nil
}

func (r *Users) UpdatePassword(ctx context.Context, id int, passwordHash string) (*model.User, error) {
	r.log.DebugContext(ctx, "update user password", "id", id)

	const q = `
		UPDATE users
		SET password_hash = $2
		WHERE id = $1
		RETURNING id, username, password_hash, created_at
	`

	var u model.User
	err := r.db.QueryRow(ctx, q, id, passwordHash).Scan(
		&u.Id,
		&u.Username,
		&u.PasswordHash,
		&u.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		r.log.ErrorContext(ctx, "update user password failed", "id", id, "err", err)
		return nil, fmt.Errorf("update user password %d: %w", id, err)
	}

	return &u, nil
}

func (r *Users) Delete(ctx context.Context, id int) error {
	r.log.DebugContext(ctx, "delete user", "id", id)

	const q = `DELETE FROM users WHERE id = $1`

	tag, err := r.db.Exec(ctx, q, id)
	if err != nil {
		r.log.ErrorContext(ctx, "delete user failed", "id", id, "err", err)
		return fmt.Errorf("delete user %d: %w", id, err)
	}

	if tag.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}
