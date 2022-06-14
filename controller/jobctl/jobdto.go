package jobctl

import (
	"github.com/NubeIO/rubix-automater/automater/model"
)

// JobBody is the data transfer object used for a job creation or update.
type JobBody struct {
	Name               string                 `json:"name"`
	Description        string                 `json:"description"`
	Disable            bool                   `json:"disable"`
	TaskName           string                 `json:"task_name"`
	SubTaskName        string                 `json:"sub_task"`
	ScheduleAt         string                 `json:"schedule_at"`
	Timeout            int                    `json:"timeout"`
	Options            *model.JobOptions      `json:"options"`
	TaskParams         map[string]interface{} `json:"task_params"`
	UsePreviousResults bool                   `json:"use_previous_results"`
}

// NewRequestBodyDTO initializes and returns a new BodyDTO instance.
func NewRequestBodyDTO() *JobBody {
	return &JobBody{}
}

// ResponseBodyDTO is the response data transfer object used for a job creation or update.
type ResponseBodyDTO *model.Job

// BuildResponseBodyDTO creates a new ResponseDTO.
func BuildResponseBodyDTO(resource *model.Job) ResponseBodyDTO {
	return ResponseBodyDTO(resource)
}
