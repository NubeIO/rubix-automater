package pipectl

import (
	"github.com/NubeIO/rubix-automater/automater/model"
	"github.com/NubeIO/rubix-automater/controller/jobctl"
)

// RequestBodyDTO is the data transfer object used for a job creation or update.
type RequestBodyDTO struct {
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	RunAt       string                   `json:"run_at"`
	Jobs        []*jobctl.RequestBodyDTO `json:"jobs"`
}

// NewRequestBodyDTO initializes and returns a new BodyDTO instance.
func NewRequestBodyDTO() *RequestBodyDTO {
	return &RequestBodyDTO{}
}

// ResponseBodyDTO is the response data transfer object used for a pipeline creation or update.
type ResponseBodyDTO *model.Pipeline

// BuildResponseBodyDTO creates a new ResponseDTO.
func BuildResponseBodyDTO(resource *model.Pipeline) ResponseBodyDTO {
	return resource
}
