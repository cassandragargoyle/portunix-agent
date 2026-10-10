# Issue #005 - VS Code Extension: Terminal content API for ptx-agent

**Type**: enhancement
**Priority**: medium
**Status**: ✅ Implemented
**GitHub**: #5
**Depends on**: Issue #003, Issue #004
**Assigned to**: Portunix VS Code Extension team

> **Note**: This issue describes a feature request for the **Portunix VS Code Extension** (separate repository). It is tracked here because ptx-agent is the consumer of this API. Implementation will be done by the VS Code Extension team.

## Problem

The current terminal reading in ptx-agent (Issues #003, #004) uses the OS accessibility layer (AT-SPI2 on Linux, UIA on Windows) to read VS Code terminal content. This works but has a fundamental limitation:

**xterm.js accessibility only exposes rows currently visible in the terminal viewport.** The scrollback buffer (command history, previous output) is not accessible via the accessibility API. This is a limitation of how xterm.js implements the accessibility tree — it mirrors the visual viewport, not the full terminal buffer.

In practice this means:

- `ptx-agent vscode terminal read` returns ~20-40 lines (one screen)
- Command output that scrolled off-screen is lost
- Long-running commands with extensive output cannot be fully captured
- The `…` character appears at truncated lines

## Proposed Solution

The Portunix VS Code Extension should expose a **local HTTP or IPC endpoint** that ptx-agent can query to get full terminal content, bypassing the accessibility layer limitation.

### VS Code Extension API capabilities

VS Code provides `vscode.window.terminals` API which gives access to:

- List of all terminal instances
- Terminal names and process IDs
- `Terminal.shellIntegration` (VS Code 1.93+) for command detection
- `window.onDidWriteTerminalData` event for real-time terminal output capture

### Proposed API endpoints

The extension should expose (via local HTTP server or named pipe):

```text
GET /api/terminals
  → [{ "id": 1, "name": "Terminal 1", "pid": 12345 }]

GET /api/terminals/{id}/content
  → { "lines": ["$ ls", "file1.txt", "file2.txt", "$ "], "lineCount": 4 }

GET /api/terminals/{id}/content?since={lineOffset}
  → { "lines": [...], "lineCount": N }  (incremental read)
```

### Data capture strategy

The extension should:

1. Listen to `onDidWriteTerminalData` for each terminal
2. Buffer the output internally (configurable max size, e.g. 10,000 lines)
3. Serve the buffered content via the local API
4. Optionally parse shell integration sequences to identify command boundaries

### Integration with ptx-agent

ptx-agent should:

1. **Try the extension API first** — if the Portunix extension is installed and running, use it for full terminal content
2. **Fall back to accessibility** — if the extension is not available, use the current AT-SPI2/UIA approach (visible rows only)
3. Report which method was used in `vscode status` output

## Scope for the VS Code Extension team

### Must have

- Capture terminal output via `onDidWriteTerminalData`
- Buffer output per terminal (configurable limit)
- Expose terminal list and content via local HTTP endpoint (localhost only)
- Authentication token to prevent unauthorized access

### Nice to have

- Shell integration support (command boundaries, exit codes)
- Named pipe transport as alternative to HTTP
- WebSocket endpoint for real-time streaming
- Configuration UI in VS Code settings

## Scope for ptx-agent (this repository)

Once the extension API is available, ptx-agent will need:

- HTTP client for the extension API
- Fallback logic (extension → accessibility)
- Updated `vscode status` to show data source

This ptx-agent work will be tracked in a separate issue.

## Acceptance Criteria

1. Extension captures full terminal output (not just visible lines)
2. ptx-agent can read terminal content that has scrolled off-screen
3. Latency for reading terminal content is under 100ms
4. Extension does not noticeably affect VS Code performance
5. Access is restricted to localhost with authentication token
