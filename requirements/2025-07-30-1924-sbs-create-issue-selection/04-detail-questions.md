# Detailed Technical Questions

## Q1: Should the new ListIssues method in pkg/issue/github.go follow the same error handling pattern as GetIssue(), returning a custom error for "no issues found"?
**Default if unknown:** Yes (consistent with existing GetIssue error handling in pkg/issue/github.go:36-42)

## Q2: Should the issue selection TUI be implemented as a separate pkg/tui/issueselect.go file or integrated into the existing pkg/tui/model.go?
**Default if unknown:** Yes, separate file (follows single responsibility principle and keeps session management TUI separate)

## Q3: Should the search functionality filter issues client-side after fetching or use GitHub's server-side search via `gh issue list --search`?
**Default if unknown:** Client-side filtering (simpler implementation, works with pagination, faster user experience)

## Q4: Should the pagination limit for `gh issue list --limit` be configurable via config.Config or hardcoded?
**Default if unknown:** Hardcoded initially (start with reasonable default like 100, add config later if needed)

## Q5: Should the issue selection TUI show issue numbers, titles, and state (open/closed) or just numbers and titles?
**Default if unknown:** Numbers and titles only (state is always "open" based on discovery answers, cleaner interface)

## Q6: Should the modification to cmd/start.go runStart() check for repository context before or after the argument count check?
**Default if unknown:** Before argument check (fail fast if not in repository, consistent with existing validation at cmd/start.go:42-47)

## Q7: Should the issue selection TUI handle empty issue lists with a specific message or fall back to prompting for manual issue number entry?
**Default if unknown:** Show specific message and exit (clear feedback, user can create issues via gh or GitHub web interface)

## Q8: Should the textinput component for search be visible by default or triggered by a key press (like '/' for search mode)?
**Default if unknown:** Visible by default (simpler UX, follows existing TUI patterns, no mode switching complexity)

## Q9: Should the issue selection preserve the existing color scheme from pkg/tui/styles.go or use different colors to distinguish it from session management?
**Default if unknown:** Reuse existing color scheme (visual consistency, already well-designed, less maintenance)

## Q10: Should canceling out of the issue selection (pressing 'q') exit the entire command or return to command prompt?
**Default if unknown:** Exit entire command (matches existing TUI behavior in pkg/tui/model.go:116, clean exit pattern)