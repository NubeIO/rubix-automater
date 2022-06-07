package jobctl

import (
	"github.com/NubeIO/rubix-automater/automater/model"
)

// RequestBodyDTO is the data transfer object used for a job creation or update.
type RequestBodyDTO struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Disable     bool                   `json:"disable"`
	TaskName    string                 `json:"task_name"`
	ScheduleAt  string                 `json:"schedule_at"`
	Timeout     int                    `json:"timeout"`
	Options     *model.JobOptions      `json:"options"`
	TaskParams  map[string]interface{} `json:"task_params"`
	//RunAt              string                 `json:"run_at"`
	UsePreviousResults bool `json:"use_previous_results"`
}

// NewRequestBodyDTO initializes and returns a new BodyDTO instance.
func NewRequestBodyDTO() *RequestBodyDTO {
	return &RequestBodyDTO{}
}

// ResponseBodyDTO is the response data transfer object used for a job creation or update.
type ResponseBodyDTO *model.Job

// BuildResponseBodyDTO creates a new ResponseDTO.
func BuildResponseBodyDTO(resource *model.Job) ResponseBodyDTO {
	return ResponseBodyDTO(resource)
}
