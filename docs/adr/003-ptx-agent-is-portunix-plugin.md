# ADR-003: ptx-agent Is a Portunix Plugin

## Status

Accepted

## Context

### Current Situation

`ptx-agent` is a standalone Go binary that provides device control capabilities (file browsing, VS Code control, keyboard simulation, window management, browser automation). It was originally designed as a "helper binary" for the Portunix ecosystem without a formal packaging and deployment mechanism.

Currently, deploying `ptx-agent` means manually building the binary (`make build`) and placing it somewhere in `$PATH`. There is no standardized way to:

- Install/uninstall `ptx-agent` as part of the Portunix system
- Track its version alongside other Portunix components
- Register its MCP server automatically during installation
- Enable/disable its functionality from the Portunix CLI

### Portunix Plugin System

Portunix provides a plugin system that manages the lifecycle of extension components. Plugins communicate with Portunix Core via two modes:

- **Helper mode** (`plugin.type: "helper"`) — lightweight CLI tools dispatched via stdio/argv as subprocesses
- **gRPC service mode** (`plugin.type: "grpc"`) — long-running daemons communicating via gRPC on ports 9000-9999

Plugins are registered in `~/.portunix/plugins/registry.json` and managed via `portunix plugin` commands:

```bash
portunix plugin install <manifest-path>   # Install from plugin.json
portunix plugin enable <name>             # Expose commands in CLI
portunix plugin disable <name>            # Hide commands from CLI
portunix plugin list                      # List installed plugins
portunix plugin remove <name>             # Uninstall
```

### Plugin Sources

Plugins can live in the bundled `portunix-plugins` repository or in **separate repositories**. The `portunix-plugins` repo maintains a list of reserved plugin names for separately-developed plugins (e.g., `ptx-codeanalysis`, `ptx-codeinfo`, `ptx-crawler`). From the dispatcher's perspective, source location makes no difference — once installed and registered, all plugins are dispatched the same way.

### Distributed MCP Ecosystem

Portunix ADR-038 introduced a distributed MCP server ecosystem where plugins can provide their own standalone MCP servers. `ptx-agent` already implements an MCP server (ADR-001) but registration currently requires manual configuration. The plugin manifest's `ai_integration` field automates MCP tool registration.

## Decision

`ptx-agent` will be formally packaged and deployed as a **Portunix helper-mode plugin**, developed in its own separate repository (`portunix-agent`).

### Plugin Type: Helper

`ptx-agent` operates as a **helper-mode plugin** (not gRPC service mode) because:

- It is invoked as a subprocess by Portunix CLI (`portunix ptx-agent <command> [args]`)
- Each invocation is a short-lived process with stdio passthrough
- No long-running daemon or persistent gRPC connection needed for CLI dispatch
- The MCP server (`ptx-agent mcp serve`) is a separate concern — it runs independently when started by an AI assistant, not via the plugin dispatcher

### Plugin Manifest

The project will include a `plugin.json` manifest following the Portunix plugin schema:

