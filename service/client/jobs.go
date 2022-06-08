package client

import (
	"encoding/json"
	"fmt"
	"github.com/NubeIO/rubix-automater/automater/model"
	"github.com/NubeIO/rubix-automater/controller/jobctl"
	"github.com/go-resty/resty/v2"
)

func (response *Response) buildResponse(restyResp *resty.Response, err error) *Response {
	response.resty = restyResp
	var msg interface{}
	if err != nil {
		msg = err.Error()
	} else {
		asJson, err := response.AsJson()
		if err != nil {
			msg = restyResp.String()
		} else {
			msg = asJson
		}
	}
	response.StatusCode = restyResp.StatusCode()
	response.Message = msg
	return response
}

func (response *Response) AsString() string {
	return response.resty.String()
}

func (response *Response) GetError() interface{} {
	return response.resty.Error()
}

func (response *Response) GetStatus() int {
	return response.resty.StatusCode()
}

// AsJson return as body as blank interface
func (response *Response) AsJson() (interface{}, error) {
	var out interface{}
	err := json.Unmarshal(response.resty.Body(), &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (inst *Client) GetJob(uuid string) (data *model.Job, response *Response) {
	path := fmt.Sprintf("%s/%s", Paths.Jobs.Path, uuid)
	response = &Response{}
	resp, err := inst.Rest.R().
		SetResult(&model.Job{}).
		Get(path)
	return resp.Result().(*model.Job), response.buildResponse(resp, err)
}

func (inst *Client) AddJob(body *jobctl.JobBody) (data *model.Job, response *Response) {
	path := fmt.Sprintf(Paths.Jobs.Path)
	response = &Response{}
	resp, err := inst.Rest.R().
		SetBody(body).
		SetResult(&model.Job{}).
		Post(path)
	return resp.Result().(*model.Job), response.buildResponse(resp, err)
}

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
