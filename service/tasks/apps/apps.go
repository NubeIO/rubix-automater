package apptask

import (
	"errors"
	automater "github.com/NubeIO/rubix-automater"
	pprint "github.com/NubeIO/rubix-automater/pkg/helpers/print"
	"github.com/NubeIO/rubix-automater/service/assitcli"
	log "github.com/sirupsen/logrus"
)

type AppTask struct {
	Description        string   `json:"description"`
	LocationName       string   `json:"locationName"`
	NetworkName        string   `json:"networkName"`
	HostName           string   `json:"hostName"`
	HostUUID           string   `json:"hostUUID"`
	AppName            string   `json:"appName"`
	SubTask            string   `json:"subTask"`
	Version            string   `json:"version"`
	ManualInstall      bool     `json:"manualInstall"`      // will not download from GitHub, and will use the app-store download path
	ManualAssetZipName string   `json:"manualAssetZipName"` // flow-framework-0.5.5-1575cf89.amd64.zip
	ManualAssetTag     string   `json:"manualAssetTag"`     // this is the release tag as in v0.0.1
	Cleanup            bool     `json:"cleanup"`
	OrderedTaskList    []string `json:"orderedTaskList"`
}

type TaskResponse struct {
	Message      interface{} `json:"message"`
	ErrorMessage string      `json:"errorMessage"`
	Error        interface{} `json:"error"`
}

func App(args ...interface{}) (interface{}, error) {
	params := &AppTask{}
	resultsMetadata := &TaskResponse{}
	automater.DecodeTaskParams(args, params)
	automater.DecodePreviousJobResults(args, &resultsMetadata)
	return runApp(params)
}

func runApp(body *AppTask) (interface{}, error) {
	cli := assitcli.New("0.0.0.0", 1662)
	b := &assitcli.AppTask{
		Description:        body.Description,
		LocationName:       body.LocationName,
		NetworkName:        body.NetworkName,
		HostName:           body.HostName,
		HostUUID:           body.HostUUID,
		AppName:            body.AppName,
		SubTask:            body.SubTask,
		Version:            body.Version,
		ManualInstall:      body.ManualInstall,
		ManualAssetZipName: body.ManualAssetZipName,
		ManualAssetTag:     body.ManualAssetTag,
		Cleanup:            body.Cleanup,
	}
	pprint.PrintJOSN(b)
	install, res := cli.AppTask(b)
	log.Infof("install app:%d", res.StatusCode)
	if res.StatusCode == 0 {
		log.Errorf("install app:%s", res.Message)
		return install, errors.New("failed to find host")
	}
	if res.StatusCode > 299 {
		log.Errorf("install app:%s", res.Message)
		return install, errors.New(install.ErrorMessage)
	}
	pprint.PrintJOSN(install)
	return install, nil
}
