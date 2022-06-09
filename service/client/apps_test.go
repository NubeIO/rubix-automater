package client

import (
	"fmt"
	"github.com/NubeIO/nubeio-rubix-lib-helpers-go/pkg/nils"
	"github.com/NubeIO/nubeio-rubix-lib-models-go/pkg/v1/model"
	"testing"
)

func TestHost(*testing.T) {

	//client := New("0.0.0.0", 8089)
	//
	//data, _ := client.GetJobs()
	//
	//if data.Jobs == nil {
	//
	//}
	//for i, job := range data.Jobs {
	//	fmt.Println(i, job)
	//}

	client := New("192.168.15.191", 1660)

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
		P16: nils.NewFloat64(2323),
	}

	point := &model.Point{
		Priority: pri,
	}

	data, res := client.FlowPointWrite("pnt_c60aa01f57b24f3a", point)

	fmt.Println(res.StatusCode)
	fmt.Println(data)

}
