package pipectl

import (
	"github.com/NubeIO/rubix-automater/automater"
	"github.com/NubeIO/rubix-automater/automater/model"
	"github.com/NubeIO/rubix-automater/controller"
	"github.com/NubeIO/rubix-automater/pkg/helpers/apperrors"
	"github.com/NubeIO/rubix-automater/pkg/helpers/timeconversion"
	"github.com/NubeIO/rubix-automater/pkg/helpers/ttime"
	"net/http"

	"github.com/gin-gonic/gin"
)

// PipelineHTTPHandler is an HTTP controller that exposes pipeline endpoints.
type PipelineHTTPHandler struct {
	controller.HTTPHandler
	pipelineService automater.PipelineService
	jobQueue        automater.JobQueue
}

// NewPipelineHTTPHandler creates and returns a new PipelineHTTPHandler.
func NewPipelineHTTPHandler(
	pipelineService automater.PipelineService,
	jobQueue automater.JobQueue) *PipelineHTTPHandler {

	return &PipelineHTTPHandler{
		pipelineService: pipelineService,
		jobQueue:        jobQueue,
	}
}

func (hdl *PipelineHTTPHandler) pushJobs(c *gin.Context, p *model.Pipeline) error {
	if !p.IsScheduled() {
		// Push it as one job into the queue.
		p.MergeJobsInOne()
		// Push only the first job of the pipeline.
		if err := hdl.jobQueue.Push(p.Jobs[0]); err != nil {
			switch err.(type) {
			case *apperrors.FullQueueErr:
				hdl.HandleError(c, http.StatusServiceUnavailable, err)
				return err
			default:
				hdl.HandleError(c, http.StatusInternalServerError, err)
				return err
			}
		}
		// Do not include next job in the response body.
		p.UnmergeJobs()
	}
	return nil

}

// Create creates a new pipeline and all of its jobs.
func (hdl *PipelineHTTPHandler) Create(c *gin.Context) {
	body := NewRequestBodyDTO()
	c.BindJSON(&body)

	jobs := make([]*model.Job, 0)
	for _, jobDTO := range body.Jobs {
		j := &model.Job{
			Name:               jobDTO.Name,
			Description:        jobDTO.Description,
			TaskName:           jobDTO.TaskName,
			Timeout:            jobDTO.Timeout,
			TaskParams:         jobDTO.TaskParams,
			UsePreviousResults: jobDTO.UsePreviousResults,
		}
		jobs = append(jobs, j)
	}

	p, err := hdl.pipelineService.Create(body.Name, body.Description, body.RunAt, jobs)
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
	err = hdl.pushJobs(c, p)
	if err != nil {
		return
	}

	c.JSON(http.StatusAccepted, BuildResponseBodyDTO(p))
}

// RecyclePipeline updates a pipeline.
func (hdl *PipelineHTTPHandler) RecyclePipeline(c *gin.Context) {
	//get the existing
	uuid := c.Param("uuid")
	getExisting, err := hdl.pipelineService.Get(uuid)
	if err != nil {
		hdl.HandleError(c, http.StatusInternalServerError, err)
		return
	}

	jobs, err := hdl.pipelineService.GetPipelineJobs(uuid)
	if err != nil {
		return
	}
	now := ttime.New().Now()
	nextRunTime, _ := timeconversion.AdjustTime(now, "30 sec")
	var recycleJobs []*model.Job
	for _, job := range jobs {
		recycleJob, err := hdl.pipelineService.RecycleJob(job.UUID, job)
		if err != nil {
			hdl.HandleError(c, http.StatusInternalServerError, err)
			return
		}
		recycleJob.RunAt = &nextRunTime
		recycleJobs = append(recycleJobs, recycleJob)
	}

	getExisting.Jobs = recycleJobs
	getExisting.Status = model.Pending
	getExisting.RunAt = &nextRunTime
	getExisting.StartedAt = nil
	getExisting.Duration = nil

	resp, err := hdl.pipelineService.RecyclePipeline(c.Param("uuid"), getExisting)
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
	err = hdl.pushJobs(c, resp)
	if err != nil {
		return
	}

	c.JSON(http.StatusOK, BuildResponseBodyDTO(resp))
}

// Get fetches a pipeline.
func (hdl *PipelineHTTPHandler) Get(c *gin.Context) {
	j, err := hdl.pipelineService.Get(c.Param("uuid"))
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

// GetPipelines fetches all pipelines, optionally filters them by status.
func (hdl *PipelineHTTPHandler) GetPipelines(c *gin.Context) {
	var status string
	value, ok := c.GetQuery("status")
	if ok {
		status = value
	}
	pipelines, err := hdl.pipelineService.GetPipelines(status)
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
		"pipelines": pipelines,
	}
	c.JSON(http.StatusOK, res)
}

// GetPipelineJobs fetches the jobs of a specified pipeline.
func (hdl *PipelineHTTPHandler) GetPipelineJobs(c *gin.Context) {
	jobs, err := hdl.pipelineService.GetPipelineJobs(c.Param("uuid"))
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
	res := map[string]interface{}{
		"jobs": jobs,
	}
	c.JSON(http.StatusOK, res)
}

// Update updates a pipeline.
func (hdl *PipelineHTTPHandler) Update(c *gin.Context) {
	body := RequestBodyDTO{}
	c.BindJSON(&body)

	err := hdl.pipelineService.Update(c.Param("uuid"), body.Name, body.Description)
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
	c.Writer.WriteHeader(http.StatusNoContent)
}

// Delete deletes a pipelines and all its jobs.
func (hdl *PipelineHTTPHandler) Delete(c *gin.Context) {
	err := hdl.pipelineService.Delete(c.Param("uuid"))
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
	c.Writer.WriteHeader(http.StatusNoContent)
}
