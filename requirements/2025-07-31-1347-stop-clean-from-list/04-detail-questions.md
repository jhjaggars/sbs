# Expert Detail Questions

Based on deep codebase analysis, these questions clarify specific technical behaviors and user experience details.

## Q1: Should the 's' key stop only the currently selected session, or should it work like the existing stop command requiring the session to be selected first?
**Default if unknown:** Yes (stop currently selected session directly - more intuitive UX)

## Q2: Should the clean operation ('c' key) work on all stale sessions globally, or only stale sessions visible in the current view mode (repository vs global)?
**Default if unknown:** Current view only (respects user's current context and view filter)

## Q3: Should the confirmation dialog for clean operations be implemented as a modal overlay, or as a status line prompt like the existing CLI clean command?
**Default if unknown:** Status line prompt (simpler implementation, consistent with existing CLI patterns)

## Q4: Should stop operations that fail (e.g., tmux session already stopped) still trigger an automatic refresh of the session list?
**Default if unknown:** Yes (ensures UI reflects current state even after failed operations)

## Q5: Should the help text in pkg/tui/model.go:242-247 be updated to include the new 's' and 'c' shortcuts in the condensed help?
**Default if unknown:** Yes (maintain discoverability pattern established in attach-from-list requirements)