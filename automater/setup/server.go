package setup

import (
	"github.com/NubeIO/rubix-automater/automater"
	"github.com/NubeIO/rubix-automater/pkg/config"
	"github.com/NubeIO/rubix-automater/pkg/router"
	"github.com/NubeIO/rubix-automater/pkg/server"
	"net/http"

	"github.com/sirupsen/logrus"
)

const (
	HTTP = "http"
)

func ServerFactory(
	cfg config.Server,
	jobService automater.JobService,
	pipelineService automater.PipelineService,
	resultService automater.ResultService,
	taskService automater.TaskService,
	jobQueue automater.JobQueue,
	storage automater.Storage, loggingFormat string,
	logger *logrus.Logger) automater.Server {

	if cfg.Protocol == HTTP {
		srv := http.Server{
			Addr: ":" + cfg.HTTP.Port,
			Handler: router.NewRouter(
				jobService, resultService,
				pipelineService, taskService,
				jobQueue, storage, loggingFormat),
		}
		httpsrv := server.NewHTTPServer(srv, logger)
		return httpsrv
	}
	return nil
}
