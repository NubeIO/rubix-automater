package jobqueue

import (
	"github.com/NubeIO/rubix-automater/automater"
	"github.com/NubeIO/rubix-automater/automater/core"
	"github.com/NubeIO/rubix-automater/pkg/helpers/apperrors"
)

var _ automater.JobQueue = &fifoqueue{}

type fifoqueue struct {
	jobs     chan *core.Job
	capacity int
}

// NewFIFOQueue creates and returns a new fifoqueue instance.
func NewFIFOQueue(capacity int) *fifoqueue {
	return &fifoqueue{
		jobs:     make(chan *core.Job, capacity),
		capacity: capacity,
	}
}

// Push adds a job to the queue.
func (q *fifoqueue) Push(j *core.Job) error {
	select {
	case q.jobs <- j:
		return nil
	default:
		return &apperrors.FullQueueErr{}
	}
}

// Pop removes and returns the head job from the queue.
func (q *fifoqueue) Pop() *core.Job {
	select {
	case j := <-q.jobs:
		return j
	default:
		return nil
	}
}

// CheckHealth checks if the job queue is alive.
func (q *fifoqueue) CheckHealth() bool {
	return q.jobs != nil
}

// Close liberates the bound resources of the job queue.
func (q *fifoqueue) Close() {
	close(q.jobs)
}
