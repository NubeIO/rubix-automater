package client

import (
	"fmt"
	"github.com/NubeIO/nubeio-rubix-lib-models-go/pkg/v1/model"
)

func (inst *Client) FlowPointWrite(uuid string, body *model.Point) (data *model.Point, response *Response) {
	path := fmt.Sprintf("%s/%s", "/api/points/write", uuid)
	response = &Response{}
	resp, err := inst.Rest.R().
		SetBody(body).
		SetResult(&model.Point{}).
		Patch(path)
	return resp.Result().(*model.Point), response.buildResponse(resp, err)
}
