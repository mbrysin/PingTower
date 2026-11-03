package repo

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mbrysin/PingTower/services/checker/model"
)

func setupCheckerSitesTest(t *testing.T) (*Sites, func()) {
	t.Helper()

	connString := os.Getenv("TEST_DATABASE_URL")
	if connString == "" {
		t.Skip("задай TEST_DATABASE_URL для интеграционных тестов PostgreSQL")
	}

	ctx := context.Background()
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		t.Fatalf("разобрать TEST_DATABASE_URL: %v", err)
	}

	schema := fmt.Sprintf("checker_test_%d", os.Getpid())
	config.ConnConfig.RuntimeParams["search_path"] = schema

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatalf("создать пул PostgreSQL: %v", err)
	}
	t.Cleanup(pool.Close)

	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("подключиться к PostgreSQL: %v", err)
	}

	// Создаём схему через соединение без search_path на неё.
	adminConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		t.Fatalf("разобрать TEST_DATABASE_URL: %v", err)
	}
	admin, err := pgxpool.NewWithConfig(ctx, adminConfig)
	if err != nil {
		t.Fatalf("создать административный пул: %v", err)
	}
	t.Cleanup(admin.Close)

	if _, err := admin.Exec(ctx, `CREATE SCHEMA `+schema); err != nil {
		t.Fatalf("создать тестовую схему: %v", err)
	}
	t.Cleanup(func() {
		_, _ = admin.Exec(context.Background(), `DROP SCHEMA IF EXISTS `+schema+` CASCADE`)
	})

	_, err = pool.Exec(ctx, `
		CREATE TABLE sites (
			id         INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			name       TEXT NOT NULL UNIQUE,
			url        TEXT NOT NULL UNIQUE,
			enabled    BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)
	`)
	if err != nil {
		t.Fatalf("создать таблицу sites: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewSites(pool, logger), func() {}
}

func TestSitesGetEnabled(t *testing.T) {
	repo, _ := setupCheckerSitesTest(t)
	ctx := context.Background()

	_, err := repo.db.Exec(ctx, `
		INSERT INTO sites (name, url, enabled) VALUES
			('Disabled', 'https://disabled.example', FALSE),
			('Second', 'https://second.example', TRUE),
			('First', 'https://first.example', TRUE)
	`)
	if err != nil {
		t.Fatalf("insert test sites: %v", err)
	}

	got, err := repo.GetEnabled(ctx)
	if err != nil {
		t.Fatalf("GetEnabled() error = %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("GetEnabled() returned %d sites, want 2", len(got))
	}
	if got[0].Name != "Second" || got[1].Name != "First" {
		t.Errorf("GetEnabled() order = [%q, %q], want [Second, First] by ID",
			got[0].Name, got[1].Name)
	}
	for _, site := range got {
		if !site.Enabled {
			t.Errorf("GetEnabled() returned disabled site %q", site.Name)
		}
	}
}

func TestSitesGetEnabledEmpty(t *testing.T) {
	repo, _ := setupCheckerSitesTest(t)

	got, err := repo.GetEnabled(context.Background())
	if err != nil {
		t.Fatalf("GetEnabled() error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("GetEnabled() returned %d sites, want 0", len(got))
	}
}

var _ = model.Site{}
