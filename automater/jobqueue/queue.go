package jobqueue

import (
	"context"
	"encoding/json"
	"github.com/NubeIO/rubix-automater/automater"
	"github.com/NubeIO/rubix-automater/automater/model"
	"github.com/NubeIO/rubix-automater/pkg/config"
	"github.com/NubeIO/rubix-automater/pkg/logger"
	rs "github.com/NubeIO/rubix-automater/pkg/redis"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

var _ automater.JobQueue = &redisQueue{}
var ctx = context.Background()

type redisQueue struct {
	*rs.Client
	logger *logrus.Logger
}

// NewRedisQueue returns a redis queue.
func NewRedisQueue(cfg config.Redis, loggingFormat string) *redisQueue {
	logger := logger.NewLogger("redisQueue", loggingFormat)
	client := rs.New(cfg.URL, cfg.PoolSize, cfg.MinIdleConns, cfg.KeyPrefix, logger)
	return &redisQueue{
		Client: client,
		logger: logger,
	}
}

func (q *redisQueue) Pop() *model.Job {
	key := q.GetRedisPrefixedKey("job-queue")
	val, err := q.RPop(ctx, key).Bytes()
	if err != nil {
		if err != redis.Nil {
			q.logger.Errorf("could not RPOP job message: %s", err)
		}
		return nil
	}
	var j *model.Job
	err = json.Unmarshal(val, &j)
	if err != nil {
		q.logger.Errorf("could not unmarshal message body: %s", err)
		return nil
	}
	return j
}

func (q *redisQueue) Close() {
	q.Client.Close()
}
