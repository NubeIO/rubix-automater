package apptask

import (
	automater "github.com/NubeIO/rubix-automater"
)

type AppTask struct {
	LocationName string `json:"locationName"`
	NetworkName  string `json:"networkName"`
	HostName     string `json:"hostName"`
	HostUUID     string `json:"hostUUID"`
	AppName      string `json:"appName"`
	SubTask      string `json:"subTask"`
	Version      string `json:"version"`
}

func App(args ...interface{}) (interface{}, error) {
	params := &AppTask{}
	//resultsMetadata := &installer.InstallResponse{}
	automater.DecodeTaskParams(args, params)
	//automater.DecodePreviousJobResults(args, &resultsMetadata)
	return runApp(params)
}

func runApp(body *AppTask) (interface{}, error) {
	//cli := assitcli.New("0.0.0.0", 1662)
	//app := &assitcli.App{
	//	LocationName: body.LocationName,
	//	NetworkName:  body.NetworkName,
	//	HostName:     body.HostName,
	//	HostUUID:     body.HostUUID,
	//	AppName:      body.AppName,
	//	Version:      body.Version,
	//}
	//install, res := cli.InstallApp(app)
	//if res.StatusCode == 0 {
	//	return install, errors.New("failed to find host")
	//}
	//if res.StatusCode > 299 {
	//	return nil, errors.New(install.Error)
	//}
	return nil, nil
}
