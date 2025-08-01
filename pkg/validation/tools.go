package validation

import (
	"fmt"
	"os/exec"
	"time"

	"sbs/pkg/cmdlog"
	"sbs/pkg/issue"
	"sbs/pkg/sandbox"
)

// CheckRequiredTools validates that all required external tools are available
func CheckRequiredTools() error {
	var errors []string

	// Check tmux
	if err := checkTmux(); err != nil {
		errors = append(errors, err.Error())
	}

	// Check git
	if err := checkGit(); err != nil {
		errors = append(errors, err.Error())
	}

	// Check gh (GitHub CLI)
	if err := issue.CheckGHInstalled(); err != nil {
		errors = append(errors, err.Error())
	}

	// Check sandbox
	if err := sandbox.CheckSandboxInstalled(); err != nil {
		errors = append(errors, err.Error())
	}

	if len(errors) > 0 {
		errorMsg := "Missing required tools:\n"
		for _, err := range errors {
			errorMsg += "  - " + err + "\n"
		}
		return fmt.Errorf("%s", errorMsg)
	}

	return nil
}

func checkTmux() error {
	return runValidationCommand("tmux", []string{"-V"}, "tmux not found. Please install tmux")
}

func checkGit() error {
	return runValidationCommand("git", []string{"--version"}, "git not found. Please install git")
}

// runValidationCommand executes a command with logging for validation purposes
func runValidationCommand(command string, args []string, errorMsg string) error {
	ctx := cmdlog.LogCommandGlobal(command, args, cmdlog.GetCaller())

	cmd := exec.Command(command, args...)
	start := time.Now()
	err := cmd.Run()
	duration := time.Since(start)

	if err != nil {
		ctx.LogCompletion(false, getExitCode(cmd), err.Error(), duration)
		return fmt.Errorf("%s", errorMsg)
	}

	ctx.LogCompletion(true, 0, "", duration)
	return nil
}

// getExitCode extracts exit code from exec.Cmd
func getExitCode(cmd *exec.Cmd) int {
	if cmd.ProcessState != nil {
		return cmd.ProcessState.ExitCode()
	}
	return -1
}
