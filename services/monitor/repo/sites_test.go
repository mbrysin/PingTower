package repo

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mbrysin/PingTower/services/monitor/model"
)

func setupSitesTest(t *testing.T) (*Sites, func()) {
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
		CREATE TABLE IF NOT EXISTS sites (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			url TEXT NOT NULL UNIQUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
	`)
	if err != nil {
		pool.Close()
		t.Fatalf("не удалось создать таблицу sites: %v", err)
	}

	_, err = pool.Exec(ctx, `TRUNCATE TABLE sites RESTART IDENTITY CASCADE;`)
	if err != nil {
		pool.Close()
		t.Fatalf("не удалось очистить таблицу sites: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := NewSites(pool, logger)

	cleanup := func() {
		_, _ = pool.Exec(ctx, `TRUNCATE TABLE sites RESTART IDENTITY CASCADE;`)
		pool.Close()
	}

	return repo, cleanup
}

func TestSites_Create(t *testing.T) {
	repo, cleanup := setupSitesTest(t)
	defer cleanup()

	ctx := context.Background()
	site := &model.Site{
		Name: "Example",
		Url:  "https://example.com",
	}

	created, err := repo.Create(ctx, site)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if created.Id == 0 {
		t.Error("Create() не установил Id")
	}
	if created.Name != "Example" {
		t.Errorf("Create() name = %q, want %q", created.Name, "Example")
	}
	if created.Url != "https://example.com" {
		t.Errorf("Create() url = %q, want %q", created.Url, "https://example.com")
	}
	if created.CreatedAt.IsZero() {
		t.Error("Create() не установил CreatedAt")
	}
}

func TestSites_Create_DuplicateURL(t *testing.T) {
	repo, cleanup := setupSitesTest(t)
	defer cleanup()

	ctx := context.Background()

	first := &model.Site{Name: "First", Url: "https://example.com"}
	if _, err := repo.Create(ctx, first); err != nil {
		t.Fatalf("первый Create() error = %v", err)
	}

	second := &model.Site{Name: "Second", Url: "https://example.com"}
	_, err := repo.Create(ctx, second)
	if !errors.Is(err, ErrSiteExists) {
		t.Errorf("повторный Create() error = %v, want ErrSiteExists", err)
	}
}

func TestSites_Get(t *testing.T) {
	repo, cleanup := setupSitesTest(t)
	defer cleanup()

	ctx := context.Background()
	created, err := repo.Create(ctx, &model.Site{
		Name: "Example",
		Url:  "https://example.com",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := repo.Get(ctx, created.Id)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if got.Id != created.Id {
		t.Errorf("Get() id = %d, want %d", got.Id, created.Id)
	}
	if got.Name != created.Name {
		t.Errorf("Get() name = %q, want %q", got.Name, created.Name)
	}
	if got.Url != created.Url {
		t.Errorf("Get() url = %q, want %q", got.Url, created.Url)
	}
	if !got.CreatedAt.Equal(created.CreatedAt) {
		t.Errorf("Get() created_at = %v, want %v", got.CreatedAt, created.CreatedAt)
	}
}

func TestSites_Get_NotFound(t *testing.T) {
	repo, cleanup := setupSitesTest(t)
	defer cleanup()

	_, err := repo.Get(context.Background(), 999999)
	if !errors.Is(err, ErrSiteNotFound) {
		t.Errorf("Get() error = %v, want ErrSiteNotFound", err)
	}
}

func TestSites_UpdateName(t *testing.T) {
	repo, cleanup := setupSitesTest(t)
	defer cleanup()

	ctx := context.Background()
	created, err := repo.Create(ctx, &model.Site{
		Name: "Old name",
		Url:  "https://example.com",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	updated, err := repo.UpdateName(ctx, created.Id, "New name")
	if err != nil {
		t.Fatalf("UpdateName() error = %v", err)
	}

	if updated.Id != created.Id {
		t.Errorf("UpdateName() id = %d, want %d", updated.Id, created.Id)
	}
	if updated.Name != "New name" {
		t.Errorf("UpdateName() name = %q, want %q", updated.Name, "New name")
	}
	if updated.Url != created.Url {
		t.Errorf("UpdateName() url = %q, want %q", updated.Url, created.Url)
	}
}

func TestSites_UpdateName_NotFound(t *testing.T) {
	repo, cleanup := setupSitesTest(t)
	defer cleanup()

	_, err := repo.UpdateName(context.Background(), 999999, "New name")
	if !errors.Is(err, ErrSiteNotFound) {
		t.Errorf("UpdateName() error = %v, want ErrSiteNotFound", err)
	}
}

func TestSites_Delete(t *testing.T) {
	repo, cleanup := setupSitesTest(t)
	defer cleanup()

	ctx := context.Background()
	created, err := repo.Create(ctx, &model.Site{
		Name: "Example",
		Url:  "https://example.com",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := repo.Delete(ctx, created.Id); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err = repo.Get(ctx, created.Id)
	if !errors.Is(err, ErrSiteNotFound) {
		t.Errorf("Get() после Delete() error = %v, want ErrSiteNotFound", err)
	}
}

func TestSites_Delete_NotFound(t *testing.T) {
	repo, cleanup := setupSitesTest(t)
	defer cleanup()

	err := repo.Delete(context.Background(), 999999)
	if !errors.Is(err, ErrSiteNotFound) {
		t.Errorf("Delete() error = %v, want ErrSiteNotFound", err)
	}
}
