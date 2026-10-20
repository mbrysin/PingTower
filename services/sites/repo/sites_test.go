package repo

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mbrysin/PingTower/services/sites/model"
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
			enabled BOOL NOT NULL DEFAULT FALSE,
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
		URL:  "https://example.com",
	}

	created, err := repo.Create(ctx, site)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if created.ID == 0 {
		t.Error("Create() не установил ID")
	}
	if created.Name != "Example" {
		t.Errorf("Create() name = %q, want %q", created.Name, "Example")
	}
	if created.URL != "https://example.com" {
		t.Errorf("Create() url = %q, want %q", created.URL, "https://example.com")
	}
	if created.CreatedAt.IsZero() {
		t.Error("Create() не установил CreatedAt")
	}
}

func TestSites_Create_DuplicateURL(t *testing.T) {
	repo, cleanup := setupSitesTest(t)
	defer cleanup()

	ctx := context.Background()

	first := &model.Site{Name: "First", URL: "https://example.com"}
	if _, err := repo.Create(ctx, first); err != nil {
		t.Fatalf("первый Create() error = %v", err)
	}

	second := &model.Site{Name: "Second", URL: "https://example.com"}
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
		URL:  "https://example.com",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := repo.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if got.ID != created.ID {
		t.Errorf("Get() id = %d, want %d", got.ID, created.ID)
	}
	if got.Name != created.Name {
		t.Errorf("Get() name = %q, want %q", got.Name, created.Name)
	}
	if got.URL != created.URL {
		t.Errorf("Get() url = %q, want %q", got.URL, created.URL)
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

func TestSites_List(t *testing.T) {
	repo, cleanup := setupSitesTest(t)
	defer cleanup()

	ctx := context.Background()

	first, err := repo.Create(ctx, &model.Site{
		Name: "First",
		URL:  "https://first.example",
	})
	if err != nil {
		t.Fatalf("Create(first) error = %v", err)
	}

	second, err := repo.Create(ctx, &model.Site{
		Name: "Second",
		URL:  "https://second.example",
	})
	if err != nil {
		t.Fatalf("Create(second) error = %v", err)
	}

	sites, err := repo.List(ctx, 10, 0)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(sites) != 2 {
		t.Fatalf("List() вернул %d сайтов, want 2", len(sites))
	}
	if sites[0].ID != first.ID {
		t.Errorf("sites[0].ID = %d, want %d", sites[0].ID, first.ID)
	}
	if sites[1].ID != second.ID {
		t.Errorf("sites[1].ID = %d, want %d", sites[1].ID, second.ID)
	}
	if sites[0].URL != first.URL {
		t.Errorf("sites[0].URL = %q, want %q", sites[0].URL, first.URL)
	}
}

func TestSites_List_Pagination(t *testing.T) {
	repo, cleanup := setupSitesTest(t)
	defer cleanup()

	ctx := context.Background()

	for i, site := range []model.Site{
		{Name: "First", URL: "https://first.example"},
		{Name: "Second", URL: "https://second.example"},
		{Name: "Third", URL: "https://third.example"},
	} {
		if _, err := repo.Create(ctx, &site); err != nil {
			t.Fatalf("Create(%d) error = %v", i, err)
		}
	}

	sites, err := repo.List(ctx, 1, 1)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(sites) != 1 {
		t.Fatalf("List() вернул %d сайтов, want 1", len(sites))
	}
	if sites[0].Name != "Second" {
		t.Errorf("List() вернул сайт %q, want %q", sites[0].Name, "Second")
	}
}

func TestSites_List_Empty(t *testing.T) {
	repo, cleanup := setupSitesTest(t)
	defer cleanup()

	sites, err := repo.List(context.Background(), 10, 0)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(sites) != 0 {
		t.Errorf("List() вернул %d сайтов, want 0", len(sites))
	}
}

func TestSites_List_InvalidPagination(t *testing.T) {
	repo, cleanup := setupSitesTest(t)
	defer cleanup()

	tests := []struct {
		name   string
		limit  int
		offset int
	}{
		{name: "zero limit", limit: 0, offset: 0},
		{name: "negative limit", limit: -1, offset: 0},
		{name: "negative offset", limit: 10, offset: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := repo.List(context.Background(), tt.limit, tt.offset)
			if err == nil {
				t.Fatal("List() ожидалась ошибка для некорректной пагинации")
			}
		})
	}
}

func TestSites_UpdateName(t *testing.T) {
	repo, cleanup := setupSitesTest(t)
	defer cleanup()

	ctx := context.Background()
	created, err := repo.Create(ctx, &model.Site{
		Name: "Old name",
		URL:  "https://example.com",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	updated, err := repo.UpdateName(ctx, created.ID, "New name")
	if err != nil {
		t.Fatalf("UpdateName() error = %v", err)
	}

	if updated.ID != created.ID {
		t.Errorf("UpdateName() id = %d, want %d", updated.ID, created.ID)
	}
	if updated.Name != "New name" {
		t.Errorf("UpdateName() name = %q, want %q", updated.Name, "New name")
	}
	if updated.URL != created.URL {
		t.Errorf("UpdateName() url = %q, want %q", updated.URL, created.URL)
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

func TestSites_SetEnabled(t *testing.T) {
	repo, cleanup := setupSitesTest(t)
	defer cleanup()

	ctx := context.Background()
	created, err := repo.Create(ctx, &model.Site{
		Name: "Example",
		URL:  "https://example.com",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.Enabled {
		t.Fatal("новый сайт должен быть выключен")
	}

	enabled, err := repo.SetEnabled(ctx, created.ID, true)
	if err != nil {
		t.Fatalf("SetEnabled(true) error = %v", err)
	}
	if !enabled.Enabled {
		t.Error("SetEnabled(true) не включил мониторинг")
	}

	disabled, err := repo.SetEnabled(ctx, created.ID, false)
	if err != nil {
		t.Fatalf("SetEnabled(false) error = %v", err)
	}
	if disabled.Enabled {
		t.Error("SetEnabled(false) не выключил мониторинг")
	}
}

func TestSites_SetEnabled_NotFound(t *testing.T) {
	repo, cleanup := setupSitesTest(t)
	defer cleanup()

	_, err := repo.SetEnabled(context.Background(), 999999, true)
	if !errors.Is(err, ErrSiteNotFound) {
		t.Errorf("SetEnabled() error = %v, want ErrSiteNotFound", err)
	}
}

func TestSites_Delete(t *testing.T) {
	repo, cleanup := setupSitesTest(t)
	defer cleanup()

	ctx := context.Background()
	created, err := repo.Create(ctx, &model.Site{
		Name: "Example",
		URL:  "https://example.com",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := repo.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err = repo.Get(ctx, created.ID)
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
