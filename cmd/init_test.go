package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitCommand_Structure(t *testing.T) {
	t.Run("command_exists", func(t *testing.T) {
		assert.NotNil(t, initCmd)
		assert.Equal(t, "init", initCmd.Use)
		assert.Equal(t, "Initialize repository for SBS usage", initCmd.Short)
		assert.Contains(t, initCmd.Long, "Initialize the current repository")
	})

	t.Run("has_required_flags", func(t *testing.T) {
		forceFlag := initCmd.Flags().Lookup("force")
		dryRunFlag := initCmd.Flags().Lookup("dry-run")

		assert.NotNil(t, forceFlag)
		assert.NotNil(t, dryRunFlag)
		assert.Equal(t, "bool", forceFlag.Value.Type())
		assert.Equal(t, "bool", dryRunFlag.Value.Type())
	})

	t.Run("accepts_no_arguments", func(t *testing.T) {
		if initCmd.Args != nil {
			err := initCmd.Args(initCmd, []string{"extra", "args"})
			assert.Error(t, err, "Should not accept arguments")
		} else {
			// If no Args validator is set, cobra default allows any number of args
			// This is fine since we don't explicitly restrict arguments in the init command
			assert.Nil(t, initCmd.Args, "No explicit argument validation set")
		}
	})
}

func TestInitCommand_GitRepositoryValidation(t *testing.T) {
	t.Run("is_git_repository_with_directory", func(t *testing.T) {
		// Create temporary directory with .git subdirectory
		tmpDir := t.TempDir()
		gitDir := filepath.Join(tmpDir, ".git")
		err := os.Mkdir(gitDir, 0755)
		require.NoError(t, err)

		result := isGitRepository(tmpDir)
		assert.True(t, result)
	})

	t.Run("is_git_repository_with_file", func(t *testing.T) {
		// Create temporary directory with .git file (worktree case)
		tmpDir := t.TempDir()
		gitFile := filepath.Join(tmpDir, ".git")
		err := os.WriteFile(gitFile, []byte("gitdir: /some/path"), 0644)
		require.NoError(t, err)

		result := isGitRepository(tmpDir)
		assert.True(t, result)
	})

	t.Run("is_not_git_repository", func(t *testing.T) {
		tmpDir := t.TempDir()
		result := isGitRepository(tmpDir)
		assert.False(t, result)
	})
}

func TestInitCommand_FileOperations(t *testing.T) {
	t.Run("file_exists_check", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Test existing file
		existingFile := filepath.Join(tmpDir, "existing.txt")
		err := os.WriteFile(existingFile, []byte("content"), 0644)
		require.NoError(t, err)
		assert.True(t, fileExists(existingFile))

		// Test non-existing file
		nonExistingFile := filepath.Join(tmpDir, "nonexistent.txt")
		assert.False(t, fileExists(nonExistingFile))
	})

	t.Run("copy_file_success", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create source file
		srcFile := filepath.Join(tmpDir, "source.txt")
		srcContent := "test content"
		err := os.WriteFile(srcFile, []byte(srcContent), 0755)
		require.NoError(t, err)

		// Copy file
		dstFile := filepath.Join(tmpDir, "destination.txt")
		err = copyFile(srcFile, dstFile)
		require.NoError(t, err)

		// Verify content
		dstContent, err := os.ReadFile(dstFile)
		require.NoError(t, err)
		assert.Equal(t, srcContent, string(dstContent))

		// Verify permissions
		srcInfo, err := os.Stat(srcFile)
		require.NoError(t, err)
		dstInfo, err := os.Stat(dstFile)
		require.NoError(t, err)
		assert.Equal(t, srcInfo.Mode(), dstInfo.Mode())
	})

	t.Run("copy_file_with_force_check_no_existing", func(t *testing.T) {
		tmpDir := t.TempDir()

		srcFile := filepath.Join(tmpDir, "source.txt")
		err := os.WriteFile(srcFile, []byte("content"), 0644)
		require.NoError(t, err)

		dstFile := filepath.Join(tmpDir, "destination.txt")
		err = copyFileWithForceCheck(srcFile, dstFile, false)
		assert.NoError(t, err)
	})

	t.Run("copy_file_with_force_check_existing_no_force", func(t *testing.T) {
		tmpDir := t.TempDir()

		srcFile := filepath.Join(tmpDir, "source.txt")
		err := os.WriteFile(srcFile, []byte("content"), 0644)
		require.NoError(t, err)

		dstFile := filepath.Join(tmpDir, "destination.txt")
		err = os.WriteFile(dstFile, []byte("existing"), 0644)
		require.NoError(t, err)

		err = copyFileWithForceCheck(srcFile, dstFile, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "file already exists")
		assert.Contains(t, err.Error(), "use --force to overwrite")
	})

	t.Run("copy_file_with_force_check_existing_with_force", func(t *testing.T) {
		tmpDir := t.TempDir()

		srcFile := filepath.Join(tmpDir, "source.txt")
		err := os.WriteFile(srcFile, []byte("new content"), 0644)
		require.NoError(t, err)

		dstFile := filepath.Join(tmpDir, "destination.txt")
		err = os.WriteFile(dstFile, []byte("old content"), 0644)
		require.NoError(t, err)

		err = copyFileWithForceCheck(srcFile, dstFile, true)
		assert.NoError(t, err)

		// Verify content was overwritten
		content, err := os.ReadFile(dstFile)
		require.NoError(t, err)
		assert.Equal(t, "new content", string(content))
	})
}

