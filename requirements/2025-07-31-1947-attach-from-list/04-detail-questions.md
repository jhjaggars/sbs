# Detail Questions

## Q6: Should we modify the condensed help text in pkg/tui/model.go:247 to include "enter to attach"?
**Default if unknown:** Yes (follows pattern from issueselect.go:335 which shows primary action)

## Q7: Should the help text change be applied to both repository and global view modes?
**Default if unknown:** Yes (both views have sessions that can be attached to)

## Q8: Should we keep the help text concise to fit on one terminal line?
**Default if unknown:** Yes (condensed help should remain readable on narrow terminals)

## Q9: Should we position "enter to attach" at the beginning of the help text for prominence?
**Default if unknown:** Yes (primary action should be most visible to users)

## Q10: Should we maintain consistency with the existing issueselect.go help text format?
**Default if unknown:** Yes (consistent UX patterns across the application)