# Expert Detail Answers

## Q1: Should we modify the existing `tmux.CreateSession()` method to accept environment variables, or create a new method variant?
**Answer:** Modify existing method to accept optional environment map

## Q2: Should the friendly name be stored in `SessionMetadata` to avoid re-sanitization on attach operations?
**Answer:** Yes (store as new field `FriendlyTitle` for performance and consistency)

## Q3: Should we create a new sanitization utility or extend the existing `repo.Manager.SanitizeName()` method?
**Answer:** Extend existing `repo.Manager.SanitizeName()` to accept a length parameter

## Q4: Should the SBS_TITLE environment variable be set only during session creation or also when executing commands via `ExecuteCommand()`?
**Answer:** Both (ensures environment consistency across all tmux operations)

## Q5: Should we handle the case where issue title is unavailable (e.g., network issues) by using a fallback friendly name?
**Answer:** Yes, fallback to "{reponame}-issue-{number}" format to avoid collisions with other projects