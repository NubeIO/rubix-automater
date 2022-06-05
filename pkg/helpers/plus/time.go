package plus

import (
	"fmt"
	"github.com/jinzhu/now"
	"time"
)

func calc() {
	a := now.New(time.Now())

	fmt.Println(a.EndOfMonth())

}
