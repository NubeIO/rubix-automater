package main

import (
	automater "github.com/NubeIO/rubix-automater"
	"github.com/NubeIO/rubix-automater/tasks"
)

func main() {
	v := automater.New("config.yaml")
	v.RegisterTask(tasks.PingHostTask, tasks.PingHost)
	v.RegisterTask(tasks.InstallAppTask, tasks.InstallApp)
	v.Run()
}
