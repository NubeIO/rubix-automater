package jobctl

import (
	"github.com/NubeIO/rubix-automater/automater"
	"github.com/NubeIO/rubix-automater/automater/model"
	"github.com/NubeIO/rubix-automater/controller"
	"github.com/NubeIO/rubix-automater/pkg/helpers/apperrors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// JobHTTPHandler is an HTTP controller that exposes job endpoints.
type JobHTTPHandler struct {
	controller.HTTPHandler
	jobService automater.JobService
	jobQueue   automater.JobQueue
}

// NewJobHTTPHandler creates and returns a new JobHTTPHandler.
func NewJobHTTPHandler(jobService automater.JobService, jobQueue automater.JobQueue) *JobHTTPHandler {
	return &JobHTTPHandler{
		jobService: jobService,
		jobQueue:   jobQueue,
	}
}

// Create creates a new job.
func (hdl *JobHTTPHandler) Create(c *gin.Context) {
	body := NewRequestBodyDTO()
	c.BindJSON(&body)

	j, err := hdl.jobService.Create(
		body.Name, body.TaskName, body.Description, body.RunAt, body.Timeout, body.Disable, body.Options, body.TaskParams)
	if err != nil {
		switch err.(type) {
		case *apperrors.ResourceValidationErr:
			hdl.HandleError(c, http.StatusBadRequest, err)
			return
		case *apperrors.ParseTimeErr:
			hdl.HandleError(c, http.StatusBadRequest, err)
			return
		default:
			hdl.HandleError(c, http.StatusInternalServerError, err)
			return
		}
	}
	c.JSON(http.StatusAccepted, BuildResponseBodyDTO(j))
}

// Get fetches a job.
func (hdl *JobHTTPHandler) Get(c *gin.Context) {
	j, err := hdl.jobService.Get(c.Param("uuid"))
	if err != nil {
		switch err.(type) {
		case *apperrors.NotFoundErr:
			hdl.HandleError(c, http.StatusNotFound, err)
			return
		default:
			hdl.HandleError(c, http.StatusInternalServerError, err)
			return
		}
	}
	c.JSON(http.StatusOK, BuildResponseBodyDTO(j))
}

// GetJobs fetches all jobs, optionally filters them by status.
func (hdl *JobHTTPHandler) GetJobs(c *gin.Context) {
	var status string
	value, ok := c.GetQuery("status")
	if ok {
		status = value
	}
	jobs, err := hdl.jobService.GetJobs(status)
	if err != nil {
		switch err.(type) {
		case *apperrors.ResourceValidationErr:
			hdl.HandleError(c, http.StatusBadRequest, err)
			return
		default:
			hdl.HandleError(c, http.StatusInternalServerError, err)
			return
		}
	}
	res := map[string]interface{}{
		"jobs": jobs,
	}
	c.JSON(http.StatusOK, res)
}

// Update updates a job.
func (hdl *JobHTTPHandler) Update(c *gin.Context) {
	body := RequestBodyDTO{}
	c.BindJSON(&body)

	j, err := hdl.jobService.Update(c.Param("uuid"), body.Name, body.Description)
	if err != nil {
		switch err.(type) {
		case *apperrors.NotFoundErr:
			hdl.HandleError(c, http.StatusNotFound, err)
			return
		default:
			hdl.HandleError(c, http.StatusInternalServerError, err)
			return
		}
	}
	c.JSON(http.StatusOK, BuildResponseBodyDTO(j))
}

// Recycle reuse a job
//	-update the run_at time to reschedule a job
func (hdl *JobHTTPHandler) Recycle(c *gin.Context) {
	body := &model.Job{}
	c.BindJSON(&body)

	j, err := hdl.jobService.Recycle(c.Param("uuid"), body)
	if err != nil {
		switch err.(type) {
		case *apperrors.NotFoundErr:
			hdl.HandleError(c, http.StatusNotFound, err)
			return
		default:
			hdl.HandleError(c, http.StatusInternalServerError, err)
			return
		}
	}
	c.JSON(http.StatusOK, BuildResponseBodyDTO(j))
}

type Delete struct {
	Message string
}

// Delete deletes a job.
func (hdl *JobHTTPHandler) Delete(c *gin.Context) {
	err := hdl.jobService.Delete(c.Param("uuid"))
	if err != nil {
		switch err.(type) {
		case *apperrors.NotFoundErr:
			hdl.HandleError(c, http.StatusNotFound, err)
			return
		case *apperrors.CannotDeletePipelineJobErr:
			hdl.HandleError(c, http.StatusConflict, err)
			return
		default:
			hdl.HandleError(c, http.StatusInternalServerError, err)
			return
		}
	}
	res := &Delete{Message: "delete ok"}
	c.JSON(http.StatusOK, res)
}

// Drop all jobs
func (hdl *JobHTTPHandler) Drop(c *gin.Context) {
	err := hdl.jobService.Drop()
	if err != nil {
		switch err.(type) {
		case *apperrors.NotFoundErr:
			hdl.HandleError(c, http.StatusNotFound, err)
			return
		case *apperrors.CannotDeletePipelineJobErr:
			hdl.HandleError(c, http.StatusConflict, err)
			return
		default:
			hdl.HandleError(c, http.StatusInternalServerError, err)
			return
		}
	}
	res := &Delete{Message: "delete ok"}
	c.JSON(http.StatusOK, res)
}
