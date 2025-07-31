# Initial Request

**Date:** 2025-07-31 08:01
**Request:** When starting a session I want to create a sandbox-friendly name and pass that to the tmux command as an environment variable. The friendly name should be generated from the title of the github issue and should only contain alphanumeric characters and hyphens. The environment variable name should be SBS_TITLE

## Context
This request is for the SBS (Sandbox Sessions) Go CLI application that orchestrates GitHub issue work environments with automatic git worktree and tmux session management.

## Key Requirements Identified
1. Generate sandbox-friendly names from GitHub issue titles
2. Sanitize names to only contain alphanumeric characters and hyphens
3. Pass the friendly name as environment variable `SBS_TITLE` to tmux commands
4. Integration with existing session start process