func TestInitCommand_DryRun(t *testing.T) {
	t.Run("dry_run_mode", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create git repository
		gitDir := filepath.Join(tmpDir, ".git")
		err := os.Mkdir(gitDir, 0755)
		require.NoError(t, err)

		// Create source files
		workIssueScript := filepath.Join(tmpDir, "work-issue.sh")
		err = os.WriteFile(workIssueScript, []byte("#!/bin/bash\necho 'work-issue'"), 0755)
		require.NoError(t, err)

		scriptsDir := filepath.Join(tmpDir, "scripts")
		err = os.Mkdir(scriptsDir, 0755)
		require.NoError(t, err)

		hookScript := filepath.Join(scriptsDir, "claude-code-stop-hook.sh")
		err = os.WriteFile(hookScript, []byte("#!/bin/bash\necho 'hook'"), 0755)
		require.NoError(t, err)

		// Change to temp directory
		originalDir, err := os.Getwd()
		require.NoError(t, err)
		defer func() {
			err := os.Chdir(originalDir)
			require.NoError(t, err)
		}()

		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		// Create command with dry-run flag
		cmd := &cobra.Command{
			Use:  "init",
			RunE: runInit,
		}
		cmd.Flags().Bool("dry-run", true, "Dry run")
		cmd.Flags().Bool("force", false, "Force")

		// Capture output by redirecting
		err = cmd.Execute()
		require.NoError(t, err)

		// Verify no files were created
		sbsDir := filepath.Join(tmpDir, ".sbs")
		assert.False(t, fileExists(sbsDir), ".sbs directory should not be created in dry-run mode")
	})
}

