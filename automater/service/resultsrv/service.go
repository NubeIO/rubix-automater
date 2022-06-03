package resultsrv

import (
	"github.com/NubeIO/rubix-automater/automater"
	"github.com/NubeIO/rubix-automater/automater/core"
)

var _ automater.ResultService = &resultService{}

type resultService struct {
	storage automater.Storage
}

// New creates a new job result service.
func New(storage automater.Storage) *resultService {
	return &resultService{
		storage: storage,
	}
}

// Get fetches a job result.
func (srv *resultService) Get(uuid string) (*core.JobResult, error) {
	return srv.storage.GetJobResult(uuid)
}

// Delete deletes a job result.
func (srv *resultService) Delete(uuid string) error {
	_, err := srv.storage.GetJobResult(uuid)
	if err != nil {
		return err
	}
	return srv.storage.DeleteJobResult(uuid)
}
