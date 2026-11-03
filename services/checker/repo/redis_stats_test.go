package repo

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"

	"github.com/mbrysin/PingTower/services/checker/model"
)

func newTestRedisStats(t *testing.T) (*RedisStats, *miniredis.Miniredis) {
	t.Helper()

	server := miniredis.RunT(t)
	client := goredis.NewClient(&goredis.Options{Addr: server.Addr()})
	t.Cleanup(func() {
		_ = client.Close()
	})

	return NewRedisStats(client, 48*time.Hour), server
}

func TestRedisStatsAddAndGetDay(t *testing.T) {
	cache, _ := newTestRedisStats(t)
	ctx := context.Background()
	checkedAt := time.Date(2026, 11, 2, 12, 0, 0, 0, time.UTC)

	want := &model.Stat{
		ID:        10,
		SiteID:    7,
		CheckedAt: checkedAt,
		IsUp:      true,
	}

	if err := cache.Add(ctx, want); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	got, err := cache.GetDay(ctx, want.SiteID, checkedAt)
	if err != nil {
		t.Fatalf("GetDay() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(GetDay()) = %d, want 1", len(got))
	}
	if got[0].ID != want.ID {
		t.Errorf("ID = %d, want %d", got[0].ID, want.ID)
	}
	if got[0].SiteID != want.SiteID {
		t.Errorf("SiteID = %d, want %d", got[0].SiteID, want.SiteID)
	}
	if !got[0].IsUp {
		t.Error("IsUp = false, want true")
	}
}

func TestRedisStatsGetDayEmpty(t *testing.T) {
	cache, _ := newTestRedisStats(t)

	got, err := cache.GetDay(
		context.Background(),
		123,
		time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("GetDay() error = %v", err)
	}
	if got == nil {
		t.Fatal("GetDay() returned nil slice; want empty slice")
	}
	if len(got) != 0 {
		t.Errorf("len(GetDay()) = %d, want 0", len(got))
	}
}

func TestRedisStatsGetDayUsesUTCDate(t *testing.T) {
	cache, _ := newTestRedisStats(t)
	ctx := context.Background()

	// В Токио это уже следующий календарный день,
	// но в UTC время всё ещё относится к 2 ноября.
	checkedAt := time.Date(2026, 11, 2, 23, 30, 0, 0, time.UTC)
	stat := &model.Stat{
		ID:        1,
		SiteID:    4,
		CheckedAt: checkedAt,
		IsUp:      true,
	}

	if err := cache.Add(ctx, stat); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	tokyoTime := checkedAt.In(time.FixedZone("JST", 9*60*60))
	got, err := cache.GetDay(ctx, stat.SiteID, tokyoTime)
	if err != nil {
		t.Fatalf("GetDay() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(GetDay()) = %d, want 1", len(got))
	}
}