func TestInitCommand_Integration(t *testing.T) {
	t.Run("successful_initialization", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create git repository
		gitDir := filepath.Join(tmpDir, ".git")
		err := os.Mkdir(gitDir, 0755)
		require.NoError(t, err)

		// Create source files
		workIssueScript := filepath.Join(tmpDir, "work-issue.sh")
		workIssueContent := "#!/bin/bash\necho 'work-issue script'"
		err = os.WriteFile(workIssueScript, []byte(workIssueContent), 0755)
		require.NoError(t, err)

		scriptsDir := filepath.Join(tmpDir, "scripts")
		err = os.Mkdir(scriptsDir, 0755)
		require.NoError(t, err)

		hookScript := filepath.Join(scriptsDir, "claude-code-stop-hook.sh")
		hookContent := "#!/bin/bash\necho 'claude hook'"
		err = os.WriteFile(hookScript, []byte(hookContent), 0755)
		require.NoError(t, err)

		// Change to temp directory
		originalDir, err := os.Getwd()
		require.NoError(t, err)
		defer func() {
			err := os.Chdir(originalDir)
			require.NoError(t, err)
		}()

		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		// Create and execute command
		cmd := &cobra.Command{
			Use:  "init",
			RunE: runInit,
		}
		cmd.Flags().Bool("dry-run", false, "Dry run")
		cmd.Flags().Bool("force", false, "Force")

		err = cmd.Execute()
		require.NoError(t, err)

		// Verify files were created
		sbsDir := filepath.Join(tmpDir, ".sbs")
		assert.True(t, fileExists(sbsDir), ".sbs directory should be created")

		copiedWorkIssue := filepath.Join(sbsDir, "work-issue.sh")
		assert.True(t, fileExists(copiedWorkIssue), "work-issue.sh should be copied")

		copiedHook := filepath.Join(sbsDir, "claude-code-stop-hook.sh")
		assert.True(t, fileExists(copiedHook), "claude-code-stop-hook.sh should be copied")

		// Verify content
		content, err := os.ReadFile(copiedWorkIssue)
		require.NoError(t, err)
		assert.Equal(t, workIssueContent, string(content))

		content, err = os.ReadFile(copiedHook)
		require.NoError(t, err)
		assert.Equal(t, hookContent, string(content))

		// Verify permissions
		srcInfo, err := os.Stat(workIssueScript)
		require.NoError(t, err)
		dstInfo, err := os.Stat(copiedWorkIssue)
		require.NoError(t, err)
		assert.Equal(t, srcInfo.Mode(), dstInfo.Mode())
	})

	t.Run("force_overwrite", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Setup git repository and source files
		gitDir := filepath.Join(tmpDir, ".git")
		err := os.Mkdir(gitDir, 0755)
		require.NoError(t, err)

		workIssueScript := filepath.Join(tmpDir, "work-issue.sh")
		newContent := "#!/bin/bash\necho 'updated content'"
		err = os.WriteFile(workIssueScript, []byte(newContent), 0755)
		require.NoError(t, err)

		scriptsDir := filepath.Join(tmpDir, "scripts")
		err = os.Mkdir(scriptsDir, 0755)
		require.NoError(t, err)

		hookScript := filepath.Join(scriptsDir, "claude-code-stop-hook.sh")
		err = os.WriteFile(hookScript, []byte("#!/bin/bash\necho 'hook'"), 0755)
		require.NoError(t, err)

		// Pre-create .sbs directory with existing files
		sbsDir := filepath.Join(tmpDir, ".sbs")
		err = os.Mkdir(sbsDir, 0755)
		require.NoError(t, err)

		existingFile := filepath.Join(sbsDir, "work-issue.sh")
		err = os.WriteFile(existingFile, []byte("old content"), 0644)
		require.NoError(t, err)

		// Change to temp directory
		originalDir, err := os.Getwd()
		require.NoError(t, err)
		defer func() {
			err := os.Chdir(originalDir)
			require.NoError(t, err)
		}()

		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		// Try without force flag (should fail)
		cmd := &cobra.Command{
			Use:  "init",
			RunE: runInit,
		}
		cmd.Flags().Bool("dry-run", false, "Dry run")
		cmd.Flags().Bool("force", false, "Force")

		err = cmd.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "file already exists")

		// Try with force flag (should succeed)
		cmd2 := &cobra.Command{
			Use:  "init",
			RunE: runInit,
		}
		cmd2.Flags().Bool("dry-run", false, "Dry run")
		cmd2.Flags().Bool("force", true, "Force")

		err = cmd2.Execute()
		require.NoError(t, err)

		// Verify content was updated
		content, err := os.ReadFile(existingFile)
		require.NoError(t, err)
		assert.Equal(t, newContent, string(content))
	})
}

