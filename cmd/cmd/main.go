package main

import (
	automater "github.com/NubeIO/rubix-automater"
	"github.com/NubeIO/rubix-automater/tasks"
)

func main() {
	v := automater.New("config.yaml")
	v.RegisterTask(tasks.PingHostName, tasks.PingHost)
	v.Run()
}
