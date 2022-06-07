package cmd

import (
	"fmt"
	automater "github.com/NubeIO/rubix-automater"
	"github.com/NubeIO/rubix-automater/service/tasks"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"time"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:           "server",
	Short:         "run automater server",
	SilenceErrors: true,
	SilenceUsage:  true,
	Run:           runRoot,
}

var rootFlags struct {
	server bool
	config string
	wipeDb bool
}

func runServer() {
	if rootFlags.server {
		v := automater.New(rootFlags.config)
		v.RegisterTask(tasks.PingHostTask, tasks.PingHost)
		v.RegisterTask(tasks.InstallAppTask, tasks.InstallApp)
		v.RegisterTask(tasks.PointWriteTask, tasks.PointWrite)
		v.Run()
	}
}

func runRoot(cmd *cobra.Command, args []string) {

	if rootFlags.server {
		v := automater.New(rootFlags.config)
		v.RegisterTask(tasks.PingHostTask, tasks.PingHost)
		v.RegisterTask(tasks.InstallAppTask, tasks.InstallApp)
		v.RegisterTask(tasks.PointWriteTask, tasks.PointWrite)
		v.Run()
	}

	//go runServer()
	//
	//if rootFlags.wipeDb {
	//	time.Sleep(5 * time.Second)
	//	cli := initRest()
	//	res := &client.Response{}
	//	res = cli.WipeDB()
	//	fmt.Println(res.StatusCode)
	//	fmt.Println(res.AsString())
	//	keepRunning()
	//} else {
	//	keepRunning()
	//}

}

func keepRunning() os.Signal {
	signalChan := make(chan os.Signal, 1)
	defer close(signalChan)
	signal.Notify(signalChan, os.Kill, os.Interrupt)
	s := <-signalChan
	signal.Stop(signalChan)
	time.Sleep(4 * time.Second)
	return s
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		//color.Magenta(err.Error())
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	pFlagSet := rootCmd.PersistentFlags()
	pFlagSet.StringVarP(&rootFlags.config, "config", "", "config.yaml", "set config path example ./config.yaml")
	pFlagSet.BoolVarP(&rootFlags.server, "server", "", false, "run server")
	pFlagSet.BoolVarP(&rootFlags.wipeDb, "wipe", "", false, "delete the db after server has started")
}
