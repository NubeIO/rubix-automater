package client

import (
	"fmt"
	"github.com/NubeIO/rubix-automater/automater/model"
	"github.com/NubeIO/rubix-automater/controller/pipectl"
)

func (inst *Client) AddPipeline(body *pipectl.PipelineBody) (data *model.Pipeline, response *Response) {
	path := fmt.Sprintf(Paths.Pipeline.Path)
	response = &Response{}
	resp, err := inst.Rest.R().
		SetBody(body).
		SetResult(&model.Pipeline{}).
		Post(path)
	return resp.Result().(*model.Pipeline), response.buildResponse(resp, err)
}
