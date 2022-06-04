package worksrv

import (
	"context"
	"fmt"
	"github.com/NubeIO/rubix-automater/automater"
	"github.com/NubeIO/rubix-automater/automater/model"
	taskRepo "github.com/NubeIO/rubix-automater/automater/service/tasksrv/taskrepo"
	"github.com/NubeIO/rubix-automater/automater/service/worksrv/work"
	"github.com/NubeIO/rubix-automater/pkg/database/storage/redis"
	intime "github.com/NubeIO/rubix-automater/pkg/helpers/ttime"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	WorkTypeTask                    = "tasks"
	WorkTypePipeline                = "pipeline"
	DefaultJobTimeout time.Duration = 5 * time.Minute
)

var _ automater.WorkService = &workService{}

type workService struct {
	// The number of go-routines that will operate concurrently.
	workers int
	// The capacity of the worker pool queue.
	queueCapacity int
	// The time unit for the calculation of the timeout interval for each tasks.
	timeoutUnit time.Duration

	storage  automater.Storage
	taskRepo *taskRepo.TaskRepository
	time     intime.Time
	queue    chan work.Work
	wg       sync.WaitGroup
	logger   *logrus.Logger
}

// New creates a new work server.
func New(
	storage automater.Storage,
	taskRepo *taskRepo.TaskRepository,
	time intime.Time, timeoutUnit time.Duration,
	workers, queueCapacity int, logger *logrus.Logger) *workService {

	return &workService{
		storage:       storage,
		taskRepo:      taskRepo,
		workers:       workers,
		queueCapacity: queueCapacity,
		timeoutUnit:   timeoutUnit,
		queue:         make(chan work.Work, queueCapacity),
		time:          time,
		logger:        logger,
	}
}

// Start starts the worker pool.
func (srv *workService) Start() {
	for i := 0; i < srv.workers; i++ {
		srv.wg.Add(1)
		go srv.startWorker(i, srv.queue, &srv.wg)
	}
	srv.logger.Infof("set up %d workers with a queue of capacity %d", srv.workers, srv.queueCapacity)
}

// Dispatch dispatches a work to the worker pool.
func (srv *workService) Dispatch(w work.Work) {
	workType := WorkTypeTask
	for job := w.Job; ; job, workType = job.Next, WorkTypePipeline {
		go func() {
			result, ok := <-w.Result
			if ok {
				fmt.Println(w.Job)
				if err := srv.storage.CreateJobResult(&result); err != nil {
					srv.logger.Errorf("could not create job result to the storage %s", err)
				}
			}
		}()

		if !job.HasNext() {
			break
		}
	}
	w.Type = workType

	srv.queue <- w
}

// CreateWork creates and return a new Work instance.
func (srv *workService) CreateWork(j *model.Job) work.Work {
	resultChan := make(chan model.JobResult, 1)
	return work.NewWork(j, resultChan, srv.timeoutUnit)
}

// Stop signals the workers to stop working gracefully.
func (srv *workService) Stop() {
	close(srv.queue)
	srv.logger.Info("waiting for ongoing tasks to finish...")
	srv.wg.Wait()
}

