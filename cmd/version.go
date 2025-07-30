package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Work Orchestrator v1.0.0")
		fmt.Println("A GitHub issue work environment orchestrator")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}