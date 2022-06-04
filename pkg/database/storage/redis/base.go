package redis

import (
	"context"
	"github.com/NubeIO/rubix-automater/automater"
	rs "github.com/NubeIO/rubix-automater/pkg/redis"
)

var _ automater.Storage = &Redis{}
var ctx = context.Background()

// Redis represents a redis client.
type Redis struct {
	*rs.Client
}

// New returns a redis client.
func New(url string, poolSize, minIdleConns int, keyPrefix string) *Redis {
	client := rs.New(url, poolSize, minIdleConns, keyPrefix, nil)
	return &Redis{
		Client: client,
	}
}

// WipeDB wipes the db.
func (rs *Redis) WipeDB() error {
	err := rs.FlushDB(ctx).Err()
	if err != nil {
		return err
	}
	return nil
}

func (rs *Redis) getRedisKeyForPipeline(id string) string {
	return rs.GetRedisPrefixedKey("pipeline:" + id)
}

func (rs *Redis) getRedisKeyForJob(id string) string {
	return rs.GetRedisPrefixedKey("job:" + id)
}

func (rs *Redis) getRedisKeyForTransaction(id string) string {
	return rs.GetRedisPrefixedKey("transaction:" + id)
}

func (rs *Redis) getRedisKeyForJobResult(id string) string {
	return rs.GetRedisPrefixedKey("jobresult:" + id)
}
