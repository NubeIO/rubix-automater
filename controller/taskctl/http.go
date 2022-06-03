package taskctl

import (
	"github.com/NubeIO/rubix-automater/automater"
	"net/http"

	"github.com/gin-gonic/gin"
)

// TaskHTTPHandler is an HTTP controller that exposes tasks endpoints.
type TaskHTTPHandler struct {
	taskService automater.TaskService
}

// NewTaskHTTPHandler creates and returns a new JobHTTPHandler.
func NewTaskHTTPHandler(taskService automater.TaskService) *TaskHTTPHandler {
	return &TaskHTTPHandler{
		taskService: taskService,
	}
}

// GetTasks returns all registered tasks.
func (hdl *TaskHTTPHandler) GetTasks(c *gin.Context) {
	taskRepo := hdl.taskService.GetTaskRepository()
	tasks := taskRepo.GetTaskNames()
	res := map[string]interface{}{
		"tasks": tasks,
	}
	c.JSON(http.StatusOK, res)
}
