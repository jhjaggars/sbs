# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

This is a minimal shell script repository containing a single bash script `work-issue.sh` that provides a wrapper for launching Claude Code in a sandboxed environment to work on GitHub issues.

## Architecture

The repository contains only one file:
- `work-issue.sh`: A bash script that sets terminal/tmux window titles and launches Claude Code in a sandbox environment with specific parameters

## Key Script Functionality

The `work-issue.sh` script:
1. Takes a GitHub issue number as an argument ($1)
2. Sets the terminal/tmux window title to include the issue title (fetched via `gh issue view`)
3. Launches Claude Code in a sandbox with network access and tmux binding
4. Uses the "sonnet" model with permissions skipped
5. Passes the issue number as a parameter to Claude Code

## Dependencies

The script requires:
- `gh` (GitHub CLI) for fetching issue information
- `tmux` support for window renaming
- `sandbox` command for containerized execution
- Claude Code CLI installed at `~/.claude/local/claude`

## Usage

Execute the script with a GitHub issue number:
```bash
./work-issue.sh <issue-number>
```

This will launch Claude Code in a sandboxed environment specifically configured to work on the specified GitHub issue.