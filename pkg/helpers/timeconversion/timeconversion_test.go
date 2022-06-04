package timeconversion

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestAdjustTime(*testing.T) {
	timeFormat := "02 Jan 06 15:04 AST"
	parse, err := time.Parse(time.RFC822, timeFormat)
	fmt.Println(parse, err)

	adjustedTime, err := AdjustTime(parse, "1 month")
	if err != nil {
		fmt.Printf("error adjusting time: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(adjustedTime)
	fmt.Printf("Adjusted is %s\n", adjustedTime.Format(timeFormat))

	timeFormat = "2022-06-04T02:23:00.426752075+10:00"
	parse, err = time.Parse(time.RFC3339Nano, timeFormat)
	adjustedTime, err = AdjustTime(parse, "1 month")
	if err != nil {
		fmt.Printf("error adjusting time: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Adjusted", adjustedTime)
}
