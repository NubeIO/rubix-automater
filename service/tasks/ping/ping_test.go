package ping

import (
	"fmt"
	"testing"
)

func Test_ping(t *testing.T) {
	found := ping("0.0.0.0", 8090)
	fmt.Println(found)

}
