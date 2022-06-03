package client

import (
	"fmt"
	"github.com/go-resty/resty/v2"
)

type Client struct {
	Rest *resty.Client
}

// New returns a new instance of the nube common apis
func New(url string, port int) *Client {
	rest := &Client{
		Rest: resty.New(),
	}
	rest.Rest.SetBaseURL(fmt.Sprintf("http://%s:%d", url, port))
	return rest
}
