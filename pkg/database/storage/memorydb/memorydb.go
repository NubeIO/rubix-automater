package memorydb

import (
	"encoding/json"
	"github.com/NubeIO/rubix-automater/automater"
	"github.com/NubeIO/rubix-automater/automater/model"
	"github.com/NubeIO/rubix-automater/pkg/helpers/apperrors"
	"sort"
	"time"
)

var _ automater.Storage = &memorydb{}

type memorydb struct {
	pipelinedb  map[string][]byte
	jobdb       map[string][]byte
	jobresultdb map[string][]byte
}

// New NewMemoryDB creates a new instance.
func New() *memorydb {
	return &memorydb{
		pipelinedb:  make(map[string][]byte),
		jobdb:       make(map[string][]byte),
		jobresultdb: make(map[string][]byte),
	}
}

// CheckHealth checks if the storage is alive.
func (mem *memorydb) CheckHealth() bool {
	return mem.jobdb != nil && mem.jobresultdb != nil && mem.pipelinedb != nil
}

// Close terminates any storage connections gracefully.
func (mem *memorydb) Close() error {
	mem.jobdb = nil
	mem.jobresultdb = nil
	mem.pipelinedb = nil
	return nil
}

// CreateJob adds a new job to the storage.
func (mem *memorydb) CreateJob(j *model.Job) error {
	serializedJob, err := json.Marshal(j)
	if err != nil {
		return err
	}
	mem.jobdb[j.UUID] = serializedJob
	return nil
}

// GetJob fetches a job from the storage.
func (mem *memorydb) GetJob(uuid string) (*model.Job, error) {
	serializedJob, ok := mem.jobdb[uuid]
	if !ok {
		return nil, &apperrors.NotFoundErr{UUID: uuid, ResourceName: "job"}
	}
	j := &model.Job{}
	if err := json.Unmarshal(serializedJob, j); err != nil {
		return nil, err
	}
	return j, nil
}

// GetJobs fetches all jobs from the storage, optionally filters the jobs by status.
func (mem *memorydb) GetJobs(status model.JobStatus) ([]*model.Job, error) {
	jobs := []*model.Job{}
	for _, serializedJob := range mem.jobdb {
		j := &model.Job{}
		if err := json.Unmarshal(serializedJob, j); err != nil {
			return nil, err
		}
		if status == model.Undefined || j.Status == status {
			jobs = append(jobs, j)
		}
	}
	// ORDER BY created_at ASC
	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].CreatedAt.Before(*jobs[j].CreatedAt)
	})
	return jobs, nil
}

// GetJobsByPipelineID fetches the jobs of the specified pipeline.
func (mem *memorydb) GetJobsByPipelineID(pipelineID string) ([]*model.Job, error) {
	serializedPipeline, ok := mem.pipelinedb[pipelineID]
	if !ok {
		// Mimic the behavior or the relational databases.
		return []*model.Job{}, nil
	}
	p := &model.Pipeline{}
	if err := json.Unmarshal(serializedPipeline, p); err != nil {
		return nil, err
	}
	return p.Jobs, nil
}

// UpdateJob updates a job to the storage.
func (mem *memorydb) UpdateJob(uuid string, j *model.Job) (*model.Job, error) {
	serializedJob, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}
	mem.jobdb[j.UUID] = serializedJob

	if j.BelongsToPipeline() {
		// Sync pipeline job
		serializedPipeline := mem.pipelinedb[j.PipelineID]
		p := &model.Pipeline{}
		if err := json.Unmarshal(serializedPipeline, p); err != nil {
			return nil, err
		}
		for i, job := range p.Jobs {
			if job.UUID == j.UUID {
				p.Jobs[i] = j
			}
		}
		serializedPipeline, err := json.Marshal(p)
		if err != nil {
			return nil, err
		}
		mem.pipelinedb[p.UUID] = serializedPipeline
	}
	return nil, nil
}

// DeleteJob deletes a job from the storage.
func (mem *memorydb) DeleteJob(uuid string) error {
	if _, ok := mem.jobdb[uuid]; !ok {
		return &apperrors.NotFoundErr{UUID: uuid, ResourceName: "job"}
	}
	delete(mem.jobdb, uuid)
	// CASCADE
	delete(mem.jobresultdb, uuid)
	return nil
}

