# Issue #007 - Add --verbose flag to `ptx-agent app list`

**Type**: enhancement
**Priority**: low
**Status**: ✅ Implemented
**Branch**: `feature/007-app-list-verbose` (merged to main)
**Closed**: 2026-02-21
**Depends on**: Issue #001

## Feature Description

Add a `--verbose` (or `-v`) flag to `ptx-agent app list` that displays additional information about each top-level window beyond the current `[role] "name"` output.

## Current Behavior

```text
$ ptx-agent app list
[application] "Code"
[application] "Firefox"
[application] "Konsole"
```

## Expected Behavior with --verbose

```text
$ ptx-agent app list --verbose
[application] "Code"
  Interfaces: Accessible, Application
  Children: 3

[application] "Firefox"
  Interfaces: Accessible, Application
  Children: 5

[application] "Konsole"
  Interfaces: Accessible, Application
  Children: 2
```

The verbose output should include:

- List of accessibility interfaces the node implements
- Number of direct children
- Any additional useful metadata available from the accessibility backend

## Technical Requirements

- Add `--verbose` / `-v` flag to `appListCmd` in `src/internal/cli/app.go`
- Use existing `Backend.Interfaces()` and `Backend.Children()` methods to retrieve extra info
- Update the MCP tool `agent_app_list_windows` in `src/internal/mcp/handlers.go` to also support a `verbose` parameter
- Works on both Linux (AT-SPI2) and Windows (UIA)

## Acceptance Criteria

1. [ ] `ptx-agent app list` output remains unchanged (backward compatible)
2. [ ] `ptx-agent app list --verbose` shows additional details per window
3. [ ] MCP tool `agent_app_list_windows` accepts optional `verbose` parameter
4. [ ] Works on both Linux and Windows

## Complexity Estimate

Low — uses existing accessibility backend methods, only presentation layer change.
