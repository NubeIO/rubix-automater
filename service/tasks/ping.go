package tasks

import (
	"errors"
	"fmt"
	automater "github.com/NubeIO/rubix-automater"
	"github.com/sirupsen/logrus"
	"net"
	"time"
)

type PingParams struct {
	URL                string `json:"url,omitempty"`
	Port               int    `json:"port,omitempty"`
	ErrorOnFailSetting int    `json:"errorOnFailSetting,omitempty"` //consider failed if count is > then the amount of times the ping failed
	DelayBetween       int    `json:"delayBetween"`
}

type PingResponse struct {
	Ok    bool  `json:"ok"`
	Error error `json:"error"`
}

func PingHost(args ...interface{}) (interface{}, error) {

	params := &PingParams{}
	var resultsMetadata string
	automater.DecodeTaskParams(args, params)
	fmt.Println("=======================================PING", params.ErrorOnFailSetting)
	automater.DecodePreviousJobResults(args, &resultsMetadata)
	metaData, err := runPingHost(params.URL, params.Port, params.ErrorOnFailSetting)
	fmt.Println("=======================================PING after", params.ErrorOnFailSetting)

	//time.Sleep(time.Duration(params.DelayBetween) * time.Second)
	return metaData, err
}

func runPingHost(url string, port int, countSetting int) (*PingResponse, error) {
	resp := &PingResponse{}
	failCount := 0
	for i := 1; i <= 1; i++ { //ping 3 times
		if ping(url, port) {
			logrus.Infoln("run task ping host ok:", fmt.Sprintf("%s:%d", url, port))
		} else {
			failCount++
			logrus.Infoln("run task ping host:", fmt.Sprintf("%s:%d", url, port), " fail count:", failCount)
		}

	}
	if failCount >= 1 {
		resp.Error = errors.New(fmt.Sprintf("ping fail count:%d was grater then the allowable ping fail count %d", failCount, countSetting))
		return resp, resp.Error
	}

	resp.Ok = true
	return resp, resp.Error
}

func ping(url string, port int) (found bool) {
	conn, err := net.DialTimeout("tcp",
		fmt.Sprintf("%s:%d", url, port),
		300*time.Millisecond)
	if err == nil {
		conn.Close()
		return true
	}
	logrus.Errorln("run task ping error:", err)
	return false

}
