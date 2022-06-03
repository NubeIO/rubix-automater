package schedulersrv

import (
	"context"
	"fmt"
	"github.com/NubeIO/rubix-automater/automater"
	intime "github.com/NubeIO/rubix-automater/pkg/helpers/intime"
	"github.com/sirupsen/logrus"
	"time"
)

var _ automater.Scheduler = &schedulerService{}

type schedulerService struct {
	jobQueue    automater.JobQueue
	storage     automater.Storage
	workService automater.WorkService
	time        intime.Time
	logger      *logrus.Logger
}

// New creates a new scheduler service.
func New(
	jobQueue automater.JobQueue,
	storage automater.Storage,
	workService automater.WorkService,
	time intime.Time,
	logger *logrus.Logger) *schedulerService {

	return &schedulerService{
		jobQueue:    jobQueue,
		storage:     storage,
		workService: workService,
		time:        time,
		logger:      logger,
	}
}

// Dispatch listens to the job queue for messages, consumes them and
// dispatches the work items to the worker pool for execution.
func (srv *schedulerService) Dispatch(ctx context.Context, duration time.Duration) {
	ticker := time.NewTicker(duration)
	go func() {
		for {
			srv.logger.Info("running loop dispatch job")
			select {
			case <-ctx.Done():
				ticker.Stop()
				srv.logger.Info("exiting...")
				return
			case <-ticker.C:
				j := srv.jobQueue.Pop()
				srv.logger.Info("dispatch no job to run")
				if j == nil {
					continue
				}
				srv.logger.Info("dispatch a new job uuid:", j.UUID)
				w := srv.workService.CreateWork(j)
				// Blocks until worker pool backlog has some space.
				srv.workService.Dispatch(w)
				message := fmt.Sprintf("job with UUID: %s", j.UUID)
				if j.BelongsToPipeline() {
					message = fmt.Sprintf("pipeline with UUID: %s", j.PipelineID)
				}
				srv.logger.Infof("sent work for %s to worker pool", message)
			}
		}
	}()
}

// Schedule polls the storage in given interval and schedules due jobs for execution.
func (srv *schedulerService) Schedule(ctx context.Context, duration time.Duration) {
	ticker := time.NewTicker(duration)
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				srv.logger.Info("exiting...")
				return
			case <-ticker.C:
				dueJobs, err := srv.storage.GetDueJobs()
				fmt.Println("GetDueJobs", dueJobs)
				if err != nil {
					srv.logger.Errorf("could not get due jobs from storage: %s", err)
					continue
				}
				for _, j := range dueJobs {
					if j.Disable {
						srv.logger.Infoln("JOB Is Disable name:", j.Name)
						continue
					} else {
						srv.logger.Infoln("JOB IS Not Disable", j.Name)
					}
					if j.BelongsToPipeline() {
						for job := j; job.HasNext(); job = job.Next {
							job.Next, err = srv.storage.GetJob(job.NextJobID)
							if err != nil {
								srv.logger.Errorf("could not get piped due job from storage: %s", err)
								continue
							}
						}
					}
					w := srv.workService.CreateWork(j)
					// Blocks until worker pool backlog has some space.
					srv.workService.Dispatch(w)

					scheduledAt := srv.time.Now()
					j.MarkScheduled(&scheduledAt)
					if _, err := srv.storage.UpdateJob(j.UUID, j); err != nil {
						srv.logger.Errorf("could not update job: %s", err)
					}
					message := fmt.Sprintf("job with UUID: %s", j.UUID)
					if j.BelongsToPipeline() {
						message = fmt.Sprintf("pipeline with UUID: %s", j.PipelineID)
					}
					srv.logger.Infof("scheduled work for %s to worker pool", message)
				}
			}
		}
	}()
}
