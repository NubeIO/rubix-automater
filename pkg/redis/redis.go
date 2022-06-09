package redis

import (
	"context"
	"io/ioutil"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

var ctx = context.Background()

// Client represents a redis autocli.
type Client struct {
	*redis.Client
	KeyPrefix string
	Logger    *logrus.Logger
}

// New returns a redis autocli.
func New(
	url string, poolSize, minIdleConns int,
	keyPrefix string, logger *logrus.Logger) *Client {

	if url == "" {
		url = "redis://localhost:6379"
	}

	if poolSize == 0 {
		poolSize = 10
	}
	if minIdleConns == 0 {
		minIdleConns = 10
	}

	rs := new(Client)

	rs.KeyPrefix = keyPrefix
	if logger != nil {
		rs.Logger = logger
	} else {
		rs.Logger = &logrus.Logger{Out: ioutil.Discard}
	}

	opt, err := redis.ParseURL(url)
	if err != nil {
		panic(err)
	}

	rs.Client = redis.NewClient(&redis.Options{
		Addr:         opt.Addr,
		Password:     opt.Password,
		DB:           opt.DB,
		PoolSize:     poolSize,
		MinIdleConns: minIdleConns,
		TLSConfig:    opt.TLSConfig,
	})
	return rs
}

// CheckHealth checks if the job queue is alive.
func (rs *Client) CheckHealth() bool {
	res, err := rs.Ping(ctx).Result()
	if err != nil {
		rs.Logger.Errorf("ping returned error: %s", err)
		return false
	}
	return res == "PONG"
}

// Close terminates any storage connections gracefully.
func (rs *Client) Close() error {
	return rs.Client.Close()
}

func (rs *Client) GetRedisPrefixedKey(key string) string {
	if rs.KeyPrefix != "" {
		return rs.KeyPrefix + ":" + key
	}
	return key
}
