package setup

import (
	"github.com/NubeIO/rubix-automater/automater"
	"github.com/NubeIO/rubix-automater/pkg/config"
	jobqueue2 "github.com/NubeIO/rubix-automater/pkg/database/jobqueue"
)

func JobQueueFactory(cfg config.JobQueue, loggingFormat string) automater.JobQueue {
	if cfg.Option == "memory" {
		return jobqueue2.NewFIFOQueue(cfg.MemoryJobQueue.Capacity)
	}
	if cfg.Option == "redis" {
		return jobqueue2.NewRedisQueue(cfg.Redis, loggingFormat)
	}
	return nil
}
