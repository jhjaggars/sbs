package validation

import (
	"fmt"
	"os/exec"

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
	cmd := exec.Command("tmux", "-V")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("tmux not found. Please install tmux")
	}
	return nil
}

func checkGit() error {
	cmd := exec.Command("git", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git not found. Please install git")
	}
	return nil
}