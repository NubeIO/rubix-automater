package flow

import (
	"errors"
	"fmt"
	"github.com/NubeIO/nubeio-rubix-lib-helpers-go/pkg/nils"
	"github.com/NubeIO/nubeio-rubix-lib-models-go/pkg/v1/model"
	automater "github.com/NubeIO/rubix-automater"
	"github.com/NubeIO/rubix-automater/service/client"
	"math/rand"
	"time"
)

type FlowAppParams struct {
	URL   string   `json:"url,omitempty"`
	Port  int      `json:"port"`
	UUID  string   `json:"uuid"`
	Delay int      `json:"delay"`
	Value *float64 `json:"value"`
}

func PointWrite(args ...interface{}) (interface{}, error) {
	fmt.Println("=======================================WRITE")
	params := &FlowAppParams{}
	resultsMetadata := &model.Point{}
	automater.DecodeTaskParams(args, params)
	automater.DecodePreviousJobResults(args, &resultsMetadata)
	write, err := runPointWrite(params)
	//time.Sleep(time.Duration(params.Delay) * time.Second)
	fmt.Println()
	return write, err
}

func runPointWrite(body *FlowAppParams) (interface{}, error) {
	cli := client.New(body.URL, body.Port)
	var writeValue *float64
	if body.Value == nil {
		writeValue = nils.NewFloat64(ran())
	} else {
		writeValue = body.Value
	}
	fmt.Println(90909090, nils.Float64IsNil(writeValue))

	pri := &model.Priority{
		P1:  nil,
		P2:  nil,
		P3:  nil,
		P4:  nil,
		P5:  nil,
		P6:  nil,
		P7:  nil,
		P8:  nil,
		P9:  nil,
		P10: nil,
		P11: nil,
		P12: nil,
		P13: nil,
		P14: nil,
		P15: nil,
		P16: writeValue,
	}

	point := &model.Point{
		Priority: pri,
	}

	install, res := cli.FlowPointWrite(body.UUID, point)
	fmt.Println(res.StatusCode)
	fmt.Println(res.StatusCode)
	fmt.Println(res.AsString())
	if res.StatusCode == 0 {
		return install, errors.New("failed to find host")
	}
	return install, nil
}

func ran() float64 {
	rand.Seed(time.Now().UnixNano())
	return float64(rand.Intn(100))

}
