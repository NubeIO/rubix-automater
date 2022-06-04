package automater

import (
	"context"
	"github.com/NubeIO/rubix-automater/automater"
	"github.com/NubeIO/rubix-automater/automater/service/jobsrv"
	"github.com/NubeIO/rubix-automater/automater/service/pipelinesrv"
	"github.com/NubeIO/rubix-automater/automater/service/resultsrv"
	"github.com/NubeIO/rubix-automater/automater/service/schedulersrv"
	"github.com/NubeIO/rubix-automater/automater/service/tasksrv"
	"github.com/NubeIO/rubix-automater/automater/service/worksrv"
	"github.com/NubeIO/rubix-automater/automater/setup"
	"github.com/NubeIO/rubix-automater/pkg/config"
	"github.com/NubeIO/rubix-automater/pkg/helpers/ttime"
	"github.com/NubeIO/rubix-automater/pkg/helpers/uuid"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/NubeIO/rubix-automater/pkg/logger"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
)

var defaultLoggingFormat = "text"

type autoMater struct {
	configPath       string
	taskService      automater.TaskService
	gracefulTermChan chan os.Signal
	logger           *logrus.Logger
}

func New(configPath string) *autoMater {
	taskService := tasksrv.New()
	gracefulTerm := make(chan os.Signal, 1)
	logger := logger.NewLogger("automater", defaultLoggingFormat)
	return &autoMater{
		configPath:       configPath,
		taskService:      taskService,
		gracefulTermChan: gracefulTerm,
		logger:           logger,
	}
}

// Run runs the service.
func (v *autoMater) Run() {
	cfg := new(config.Config)
	filePath, _ := filepath.Abs(v.configPath)
	if err := cfg.Load(filePath); err != nil {
		v.logger.Fatalf("could not load config: %s", err)
	}
	if cfg.LoggingFormat != defaultLoggingFormat {
		v.logger = logger.NewLogger("automater", cfg.LoggingFormat)
	}
	taskRepo := v.taskService.GetTaskRepository()

	for _, name := range taskRepo.GetTaskNames() {
		v.logger.Infof("registered tasks with name: %s", name)
	}
	jobQueue := setup.JobQueueFactory(cfg.JobQueue, cfg.LoggingFormat)
	v.logger.Infof("initialized [%s] as a job queue", cfg.JobQueue.Option)
	storage := setup.StorageFactory(cfg.Storage)
	v.logger.Infof("initialized [%s] as a storage", cfg.Storage.Option)
	pipelineService := pipelinesrv.New(storage, taskRepo, uuid.New(), ttime.New())
	jobService := jobsrv.New(storage, taskRepo, uuid.New(), ttime.New())
	resultService := resultsrv.New(storage)

	workPoolLogger := logger.NewLogger("workerpool", cfg.LoggingFormat)
	workService := worksrv.New(
		storage, taskRepo, ttime.New(), cfg.TimeoutUnit,
		cfg.WorkerPool.Workers, cfg.WorkerPool.QueueCapacity, workPoolLogger)
	workService.Start()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	schedulerLogger := logger.NewLogger("scheduler", cfg.LoggingFormat)
	schedulerService := schedulersrv.New(jobQueue, storage, workService, ttime.New(), schedulerLogger)
	schedulerService.Schedule(ctx, time.Duration(cfg.Scheduler.StoragePollingInterval)*cfg.TimeoutUnit)
	schedulerService.Dispatch(ctx, time.Duration(cfg.Scheduler.JobQueuePollingInterval)*cfg.TimeoutUnit)

	server := setup.ServerFactory(
		cfg.Server, jobService, pipelineService, resultService,
		v.taskService, jobQueue, storage, cfg.LoggingFormat, v.logger)
	server.Serve()
	v.logger.Infof("initialized [%s] server", cfg.Server.Protocol)

	signal.Notify(v.gracefulTermChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-v.gracefulTermChan
	v.logger.Printf("server notified %+v", sig)
	server.GracefullyStop()

	jobQueue.Close()
	workService.Stop()
	storage.Close()
}

func (v *autoMater) Stop() {
	v.gracefulTermChan <- os.Interrupt
}

// RegisterTask registers a tack callback to the tasks database under the specified name.
func (v *autoMater) RegisterTask(name string, callback func(...interface{}) (interface{}, error)) {
	v.taskService.Register(name, callback)
}

// DecodeTaskParams uses https://github.com/mitchellh/mapstructure
// to decode tasks params to a pointer of map or struct.
func DecodeTaskParams(args []interface{}, params interface{}) {
	mapstructure.Decode(args[0], params)
}

// DecodePreviousJobResults uses https://github.com/mitchellh/mapstructure
// to safely decode previous job's results metadata to a pointer of map or struct.
func DecodePreviousJobResults(args []interface{}, results interface{}) {
	if len(args) == 2 {
		mapstructure.Decode(args[1], results)
	}
}
