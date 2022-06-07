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

type Path struct {
	Path string
}

var Paths = struct {
	Pipeline Path
	Jobs     Path
	Store    Path
	System   Path
	Admin    Path
}{
	Pipeline: Path{Path: "/api/pipelines"},
	Jobs:     Path{Path: "/api/jobs"},
	Store:    Path{Path: "/api/store"},
	System:   Path{Path: "/api/system"},
	Admin:    Path{Path: "/api/admin"},
}

type Response struct {
	StatusCode int         `json:"status_code"`
	Message    interface{} `json:"message"`
	resty      *resty.Response
}
