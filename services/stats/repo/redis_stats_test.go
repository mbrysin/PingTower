package repo

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"github.com/mbrysin/PingTower/services/stats/model"
)

func setupRedisStats(t *testing.T) (*RedisStats, *redis.Client, *miniredis.Miniredis) {
	t.Helper()

	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() {
		_ = client.Close()
	})

	return NewRedisStats(client), client, server
}

func addRedisStat(t *testing.T, client *redis.Client, siteID int, checkedAt time.Time, id int) {
	t.Helper()

	ctx := context.Background()
	day := checkedAt.UTC()

	data, err := json.Marshal(model.Stat{
		ID:        id,
		SiteID:    siteID,
		CheckedAt: checkedAt,
	})
	if err != nil {
		t.Fatalf("marshal stat: %v", err)
	}

	if err := client.ZAdd(ctx, dailyStatsKey(siteID, day), redis.Z{
		Score:  float64(checkedAt.UnixMilli()),
		Member: idString(id),
	}).Err(); err != nil {
		t.Fatalf("zadd stat: %v", err)
	}

	if err := client.HSet(ctx, statsDataKey(siteID, day), idString(id), data).Err(); err != nil {
		t.Fatalf("hset stat: %v", err)
	}
}

func TestRedisStatsGetBetween(t *testing.T) {
	repo, client, _ := setupRedisStats(t)

	from := time.Date(2026, 3, 10, 10, 0, 0, 0, time.UTC)
	to := from.Add(time.Hour)

	addRedisStat(t, client, 7, from.Add(-time.Millisecond), 1) // До начала — не вернуть.
	addRedisStat(t, client, 7, from, 2)                        // Начало включается.
	addRedisStat(t, client, 7, from.Add(30*time.Minute), 3)
	addRedisStat(t, client, 7, to, 4)                       // Конец не включается.
	addRedisStat(t, client, 8, from.Add(10*time.Minute), 5) // Другой сайт.

	got, err := repo.GetBetween(context.Background(), 7, from, to)
	if err != nil {
		t.Fatalf("GetBetween returned error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 stats, got %d", len(got))
	}
	if got[0].ID != 2 || got[1].ID != 3 {
		t.Fatalf("unexpected stat IDs: got %d and %d", got[0].ID, got[1].ID)
	}
}

func TestRedisStatsGetBetweenAcrossDays(t *testing.T) {
	repo, client, _ := setupRedisStats(t)

	from := time.Date(2026, 3, 10, 23, 30, 0, 0, time.UTC)
	to := time.Date(2026, 3, 11, 0, 30, 0, 0, time.UTC)

	addRedisStat(t, client, 7, from, 10)
	addRedisStat(t, client, 7, time.Date(2026, 3, 11, 0, 15, 0, 0, time.UTC), 11)

	got, err := repo.GetBetween(context.Background(), 7, from, to)
	if err != nil {
		t.Fatalf("GetBetween returned error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 stats across days, got %d", len(got))
	}
}

func TestRedisStatsGetBetweenEmpty(t *testing.T) {
	repo, _, _ := setupRedisStats(t)

	from := time.Date(2026, 3, 10, 10, 0, 0, 0, time.UTC)
	got, err := repo.GetBetween(context.Background(), 7, from, from.Add(time.Hour))
	if err != nil {
		t.Fatalf("GetBetween returned error: %v", err)
	}

	if len(got) != 0 {
		t.Fatalf("expected no stats, got %d", len(got))
	}
}

func TestRedisStatsGetBetweenInvalidRange(t *testing.T) {
	repo, _, _ := setupRedisStats(t)

	from := time.Date(2026, 3, 10, 10, 0, 0, 0, time.UTC)
	_, err := repo.GetBetween(context.Background(), 7, from, from)

	if err == nil {
		t.Fatal("expected error for empty range")
	}
}

func idString(id int) string {
	return strconv.Itoa(id)
}
