package apptask

import (
	"errors"
	automater "github.com/NubeIO/rubix-automater"
	"github.com/NubeIO/rubix-automater/service/assitcli"
)

type AppTask struct {
	LocationName    string   `json:"locationName"`
	NetworkName     string   `json:"networkName"`
	HostName        string   `json:"hostName"`
	HostUUID        string   `json:"hostUUID"`
	AppName         string   `json:"appName"`
	SubTask         string   `json:"subTask"`
	Version         string   `json:"version"`
	OrderedTaskList []string `json:"orderedTaskList"`
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
		LocationName: body.LocationName,
		NetworkName:  body.NetworkName,
		HostName:     body.HostName,
		HostUUID:     body.HostUUID,
		AppName:      body.AppName,
		SubTask:      body.SubTask,
		Version:      body.Version,
	}
	install, res := cli.AppTask(b)
	if res.StatusCode == 0 {
		return install, errors.New("failed to find host")
	}
	if res.StatusCode > 299 {
		return install, errors.New(install.ErrorMessage)
	}
	return install, nil
}
