package apps

import (
	"errors"
	"github.com/NubeIO/rubix-assist/service/client"
	"github.com/NubeIO/rubix-assist/service/em"
	automater "github.com/NubeIO/rubix-automater"
	"github.com/NubeIO/rubix-cli-app/service/apps/installer"
)

type InstallAppParams struct {
	HostUUID string `json:"hostUUID"`
	HostName string `json:"hostName,omitempty"`
	AppName  string `json:"appName"`
	Version  string `json:"version"`
}

func InstallApp(args ...interface{}) (interface{}, error) {
	params := &InstallAppParams{}
	resultsMetadata := &installer.InstallResponse{}
	automater.DecodeTaskParams(args, params)
	automater.DecodePreviousJobResults(args, &resultsMetadata)
	return runAppInstall(params)
}

func runAppInstall(body *InstallAppParams) (interface{}, error) {
	cli := client.New("0.0.0.0", 1662)
	app := &em.App{
		HostUUID: body.HostUUID,
		HostName: body.HostName,
		AppName:  body.AppName,
		Version:  body.Version,
	}
	install, res := cli.InstallApp(app)
	if res.GetStatus() == 0 {
		return install, errors.New("failed to find host")
	}
	if res.GetStatus() > 299 {
		return nil, errors.New(res.GetMessage().(string))
	}
	return install, nil
}
