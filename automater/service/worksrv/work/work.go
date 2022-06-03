package work

import (
	"github.com/NubeIO/rubix-automater/automater/model"
	"time"
)

// Work is the tasks to be executed by the workers.
type Work struct {
	Type        string
	Job         *model.Job
	Result      chan model.JobResult
	TimeoutUnit time.Duration
}

// NewWork initializes and returns a new Work instance.
func NewWork(
	j *model.Job,
	resultChan chan model.JobResult,
	timeoutUnit time.Duration) Work {

	return Work{
		Job:         j,
		Result:      resultChan,
		TimeoutUnit: timeoutUnit,
	}
}
