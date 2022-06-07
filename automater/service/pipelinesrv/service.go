package pipelinesrv

import (
	"encoding/json"
	"github.com/NubeIO/rubix-automater/automater"
	"github.com/NubeIO/rubix-automater/automater/model"
	taskRepo "github.com/NubeIO/rubix-automater/automater/service/tasksrv/taskrepo"
	"github.com/NubeIO/rubix-automater/pkg/helpers/apperrors"
	"github.com/NubeIO/rubix-automater/pkg/helpers/timeconversion"
	ttime "github.com/NubeIO/rubix-automater/pkg/helpers/ttime"
	"github.com/NubeIO/rubix-automater/pkg/helpers/uuid"
	"github.com/jmattheis/go-timemath"
	"strings"
	"time"
)

var _ automater.PipelineService = &pipeLineService{}

type pipeLineService struct {
	storage  automater.Storage
	taskRepo *taskRepo.TaskRepository
	uuidGen  uuid.Generator
	time     ttime.Time
}

// New creates a new pipeline server.
func New(
	storage automater.Storage,
	taskRepo *taskRepo.TaskRepository,
	uuidGen uuid.Generator,
	time ttime.Time) *pipeLineService {
	return &pipeLineService{
		storage:  storage,
		taskRepo: taskRepo,
		uuidGen:  uuidGen,
		time:     time,
	}
}

func RunAt(runAt string, p *model.PipelineOptions, index int) (time.Time, error) {
	now := ttime.New().Now()
	var addDelay bool
	if p != nil && index > 0 {
		if p.DelayBetweenTask > 0 {
			addDelay = true
		}
	}
	if runAt != "" {
		timeChecked, err := time.Parse(time.RFC3339Nano, runAt)
		if err != nil {
			nextRunTime, err := timeconversion.AdjustTime(ttime.New().Now(), runAt) // user can just pass in 15 sec on the HTTP POST RunAt
			if addDelay {
				nextRunTime = timemath.Minute.Add(nextRunTime, p.DelayBetweenTask*index)
			}
			if err != nil {
				return now, &apperrors.ParseTimeErr{Message: err.Error()}
			} else {
				return nextRunTime, nil
			}
		} else {
			return timeChecked, nil
		}
	} else {
		return now, nil
	}

}

// Create creates a new pipeline.
func (srv *pipeLineService) Create(name, description, runAt string, pipelineOptions *model.PipelineOptions, jobs []*model.Job) (*model.Pipeline, error) {
	pipelineUUID, err := srv.uuidGen.Make("pip")
	if err != nil {
		return nil, err
	}
	jobIDs := make([]string, 0)
	for i := 0; i < len(jobs); i++ { //make the pipeline jobs
		jobUUID, err := srv.uuidGen.Make("job")
		if err != nil {
			return nil, err
		}
		jobIDs = append(jobIDs, jobUUID)
	}
	jobsToCreate := make([]*model.Job, 0)
	for i, job := range jobs {
		now, err := RunAt(runAt, pipelineOptions, i)
		runAtTime := now.Add(time.Millisecond * time.Duration(i+2)) // in db GetDueJobs it orders by time desc, so we need a small buffer (this is a hack)
		if err != nil {
			return nil, err
		}
		jobID := jobIDs[i]
		nextJobID := ""
		if i < len(jobs)-1 {
			nextJobID = jobIDs[i+1]
		}
		createdAt := srv.time.Now()
		j := model.NewJob(
			jobID, job.Name, job.TaskName, job.Description, pipelineUUID, nextJobID,
			job.Timeout, &runAtTime, &createdAt, job.UsePreviousResults, job.Disable, job.JobOptions, job.TaskParams)
		if err := j.Validate(srv.taskRepo); err != nil {
			return nil, &apperrors.ResourceValidationErr{Message: err.Error()}
		}
		jobsToCreate = append(jobsToCreate, j)
	}
	createdAt := srv.time.Now()
	p := model.NewPipeline(pipelineUUID, name, description, pipelineOptions, jobsToCreate, &createdAt)

	if err := p.Validate(); err != nil {
		return nil, &apperrors.ResourceValidationErr{Message: err.Error()}
	}
	// Inherit first job's schedule timestamp.
	p.RunAt = jobsToCreate[0].RunAt

	if err := srv.storage.CreatePipeline(p); err != nil {
		return nil, err
	}
	return p, nil
}

func (srv *pipeLineService) RecyclePipeline(uuid string, p *model.Pipeline) (*model.Pipeline, error) {
	return srv.storage.RecyclePipeline(uuid, p)

}

func (srv *pipeLineService) RecycleJob(uuid string, p *model.Job) (*model.Job, error) {
	return srv.storage.Recycle(uuid, p)
}

// Get fetches a pipeline.
func (srv *pipeLineService) Get(uuid string) (*model.Pipeline, error) {
	p, err := srv.storage.GetPipeline(uuid)
	if err != nil {
		return nil, err
	}
	p.SetDuration()
	// Do not marshal pipeline jobs cause NoSQL databases
	// store them along with the pipeline.
	p.Jobs = nil
	return p, nil
}

// GetPipelineJobs fetches the jobs of a specified pipeline.
func (srv *pipeLineService) GetPipelineJobs(uuid string) ([]*model.Job, error) {
	_, err := srv.storage.GetPipeline(uuid)
	if err != nil {
		return nil, err
	}
	jobs, err := srv.storage.GetJobsByPipelineID(uuid)
	if err != nil {
		return nil, err
	}

	for _, j := range jobs {
		j.SetDuration()
		j.Next = nil
	}
	return jobs, nil
}

// GetPipelines fetches all pipelines, optionally filters the pipelines by status.
func (srv *pipeLineService) GetPipelines(status string) ([]*model.Pipeline, error) {
	var pipelineStatus model.JobStatus
	if status == "" {
		pipelineStatus = model.Undefined
	} else {
		err := json.Unmarshal([]byte("\""+strings.ToUpper(status)+"\""), &pipelineStatus)
		if err != nil {
			return nil, err
		}
	}
	pipelines, err := srv.storage.GetPipelines(pipelineStatus)
	if err != nil {
		return nil, err
	}
	for _, p := range pipelines {
		p.SetDuration()
		// Do not marshal pipeline jobs cause NoSQL databases
		// store them along with the pipeline.
		p.Jobs = nil
	}

	return pipelines, nil
}

// Update updates a pipeline.
func (srv *pipeLineService) Update(uuid, name, description string) error {
	p, err := srv.storage.GetPipeline(uuid)
	if err != nil {
		return err
	}
	p.Name = name
	p.Description = description
	return srv.storage.UpdatePipeline(uuid, p)
}

// Delete deletes a pipeline.
func (srv *pipeLineService) Delete(uuid string) error {
	_, err := srv.storage.GetPipeline(uuid)
	if err != nil {
		return err
	}
	return srv.storage.DeletePipeline(uuid)
}
