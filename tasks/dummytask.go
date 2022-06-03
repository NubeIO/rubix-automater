package tasks

import (
	"fmt"
	"github.com/NubeIO/rubix-automater"
	"time"
)

// DummyParams is an example of a tasks params structure.
type DummyParams struct {
	URL string `json:"url,omitempty"`
}

// DummyTask is a dummy tasks callback.
func DummyTask(args ...interface{}) (interface{}, error) {
	dummyParams := &DummyParams{}
	var resultsMetadata string
	automater.DecodeTaskParams(args, dummyParams)
	automater.DecodePreviousJobResults(args, &resultsMetadata)
	fmt.Println("DummyTask")
	time.Sleep(1 * time.Minute)
	metadata := downloadContent(dummyParams.URL, resultsMetadata)
	fmt.Println("DummyTask after")
	return metadata, nil
}

func downloadContent(URL, bucket string) string {
	fmt.Println("downloadContent func test", URL, bucket)
	return "some metadata"
}
