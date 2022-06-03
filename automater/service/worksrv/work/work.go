package work

import (
	"github.com/NubeIO/rubix-automater/automater/core"
	"time"
)

// Work is the tasks to be executed by the workers.
type Work struct {
	Type        string
	Job         *core.Job
	Result      chan core.JobResult
	TimeoutUnit time.Duration
}

// NewWork initializes and returns a new Work instance.
func NewWork(
	j *core.Job,
	resultChan chan core.JobResult,
	timeoutUnit time.Duration) Work {

	return Work{
		Job:         j,
		Result:      resultChan,
		TimeoutUnit: timeoutUnit,
	}
}