```json
{
  "name": "ptx-agent",
  "version": "0.1.0",
  "description": "Remote device control agent — file browsing, accessibility, keyboard, browser automation",
  "author": "CassandraGargoyle",
  "plugin": {
    "type": "helper",
    "binary": "ptx-agent",
    "runtime": "native",
    "mode": "helper"
  },
  "dependencies": {
    "portunix_min_version": "1.10.0",
    "os_support": ["linux", "windows"]
  },
  "ai_integration": {
    "mcp_tools": [
      {"name": "agent_files_list", "description": "List files and directories at a given path"},
      {"name": "agent_vscode_status", "description": "Get VS Code accessibility status"},
      {"name": "agent_vscode_terminal_read", "description": "Read VS Code terminal content"},
      {"name": "agent_vscode_tree", "description": "Dump VS Code accessibility tree"},
      {"name": "agent_app_type", "description": "Simulate keyboard text input"},
      {"name": "agent_app_key", "description": "Send key combinations to focused application"},
      {"name": "agent_app_focus", "description": "Focus a window by name"},
      {"name": "agent_app_list_windows", "description": "List open desktop windows"},
      {"name": "agent_browser_launch", "description": "Launch browser for web automation"},
      {"name": "agent_browser_goto", "description": "Navigate browser to URL"},
      {"name": "agent_browser_click", "description": "Click interactive element by index"},
      {"name": "agent_browser_type", "description": "Type text into form element"},
      {"name": "agent_browser_screenshot", "description": "Capture page screenshot"},
      {"name": "agent_browser_content", "description": "Extract text/HTML from page"},
      {"name": "agent_browser_elements", "description": "List interactive page elements"},
      {"name": "agent_browser_close", "description": "Close browser instance"}
    ]
  },
  "permissions": {
    "level": "full",
    "filesystem": true,
    "network": true,
    "system": true
  },
  "commands": [
    {"name": "files", "description": "File browsing on target machine", "subcommands": ["list"]},
    {"name": "vscode", "description": "VS Code control via accessibility APIs", "subcommands": ["status", "terminal", "tree"]},
    {"name": "app", "description": "Application control — keyboard, windows", "subcommands": ["type", "key", "focus", "list-windows"]},
    {"name": "mcp", "description": "MCP server for AI assistant integration", "subcommands": ["serve", "configure"]},
    {"name": "browser", "description": "Browser automation via Playwright", "subcommands": ["launch", "close"]}
  ]
}
```

### Two-Level Dispatch

When installed as a plugin, `ptx-agent` participates in the standard two-level dispatch:

```text
portunix ptx-agent files list /home
    │
    ├─ Level 1: Core Helper Dispatch (dispatcher.go)
    │  Is "ptx-agent" in registerHelpers()? → NO
    │
    └─ Level 2: Plugin Command Dispatch (plugin_dispatcher.go)
       Is "ptx-agent" enabled in registry.json? → YES
       → exec ptx-agent files list /home (native runtime, stdio passthrough)
```

### Build Target

The `Makefile` will include a `dist` target:

```makefile
dist: build
	mkdir -p dist
	cp $(BINARY) dist/
	cp plugin.json dist/
```

### Deployment Flow

```text
make dist → dist/ptx-agent + dist/plugin.json
         → portunix plugin install dist/plugin.json
         → portunix plugin enable ptx-agent
```

### Reserved Plugin Name

`ptx-agent` should be added to the reserved plugin names in `portunix-plugins` repository, alongside `ptx-codeanalysis`, `ptx-codeinfo`, `ptx-crawler`, and `ptx-query`.

### What Does NOT Change

- `ptx-agent` remains a standalone binary — it can still run independently without Portunix
- CLI interface stays the same (`ptx-agent files`, `ptx-agent app`, etc.)
- MCP server capabilities are unchanged (ADR-001)
- The binary is not loaded into the Portunix process — it runs as a separate subprocess

## Consequences

### Positive

1. Standardized lifecycle management via `portunix plugin install/enable/disable/remove`
2. Dynamic command registration — `portunix ptx-agent` works without modifying `dispatcher.go`
3. MCP tools declared in manifest's `ai_integration` field for automatic discovery
4. Plugin can be enabled/disabled without rebuilding Portunix
5. Consistent with other separately-developed plugins (ptx-codeanalysis, ptx-codeinfo)
6. Version and permission tracking in `~/.portunix/plugins/registry.json`

### Negative

1. Requires a `plugin.json` manifest to be maintained in sync with project changes
2. `make dist` becomes the required build target for deployment (not just `make build`)
3. Deployment depends on a running Portunix installation
4. `ptx-agent` requires `full` permission level due to system-level operations (keyboard, windows)

### Mitigation

- `plugin.json` is mostly static — version can be updated via build scripts
- `make dist` delegates to `make build` — no duplication of build logic
- `ptx-agent` remains usable standalone — Portunix plugin packaging is optional
- Full permission level is justified by the nature of device control operations

## Related Decisions

- **ADR-001**: ptx-agent Provides Its Own MCP Server
- **ADR-002**: Browser Control via Playwright
- **Portunix ADR-038**: Distributed MCP Server Ecosystem for Plugins

## Related Issues

- **portunix-agent Issue #011**: Plugin packaging and deployment (implementation issue)

## Date

2026-02-22

## Author

ZK
