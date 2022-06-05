package setup

import (
	"github.com/NubeIO/rubix-automater/automater"
	"github.com/NubeIO/rubix-automater/automater/jobqueue"
	"github.com/NubeIO/rubix-automater/pkg/config"
)

func JobQueueFactory(cfg config.JobQueue, loggingFormat string) automater.JobQueue {
	if cfg.Option == "redis" {
		return jobqueue.NewRedisQueue(cfg.Redis, loggingFormat)
	}
	return nil
}
