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
and populating it with the necessary scripts (work-issue.sh and loghook script).

This command creates:
- .sbs/work-issue.sh (copied from repository root or template)
- .sbs/claude-code-stop-hook.sh (copied from scripts/ directory)

After running this command, 'sbs start' will use the local .sbs/work-issue.sh
script by default instead of the global configuration.`,
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
	workIssueScript := filepath.Join(sbsDir, "work-issue.sh")
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
		fmt.Printf("Created directory: %s\n", sbsDir)
	}

	// Copy work-issue.sh
	srcWorkIssue := filepath.Join(cwd, "work-issue.sh")
	if !fileExists(srcWorkIssue) {
		return fmt.Errorf("work-issue.sh not found in repository root")
	}

	if dryRun {
		fmt.Printf("- Copy %s -> %s\n", srcWorkIssue, workIssueScript)
	} else {
		if err := copyFileWithForceCheck(srcWorkIssue, workIssueScript, force); err != nil {
			return fmt.Errorf("failed to copy work-issue.sh: %w", err)
		}
		fmt.Printf("Copied: %s -> %s\n", srcWorkIssue, workIssueScript)
	}

	// Copy claude-code-stop-hook.sh
	srcHook := filepath.Join(cwd, "scripts", "claude-code-stop-hook.sh")
	if !fileExists(srcHook) {
		return fmt.Errorf("scripts/claude-code-stop-hook.sh not found")
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
	fmt.Println("- 'sbs start' will now use .sbs/work-issue.sh by default")
	fmt.Println("- You can customize .sbs/work-issue.sh for this repository's specific needs")
	fmt.Println("- Consider adding .sbs/ to version control to share team configuration")

	return nil
}

func isGitRepository(path string) bool {
	gitDir := filepath.Join(path, ".git")
	if stat, err := os.Stat(gitDir); err == nil {
		return stat.IsDir()
	}

	// Check if .git is a file (in case of worktrees)
	if _, err := os.Stat(gitDir); err == nil {
		return true
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
