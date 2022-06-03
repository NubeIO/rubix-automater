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
}

func PingHost(args ...interface{}) (interface{}, error) {
	params := &PingParams{}
	var resultsMetadata string
	automater.DecodeTaskParams(args, params)
	automater.DecodePreviousJobResults(args, &resultsMetadata)
	time.Sleep(1 * time.Second)
	metaData := runPingHost(params.URL, params.Port, params.ErrorOnFailSetting)
	return metaData, metaData
}

func runPingHost(url string, port int, countSetting int) error {
	failCount := 0
	for i := 1; i <= 3; i++ { //ping 3 times
		if ping(url, port) {
			logrus.Infoln("run task ping host ok:", fmt.Sprintf("%s:%d", url, port))
		} else {
			failCount++
			logrus.Infoln("run task ping host:", fmt.Sprintf("%s:%d", url, port), " fail count:", failCount)
		}
		time.Sleep(5 * time.Second)
	}
	if failCount >= countSetting {
		return errors.New(fmt.Sprintf("ping fail count:%d was grater then the allowable ping fail count %d", failCount, countSetting))
	}
	return nil
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
