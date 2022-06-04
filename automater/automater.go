package automater

import (
	"context"
	"github.com/NubeIO/rubix-automater/automater/model"
	taskRepo "github.com/NubeIO/rubix-automater/automater/service/tasksrv/taskrepo"
	"github.com/NubeIO/rubix-automater/automater/service/worksrv/work"
	"time"
)

// Storage represents a driven actor storage interface.
type Storage interface {
	CreateJob(j *model.Job) error
	GetJob(uuid string) (*model.Job, error)
	GetJobs(status model.JobStatus) ([]*model.Job, error)
	GetDueJobs() ([]*model.Job, error)
	GetJobsByPipelineID(pipelineID string) ([]*model.Job, error)
	UpdateJob(uuid string, j *model.Job) (*model.Job, error)
	Recycle(uuid string, body *model.Job) (*model.Job, error)
	DeleteJob(uuid string) error
	CreateJobResult(result *model.JobResult) error
	GetJobResult(jobID string) (*model.JobResult, error)
	UpdateJobResult(jobID string, result *model.JobResult) error
	DeleteJobResult(jobID string) error
	CreatePipeline(p *model.Pipeline) error
	GetPipeline(uuid string) (*model.Pipeline, error)

	CreateTransaction(jobID string, job *model.Job) (*model.Transaction, error)
	GetTransactions(status model.JobStatus) ([]*model.Transaction, error)

	GetPipelines(status model.JobStatus) ([]*model.Pipeline, error)
	UpdatePipeline(uuid string, p *model.Pipeline) error
	DeletePipeline(uuid string) error
	CheckHealth() bool
	Close() error

	// WipeDB wipes the DB
	WipeDB() error
}

// JobQueue represents a driven actor queue interface.
type JobQueue interface {
	// Push adds a job to the queue.
	Push(j *model.Job) error

	// Pop removes and returns the head job from the queue.
	Pop() *model.Job

	// CheckHealth checks if the job queue is alive.
	CheckHealth() bool

	// Close liberates the bound resources of the job queue.
	Close()
}

// JobService represents a driver actor server interface.
type JobService interface {
	Create(name, taskName, description, runAt string, timeout int, disable bool, options *model.JobOptions, taskParams map[string]interface{}) (*model.Job, error)
	Get(uuid string) (*model.Job, error)
	GetJobs(status string) ([]*model.Job, error)
	Update(uuid, name, description string) (*model.Job, error)
	Recycle(uuid string, body *model.Job) (*model.Job, error)
	Delete(uuid string) error
	Drop() error
}

type TransactionService interface {
	GetTransactions(status string) ([]*model.Transaction, error)
}

// ResultService represents a driver actor server interface.
type ResultService interface {
	// Get fetches a job result.
	Get(uuid string) (*model.JobResult, error)
	// Delete deletes a job result.
	Delete(uuid string) error
}

// PipelineService represents a driver actor server interface.
type PipelineService interface {
	// Create creates a new pipeline.
	Create(name, description, runAt string, jobs []*model.Job) (*model.Pipeline, error)
	// Get fetches a pipeline.
	Get(uuid string) (*model.Pipeline, error)
	// GetPipelines fetches all pipelines, optionally filters the pipelines by status.
	GetPipelines(status string) ([]*model.Pipeline, error)
	// GetPipelineJobs fetches the jobs of a specified pipeline.
	GetPipelineJobs(uuid string) ([]*model.Job, error)
	// Update updates a pipeline.
	Update(uuid, name, description string) error
	// Delete deletes a pipeline.
	Delete(uuid string) error
}

// WorkService represents a driver actor server interface.
type WorkService interface {
	// Start starts the worker pool.
	Start()

	// Stop signals the workers to stop working gracefully.
	Stop()

	// Dispatch dispatches a work to the worker pool.
	Dispatch(w work.Work)

	// CreateWork creates and return a new Work instance.
	CreateWork(j *model.Job) work.Work

	// Exec executes a work.
	Exec(ctx context.Context, w work.Work) error
}

// TaskService represents a driver actor server interface.
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

// Server represents a driver actor server interface.
type Server interface {
	// Serve start the server.
	Serve()
	// GracefullyStop gracefully stops the server.
	GracefullyStop()
}
