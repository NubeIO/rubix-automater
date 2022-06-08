package pipectl

import (
	"github.com/NubeIO/rubix-automater/automater/model"
	"github.com/NubeIO/rubix-automater/controller/jobctl"
)

// PipelineBody is the data transfer object used for a job creation or update.
type PipelineBody struct {
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	ScheduleAt      string                 `json:"schedule_at"`
	PipelineOptions *model.PipelineOptions `json:"options"`
	Jobs            []*jobctl.JobBody      `json:"jobs"`
}

// NewRequestBodyDTO initializes and returns a new BodyDTO instance.
func NewRequestBodyDTO() *PipelineBody {
	return &PipelineBody{}
}

// ResponseBodyDTO is the response data transfer object used for a pipeline creation or update.
type ResponseBodyDTO *model.Pipeline

// BuildResponseBodyDTO creates a new ResponseDTO.
func BuildResponseBodyDTO(resource *model.Pipeline) ResponseBodyDTO {
	return resource
}
