# Expert Detail Answers

## Q1: Should the 's' key stop only the currently selected session, or should it work like the existing stop command requiring the session to be selected first?
**Answer:** Yes (stop currently selected session directly)

## Q2: Should the clean operation ('c' key) work on all stale sessions globally, or only stale sessions visible in the current view mode (repository vs global)?
**Answer:** Current view only

## Q3: Should the confirmation dialog for clean operations be implemented as a modal overlay, or as a status line prompt like the existing CLI clean command?
**Answer:** Modal overlay

## Q4: Should stop operations that fail (e.g., tmux session already stopped) still trigger an automatic refresh of the session list?
**Answer:** Yes, be sure to display the error to the user

## Q5: Should the help text in pkg/tui/model.go:242-247 be updated to include the new 's' and 'c' shortcuts in the condensed help?
**Answer:** Yes