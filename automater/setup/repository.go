package setup

import (
	"github.com/NubeIO/rubix-automater/automater"
	"github.com/NubeIO/rubix-automater/pkg/config"
	"github.com/NubeIO/rubix-automater/pkg/database/storage/memorydb"
	"github.com/NubeIO/rubix-automater/pkg/database/storage/redis"
	"github.com/NubeIO/rubix-automater/pkg/database/storage/relational"
	"github.com/NubeIO/rubix-automater/pkg/database/storage/relational/postgres"
)

func StorageFactory(cfg config.Storage) automater.Storage {
	if cfg.Option == "memory" {
		return memorydb.New()
	}
	if cfg.Option == "redis" {
		return redis.New(
			cfg.Redis.URL, cfg.Redis.PoolSize, cfg.Redis.MinIdleConns, cfg.Redis.KeyPrefix)
	}

	// Init common options for relational databases.
	options := new(relational.DBOptions)
	if cfg.Option == "postgres" {
		options.ConnectionMaxLifetime = cfg.Postgres.ConnectionMaxLifetime
		options.MaxOpenConnections = cfg.Postgres.MaxOpenConnections
		options.MaxIdleConnections = cfg.Postgres.MaxIdleConnections
		return postgres.New(cfg.Postgres.DSN, options)
	}
	return nil
}
