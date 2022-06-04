package tasks

import (
	"errors"
	automater "github.com/NubeIO/rubix-automater"
	dbase "github.com/NubeIO/rubix-cli-app/database"
	"github.com/NubeIO/rubix-cli-app/service/client"
	"time"
)

type InstallAppParams struct {
	URL     string `json:"url,omitempty"`
	Port    int    `json:"port"`
	AppName string `json:"AppName"`
	Version string `json:"version"`
	Token   string `json:"token,omitempty"`
}

func InstallApp(args ...interface{}) (interface{}, error) {
	params := &InstallAppParams{}
	resultsMetadata := &PingResponse{}
	automater.DecodeTaskParams(args, params)
	automater.DecodePreviousJobResults(args, &resultsMetadata)
	time.Sleep(1 * time.Second)
	return runAppInstall(params)
}

func runAppInstall(app *InstallAppParams) (interface{}, error) {
	cli := client.New(app.URL, app.Port)
	install, res := cli.InstallApp(&dbase.App{AppName: app.AppName, Version: app.Version, Token: app.Token})
	if res.StatusCode == 0 {
		return install, errors.New("failed to find host")
	}
	return install, nil
}
