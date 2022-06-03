package jobctl

import (
	"github.com/NubeIO/rubix-automater/automater/core"
)

// RequestBodyDTO is the data transfer object used for a job creation or update.
type RequestBodyDTO struct {
	Name               string                 `json:"name"`
	Description        string                 `json:"description"`
	TaskName           string                 `json:"task_name"`
	Timeout            int                    `json:"timeout"`
	TaskParams         map[string]interface{} `json:"task_params"`
	RunAt              string                 `json:"run_at"`
	UsePreviousResults bool                   `json:"use_previous_results"`
}

// NewRequestBodyDTO initializes and returns a new BodyDTO instance.
func NewRequestBodyDTO() *RequestBodyDTO {
	return &RequestBodyDTO{}
}

// ResponseBodyDTO is the response data transfer object used for a job creation or update.
type ResponseBodyDTO *core.Job

// BuildResponseBodyDTO creates a new ResponseDTO.
func BuildResponseBodyDTO(resource *core.Job) ResponseBodyDTO {
	return ResponseBodyDTO(resource)
}
