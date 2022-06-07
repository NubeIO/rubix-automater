package model

import (
	"fmt"
	"strings"
	"time"
)

type PipelineOptions struct {
	EnableInterval    bool   `json:"enable_interval"`
	RunOnInterval     string `json:"run_on_interval"`
	EnableOnFailRetry bool   `json:"enable_on_fail_retry"`
	DelayBetweenTask  int    `json:"delay_between_task_in_min"`
	HowTimesToRetry   bool   `json:"how_times_to_retry"`
	OnFailRetryDelay  string `json:"birth,omitempty"`
}

// Pipeline represents a sequence of async tasks.
type Pipeline struct {
	UUID        string `json:"uuid"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`

	PipelineOptions *PipelineOptions `json:"options"`

	Jobs []*Job `json:"jobs,omitempty"`

	Status JobStatus `json:"status"`

	// RunAt is the UTC timestamp indicating the ttime for the pipeline to run.
	RunAt *time.Time `json:"run_at,omitempty"`

	// CreatedAt is the UTC timestamp of the pipeline creation.
	CreatedAt *time.Time `json:"created_at,omitempty"`

	// StartedAt is the UTC timestamp of the moment the pipeline started.
	StartedAt *time.Time `json:"started_at,omitempty"`

	// CompletedAt is the UTC timestamp of the moment the pipeline finished.
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// Duration indicates how much the pipeline took to complete.
	Duration *time.Duration `json:"duration,omitempty"`
}

func NewPipeline(uuid, name, description string, pipelineOptions *PipelineOptions, jobs []*Job, createdAt *time.Time) *Pipeline {
	p := &Pipeline{
		UUID:            uuid,
		Name:            name,
		Description:     description,
		PipelineOptions: pipelineOptions,
		Jobs:            jobs,
		Status:          Pending,
		CreatedAt:       createdAt,
	}
	return p
}

// MarkStarted updates the status and timestamp at the moment the pipeline started.
func (p *Pipeline) MarkStarted(startedAt *time.Time) {
	p.Status = InProgress
	p.StartedAt = startedAt
}

// MarkCompleted updates the status and timestamp at the moment the pipeline finished.
func (p *Pipeline) MarkCompleted(completedAt *time.Time) {
	p.Status = Completed
	p.CompletedAt = completedAt
}

// MarkFailed updates the status and timestamp at the moment the pipeline failed.
func (p *Pipeline) MarkFailed(failedAt *time.Time) {
	p.Status = Failed
	p.CompletedAt = failedAt
}

// SetDuration sets the duration of the pipeline if it's completed of failed.
func (p *Pipeline) SetDuration() {
	if p.Status == Completed || p.Status == Failed {
		duration := p.CompletedAt.Sub(*p.StartedAt) / time.Millisecond
		p.Duration = &duration
	}
}

// Validate performs basic sanity checks on the pipeline request payload.
func (p *Pipeline) Validate() error {
	var required []string

	if p.Name == "" {
		required = append(required, "name")
	}

	if len(p.Jobs) == 0 {
		required = append(required, "jobs")
	}

	if len(required) > 0 {
		return fmt.Errorf(strings.Join(required, ", ") + " required")
	}

	if len(p.Jobs) == 1 {
		return fmt.Errorf("pipeline shoud have at least 2 jobs, %d given", len(p.Jobs))
	}

	if p.Status != Undefined {
		err := p.Status.Validate()
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Pipeline) IsScheduled() bool {
	return p.RunAt != nil
}

func (p *Pipeline) MergeJobsInOne() {
	for i := 0; i < len(p.Jobs)-1; i++ {
		p.Jobs[i].Next = p.Jobs[i+1]
	}
}

func (p *Pipeline) UnmergeJobs() {
	for _, j := range p.Jobs {
		j.Next = nil
	}
}
