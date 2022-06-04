package transactionctl

import (
	"github.com/NubeIO/rubix-automater/automater/model"
)

// ResponseBodyDTO is the response data transfer object used for a job result retrieval.
type ResponseBodyDTO *model.Transaction

// BuildResponseBodyDTO creates a new ResponseDTO.
func BuildResponseBodyDTO(resource *model.Transaction) ResponseBodyDTO {
	return resource
}
