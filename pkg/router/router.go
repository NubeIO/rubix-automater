package router

import (
	"bytes"
	"encoding/json"
	"github.com/NubeIO/rubix-automater/automater"
	"github.com/NubeIO/rubix-automater/controller/jobctl"
	"github.com/NubeIO/rubix-automater/controller/pipectl"
	"github.com/NubeIO/rubix-automater/controller/resultctl"
	"github.com/NubeIO/rubix-automater/controller/taskctl"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// NewRouter initializes and returns a new gin.Engine instance.
func NewRouter(
	jobService automater.JobService,
	resultService automater.ResultService,
	pipelineService automater.PipelineService,
	taskService automater.TaskService,
	jobQueue automater.JobQueue,
	storage automater.Storage, loggingFormat string) *gin.Engine {

	jobHandler := jobctl.NewJobHTTPHandler(jobService, jobQueue)
	resultHandler := resultctl.NewResultHTTPHandler(resultService)
	pipelineHandler := pipectl.NewPipelineHTTPHandler(pipelineService, jobQueue)
	taskHandler := taskctl.NewTaskHTTPHandler(taskService)

	r := gin.New()
	if loggingFormat == "text" {
		r.Use(gin.Logger())
	} else {
		r.Use(JSONLogMiddleware())
	}
	r.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err})
		}
		c.AbortWithStatus(http.StatusInternalServerError)
	}))
	// CORS: Allow all origins - Revisit this.
	r.Use(cors.Default())

	r.GET("/api/status", HandleStatus(jobQueue, storage))

	r.POST("/api/jobs", jobHandler.Create)
	r.GET("/api/jobs", jobHandler.GetJobs)
	r.GET("/api/jobs/:uuid", jobHandler.Get)
	r.PATCH("/api/jobs/:uuid", jobHandler.Update)
	r.PATCH("/api/job/:uuid", jobHandler.UpdateAll)
	r.DELETE("/api/jobs/:uuid", jobHandler.Delete)
	r.DELETE("/api/jobs/drop", jobHandler.Drop)

	r.GET("/api/jobs/:uuid/results", resultHandler.Get)
	r.DELETE("/api/jobs/:uuid/results", resultHandler.Delete)

	r.POST("/api/pipelines", pipelineHandler.Create)
	r.GET("/api/pipelines", pipelineHandler.GetPipelines)
	r.GET("/api/pipelines/:uuid", pipelineHandler.Get)
	r.PATCH("/api/pipelines/:uuid", pipelineHandler.Update)
	r.DELETE("/api/pipelines/:uuid", pipelineHandler.Delete)

	r.GET("/api/pipelines/:uuid/jobs", pipelineHandler.GetPipelineJobs)

	r.GET("/api/tasks", taskHandler.GetTasks)
	return r
}

// HandleStatus is an endpoint providing information and the status of the server,
func HandleStatus(jobQueue automater.JobQueue, storage automater.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		now := time.Now().UTC()
		res := map[string]interface{}{
			"job_queue_healthy": jobQueue.CheckHealth(),
			"storage_healthy":   storage.CheckHealth(),
			"intime":            now,
		}
		c.JSON(http.StatusOK, res)
	}
}

func JSONLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		start := time.Now()
		// Process Request
		c.Next()
		// Stop timer
		elapsed := time.Since(start)

		entry := logrus.WithFields(logrus.Fields{
			"method":         c.Request.Method,
			"path":           c.Request.RequestURI,
			"status":         c.Writer.Status(),
			"referrer":       c.Request.Referer(),
			"request_id":     c.Writer.Header().Get("Request-Id"),
			"remote_address": c.Request.RemoteAddr,
			"elapsed":        elapsed.String(),
		})

		var message string
		status := c.Writer.Status()
		if status >= 400 {
			errResponse := struct {
				Code    int    `json:"code"`
				Error   bool   `json:"error"`
				Message string `json:"message"`
			}{}

			json.Unmarshal(blw.body.Bytes(), &errResponse)
			message = errResponse.Message
		}

		entry.Logger.Formatter = &logrus.JSONFormatter{}
		if status >= 500 {
			entry.Error(c.Errors.String())
		} else {
			entry.WithTime(start).Info(message)
		}
	}
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}
