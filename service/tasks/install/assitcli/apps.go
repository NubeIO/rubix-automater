package assitcli

import (
	"fmt"
	"github.com/NubeIO/edge/service/apps/installer"
)

type App struct {
	LocationName string `json:"locationName"`
	NetworkName  string `json:"networkName"`
	HostName     string `json:"hostName"`
	HostUUID     string `json:"hostUUID"`
	AppName      string `json:"appName"`
	Version      string `json:"version"`
}

func (inst *Client) InstallApp(body *App) (data *installer.InstallResponse, response *Response) {
	path := fmt.Sprintf("%s/%s", Paths.Apps.Path, "install")
	response = &Response{}
	resp, err := inst.Rest.R().
		SetBody(body).
		SetResult(&installer.InstallResponse{}).
		SetError(&Response{}).
		Post(path)

	fmt.Println(22222)
	fmt.Println(resp.StatusCode())
	fmt.Println(resp.String())
	response = response.buildResponse(resp, err)
	if resp.IsSuccess() {
		data = resp.Result().(*installer.InstallResponse)
		response.Message = resp.Result().(*installer.InstallResponse)
	}
	return data, response
}
