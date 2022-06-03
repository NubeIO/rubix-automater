package taskrepo

import (
	"fmt"
)

// TaskFunc is the type of the tasks callback.
type TaskFunc func(...interface{}) (interface{}, error)

// TaskRepository is the in memory tasks database.
type TaskRepository map[string]TaskFunc

// New initializes and returns a new TaskRepository instance.
func New() *TaskRepository {
	repo := make(map[string]TaskFunc)
	taskRepo := TaskRepository(repo)
	return &taskRepo
}

// GetTaskFunc returns the TaskFunc for a specified name if that exists in the tasks database.
func (repo TaskRepository) GetTaskFunc(name string) (TaskFunc, error) {
	task, ok := repo[name]
	if !ok {
		return nil, fmt.Errorf("tasks with name: %s is not registered", name)
	}
	return task, nil
}

// GetTaskNames returns all the names of the tasks currently in the tasks database.
func (repo TaskRepository) GetTaskNames() []string {
	var names []string
	for name := range repo {
		names = append(names, name)
	}
	return names
}

// Register adds a new tasks in the database.
func (repo TaskRepository) Register(name string, taskFunc TaskFunc) {
	repo[name] = taskFunc
}
