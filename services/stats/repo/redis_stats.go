package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/mbrysin/PingTower/services/stats/model"
	"github.com/redis/go-redis/v9"
)

type RedisStats struct {
	client *redis.Client
}

func NewRedisStats(client *redis.Client) *RedisStats {
	return &RedisStats{client: client}
}

func dailyStatsKey(siteID int, day time.Time) string {
	return fmt.Sprintf(
		"pingtower:stats:{%d}:%s",
		siteID,
		day.UTC().Format("2006-01-02"),
	)
}

func statsDataKey(siteID int, day time.Time) string {
	return fmt.Sprintf(
		"pingtower:stats-data:{%d}:%s",
		siteID,
		day.UTC().Format("2006-01-02"),
	)
}

func (r *RedisStats) GetBetween(
	ctx context.Context,
	siteID int,
	from time.Time,
	to time.Time,
) ([]model.Stat, error) {
	if !from.Before(to) {
		return nil, fmt.Errorf("from must be before to")
	}

	from = from.UTC()
	to = to.UTC()

	result := make([]model.Stat, 0)

	day := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)

	for day.Before(to) {

		dayStart := day
		dayEnd := day.AddDate(0, 0, 1)

		rangeStart := from
		if rangeStart.Before(dayStart) {
			rangeStart = dayStart
		}

		rangeEnd := to
		if rangeEnd.After(dayEnd) {
			rangeEnd = dayEnd
		}

		minScore := strconv.FormatInt(rangeStart.UnixMilli(), 10)
		maxScore := "(" + strconv.FormatInt(rangeEnd.UnixMilli(), 10)

		zsetKey := dailyStatsKey(siteID, day)
		hashKey := statsDataKey(siteID, day)

		ids, err := r.client.ZRangeByScore(ctx, zsetKey, &redis.ZRangeBy{
			Min: minScore,
			Max: maxScore,
		}).Result()
		if err != nil {
			return nil, fmt.Errorf("get stat IDs from redis: %w", err)
		}

		day = dayEnd

		if len(ids) == 0 {
			continue
		}

		values, err := r.client.HMGet(ctx, hashKey, ids...).Result()
		if err != nil {
			return nil, fmt.Errorf("get stats from redis: %w", err)
		}

		for _, value := range values {
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
	}

	return result, nil
}
