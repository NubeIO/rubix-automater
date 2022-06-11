package model

import "time"

// Transaction represents a result of a job.
type Transaction struct {
	UUID string `json:"uuid"`

	// JobID is the auto-generated pipeline identifier in UUID4 format.
	JobID string `json:"job_id"`

	IsPipeLine bool `json:"is_pipe_line"`

	TaskType string `json:"task_type"`

	// PipelineID is the auto-generated pipeline identifier in UUID4 format.
	PipelineID string `json:"pipeline_id,omitempty"`

	Status JobStatus `json:"status"`

	FailureReason string `json:"failure_reason,omitempty"`

	CreatedAt *time.Time `json:"created_at,omitempty"`
	// StartedAt is the UTC timestamp of the moment the job started.
	StartedAt *time.Time `json:"started_at,omitempty"`
	// CompletedAt is the UTC timestamp of the moment the job finished.
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	// LastRecyclerCreation last time the job was recycled at
	Duration *time.Duration `json:"duration,omitempty"`
}

// PublishTransaction represents a result of a job.
type PublishTransaction struct {
	UUID string `json:"uuid"`

	// JobID is the auto-generated pipeline identifier in UUID4 format.
	JobID string `json:"job_id"`

	IsPipeLine bool `json:"is_pipe_line"`

	TaskType string `json:"task_type"`

	// PipelineID is the auto-generated pipeline identifier in UUID4 format.
	PipelineID string `json:"pipeline_id,omitempty"`

	Status string `json:"status"`

	FailureReason string `json:"failure_reason,omitempty"`

	CreatedAt *time.Time `json:"created_at,omitempty"`
	// StartedAt is the UTC timestamp of the moment the job started.
	StartedAt *time.Time `json:"started_at,omitempty"`
	// CompletedAt is the UTC timestamp of the moment the job finished.
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	// LastRecyclerCreation last time the job was recycled at
	Duration *time.Duration `json:"duration,omitempty"`
}
