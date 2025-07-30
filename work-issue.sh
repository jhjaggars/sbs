#!/bin/bash

set_title() {
    if [ -n "$TMUX" ]; then
        # We're in tmux - use tmux command
        tmux rename-window "$1"
    else
        # Regular terminal - use ANSI escape sequence
        printf '\033]0;%s\007' "$1"
    fi
}

TITLE=(gh issue view $1 --json title -q .title)

set_title "✻ $TITLE"

sandbox \
  --net="host" \
  --bind /tmp/tmux-1000 \
  --name "work-issue-$1" \
  ~/.claude/local/claude \
  --model sonnet \
  --dangerously-skip-permissions \
  "/work-issue $1"
