package schedulersrv

import (
	"context"
	"fmt"
	"github.com/NubeIO/rubix-automater/automater"
	"github.com/NubeIO/rubix-automater/automater/model"
	intime "github.com/NubeIO/rubix-automater/pkg/helpers/ttime"
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

// New creates a new scheduler server.
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
				if j == nil {
					srv.logger.Info("dispatch no job to run")
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
				srv.logger.Info("exiting schedule...")
				return
			case <-ticker.C:
				dueJobs, err := srv.storage.GetDueJobs()
				srv.logger.Infoln("schedule loop job count:", len(dueJobs))
				if err != nil {
					srv.logger.Errorf("could not get due jobs from storage: %s", err)
					continue
				}
				for _, j := range dueJobs {
					if !j.BelongsToPipeline() {
						if j.Disable {
							srv.logger.Infoln("schedule JOB Is Disable name:", j.Name)
							continue
						} else {
							srv.logger.Infoln("schedule JOB IS Not Disable", j.Name)
						}
					} else {
						p, _ := srv.storage.GetPipeline(j.PipelineID)
						if p.IsDisabled() { // reset the pipeline
							srv.storage.RecyclePipeline(j.PipelineID, p) // reset the pipeline
							continue
						}
						if p.CancelOnFailure() { // quite pipeline if a job has failed
							if p.Status == model.Failed {
								srv.storage.RecyclePipeline(j.PipelineID, p) // reset the pipeline
								continue
							}
						}
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
