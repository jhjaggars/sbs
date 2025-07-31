package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"sbs/pkg/config"
	"sbs/pkg/validation"
)

var rootCmd = &cobra.Command{
	Use:   "sbs",
	Short: "Sandbox Sessions - Manage GitHub issue work environments",
	Long: `SBS (Sandbox Sessions) creates and manages isolated work environments for GitHub issues.
It automatically handles:
- Git branch creation and worktrees
- Tmux session management
- Integration with work-issue.sh script

Each issue gets its own branch, worktree, and tmux session for organized development.`,
}

var cfg *config.Config

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringP("config", "c", "", "config file (default is ~/.config/sbs/config.json)")
}

func initConfig() {
	var err error
	cfg, err = config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Validate required tools are available
	if err := validation.CheckRequiredTools(); err != nil {
		fmt.Printf("Tool validation failed:\n%v", err)
		os.Exit(1)
	}
}
