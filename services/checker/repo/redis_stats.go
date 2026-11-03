package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/mbrysin/PingTower/services/checker/model"
)

type RedisStats struct {
	client    *redis.Client
	retention time.Duration
}

func NewRedisStats(client *redis.Client, retention time.Duration) *RedisStats {
	return &RedisStats{
		client:    client,
		retention: retention,
	}
}

func dailyStatsKey(siteID int, day time.Time) string {
	date := day.UTC().Format("2006-01-02")
	return fmt.Sprintf("pingtower:stats:{%d}:%s", siteID, date)
}

func statsDataKey(siteID int, day time.Time) string {
	date := day.UTC().Format("2006-01-02")
	return fmt.Sprintf("pingtower:stats-data:{%d}:%s", siteID, date)
}

func (r *RedisStats) Add(ctx context.Context, stat *model.Stat) error {
	data, err := json.Marshal(stat)
	if err != nil {
		return fmt.Errorf("marshal stat: %w", err)
	}

	checkedAt := stat.CheckedAt.UTC()
	zsetKey := dailyStatsKey(stat.SiteID, checkedAt)
	hashKey := statsDataKey(stat.SiteID, checkedAt)
	member := strconv.Itoa(stat.ID)

	nextDay := checkedAt.Truncate(24 * time.Hour).Add(24 * time.Hour)
	ttl := time.Until(nextDay) + r.retention
	if ttl <= 0 {
		ttl = r.retention
	}

	pipe := r.client.Pipeline()
	pipe.ZAdd(ctx, zsetKey, redis.Z{
		Score:  float64(checkedAt.UnixMilli()),
		Member: member,
	})
	pipe.HSet(ctx, hashKey, member, data)
	pipe.Expire(ctx, zsetKey, ttl)
	pipe.Expire(ctx, hashKey, ttl)

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("save stat to redis: %w", err)
	}
	return nil
}

func (r *RedisStats) GetDay(
	ctx context.Context,
	siteID int,
	day time.Time,
) ([]model.Stat, error) {
	day = day.UTC().Truncate(24 * time.Hour)

	zsetKey := dailyStatsKey(siteID, day)
	hashKey := statsDataKey(siteID, day)

	ids, err := r.client.ZRange(ctx, zsetKey, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("get stat IDs from redis: %w", err)
	}
	if len(ids) == 0 {
		return []model.Stat{}, nil
	}

	values, err := r.client.HMGet(ctx, hashKey, ids...).Result()
	if err != nil {
		return nil, fmt.Errorf("get stats from redis: %w", err)
	}

	result := make([]model.Stat, 0, len(values))
	for _, value := range values {
		if value == nil {
			continue
		}

		data, ok := value.(string)
		if !ok {
			continue
		}

		var stat model.Stat
		if err := json.Unmarshal([]byte(data), &stat); err != nil {
			return nil, fmt.Errorf("unmarshal stat from redis: %w", err)
		}
		result = append(result, stat)
	}

	return result, nil
}
