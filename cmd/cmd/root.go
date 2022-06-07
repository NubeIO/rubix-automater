package cmd

import (
	"fmt"
	automater "github.com/NubeIO/rubix-automater"
	"github.com/NubeIO/rubix-automater/service/tasks"
	"github.com/spf13/cobra"
	"os"
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
}

func runRoot(cmd *cobra.Command, args []string) {

	if rootFlags.server {
		v := automater.New(rootFlags.config)
		v.RegisterTask(tasks.PingHostTask, tasks.PingHost)
		v.RegisterTask(tasks.InstallAppTask, tasks.InstallApp)
		v.RegisterTask(tasks.PointWriteTask, tasks.PointWrite)
		v.Run()
	}

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
}
