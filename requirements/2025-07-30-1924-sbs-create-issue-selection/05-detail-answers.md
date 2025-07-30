# Detailed Technical Answers

## Q1: Should the new ListIssues method in pkg/issue/github.go follow the same error handling pattern as GetIssue(), returning a custom error for "no issues found"?
**Answer:** Yes

## Q2: Should the issue selection TUI be implemented as a separate pkg/tui/issueselect.go file or integrated into the existing pkg/tui/model.go?
**Answer:** Yes

## Q3: Should the search functionality filter issues client-side after fetching or use GitHub's server-side search via `gh issue list --search`?
**Answer:** Server-side

## Q4: Should the pagination limit for `gh issue list --limit` be configurable via config.Config or hardcoded?
**Answer:** Hardcoded

## Q5: Should the issue selection TUI show issue numbers, titles, and state (open/closed) or just numbers and titles?
**Answer:** Numbers and titles

## Q6: Should the modification to cmd/start.go runStart() check for repository context before or after the argument count check?
**Answer:** Before

## Q7: Should the issue selection TUI handle empty issue lists with a specific message or fall back to prompting for manual issue number entry?
**Answer:** Show specific message and exit

## Q8: Should the textinput component for search be visible by default or triggered by a key press (like '/' for search mode)?
**Answer:** Visible by default

## Q9: Should the issue selection preserve the existing color scheme from pkg/tui/styles.go or use different colors to distinguish it from session management?
**Answer:** Reuse existing color scheme

## Q10: Should canceling out of the issue selection (pressing 'q') exit the entire command or return to command prompt?
**Answer:** Exit entirely