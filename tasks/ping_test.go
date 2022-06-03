package tasks

import (
	"fmt"
	"testing"
)

func Test_ping(t *testing.T) {
	found := ping("nube-io.com", 443)
	fmt.Println(found)

}
