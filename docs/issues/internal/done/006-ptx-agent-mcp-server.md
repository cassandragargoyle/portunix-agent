# Issue #006 - Implement MCP Server for ptx-agent

**Type**: enhancement
**Priority**: high
**Status**: ✅ Implemented
**Branch**: `feature/006-mcp-server` (merged to main)
**Closed**: 2026-02-21
**Depends on**: Issue #001
**ADR**: [ADR-001](../../../adr/001-ptx-agent-mcp-server.md)

## Feature Description

Implement a standalone MCP (Model Context Protocol) server in `ptx-agent` that exposes device control capabilities to AI assistants (Claude Code, Cursor, etc.) via the standard MCP protocol.

This follows the Portunix distributed MCP ecosystem architecture (Portunix ADR-038) where helper binaries provide their own MCP servers in **standalone** mode.

## Use Case

AI assistants currently cannot directly interact with `ptx-agent` capabilities. They must shell out to CLI commands, which is:

- **Fragile** — output parsing breaks on format changes
- **Undiscoverable** — AI assistants don't know what tools are available without documentation
- **Unvalidated** — no input schema validation before execution
- **Inconsistent** — different error formats between CLI and MCP

With an MCP server, AI assistants can:

1. Discover available tools via `tools/list`
2. Call tools with validated JSON input
3. Receive structured JSON responses
4. Integrate seamlessly with the Portunix MCP ecosystem

## Technical Requirements

### MCP Protocol Implementation

- Implement MCP 1.0 protocol over stdio (newline-delimited JSON-RPC)
- Handle standard MCP methods: `initialize`, `initialized`, `tools/list`, `tools/call`
- Return proper JSON-RPC error responses for invalid requests
- Graceful shutdown on SIGINT/SIGTERM

### Command Structure

```bash
# New cobra commands
ptx-agent mcp serve     # Start MCP server (stdio mode)
ptx-agent mcp configure # Register with Claude Code
ptx-agent mcp remove    # Unregister from Claude Code
ptx-agent mcp status    # Check registration status
```

### MCP Tools to Expose

Each tool must have a complete JSON Schema `inputSchema`:

| Tool | Description | Parameters |
| ---- | ----------- | ---------- |
| `agent_files_list` | List files and directories | `path` (string, required) |
| `agent_vscode_status` | Get VS Code accessibility status | — |
| `agent_vscode_terminal_read` | Read VS Code terminal content | `terminal_id` (int, optional) |
| `agent_vscode_tree` | Read accessibility tree | `depth` (int, optional) |
| `agent_app_type` | Simulate keyboard text input | `text` (string, required) |
| `agent_app_key` | Send key combination | `keys` (string, required), e.g. `"ctrl+c"` |
| `agent_app_focus` | Focus a window | `name` (string) or `pid` (int) |
| `agent_app_list_windows` | List open windows | — |

### Architecture

```text
┌──────────────────────┐
│  AI Assistant        │
│  (Claude Code)       │
│                      │
│  MCP Server:         │
│  └── ptx-agent       │
└──────┬───────────────┘
       │ stdio (JSON-RPC)
       ▼
┌──────────────────────────────────────┐
│  ptx-agent mcp serve                 │
│                                      │
│  ┌──────────────┐  ┌──────────────┐  │
│  │ MCP Handler  │  │ Tool Router  │  │
│  │ (JSON-RPC)   │──│              │  │
│  └──────────────┘  └──────┬───────┘  │
│                           │          │
│  ┌────────┬───────┬───────┴───────┐  │
│  │ files  │vscode │ app (kbd/win) │  │
│  │ pkg    │ pkg   │ pkg           │  │
│  └────────┴───────┴───────────────┘  │
└──────────────────────────────────────┘
```

### Implementation Approach

1. **New package `src/internal/mcp/`** — MCP server implementation
   - `server.go` — MCP server struct, stdio loop, JSON-RPC dispatch
   - `handlers.go` — tool definitions and handler functions
   - `tools.go` — tool registry with input schemas

2. **New CLI commands `src/internal/cli/mcp.go`** — cobra commands for `mcp serve`, `mcp configure`, `mcp remove`, `mcp status`

3. **Tool handlers** reuse existing internal packages (`files`, `vscode`, `keyboard`, `windows`) — no duplication of business logic

### Claude Code Registration

`ptx-agent mcp configure` should:

1. Detect Claude Code installation (check for `~/.claude/` or `claude` binary)
2. Register via `claude mcp add ptx-agent -- /path/to/ptx-agent mcp serve`
3. Confirm registration to user

## Security Considerations

- MCP server inherits the same access control as CLI commands (Issue #002)
- stdio mode only — no network listeners, no additional attack surface
- Tool execution respects existing security boundaries
- No credentials or secrets exposed via MCP tools

## Acceptance Criteria

1. [ ] `ptx-agent mcp serve` starts an MCP server that handles `initialize` and `tools/list`
2. [ ] All existing CLI capabilities are exposed as MCP tools with proper input schemas
3. [ ] `tools/call` correctly invokes internal packages and returns structured results
4. [ ] `ptx-agent mcp configure` registers the server with Claude Code
5. [ ] `ptx-agent mcp remove` unregisters the server from Claude Code
6. [ ] Error responses follow JSON-RPC 2.0 error format
7. [ ] Graceful shutdown on SIGINT/SIGTERM
8. [ ] Works on both Linux and Windows

## Out of Scope

- TCP/WebSocket transport modes (stdio only for initial implementation)
- MCP `resources` or `prompts` capabilities
- Plugin manifest integration (handled by Portunix core, not ptx-agent)
- Authentication/authorization beyond existing access control (Issue #002)

## Complexity Estimate

Medium — MCP protocol is well-defined, and tool handlers delegate to existing packages. Main effort is the JSON-RPC stdio loop and tool schema definitions.

## Implementation Notes

- Reference implementation: `src/app/mcp/` in the main Portunix project (`ptx-mcp`)
- MCP specification: <https://modelcontextprotocol.io>
- Reuse the same JSON-RPC patterns as `ptx-mcp` for consistency across the ecosystem
- Consider using a shared MCP library if one is extracted from `ptx-mcp` in the future
