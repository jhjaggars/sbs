# Initial Requirements Request

**Date:** 2025-07-31 13:47
**Request:** from the list we should be able to stop and clean sessions

## Context
User wants the ability to stop and clean sessions directly from the interactive list interface, rather than having to exit the list and run separate commands.

## Initial Understanding
Currently users can:
- View active sessions with `./sbs list` (interactive TUI)
- Stop sessions with `./sbs stop {number}`
- Clean stale sessions with `./sbs clean`

The request is to integrate stop and clean functionality into the list interface itself.