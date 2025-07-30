# Initial Request

**Date:** 2025-07-30 19:24
**Request:** When running `sbs start` with no argument, we should fetch the list of github issues with the `gh issues` command. We can then present those to the user so that they can select. The user should be able to cancel out of the list and do nothing.

## Context
This request is for enhancing the existing `sbs start` command to provide an interactive issue selection workflow when no specific issue number is provided as an argument. Currently `sbs start` requires an issue number - this would make it optional and interactive.