package assitcli

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

func (inst *Client) AppTask(body *AppTask) (*TaskResponse, *Response) {
	response := &Response{}
	resp, _ := inst.Rest.R().
		SetBody(body).
		SetResult(&TaskResponse{}).
		SetError(&TaskResponse{}).
		Post(Paths.PipelineTask.Path)
	response.StatusCode = resp.StatusCode()
	if resp.IsSuccess() {
		data := resp.Result().(*TaskResponse)
		response.Message = resp.Result().(*TaskResponse)
		return data, response
	}
	return resp.Error().(*TaskResponse), response
}

//func (inst *Client) InstallApp(body *App) (data *installer.InstallResponse, response *Response) {
//	path := fmt.Sprintf("%s/%s", Paths.Apps.Path, "install")
//	response = &Response{}
//	resp, err := inst.Rest.R().
//		SetBody(body).
//		SetResult(&installer.InstallResponse{}).
//		SetError(&Response{}).
//		Post(path)
//	response = response.buildResponse(resp, err)
//	if resp.IsSuccess() {
//		data = resp.Result().(*installer.InstallResponse)
//		response.Message = resp.Result().(*installer.InstallResponse)
//	}
//	return data, response
//}
