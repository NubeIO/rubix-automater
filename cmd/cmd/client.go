package cmd

import (
	"fmt"
	"github.com/NubeIO/rubix-automater/automater/model"
	"github.com/NubeIO/rubix-automater/controller/jobctl"
	"github.com/NubeIO/rubix-automater/controller/pipectl"
	"github.com/NubeIO/rubix-automater/service/client"
	"github.com/spf13/cobra"
)

type T struct {
	Jobs []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		TaskName    string `json:"task_name"`
		TaskParams  struct {
			Url                string `json:"url"`
			Port               int    `json:"port"`
			ErrorOnFailSetting int    `json:"errorOnFailSetting"`
			DelayBetween       int    `json:"delayBetween"`
			Uuid               string `json:"uuid,omitempty"`
			Delay              int    `json:"delay,omitempty"`
			Value              int    `json:"value,omitempty"`
		} `json:"task_params"`
		UsePreviousResults bool `json:"use_previous_results"`
		Options            struct {
		} `json:"options"`
		Timeout int `json:"timeout,omitempty"`
	} `json:"jobs"`
}

var clientCmd = &cobra.Command{
	Use:           "client",
	Short:         "rest client",
	SilenceErrors: true,
	SilenceUsage:  true,
	Run:           runClient,
}

var clientFlags struct {
	wipDB           bool
	addPingPipeline bool
}

func initRest() *client.Client {
	return client.New("0.0.0.0", 8089)
}

func runClient(cmd *cobra.Command, args []string) {
	cli := initRest()
	if clientFlags.addPingPipeline {
		res := &client.Response{}
		jobOne := &jobctl.RequestBodyDTO{
			Name:     "ping 1",
			TaskName: "pingHost",
			Options: &model.JobOptions{
				EnableInterval: false,
				RunOnInterval:  "",
			},
			TaskParams: map[string]interface{}{"url": "0.0.0.0", "port": 1660},
		}

		var jobs []*jobctl.RequestBodyDTO

		jobs = append(jobs, jobOne)
		jobOne.Name = "ping 2"
		jobs = append(jobs, jobOne)

		body := &pipectl.RequestBodyDTO{
			Name:       "ping pipeline",
			Jobs:       jobs,
			ScheduleAt: "10 sec",
			PipelineOptions: &model.PipelineOptions{
				EnableInterval:   true,
				RunOnInterval:    "10 sec",
				DelayBetweenTask: 0,
				CancelOnFailure:  false,
			},
		}

		pipeline, res := cli.AddPipeline(body)
		fmt.Println(res.StatusCode)
		if res.StatusCode > 299 {
			fmt.Println(res.AsString())
		}
		fmt.Println("ADDED NEW PIPELINE", pipeline.Name)
	}

}

func init() {
	rootCmd.AddCommand(clientCmd)
	flagSet := clientCmd.Flags()
	flagSet.BoolVarP(&clientFlags.addPingPipeline, "add-ping", "", false, "add one ping job to the pipeline")
}
