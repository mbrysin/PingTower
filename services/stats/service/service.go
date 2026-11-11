package service

import (
	"context"
	"fmt"
	"time"

	"github.com/mbrysin/PingTower/services/stats/model"
)

type StatsRepository interface {
	GetStatsBetween(
		ctx context.Context,
		siteID int,
		from time.Time,
		to time.Time,
	) ([]model.Stat, error)

	GetDailyStatsBetween(
		ctx context.Context,
		siteID int,
		from time.Time,
		to time.Time,
	) ([]model.DailyStat, error)

	GetLatest(ctx context.Context, siteID int) (*model.Stat, error)

	GetLatestForSites(ctx context.Context, siteIDs []int) (map[int]model.Stat, error)
}

type RedisStatsRepository interface {
	GetBetween(
		ctx context.Context,
		siteID int,
		from time.Time,
		to time.Time,
	) ([]model.Stat, error)
}

type Stats struct {
	repo      StatsRepository
	redisRepo RedisStatsRepository
}

func NewStats(repo StatsRepository, redisRepo RedisStatsRepository) *Stats {
	return &Stats{
		repo:      repo,
		redisRepo: redisRepo,
	}
}

func (s *Stats) GetStatsBetween(
	ctx context.Context,
	siteID int,
	from time.Time,
	to time.Time,
) ([]model.Stat, error) {
	if siteID <= 0 {
		return nil, fmt.Errorf("site_id must be positive")
	}
	if !from.Before(to) {
		return nil, fmt.Errorf("from must be before to")
	}

	stats, err := s.redisRepo.GetBetween(ctx, siteID, from, to)
	if err != nil {
		return nil, fmt.Errorf("get stats from redis: %w", err)
	}
	if len(stats) > 0 {
		return stats, nil
	}

	return s.repo.GetStatsBetween(ctx, siteID, from, to)
}

func (s *Stats) GetDailyStatsBetween(
	ctx context.Context,
	siteID int,
	from time.Time,
	to time.Time,
) ([]model.DailyStat, error) {
	if siteID <= 0 {
		return nil, fmt.Errorf("site_id must be positive")
	}

	from = dateOnly(from)
	to = dateOnly(to)

	if !from.Before(to) {
		return nil, fmt.Errorf("from must be before to")
	}

	return s.repo.GetDailyStatsBetween(ctx, siteID, from, to)
}

func dateOnly(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func (s *Stats) GetLatest(
	ctx context.Context,
	siteID int,
) (*model.Stat, error) {
	if siteID <= 0 {
		return nil, fmt.Errorf("site_id must be positive")
	}

	stat, err := s.repo.GetLatest(ctx, siteID)
	if err != nil {
		return nil, fmt.Errorf("get latest stat: %w", err)
	}

	return stat, nil
}

func (s *Stats) GetLatestForSites(
	ctx context.Context,
	siteIDs []int,
) (map[int]model.Stat, error) {
	seen := make(map[int]struct{}, len(siteIDs))
	for _, id := range siteIDs {
		if id <= 0 {
			return nil, fmt.Errorf("site_id must be positive")
		}
		if _, exists := seen[id]; exists {
			return nil, fmt.Errorf("duplicate site_id: %d", id)
		}
		seen[id] = struct{}{}
	}

	return s.repo.GetLatestForSites(ctx, siteIDs)
}
