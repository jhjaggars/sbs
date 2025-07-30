# Discovery Questions

## Q1: Should `sbs start` with no arguments change the command signature from requiring an issue number to making it optional?
**Default if unknown:** Yes (backward compatibility maintained, new interactive mode when no args provided)

## Q2: Should the issue list show all repository issues or only open issues by default?
**Default if unknown:** Yes, only open issues (most relevant for active development work)

## Q3: Should users be able to filter issues by labels, assignee, or other criteria in the selection interface?
**Default if unknown:** No (keep initial implementation simple, add filtering later if needed)

## Q4: Should the issue selection interface show issue descriptions/bodies in addition to titles?
**Default if unknown:** No (keep interface clean and fast, titles are usually sufficient for selection)

## Q5: Should there be a search/filter capability within the issue list?
**Default if unknown:** Yes (essential for repositories with many issues)

## Q6: Should the interface support multi-repository issue selection if run from outside a git repository?
**Default if unknown:** No (scope to single repository to match existing session management patterns)

## Q7: Should the issue list be cached locally to improve performance on subsequent runs?
**Default if unknown:** No (always fetch fresh data to ensure accuracy, caching can be added later)

## Q8: Should the interface support pagination for repositories with hundreds of issues?
**Default if unknown:** Yes (GitHub API typically paginates, but start simple with reasonable limit)

## Q9: Should keyboard shortcuts match the existing TUI interface (j/k for navigation, q to quit)?
**Default if unknown:** Yes (consistency with existing pkg/tui patterns)

## Q10: Should users be able to create new issues directly from this interface if no suitable issue exists?
**Default if unknown:** No (separate concern, `sbs start` should focus on selecting existing issues)