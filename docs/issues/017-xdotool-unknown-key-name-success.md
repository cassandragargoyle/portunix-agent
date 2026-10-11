# Issue #17: xdotool backend reports success for an unknown key name

**Status**: Open
**Priority**: Low
**Type**: Bug
**Created**: 2026-10-11
**GitHub**: #17
**Related**: #14, portunix-archon #1

## Summary

With the xdotool backend, a key combination with an unknown key name returns success. Only the
valid part of the combination is sent.

## Problem Description

Found during the acceptance test of portunix-archon issue #1 (Ubuntu 26.04, X11, xdotool
backend).

```bash
ptx-agent keyboard key ctrl+nosuchkey
```

- The command reports `Sent key combination: ctrl+nosuchkey` and exits with 0.
- The focused window receives only `Control_L`.
- The uinput backend handles the same input correctly. It returns
  `unknown key "nosuchkey" in combination "ctrl+nosuchkey"`.

Cause: `XdotoolBackend.Key` in `src/pkg/keyboard/backend_xdotool.go` checks only the exit code
of `xdotool key`. xdotool prints a warning for an unknown key name, ignores that name, sends
the rest, and exits with 0.

The two backends behave differently for the same input. A caller cannot detect a typing
mistake in a key combination when xdotool is active.

## Proposed Solution

Validate the combination before xdotool is called. Options:

- Validate every part of the combination against the known key names (the same check that the
  uinput backend does in `comboToKeys`), or
- capture the stderr of xdotool and return an error when it reports an unknown key name.

The first option gives the same error message for both backends and sends no key at all for a
bad combination.

## Acceptance Criteria

1. `Key` with an unknown key name returns an error with the xdotool backend.
2. No key is sent when the combination is not valid.
3. The error message has the same form for both backends.
4. Valid combinations (`ctrl+s`, `alt+Tab`, `Return`) work as before.
5. Unit tests cover the validation.

## Testing

- **Unit tests**: validation of valid and invalid combinations. No X server is needed.
- **Manual test on Linux (X11)**: without access to `/dev/uinput`, run
  `ptx-agent keyboard key ctrl+nosuchkey` and check the error and the exit code. Run
  `ptx-agent keyboard key ctrl+s` and check that it still works.
