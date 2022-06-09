package install

import (
	"errors"
	"github.com/NubeIO/rubix-automater/service/tasks/install/assitcli"

	automater "github.com/NubeIO/rubix-automater"
	"github.com/NubeIO/rubix-cli-app/service/apps/installer"
)

type AppParams struct {
	LocationName string `json:"locationName"`
	NetworkName  string `json:"networkName"`
	HostName     string `json:"hostName"`
	HostUUID     string `json:"hostUUID"`
	AppName      string `json:"appName"`
	Version      string `json:"version"`
}

func App(args ...interface{}) (interface{}, error) {
	params := &AppParams{}
	resultsMetadata := &installer.InstallResponse{}
	automater.DecodeTaskParams(args, params)
	automater.DecodePreviousJobResults(args, &resultsMetadata)
	return runAppInstall(params)
}

func runAppInstall(body *AppParams) (interface{}, error) {
	cli := assitcli.New("0.0.0.0", 1662)
	app := &assitcli.App{
		LocationName: body.LocationName,
		NetworkName:  body.NetworkName,
		HostName:     body.HostName,
		HostUUID:     body.HostUUID,
		AppName:      body.AppName,
		Version:      body.Version,
	}
	install, res := cli.InstallApp(app)
	if res.StatusCode == 0 {
		return install, errors.New("failed to find host")
	}
	if res.StatusCode > 299 {
		return nil, errors.New(install.Error)
	}
	return install, nil
}
