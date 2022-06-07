package main

import (
	automater "github.com/NubeIO/rubix-automater"
	tasks "github.com/NubeIO/rubix-automater/service/tasks"
)

func main() {
	v := automater.New("config.yaml")
	v.RegisterTask(tasks.PingHostTask, tasks.PingHost)
	v.RegisterTask(tasks.InstallAppTask, tasks.InstallApp)
	v.RegisterTask(tasks.PointWriteTask, tasks.PointWrite)
	v.Run()
}
