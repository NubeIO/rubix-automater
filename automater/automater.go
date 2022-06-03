package automater

import (
	"context"
	"github.com/NubeIO/rubix-automater/automater/core"
	taskRepo "github.com/NubeIO/rubix-automater/automater/service/tasksrv/taskrepo"
	"github.com/NubeIO/rubix-automater/automater/service/worksrv/work"
	"time"
)

// Storage represents a driven actor storage interface.
type Storage interface {
	CreateJob(j *core.Job) error
	GetJob(uuid string) (*core.Job, error)
	GetJobs(status core.JobStatus) ([]*core.Job, error)
	GetDueJobs() ([]*core.Job, error)
	GetJobsByPipelineID(pipelineID string) ([]*core.Job, error)
	UpdateJob(uuid string, j *core.Job) error
	DeleteJob(uuid string) error
	CreateJobResult(result *core.JobResult) error
	GetJobResult(jobID string) (*core.JobResult, error)
	UpdateJobResult(jobID string, result *core.JobResult) error
	DeleteJobResult(jobID string) error
	CreatePipeline(p *core.Pipeline) error
	GetPipeline(uuid string) (*core.Pipeline, error)

	GetPipelines(status core.JobStatus) ([]*core.Pipeline, error)
	UpdatePipeline(uuid string, p *core.Pipeline) error
	DeletePipeline(uuid string) error
	CheckHealth() bool
	Close() error
}

// JobQueue represents a driven actor queue interface.
type JobQueue interface {
	// Push adds a job to the queue.
	Push(j *core.Job) error

	// Pop removes and returns the head job from the queue.
	Pop() *core.Job

	// CheckHealth checks if the job queue is alive.
	CheckHealth() bool

	// Close liberates the bound resources of the job queue.
	Close()
}

// JobService represents a driver actor service interface.
type JobService interface {
	Create(name, taskName, description, runAt string, timeout int, taskParams map[string]interface{}) (*core.Job, error)
	Get(uuid string) (*core.Job, error)
	GetJobs(status string) ([]*core.Job, error)
	Update(uuid, name, description string) error
	UpdateAll(uuid string, body *core.Job) error
	Delete(uuid string) error
	Drop() error
}

// ResultService represents a driver actor service interface.
type ResultService interface {
	// Get fetches a job result.
	Get(uuid string) (*core.JobResult, error)
	// Delete deletes a job result.
	Delete(uuid string) error
}

// PipelineService represents a driver actor service interface.
type PipelineService interface {
	// Create creates a new pipeline.
	Create(name, description, runAt string, jobs []*core.Job) (*core.Pipeline, error)
	// Get fetches a pipeline.
	Get(uuid string) (*core.Pipeline, error)
	// GetPipelines fetches all pipelines, optionally filters the pipelines by status.
	GetPipelines(status string) ([]*core.Pipeline, error)
	// GetPipelineJobs fetches the jobs of a specified pipeline.
	GetPipelineJobs(uuid string) ([]*core.Job, error)
	// Update updates a pipeline.
	Update(uuid, name, description string) error
	// Delete deletes a pipeline.
	Delete(uuid string) error
}

// WorkService represents a driver actor service interface.
type WorkService interface {
	// Start starts the worker pool.
	Start()

	// Stop signals the workers to stop working gracefully.
	Stop()

	// Dispatch dispatches a work to the worker pool.
	Dispatch(w work.Work)

	// CreateWork creates and return a new Work instance.
	CreateWork(j *core.Job) work.Work

	// Exec executes a work.
	Exec(ctx context.Context, w work.Work) error
}

// TaskService represents a driver actor service interface.
type TaskService interface {
	// Register registers a new tasks in the tasks database.
	Register(name string, taskFunc taskRepo.TaskFunc)
	// GetTaskRepository returns the tasks database.
	GetTaskRepository() *taskRepo.TaskRepository
}

// Scheduler represents a domain event listener.
type Scheduler interface {
	// Schedule polls the storage in given interval and schedules due jobs for execution.
	Schedule(ctx context.Context, duration time.Duration)
	// Dispatch listens to the job queue for messages, consumes them and
	// dispatches the jobs for execution.
	Dispatch(ctx context.Context, duration time.Duration)
}

// Server represents a driver actor service interface.
type Server interface {
	// Serve start the server.
	Serve()
	// GracefullyStop gracefully stops the server.
	GracefullyStop()
}
