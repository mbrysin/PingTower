package repo

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mbrysin/PingTower/services/auth/model"
)

func setupTest(t *testing.T) (*Users, func()) {
	t.Helper()

	connString := os.Getenv("TEST_DATABASE_URL")
	if connString == "" {
		connString = "postgres://postgres:postgres@localhost:5432/testdb?sslmode=disable"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		t.Fatalf("не удалось подключиться к БД: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Fatalf("БД недоступна: %v", err)
	}

	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
	`)
	if err != nil {
		pool.Close()
		t.Fatalf("не удалось создать таблицу users: %v", err)
	}

	_, err = pool.Exec(ctx, `TRUNCATE TABLE users RESTART IDENTITY CASCADE;`)
	if err != nil {
		pool.Close()
		t.Fatalf("не удалось очистить таблицу users: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := NewUsers(pool, logger)

	cleanup := func() {
		_, _ = pool.Exec(ctx, `TRUNCATE TABLE users RESTART IDENTITY CASCADE;`)
		pool.Close()
	}

	return repo, cleanup
}

func TestUsers_Create(t *testing.T) {
	repo, cleanup := setupTest(t)
	defer cleanup()

	ctx := context.Background()

	u := &model.User{
		Username:     "alice",
		PasswordHash: "hash1",
	}

	created, err := repo.Create(ctx, u)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.Id == 0 {
		t.Error("Create() не установил Id")
	}
	if created.Username != "alice" {
		t.Errorf("Create() username = %s, want alice", created.Username)
	}
	if created.PasswordHash != "hash1" {
		t.Errorf("Create() password_hash = %s, want hash1", created.PasswordHash)
	}
	if created.CreatedAt.IsZero() {
		t.Error("Create() не установил CreatedAt")
	}

	u2 := &model.User{
		Username:     "alice",
		PasswordHash: "hash2",
	}
	_, err = repo.Create(ctx, u2)
	if !errors.Is(err, ErrUsernameTaken) {
		t.Errorf("Create() duplicate error = %v, want ErrUsernameTaken", err)
	}
}

func TestUsers_Get(t *testing.T) {
	repo, cleanup := setupTest(t)
	defer cleanup()

	ctx := context.Background()

	u := &model.User{Username: "bob", PasswordHash: "hash"}
	created, err := repo.Create(ctx, u)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := repo.Get(ctx, created.Id)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Id != created.Id || got.Username != created.Username {
		t.Errorf("Get() = %+v, want %+v", got, created)
	}

	_, err = repo.Get(ctx, 9999)
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("Get() error = %v, want ErrUserNotFound", err)
	}
}

func TestUsers_GetByUsername(t *testing.T) {
	repo, cleanup := setupTest(t)
	defer cleanup()

	ctx := context.Background()

	u := &model.User{Username: "Charlie", PasswordHash: "hash"}
	created, err := repo.Create(ctx, u)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := repo.GetByUsername(ctx, "charlie")
	if err != nil {
		t.Fatalf("GetByUsername() error = %v", err)
	}
	if got.Id != created.Id {
		t.Errorf("GetByUsername() id = %d, want %d", got.Id, created.Id)
	}

	_, err = repo.GetByUsername(ctx, "nonexistent")
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("GetByUsername() error = %v, want ErrUserNotFound", err)
	}
}

func TestUsers_List(t *testing.T) {
	repo, cleanup := setupTest(t)
	defer cleanup()

	ctx := context.Background()

	users := []*model.User{
		{Username: "user1", PasswordHash: "h1"},
		{Username: "user2", PasswordHash: "h2"},
		{Username: "user3", PasswordHash: "h3"},
		{Username: "user4", PasswordHash: "h4"},
		{Username: "user5", PasswordHash: "h5"},
	}
	for _, u := range users {
		_, err := repo.Create(ctx, u)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		time.Sleep(time.Millisecond)
	}

	list, err := repo.List(ctx, ListParams{})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(list) != 5 {
		t.Errorf("List() len = %d, want 5", len(list))
	}

	if list[0].Username != "user5" {
		t.Errorf("List() first = %s, want user5", list[0].Username)
	}

	list, err = repo.List(ctx, ListParams{Limit: 2, Offset: 1})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(list) != 2 {
		t.Errorf("List() len = %d, want 2", len(list))
	}
	if list[0].Username != "user4" || list[1].Username != "user3" {
		t.Errorf("List() = %v, %v, want user4, user3", list[0].Username, list[1].Username)
	}

	list, err = repo.List(ctx, ListParams{Offset: -10})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(list) != 5 {
		t.Errorf("List() with negative offset len = %d, want 5", len(list))
	}
}

func TestUsers_UpdateUsername(t *testing.T) {
	repo, cleanup := setupTest(t)
	defer cleanup()

	ctx := context.Background()

	u := &model.User{Username: "dave", PasswordHash: "hash"}
	created, err := repo.Create(ctx, u)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	updated, err := repo.UpdateUsername(ctx, created.Id, "dave_new")
	if err != nil {
		t.Fatalf("UpdateUsername() error = %v", err)
	}
	if updated.Username != "dave_new" {
		t.Errorf("UpdateUsername() username = %s, want dave_new", updated.Username)
	}

	_, err = repo.UpdateUsername(ctx, 9999, "nobody")
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("UpdateUsername() error = %v, want ErrUserNotFound", err)
	}

	u2 := &model.User{Username: "eve", PasswordHash: "hash"}
	_, err = repo.Create(ctx, u2)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	_, err = repo.UpdateUsername(ctx, created.Id, "eve")
	if !errors.Is(err, ErrUsernameTaken) {
		t.Errorf("UpdateUsername() error = %v, want ErrUsernameTaken", err)
	}
}

func TestUsers_UpdatePassword(t *testing.T) {
	repo, cleanup := setupTest(t)
	defer cleanup()

	ctx := context.Background()

	u := &model.User{Username: "frank", PasswordHash: "old_hash"}
	created, err := repo.Create(ctx, u)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	updated, err := repo.UpdatePassword(ctx, created.Id, "new_hash")
	if err != nil {
		t.Fatalf("UpdatePassword() error = %v", err)
	}
	if updated.PasswordHash != "new_hash" {
		t.Errorf("UpdatePassword() password_hash = %s, want new_hash", updated.PasswordHash)
	}

	_, err = repo.UpdatePassword(ctx, 9999, "hash")
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("UpdatePassword() error = %v, want ErrUserNotFound", err)
	}
}

func TestUsers_Delete(t *testing.T) {
	repo, cleanup := setupTest(t)
	defer cleanup()

	ctx := context.Background()

	u := &model.User{Username: "grace", PasswordHash: "hash"}
	created, err := repo.Create(ctx, u)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	err = repo.Delete(ctx, created.Id)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	err = repo.Delete(ctx, created.Id)
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("Delete() error = %v, want ErrUserNotFound", err)
	}
}