// GetDueJobs fetches all jobs scheduled to run before now and have not been scheduled yet.
func (mem *memorydb) GetDueJobs() ([]*model.Job, error) {
	dueJobs := []*model.Job{}
	for _, serializedJob := range mem.jobdb {
		j := &model.Job{}
		if err := json.Unmarshal(serializedJob, j); err != nil {
			return nil, err
		}
		if j.IsScheduled() {
			if j.RunAt.Before(time.Now()) && j.Status == model.Pending {
				dueJobs = append(dueJobs, j)
			}
		}
	}
	// ORDER BY run_at ASC
	sort.Slice(dueJobs, func(i, j int) bool {
		return dueJobs[i].RunAt.Before(*dueJobs[j].RunAt)
	})
	return dueJobs, nil
}

// CreateJobResult adds new job result to the storage.
func (mem *memorydb) CreateJobResult(result *model.JobResult) error {
	serializedJobResult, err := json.Marshal(result)
	if err != nil {
		return err
	}
	mem.jobresultdb[result.JobID] = serializedJobResult
	return nil
}

// GetJobResult fetches a job result from the storage.
func (mem *memorydb) GetJobResult(jobID string) (*model.JobResult, error) {
	serializedJobResult, ok := mem.jobresultdb[jobID]
	if !ok {
		return nil, &apperrors.NotFoundErr{UUID: jobID, ResourceName: "job result"}
	}
	result := &model.JobResult{}
	json.Unmarshal(serializedJobResult, result)
	return result, nil
}

// UpdateJobResult updates a job result to the storage.
func (mem *memorydb) UpdateJobResult(jobID string, result *model.JobResult) error {
	serializedJobResult, err := json.Marshal(result)
	if err != nil {
		return err
	}
	mem.jobresultdb[result.JobID] = serializedJobResult
	return nil
}

// DeleteJobResult deletes a job result from the storage.
func (mem *memorydb) DeleteJobResult(uuid string) error {
	if _, ok := mem.jobresultdb[uuid]; !ok {
		return &apperrors.NotFoundErr{UUID: uuid, ResourceName: "job result"}
	}
	delete(mem.jobresultdb, uuid)
	return nil
}

// CreatePipeline adds a new pipeline and of its jobs to the storage.
func (mem *memorydb) CreatePipeline(p *model.Pipeline) error {
	serializedPipeline, err := json.Marshal(p)
	if err != nil {
		return err
	}
	mem.pipelinedb[p.UUID] = serializedPipeline
	for _, j := range p.Jobs {
		serializedJob, err := json.Marshal(j)
		if err != nil {
			return err
		}
		mem.jobdb[j.UUID] = serializedJob
	}
	return nil
}

// GetPipeline fetches a pipeline from the storage.
func (mem *memorydb) GetPipeline(uuid string) (*model.Pipeline, error) {
	serializedPipeline, ok := mem.pipelinedb[uuid]
	if !ok {
		return nil, &apperrors.NotFoundErr{UUID: uuid, ResourceName: "pipeline"}
	}
	p := &model.Pipeline{}
	if err := json.Unmarshal(serializedPipeline, p); err != nil {
		return nil, err
	}
	return p, nil
}

// GetPipelines fetches all pipelines from the storage optionally filters the pipelines by status.
func (mem *memorydb) GetPipelines(status model.JobStatus) ([]*model.Pipeline, error) {
	pipelines := []*model.Pipeline{}
	for _, serializedPipeline := range mem.pipelinedb {
		p := &model.Pipeline{}
		if err := json.Unmarshal(serializedPipeline, p); err != nil {
			return nil, err
		}
		if status == model.Undefined || p.Status == status {
			pipelines = append(pipelines, p)
		}
	}
	// ORDER BY created_at ASC
	sort.Slice(pipelines, func(i, j int) bool {
		return pipelines[i].CreatedAt.Before(*pipelines[j].CreatedAt)
	})
	return pipelines, nil
}

// UpdatePipeline updates a pipeline to the storage.
func (mem *memorydb) UpdatePipeline(uuid string, p *model.Pipeline) error {
	serializedPipeline, err := json.Marshal(p)
	if err != nil {
		return err
	}
	mem.pipelinedb[p.UUID] = serializedPipeline
	return nil
}

// DeletePipeline deletes a pipeline and all its jobs from the storage.
func (mem *memorydb) DeletePipeline(uuid string) error {
	serializedPipeline, ok := mem.pipelinedb[uuid]
	if !ok {
		return &apperrors.NotFoundErr{UUID: uuid, ResourceName: "pipeline"}
	}
	p := &model.Pipeline{}
	err := json.Unmarshal(serializedPipeline, &p)
	if err != nil {
		return err
	}
	for _, j := range p.Jobs {
		// CASCADE
		delete(mem.jobdb, j.UUID)
		delete(mem.jobresultdb, j.UUID)
	}
	delete(mem.pipelinedb, uuid)
	return nil
}
