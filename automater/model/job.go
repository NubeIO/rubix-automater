package model

import (
	"fmt"
	taskRepo "github.com/NubeIO/rubix-automater/automater/service/tasksrv/taskrepo"
	"strings"
	"time"
)

/*
Recycle job
if the job is to be marked as completed and the job is set to be a CRON job then do the following

- set state back to IN_PROGRESS
- bump the Job.RunCount up one
- reset the Job.RunAt time to completed plus the next interval eg: completed at 3pm and interval is 15min then set RunAt=03:15

Rest on fail
- if the job has failed and the OnFailRetry is true
- bump the Job.FailCount up one
- reset the Job.RunAt time to completed plus the next interval eg: completed at 3pm and interval is 15min then set RunAt=03:15

*/

type JobOptions struct {
	EnableInterval    bool       `json:"enable_interval"`
	RunOnInterval     string     `json:"run_on_interval"`
	EnableOnFailRetry bool       `json:"enable_on_fail_retry"`
	HowTimesToRetry   bool       `json:"how_times_to_retry"`
	OnFailRetryDelay  *time.Time `json:"birth,omitempty"`
}

// Job represents an async tasks.
type Job struct {
	UUID    string `json:"uuid"`
	Name    string `json:"name"`
	Disable bool   `json:"disable"`

	//stats on run and error count
	RunCount  int `json:"run_count"`
	FailCount int `json:"fail_count"`

	JobOptions *JobOptions `json:"job_options"`

	// PipelineID is the auto-generated pipeline identifier in UUID4 format.
	PipelineID string `json:"pipeline_id,omitempty"`

	// NextJobID is the UUID of the job that should run next in the pipeline, if any.
	NextJobID string `json:"next_job_id,omitempty"`

	// UsePreviousResults indicates where the job should use the
	UsePreviousResults bool `json:"use_previous_results,omitempty"`

	// Next points to the next job of the pipeline, if any.
	Next *Job `json:"next,omitempty"`

	// TaskName is the name of the tasks to be executed.
	TaskName string `json:"task_name"`

	// TaskParams are the required parameters for the tasks assigned to the specific job.
	TaskParams map[string]interface{} `json:"task_params,omitempty"`

	// Timeout is the ttime in seconds after which the job tasks will be interrupted.
	Timeout int `json:"timeout_in_sec,omitempty"`

	Description string `json:"description,omitempty"`
	// Status represents the status of the job.

	Status JobStatus `json:"status"`

	FailureReason string `json:"failure_reason,omitempty"`
	// RunAt is the UTC timestamp indicating the ttime for the job to run.
	RunAt *time.Time `json:"run_at,omitempty"`
	// RunAt is like run every 15min
	// ScheduledAt is the UTC timestamp indicating the ttime that the job got scheduled.
	ScheduledAt *time.Time `json:"scheduled_at"`
	// CreatedAt is the UTC timestamp of the job creation.
	CreatedAt *time.Time `json:"created_at,omitempty"`
	// StartedAt is the UTC timestamp of the moment the job started.
	StartedAt *time.Time `json:"started_at,omitempty"`
	// CompletedAt is the UTC timestamp of the moment the job finished.
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	// Duration indicates how much the job took to complete.
	Duration *time.Duration `json:"duration,omitempty"`
}

// NewJob initializes and returns a new Job instance.
func NewJob(
	uuid, name, taskName, description, pipelineID, nextJobID string,
	timeout int, runAt *time.Time, createdAt *time.Time,
	usePreviousResults bool, disable bool, options *JobOptions, taskParams map[string]interface{}) *Job {

	if runAt.IsZero() {
		runAt = nil
	}
	return &Job{
		UUID:               uuid,
		Name:               name,
		TaskName:           taskName,
		PipelineID:         pipelineID,
		NextJobID:          nextJobID,
		UsePreviousResults: usePreviousResults,
		Timeout:            timeout,
		Description:        description,
		Disable:            disable,
		JobOptions:         options,
		TaskParams:         taskParams,
		Status:             Pending,
		RunAt:              runAt,
		CreatedAt:          createdAt,
	}
}

func (j *Job) MarkPending() {
	j.Status = Pending
}

// MarkStarted updates the status and timestamp at the moment the job started.
func (j *Job) MarkStarted(startedAt *time.Time) {
	j.Status = InProgress
	j.StartedAt = startedAt
}

// MarkScheduled updates the status and timestamp at the moment the job got scheduled.
func (j *Job) MarkScheduled(scheduledAt *time.Time) {
	j.Status = Scheduled
	j.ScheduledAt = scheduledAt
}

// MarkCompleted updates the status and timestamp at the moment the job finished.
func (j *Job) MarkCompleted(completedAt *time.Time) {
	j.Status = Completed
	j.CompletedAt = completedAt
}

// MarkRecycle will set the job back to pending
func (j *Job) MarkRecycle() {
	j.Status = InProgress
}

// MarkFailed updates the status and timestamp at the moment the job failed.
func (j *Job) MarkFailed(failedAt *time.Time, reason string) {
	j.Status = Failed
	j.FailureReason = reason
	j.CompletedAt = failedAt
}

// SetDuration sets the duration of the job if it's completed of failed.
func (j *Job) SetDuration() {
	if j.Status == Completed || j.Status == Failed {
		duration := j.CompletedAt.Sub(*j.StartedAt) / time.Millisecond
		j.Duration = &duration
	}
}

// Validate perfoms basic sanity checks on the job request payload.
func (j *Job) Validate(taskRepo *taskRepo.TaskRepository) error {
	var required []string

	if j.Name == "" {
		required = append(required, "name")
	}

	if j.RunAt == nil {
		required = append(required, "run_at")
	}

	if j.TaskName == "" {
		required = append(required, "task_name")
	}

	if len(required) > 0 {
		return fmt.Errorf(strings.Join(required, ", ") + " required")
	}

	_, err := taskRepo.GetTaskFunc(j.TaskName)
	if err != nil {
		taskNames := taskRepo.GetTaskNames()
		return fmt.Errorf("%s is not a valid tasks name - valid tasks: %v", j.TaskName, taskNames)
	}

	if j.Status != Undefined {
		err := j.Status.Validate()
		if err != nil {
			return err
		}
	}
	return nil
}

func (j *Job) IsScheduled() bool {
	return j.RunAt != nil
}

func (j *Job) HasNext() bool {
	return j.NextJobID != ""
}

func (j *Job) BelongsToPipeline() bool {
	return j.PipelineID != ""
}
