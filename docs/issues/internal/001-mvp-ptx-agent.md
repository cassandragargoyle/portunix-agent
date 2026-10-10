# Issue #001 - MVP: ptx-agent helper for Portunix

**Type**: enhancement
**Priority**: high
**Status**: needs-triage
**Branch**: `feature/issue-001-mvp-ptx-agent`

## Feature Description

Create an MVP of `ptx-agent` — a Go helper binary for the Portunix system that enables remote device control on machines where it is installed. The agent follows the established Portunix dispatcher architecture (`ptx-*` helper pattern, see ADR-014).

## Use Case

Portunix needs a lightweight agent running on target machines that allows remote control of the desktop environment. This is essential for scenarios such as remote assistance, automated workflows, and cross-platform device management. The agent acts as the "hands" of the Portunix system on the target device.

## Scope (MVP)

### 1. Application switching

- List running applications/windows
- Search for an application by name or title
- Switch focus to a specified application window

### 2. Keyboard input

- Send keystrokes to the currently focused application
- Support key combinations (Ctrl+C, Alt+Tab, etc.)
- Support text string input with configurable typing speed

### 3. File browsing

- List files and directories within allowed (whitelisted) paths
- Navigate directory structure (up/down)
- Return file metadata (name, size, modified date)
- Security: enforce allowed directory whitelist, deny access outside configured paths

## Technical Requirements

- **Language**: Go (consistent with Portunix ecosystem)
- **Binary name**: `ptx-agent`
- **CLI framework**: `github.com/spf13/cobra` (consistent with other ptx-* helpers)
- **Dispatcher integration**: register commands with main `portunix` dispatcher
- **Platforms**: Linux (primary MVP), Windows (secondary)

### Linux implementation

- Use `xdotool` for window management and keyboard simulation (similar to `presentation_toolkit` xdotool backend)
- Fallback detection when xdotool is unavailable

### Windows implementation (post-MVP or parallel)

- Use Windows API (user32.dll) for window enumeration and keyboard input
- Use `FindWindow`, `SetForegroundWindow`, `SendInput` APIs

### Inspiration

- Architecture pattern: existing `ptx-*` helpers in Portunix (ptx-container, ptx-installer, etc.)
- Keyboard/app control: `tools/presentation_toolkit` from portunix-marketing (Python/xdotool backend)

## Proposed Commands

```bash
portunix agent windows list              # list open windows
portunix agent windows find <query>      # search window by name
portunix agent windows focus <window-id> # switch to window
portunix agent keyboard type <text>      # type text string
portunix agent keyboard key <combo>      # send key combination
portunix agent files list <path>         # list directory contents
portunix agent files info <path>         # file metadata
```

## Security Considerations

- File browsing restricted to configured allowed directories only
- No file content read/write in MVP (listing and metadata only)
- Agent should validate and sanitize all input paths
- Consider authentication mechanism for remote command acceptance (post-MVP)

## Acceptance Criteria

- [ ] `ptx-agent` binary builds on Linux
- [ ] Integrates with Portunix dispatcher (`portunix agent ...` commands work)
- [ ] Can list and search open windows on Linux (X11/xdotool)
- [ ] Can switch focus to a specific window
- [ ] Can send keyboard text input to focused application
- [ ] Can send key combinations (e.g. Ctrl+C)
- [ ] Can list files in allowed directories
- [ ] Denies access to paths outside allowed directories
- [ ] Graceful error handling when xdotool is not installed
- [ ] Basic unit tests for core functionality
- [ ] Command help text for all subcommands

## Out of Scope (MVP)

- Mouse control
- Screen capture / streaming
- File content read/write
- Remote network communication protocol
- Authentication / authorization
- Windows platform support (tracked separately)
- Wayland support (X11 only for MVP)

## Complexity Estimate

Medium — established patterns exist in Portunix helpers and presentation_toolkit. Main effort is xdotool Go wrapper and file access security layer.

---

**Created**: 2026-02-16
**Author**: Architect
