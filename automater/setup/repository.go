package setup

import (
	"github.com/NubeIO/rubix-automater/automater"
	"github.com/NubeIO/rubix-automater/pkg/config"
	"github.com/NubeIO/rubix-automater/pkg/database/storage/db"
)

func StorageFactory(cfg config.Storage) automater.Storage {
	if cfg.Option == "redis" {
		return redis.New(
			cfg.Redis.URL, cfg.Redis.PoolSize, cfg.Redis.MinIdleConns, cfg.Redis.KeyPrefix)
	}
	return nil
}
