package cmd

import (
	"log"
	"os"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "radSched",
	Short: "Rad-Sched is a Scheduler for Radical",
	Long:  `Rad-Sched is a CLI tool to schedule functions efficiently on Radical.`,
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		log.Fatalf("Error starting RadSched: %v", err)
		os.Exit(1)
	}
}

func InitRootCmd() {
	RootCmd.AddCommand(BootstrapCmd)
	RootCmd.AddCommand(PrepareCmd)
	RootCmd.AddCommand(RunCmd)
	RunCmd.Flags().Bool("with-weight", false, "Run the function with weight")
}