// ExecJobWork executes the job work.
func (srv *workService) ExecJobWork(ctx context.Context, w work.Work) error {
	// Do not let the go-routines wait for result in case of early exit.
	defer close(w.Result)
	srv.logger.Info("executes the job worker", w.Job.Name)

	startedAt := srv.time.Now()
	w.Job.MarkStarted(&startedAt)
	if _, err := srv.storage.UpdateJob(w.Job.UUID, w.Job); err != nil {
		return err
	}
	timeout := DefaultJobTimeout
	if w.Job.Timeout > 0 {
		timeout = time.Duration(w.Job.Timeout) * w.TimeoutUnit
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	jobResultChan := make(chan model.JobResult, 1)

	srv.work(w.Job, jobResultChan, nil)

	var jobResult model.JobResult
	select {
	case <-ctx.Done():
		failedAt := srv.time.Now()
		w.Job.MarkFailed(&failedAt, ctx.Err().Error())
		jobResult = model.JobResult{
			JobID:    w.Job.UUID,
			Metadata: nil,
			Error:    ctx.Err().Error(),
		}
	case jobResult = <-jobResultChan:
		if jobResult.Error != "" {
			failedAt := srv.time.Now()
			srv.logger.Errorln("RAN FAILED", w.Job.Name)
			w.Job.MarkFailed(&failedAt, jobResult.Error)
		} else {
			srv.logger.Info("RAN JOB", w.Job.Name)
			completedAt := srv.time.Now()
			w.Job.MarkCompleted(&completedAt)
		}
	}
	if w.Job.JobOptions.EnableInterval {
		if err := srv.storage.Pub(redis.TopicJob, w.Job); err != nil {
			return err
		}
		if _, err := srv.storage.CreateTransaction(w.Job); err != nil {
			w.Job.MarkPending() //rest back to pending
			if _, err := srv.storage.Recycle(w.Job.UUID, w.Job); err != nil {
				return err
			}
		}
		if _, err := srv.storage.Recycle(w.Job.UUID, w.Job); err != nil {
			return err
		}
	} else {
		if _, err := srv.storage.UpdateJob(w.Job.UUID, w.Job); err != nil {
			return err
		}
	}

	w.Result <- jobResult

	return nil
}

// ExecPipelineWork executes the pipeline work.
func (srv *workService) ExecPipelineWork(ctx context.Context, w work.Work) error {
	// Do not let the go-routines wait for result in case of early exit.
	defer close(w.Result)

	p, err := srv.storage.GetPipeline(w.Job.PipelineID)
	if err != nil {
		return err
	}
	var jobResult model.JobResult

	for job, i := w.Job, 0; ; job, i = job.Next, i+1 {
		startedAt := srv.time.Now()
		job.MarkStarted(&startedAt)
		if _, err := srv.storage.UpdateJob(job.UUID, job); err != nil {
			return err
		}
		if i == 0 {
			p.MarkStarted(&startedAt)
			if err := srv.storage.UpdatePipeline(p.UUID, p); err != nil {
				return err
			}
		}

		timeout := DefaultJobTimeout
		if job.Timeout > 0 {
			timeout = time.Duration(job.Timeout) * w.TimeoutUnit
		}
		ctx, cancel := context.WithTimeout(ctx, timeout)
		jobResultChan := make(chan model.JobResult, 1)

		srv.work(job, jobResultChan, jobResult.Metadata)

		select {
		case <-ctx.Done():
			failedAt := srv.time.Now()
			job.MarkFailed(&failedAt, ctx.Err().Error())
			p.MarkFailed(&failedAt)

			jobResult = model.JobResult{
				JobID:    job.UUID,
				Metadata: nil,
				Error:    ctx.Err().Error(),
			}
		case jobResult = <-jobResultChan:
			if jobResult.Error != "" {
				failedAt := srv.time.Now()
				job.MarkFailed(&failedAt, jobResult.Error)
				p.MarkFailed(&failedAt)
			} else {
				completedAt := srv.time.Now()
				job.MarkCompleted(&completedAt)
				if !job.HasNext() {
					p.MarkCompleted(&completedAt)
				}
			}
		}
		// Reset timeout.
		cancel()
		if job.JobOptions.EnableInterval {
			if _, err := srv.storage.Recycle(job.UUID, job); err != nil {
				return err
			}

		} else {
			if _, err := srv.storage.UpdateJob(job.UUID, job); err != nil {
				return err
			}
			if p.Status == model.Failed || p.Status == model.Completed {
				if err := srv.storage.UpdatePipeline(p.UUID, p); err != nil {
					return err
				}
			}
		}

		w.Result <- jobResult

		// Stop the pipeline execution on failure.
		if job.Status == model.Failed {
			break
		}

		// Stop the pipeline execution if there's no other job.
		if !job.HasNext() {
			break
		}
	}

	return nil
}

func (srv *workService) work(
	job *model.Job,
	jobResultChan chan model.JobResult,
	previousJobResultsMetadata interface{}) {

	go func() {
		defer func() {
			if p := recover(); p != nil {
				result := model.JobResult{
					JobID:    job.UUID,
					Metadata: nil,
					Error:    fmt.Errorf("%v", p).Error(),
				}
				jobResultChan <- result
				close(jobResultChan)
			}
		}()
		var errMsg string

		// Should be already validated.
		taskFunc, _ := srv.taskRepo.GetTaskFunc(job.TaskName)

		params := []interface{}{
			job.TaskParams,
		}
		if job.UsePreviousResults && previousJobResultsMetadata != nil {
			params = append(params, previousJobResultsMetadata)
		}
		// Perform the actual work.
		resultMetadata, jobErr := taskFunc(params...)
		if jobErr != nil {
			errMsg = jobErr.Error()
		}

		result := model.JobResult{
			JobID:    job.UUID,
			Metadata: resultMetadata,
			Error:    errMsg,
		}
		jobResultChan <- result
		close(jobResultChan)
	}()
}

func (srv *workService) Exec(ctx context.Context, w work.Work) error {
	if w.Type == WorkTypePipeline {
		return srv.ExecPipelineWork(ctx, w)
	}
	return srv.ExecJobWork(ctx, w)
}

func (srv *workService) startWorker(uuid int, queue <-chan work.Work, wg *sync.WaitGroup) {
	defer wg.Done()
	logPrefix := fmt.Sprintf("[worker] %d", uuid)
	for work := range queue {
		srv.logger.Infof("%s executing %s...", logPrefix, work.Type)
		if err := srv.Exec(context.Background(), work); err != nil {
			srv.logger.Errorf("could not update job status: %s", err)
		}
		srv.logger.Infof("%s %s finished!", logPrefix, work.Type)
	}
	srv.logger.Infof("%s exiting...", logPrefix)
}
