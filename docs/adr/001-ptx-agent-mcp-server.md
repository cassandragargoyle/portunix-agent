# ADR-001: ptx-agent Provides Its Own MCP Server

## Status

Accepted

## Context

### Current Situation

The Portunix ecosystem has a central MCP server (`ptx-mcp`) that exposes system tools, package management, and container operations to AI assistants. The main Portunix project recently implemented a distributed MCP server ecosystem (see Portunix ADR-038) that allows plugins and helper binaries to provide their own standalone MCP servers.

`ptx-agent` is a helper binary that provides remote device control capabilities:

- **File browsing** — list and navigate files on the target machine
- **VS Code control** — read terminal content, navigate UI elements via accessibility APIs (AT-SPI2 on Linux, UIA on Windows)
- **Application control** — keyboard input simulation, window management, application switching

These capabilities are currently only accessible via the CLI (`ptx-agent files`, `ptx-agent vscode`, `ptx-agent app`). AI assistants cannot use them directly — they must invoke shell commands, which is less reliable, harder to discover, and bypasses the MCP tool schema validation.

### Motivation

With the distributed MCP ecosystem in place (Portunix ADR-038), `ptx-agent` should participate as a standalone MCP server. This enables AI assistants (Claude Code, Cursor, etc.) to directly invoke device control operations with structured input/output, proper error handling, and tool discoverability via `tools/list`.

### Why Standalone Mode (Not Extend)

`ptx-agent` is a separate binary with its own process lifecycle, platform-specific dependencies (AT-SPI2, UIA, uinput), and OS-level permissions. Running its tools inside the central MCP server process would:

- Require the central server to load platform-specific accessibility libraries
- Create a single point of failure — an AT-SPI2 crash would take down the central MCP server
- Prevent `ptx-agent` from running on a different machine than the central server (remote control use case)

The standalone mode aligns with `ptx-agent`'s design as an independently deployable helper binary.

## Decision

`ptx-agent` will provide its own MCP server using the **standalone** mode from the Portunix distributed MCP ecosystem (ADR-038).

### Command Interface

```bash
# Start MCP server in stdio mode (for AI assistant integration)
ptx-agent mcp serve

# Register with Claude Code
ptx-agent mcp configure

# Or register via central Portunix MCP
portunix mcp configure --plugin ptx-agent
```

### Exposed MCP Tools

Based on current `ptx-agent` capabilities, the following tools will be exposed:

| Tool | Description | Maps to CLI |
| ---- | ----------- | ----------- |
| `agent_files_list` | List files and directories at a given path | `ptx-agent files list` |
| `agent_vscode_status` | Get VS Code instance status and accessibility info | `ptx-agent vscode status` |
| `agent_vscode_terminal_read` | Read terminal content from VS Code | `ptx-agent vscode terminal read` |
| `agent_vscode_tree` | Read VS Code accessibility tree | `ptx-agent vscode tree` |
| `agent_app_type` | Simulate keyboard input in the focused application | `ptx-agent app type` |
| `agent_app_key` | Send key combinations to the focused application | `ptx-agent app key` |
| `agent_app_focus` | Focus a window by name or PID | `ptx-agent app focus` |
| `agent_app_list_windows` | List open windows | `ptx-agent app list-windows` |

Tool names use the `agent_` prefix to avoid collisions with central MCP tools when both are registered.

### Protocol

- **Transport**: stdio (newline-delimited JSON-RPC) — consistent with Portunix MCP conventions (ADR-005)
- **Protocol version**: MCP 1.0
- **Capabilities**: `tools` (no `resources` or `prompts` in initial implementation)

### Security

- Inherits `ptx-agent`'s existing security boundaries (Issue #002)
- MCP server respects the same access control configuration as CLI commands
- No additional network exposure — stdio mode only communicates with the parent process

## Consequences

### Positive

1. AI assistants can directly discover and invoke device control tools via MCP protocol
2. Structured input/output with JSON Schema validation replaces fragile CLI parsing
3. Aligns with Portunix distributed MCP ecosystem — `ptx-agent` becomes a first-class MCP participant
4. Independent lifecycle — `ptx-agent` MCP server can run on remote machines separately from central MCP
5. Tool discoverability — AI assistants see all available tools via `tools/list`

### Negative

1. Additional MCP server entry in AI assistant configuration (one more server to manage)
2. Implementation effort for MCP protocol handling in `ptx-agent`
3. Must keep MCP tool definitions in sync with CLI command changes

### Mitigation

- Registration is automated via `ptx-agent mcp configure` or `portunix mcp configure --plugin ptx-agent`
- MCP protocol implementation can reuse patterns from the central `ptx-mcp` server
- Tool definitions are generated from the same internal command structures as CLI commands

## Related Decisions

- **Portunix ADR-005**: Revert Default stdio Mode for MCP (MCP command conventions)
- **Portunix ADR-038**: Distributed MCP Server Ecosystem for Plugins (ecosystem architecture)

## Related Issues

- **portunix-agent Issue #001**: MVP ptx-agent helper for Portunix
- **portunix-agent Issue #002**: Security access control for ptx-agent
- **portunix-agent Issue #006**: Implement MCP server for ptx-agent (implementation issue)

## Date

2026-02-21

## Author

ZK
