package client

import "fmt"

func (inst *Client) WipeDB() (response *Response) {
	path := fmt.Sprintf("%s/%s", Paths.Admin.Path, "flush")
	response = &Response{}
	resp, err := inst.Rest.R().
		Delete(path)
	return response.buildResponse(resp, err)
}
