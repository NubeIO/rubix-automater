package jobsrv

import (
	"encoding/json"
	"fmt"
	"github.com/NubeIO/rubix-automater/automater"
	"github.com/NubeIO/rubix-automater/automater/core"
	taskRepo "github.com/NubeIO/rubix-automater/automater/service/tasksrv/taskrepo"
	"github.com/NubeIO/rubix-automater/pkg/helpers/apperrors"
	intime "github.com/NubeIO/rubix-automater/pkg/helpers/intime"
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

// New creates a new job service.
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
	timeout int, taskParams map[string]interface{}) (*core.Job, error) {
	var runAtTime time.Time

	uuid, err := srv.uuidGen.Make("job")
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
	j := core.NewJob(
		uuid, name, taskName, description,
		"", "", timeout, &runAtTime,
		&createdAt, false, taskParams)

	if err := j.Validate(srv.taskRepo); err != nil {
		return nil, &apperrors.ResourceValidationErr{Message: err.Error()}
	}

	if err := srv.storage.CreateJob(j); err != nil {
		return nil, err
	}
	return j, nil
}

// Get fetches a job.
func (srv *jobService) Get(uuid string) (*core.Job, error) {
	j, err := srv.storage.GetJob(uuid)
	if err != nil {
		return nil, err
	}
	j.SetDuration()
	// Do not marshal job next, cause it's stored in NoSQL databases.
	j.Next = nil
	return j, nil
}

// GetJobs fetches all jobs, optionally filters the jobs by status.
func (srv *jobService) GetJobs(status string) ([]*core.Job, error) {
	var jobStatus core.JobStatus
	if status == "" {
		jobStatus = core.Undefined
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
func (srv *jobService) Update(uuid, name, description string) error {
	j, err := srv.storage.GetJob(uuid)
	if err != nil {
		return err
	}
	j.Name = name
	j.Description = description
	return srv.storage.UpdateJob(uuid, j)
}

// UpdateAll updates a job.
func (srv *jobService) UpdateAll(uuid string, body *core.Job) error {
	j, err := srv.storage.GetJob(uuid)
	if err != nil {
		return err
	}
	createdAt := srv.time.Now()
	body.UUID = j.UUID
	body.CreatedAt = &createdAt
	return srv.storage.UpdateJob(uuid, body)
}

// Delete deletes a job.
func (srv *jobService) Delete(uuid string) error {
	j, err := srv.storage.GetJob(uuid)
	if err != nil {
		return err
	}
	if j.BelongsToPipeline() {
		// TODO: Add a test case for this scenario.
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
