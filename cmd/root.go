package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"sbs/pkg/cmdlog"
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
var verbose bool

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringP("config", "c", "", "config file (default is ~/.config/sbs/config.json)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose command logging")
}

func initConfig() {
	var err error
	cfg, err = config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Initialize command logging based on configuration and verbose flag
	enableLogging := cfg.CommandLogging || verbose
	if enableLogging {
		logConfig := cmdlog.Config{
			Enabled:  true,
			Level:    cfg.CommandLogLevel,
			FilePath: cfg.CommandLogPath,
		}

		// Override settings when verbose flag is used
		if verbose {
			logConfig.Level = "debug"
			// If no file path is configured, verbose output goes to stderr
			if logConfig.FilePath == "" {
				logConfig.FilePath = "" // Empty string means stderr output
			}
		}

		// Set default log level if not specified
		if logConfig.Level == "" {
			logConfig.Level = "info"
		}

		logger := cmdlog.NewCommandLogger(logConfig)
		cmdlog.SetGlobalLogger(logger)
	}

	// Validate required tools are available
	if err := validation.CheckRequiredTools(); err != nil {
		fmt.Printf("Tool validation failed:\n%v", err)
		os.Exit(1)
	}
}
