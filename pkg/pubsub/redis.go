package pubsub

import (
	"github.com/NubeIO/lib-redis/libredis"
)

func New(config *libredis.Config) libredis.Client {
	config.Addr = ""
	client, err := libredis.New(config)
	if err != nil {
		return nil
	}
	return client
}