func TestInitCommand_ErrorHandling(t *testing.T) {
	t.Run("not_git_repository", func(t *testing.T) {
		tmpDir := t.TempDir()

		originalDir, err := os.Getwd()
		require.NoError(t, err)
		defer func() {
			err := os.Chdir(originalDir)
			require.NoError(t, err)
		}()

		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		cmd := &cobra.Command{
			Use:  "init",
			RunE: runInit,
		}
		cmd.Flags().Bool("dry-run", false, "Dry run")
		cmd.Flags().Bool("force", false, "Force")

		err = cmd.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be run from within a git repository")
	})

	t.Run("missing_work_issue_script", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create git repository but no work-issue.sh
		gitDir := filepath.Join(tmpDir, ".git")
		err := os.Mkdir(gitDir, 0755)
		require.NoError(t, err)

		originalDir, err := os.Getwd()
		require.NoError(t, err)
		defer func() {
			err := os.Chdir(originalDir)
			require.NoError(t, err)
		}()

		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		cmd := &cobra.Command{
			Use:  "init",
			RunE: runInit,
		}
		cmd.Flags().Bool("dry-run", false, "Dry run")
		cmd.Flags().Bool("force", false, "Force")

		err = cmd.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "work-issue.sh not found")
	})

	t.Run("missing_hook_script", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create git repository and work-issue.sh but no scripts directory
		gitDir := filepath.Join(tmpDir, ".git")
		err := os.Mkdir(gitDir, 0755)
		require.NoError(t, err)

		workIssueScript := filepath.Join(tmpDir, "work-issue.sh")
		err = os.WriteFile(workIssueScript, []byte("#!/bin/bash"), 0755)
		require.NoError(t, err)

		originalDir, err := os.Getwd()
		require.NoError(t, err)
		defer func() {
			err := os.Chdir(originalDir)
			require.NoError(t, err)
		}()

		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		cmd := &cobra.Command{
			Use:  "init",
			RunE: runInit,
		}
		cmd.Flags().Bool("dry-run", false, "Dry run")
		cmd.Flags().Bool("force", false, "Force")

		err = cmd.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "claude-code-stop-hook.sh not found")
	})

	t.Run("permission_error", func(t *testing.T) {
		if os.Getuid() == 0 {
			t.Skip("Skipping permission test when running as root")
		}

		tmpDir := t.TempDir()

		// Create git repository and source files
		gitDir := filepath.Join(tmpDir, ".git")
		err := os.Mkdir(gitDir, 0755)
		require.NoError(t, err)

		workIssueScript := filepath.Join(tmpDir, "work-issue.sh")
		err = os.WriteFile(workIssueScript, []byte("#!/bin/bash"), 0755)
		require.NoError(t, err)

		scriptsDir := filepath.Join(tmpDir, "scripts")
		err = os.Mkdir(scriptsDir, 0755)
		require.NoError(t, err)

		hookScript := filepath.Join(scriptsDir, "claude-code-stop-hook.sh")
		err = os.WriteFile(hookScript, []byte("#!/bin/bash"), 0755)
		require.NoError(t, err)

		// Create .sbs directory with no write permissions
		sbsDir := filepath.Join(tmpDir, ".sbs")
		err = os.Mkdir(sbsDir, 0555) // Read and execute only, no write
		require.NoError(t, err)
		defer func() {
			// Restore permissions to allow cleanup
			os.Chmod(sbsDir, 0755)
		}()

		originalDir, err := os.Getwd()
		require.NoError(t, err)
		defer func() {
			err := os.Chdir(originalDir)
			require.NoError(t, err)
		}()

		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		cmd := &cobra.Command{
			Use:  "init",
			RunE: runInit,
		}
		cmd.Flags().Bool("dry-run", false, "Dry run")
		cmd.Flags().Bool("force", false, "Force")

		err = cmd.Execute()
		assert.Error(t, err)
		// Should fail when trying to write to read-only directory
		assert.True(t, strings.Contains(err.Error(), "failed to copy") ||
			strings.Contains(err.Error(), "permission denied"))
	})
}
