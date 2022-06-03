package resultctl

import (
	"github.com/NubeIO/rubix-automater/automater/core"
)

// ResponseBodyDTO is the response data transfer object used for a job result retrieval.
type ResponseBodyDTO *core.JobResult

// BuildResponseBodyDTO creates a new ResponseDTO.
func BuildResponseBodyDTO(resource *core.JobResult) ResponseBodyDTO {
	return resource
}
