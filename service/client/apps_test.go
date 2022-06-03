package client

import (
	"fmt"
	"testing"
)

func TestHost(*testing.T) {

	client := New("0.0.0.0", 8089)

	data, _ := client.GetJobs()

	if data.Jobs == nil {

	}
	for i, job := range data.Jobs {
		fmt.Println(i, job)
	}

}
