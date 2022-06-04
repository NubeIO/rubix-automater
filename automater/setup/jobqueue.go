package setup

import (
	"github.com/NubeIO/rubix-automater/automater"
	"github.com/NubeIO/rubix-automater/pkg/config"
	"github.com/NubeIO/rubix-automater/pkg/database/jobqueue"
)

func JobQueueFactory(cfg config.JobQueue, loggingFormat string) automater.JobQueue {
	if cfg.Option == "redis" {
		return jobqueue.NewRedisQueue(cfg.Redis, loggingFormat)
	}
	return nil
}
