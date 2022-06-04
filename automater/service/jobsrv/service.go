package jobsrv

import (
	"encoding/json"
	"fmt"
	"github.com/NubeIO/rubix-automater/automater"
	"github.com/NubeIO/rubix-automater/automater/model"
	taskRepo "github.com/NubeIO/rubix-automater/automater/service/tasksrv/taskrepo"
	"github.com/NubeIO/rubix-automater/pkg/helpers/apperrors"
	"github.com/NubeIO/rubix-automater/pkg/helpers/timeconversion"
	intime "github.com/NubeIO/rubix-automater/pkg/helpers/ttime"
	"github.com/NubeIO/rubix-automater/pkg/helpers/uuid"
	"strings"
	"time"
)

var _ automater.JobService = &jobService{}

type jobService struct {
	storage  automater.Storage
	taskRepo *taskRepo.TaskRepository
	uuidGen  uuid.Generator
	time     intime.Time
}

// New creates a new job server.
func New(
	storage automater.Storage,
	taskRepo *taskRepo.TaskRepository,
	uuidGen uuid.Generator,
	time intime.Time) *jobService {
	return &jobService{
		storage:  storage,
		taskRepo: taskRepo,
		uuidGen:  uuidGen,
		time:     time,
	}
}

// Create creates a new job.
func (srv *jobService) Create(
	name, taskName, description, runAt string,
	timeout int, disable bool, options *model.JobOptions, taskParams map[string]interface{}) (*model.Job, error) {
	var runAtTime time.Time

	id, err := srv.uuidGen.Make("job")
	if err != nil {
		return nil, err
	}
	if runAt != "" {
		runAtTime, err = time.Parse(time.RFC3339Nano, runAt)
		if err != nil {
			return nil, &apperrors.ParseTimeErr{Message: err.Error()}
		}
	}
	createdAt := srv.time.Now()
	j := model.NewJob(
		id, name, taskName, description,
		"", "", timeout, &runAtTime,
		&createdAt, false, disable, options, taskParams)

	if err := j.Validate(srv.taskRepo); err != nil {
		return nil, &apperrors.ResourceValidationErr{Message: err.Error()}
	}

	if err := srv.storage.CreateJob(j); err != nil {
		return nil, err
	}
	return j, nil
}

// Get fetches a job.
func (srv *jobService) Get(uuid string) (*model.Job, error) {
	j, err := srv.storage.GetJob(uuid)
	if err != nil {
		return nil, err
	}
	j.SetDuration()
	// Do not marshal job next, because it's stored in NoSQL databases.
	j.Next = nil
	return j, nil
}

// GetJobs fetches all jobs, optionally filters the jobs by status.
func (srv *jobService) GetJobs(status string) ([]*model.Job, error) {
	var jobStatus model.JobStatus
	if status == "" {
		jobStatus = model.Undefined
	} else {
		err := json.Unmarshal([]byte("\""+strings.ToUpper(status)+"\""), &jobStatus)
		if err != nil {
			return nil, err
		}
	}
	jobs, err := srv.storage.GetJobs(jobStatus)
	if err != nil {
		return nil, err
	}
	for _, j := range jobs {
		j.SetDuration()
		// Do not marshal job next, cause it's stored in NoSQL databases.
		j.Next = nil
	}
	return jobs, nil
}

// Update updates a job.
func (srv *jobService) Update(uuid, name, description string) (*model.Job, error) {
	j, err := srv.storage.GetJob(uuid)
	if err != nil {
		return nil, err
	}
	j.Name = name
	j.Description = description
	return srv.storage.UpdateJob(uuid, j)
}

// Recycle reuse a job.
func (srv *jobService) Recycle(uuid string, body *model.Job) (*model.Job, error) {
	fmt.Println(22222, uuid)
	j, err := srv.storage.GetJob(uuid)
	if err != nil {
		return nil, err
	}
	var nextRunTime time.Time
	//if the job was completed and is enabled as cron
	if j.JobOptions != nil {
		if j.CompletedAt != nil {
			nextRunTime, err = timeconversion.AdjustTime(*j.CompletedAt, j.JobOptions.RunOnInterval)
			if err != nil {
				return nil, err
			}
			j.RunAt = &nextRunTime
		}
	}

	if j.RunAt != nil {
		//runAtTime, err := time.Parse(time.RFC3339Nano, body.RunAt.String())
		//if err != nil {
		//	return nil, &apperrors.ParseTimeErr{Message: err.Error()}
		//}
		//body.RunAt = &runAtTime
	} else {
		//body.RunAt = &now
	}

	j.Status = model.Pending
	j.ScheduledAt = nil
	j.StartedAt = nil
	j.CompletedAt = nil
	return srv.storage.Recycle(uuid, j)
}

// Delete deletes a job.
func (srv *jobService) Delete(uuid string) error {
	j, err := srv.storage.GetJob(uuid)
	if err != nil {
		return err
	}
	if j.BelongsToPipeline() {
		// TODO: Add a pubsub case for this scenario.
		return &apperrors.CannotDeletePipelineJobErr{
			Message: fmt.Sprintf(
				`job with UUID: %s can not be deleted because it belongs 
				to a pipeline - try to delete the pipeline instead`, uuid),
		}
	}
	return srv.storage.DeleteJob(uuid)
}

func (srv *jobService) Drop() error {
	jobs, err := srv.GetJobs("")
	if err != nil {
		return err
	}
	for _, job := range jobs {
		srv.Delete(job.UUID)
	}
	return nil
}
