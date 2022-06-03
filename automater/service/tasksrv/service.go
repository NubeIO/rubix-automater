package tasksrv

import (
	taskRepo "github.com/NubeIO/rubix-automater/automater/service/tasksrv/taskrepo"
)

type taskService struct {
	taskRepo *taskRepo.TaskRepository
}

// New creates a new job service.
func New() *taskService {
	return &taskService{
		taskRepo: taskRepo.New(),
	}
}

// Register registers a new tasks in the tasks database.
func (srv *taskService) Register(name string, taskFunc taskRepo.TaskFunc) {
	srv.taskRepo.Register(name, taskFunc)
}

// GetTaskRepository returns the tasks database.
func (srv *taskService) GetTaskRepository() *taskRepo.TaskRepository {
	return srv.taskRepo
}
