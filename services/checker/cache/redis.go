package cache

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mbrysin/PingTower/services/checker/model"
	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
}

func NewRedis(client *redis.Client) *Redis {
	return &Redis{client: client}
}

func (r *Redis) SetLatest(ctx context.Context, siteID int, result model.CheckResult) error {
	if result.SiteID != 0 && result.SiteID != siteID {
		return fmt.Errorf(
			"set latest status: result site ID %d does not match argument %d",
			result.SiteID,
			siteID,
		)
	}

	result.SiteID = siteID

	value, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("marshal latest status: %w", err)
	}

	key := fmt.Sprintf("sites:%d:latest", siteID)

	if err := r.client.Set(ctx, key, value, 0).Err(); err != nil {
		return fmt.Errorf("set latest status in redis: %w", err)
	}

	return nil
}
