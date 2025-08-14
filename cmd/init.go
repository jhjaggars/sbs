package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize repository for SBS usage",
	Long: `Initialize the current repository for use with SBS by creating a .sbs directory
and populating it with the necessary scripts.

This command creates:
- .sbs/start (template start script for this repository)
- .sbs/claude-code-stop-hook.sh (copied from scripts/ directory)

After running this command, 'sbs start' will use the local .sbs/start
script if it exists, otherwise will start the session without any script.`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().Bool("force", false, "Overwrite existing files")
	initCmd.Flags().Bool("dry-run", false, "Show what would be created without making changes")
}

func runInit(cmd *cobra.Command, args []string) error {
	force, _ := cmd.Flags().GetBool("force")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	// Check if we're in a git repository
	if !isGitRepository(cwd) {
		return fmt.Errorf("must be run from within a git repository")
	}

	sbsDir := filepath.Join(cwd, ".sbs")
	startScript := filepath.Join(sbsDir, "start")
	hookScript := filepath.Join(sbsDir, "claude-code-stop-hook.sh")

	fmt.Printf("Initializing SBS in repository: %s\n", cwd)

	if dryRun {
		fmt.Println("\n[DRY RUN] Would perform the following actions:")
	}

	// Create .sbs directory
	if dryRun {
		fmt.Printf("- Create directory: %s\n", sbsDir)
	} else {
		if err := os.MkdirAll(sbsDir, 0755); err != nil {
			return fmt.Errorf("failed to create .sbs directory: %w", err)
		}

		// Validate directory creation
		if !fileExists(sbsDir) {
			return fmt.Errorf("failed to verify .sbs directory creation at: %s", sbsDir)
		}

		fmt.Printf("Created directory: %s\n", sbsDir)
	}

	// Create template start script
	if dryRun {
		fmt.Printf("- Create template start script: %s\n", startScript)
	} else {
		if err := createStartScriptWithForceCheck(startScript, force); err != nil {
			return fmt.Errorf("failed to create start script: %w", err)
		}
		fmt.Printf("Created template start script: %s\n", startScript)
	}

	// Copy claude-code-stop-hook.sh
	srcHook := filepath.Join(cwd, "scripts", "claude-code-stop-hook.sh")
	if !fileExists(srcHook) {
		return fmt.Errorf("claude-code-stop-hook.sh not found at expected path: %s", srcHook)
	}

	if dryRun {
		fmt.Printf("- Copy %s -> %s\n", srcHook, hookScript)
	} else {
		if err := copyFileWithForceCheck(srcHook, hookScript, force); err != nil {
			return fmt.Errorf("failed to copy claude-code-stop-hook.sh: %w", err)
		}
		fmt.Printf("Copied: %s -> %s\n", srcHook, hookScript)
	}

	if dryRun {
		fmt.Println("\n[DRY RUN] No changes made.")
		return nil
	}

	fmt.Println("\nSBS initialization completed successfully!")
	fmt.Println("- 'sbs start' will use .sbs/start script if it exists")
	fmt.Println("- You can customize .sbs/start for this repository's specific needs")
	fmt.Println("- Consider adding .sbs/ to version control to share team configuration")

	return nil
}

func isGitRepository(path string) bool {
	gitDir := filepath.Join(path, ".git")
	if stat, err := os.Stat(gitDir); err == nil {
		// .git exists - it can be either a directory (normal repo) or file (worktree)
		return stat.IsDir() || stat.Mode().IsRegular()
	}

	return false
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func copyFileWithForceCheck(src, dst string, force bool) error {
	// Check if destination exists and force flag
	if fileExists(dst) && !force {
		return fmt.Errorf("file already exists: %s (use --force to overwrite)", dst)
	}

	return copyFile(src, dst)
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}

	// Copy permissions
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, srcInfo.Mode())
}

// createStartScriptWithForceCheck creates a template start script with force check
func createStartScriptWithForceCheck(dst string, force bool) error {
	// Check if destination exists and force flag
	if fileExists(dst) && !force {
		return fmt.Errorf("file already exists: %s (use --force to overwrite)", dst)
	}

	return createStartScript(dst)
}

// createStartScript creates a template start script
func createStartScript(dst string) error {
	templateContent := `#!/bin/bash
# SBS Start Script
# This script runs when starting a new SBS session for this repository.
# Customize this script to set up your development environment.

set -e

echo "Starting SBS session..."

# Example: Install dependencies
# npm install

# Example: Start development servers
# npm run dev &

# Example: Open editor
# code .

echo "SBS session setup complete!"
`

	if err := os.WriteFile(dst, []byte(templateContent), 0755); err != nil {
		return err
	}

	return nil
}
