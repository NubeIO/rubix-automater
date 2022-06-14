package jobsrv

import (
	"encoding/json"
	"fmt"
	"github.com/NubeIO/rubix-automater/automater"
	"github.com/NubeIO/rubix-automater/automater/model"
	taskRepo "github.com/NubeIO/rubix-automater/automater/service/tasksrv/taskrepo"
	"github.com/NubeIO/rubix-automater/pkg/helpers/apperrors"
	intime "github.com/NubeIO/rubix-automater/pkg/helpers/ttime"
	"github.com/NubeIO/rubix-automater/pkg/helpers/uuid"
	"strings"
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
	name, taskName, subTaskName, description string, scheduleAt string,
	timeout int, disable bool, options *model.JobOptions, taskParams map[string]interface{}) (*model.Job, error) {
	id, _ := srv.uuidGen.Make("job")

	runAt, err := automater.RunAt(scheduleAt)
	if err != nil {
		return nil, err
	}

	fmt.Println("RUN, AT", runAt)

	createdAt := srv.time.Now()
	j := model.NewJob(
		id, name, taskName, subTaskName, description,
		"", "", timeout, &runAt,
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
func (srv *jobService) Update(uuid string, body *model.Job) (*model.Job, error) {
	_, err := srv.storage.GetJob(uuid)
	if err != nil {
		return nil, err
	}
	return srv.storage.UpdateJob(uuid, body)
}

// Recycle reuse a job.
func (srv *jobService) Recycle(uuid string, body *model.Job) (*model.Job, error) {
	_, err := srv.storage.GetJob(uuid)
	if err != nil {
		return nil, err
	}
	return srv.storage.Recycle(uuid, body)
}

// Delete deletes a job.
func (srv *jobService) Delete(uuid string) error {
	j, err := srv.storage.GetJob(uuid)
	if err != nil {
		return err
	}
	if j.BelongsToPipeline() {
		// TODO: Add a testing case for this scenario.
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
