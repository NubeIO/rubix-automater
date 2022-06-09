package client

import (
	"fmt"
	"github.com/NubeIO/rubix-automater/automater/model"
)

type Jobs struct {
	Jobs []model.Job `json:"jobs"`
}

func (inst *Client) GetJobs() (data *Jobs, response *Response) {
	path := fmt.Sprintf(Paths.Jobs.Path)
	response = &Response{}
	resp, err := inst.Rest.R().
		SetResult(&Jobs{}).
		Get(path)
	return resp.Result().(*Jobs), response.buildResponse(resp, err)
}

func (inst *Client) DeleteJob(uuid string) (response *Response) {
	path := fmt.Sprintf("%s/%s", Paths.Jobs.Path, uuid)
	response = &Response{}
	resp, err := inst.Rest.R().
		Delete(path)
	return response.buildResponse(resp, err)
}
