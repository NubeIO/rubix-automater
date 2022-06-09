package redis

import (
	"context"
	"fmt"
	"github.com/NubeIO/lib-redis/libredis"
	"github.com/NubeIO/rubix-automater/automater"
	"github.com/NubeIO/rubix-automater/pkg/pubsub"
	rs "github.com/NubeIO/rubix-automater/pkg/redis"
)

var _ automater.Storage = &Redis{}
var ctx = context.Background()

// Redis represents a redis autocli.
type Redis struct {
	*rs.Client
	pub libredis.Client
}

// New returns a redis autocli.
func New(url string, poolSize, minIdleConns int, keyPrefix string) *Redis {
	client := rs.New(url, poolSize, minIdleConns, keyPrefix, nil)
	c := &libredis.Config{Addr: url}
	pub := pubsub.New(c)
	return &Redis{
		Client: client,
		pub:    pub,
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

const (
	pipeline    = "pipeline"
	job         = "job"
	transaction = "transaction"
	jobresult   = "jobresult"
)

func (rs *Redis) getRedisKeyForPipeline(id string) string {
	return rs.GetRedisPrefixedKey(fmt.Sprintf("%s:%s", pipeline, id))
}

func (rs *Redis) getRedisKeyForJob(id string) string {
	return rs.GetRedisPrefixedKey(fmt.Sprintf("%s:%s", job, id))
}

func (rs *Redis) getRedisKeyForTransaction(id string) string {
	return rs.GetRedisPrefixedKey(fmt.Sprintf("%s:%s", transaction, id))
}

func (rs *Redis) getRedisKeyForJobResult(id string) string {
	return rs.GetRedisPrefixedKey(fmt.Sprintf("%s:%s", jobresult, id))
}
