package client

import (
	"encoding/json"
	"fmt"
	"github.com/NubeIO/rubix-automater/automater/core"
	"github.com/go-resty/resty/v2"
)

type Path struct {
	Path string
}

var Paths = struct {
	Jobs   Path
	Store  Path
	System Path
}{
	Jobs:   Path{Path: "/api/jobs"},
	Store:  Path{Path: "/api/store"},
	System: Path{Path: "/api/system"},
}

type Response struct {
	StatusCode int         `json:"status_code"`
	Message    interface{} `json:"message"`
	resty      *resty.Response
}

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

func (inst *Client) GetHost(uuid string) (data *core.Job, response *Response) {
	path := fmt.Sprintf("%s/%s", Paths.Jobs.Path, uuid)
	response = &Response{}
	resp, err := inst.Rest.R().
		SetResult(&core.Job{}).
		Get(path)
	return resp.Result().(*core.Job), response.buildResponse(resp, err)
}

func (inst *Client) AddHost(body *core.Job) (data *core.Job, response *Response) {
	path := fmt.Sprintf(Paths.Jobs.Path)
	response = &Response{}
	resp, err := inst.Rest.R().
		SetBody(body).
		SetResult(&core.Job{}).
		Post(path)
	return resp.Result().(*core.Job), response.buildResponse(resp, err)
}

type Jobs struct {
	Jobs []core.Job `json:"jobs"`
}

func (inst *Client) GetHosts() (data *Jobs, response *Response) {
	path := fmt.Sprintf(Paths.Jobs.Path)
	response = &Response{}
	resp, err := inst.Rest.R().
		SetResult(&Jobs{}).
		Get(path)
	return resp.Result().(*Jobs), response.buildResponse(resp, err)
}

func (inst *Client) DeleteHost(uuid string) (response *Response) {
	path := fmt.Sprintf("%s/%s", Paths.Jobs.Path, uuid)
	response = &Response{}
	resp, err := inst.Rest.R().
		Delete(path)
	return response.buildResponse(resp, err)
}
