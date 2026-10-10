# Issue #011 - Plugin packaging and deployment

**Type**: enhancement
**Priority**: high
**Status**: ✅ Implemented
**Branch**: feature/011-plugin-packaging (merged to main)
**Depends on**: Issue #001
**Related ADR**: ADR-003

## Problem

`ptx-agent` has no standardized packaging or deployment mechanism. Deploying it currently requires manually building the binary and placing it in `$PATH`. There is no `plugin.json` manifest, no `make dist` target, and no integration with the Portunix plugin system.

## Feature Description

Package `ptx-agent` as a proper Portunix helper-mode plugin so it can be installed, enabled, disabled, and removed via standard `portunix plugin` commands. The plugin is developed in a separate repository (`portunix-agent`) and should be registered as a reserved plugin name in `portunix-plugins`.

## Implementation Tasks

### 1. Create `plugin.json` manifest

Create `plugin.json` in the project root following the Portunix plugin manifest schema:

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
    "type": "mcp-server",
    "transport": "stdio"
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

### 2. Add `make dist` target to Makefile

Add a `dist` target that creates a distribution directory:

```makefile
DIST_DIR := dist

dist: build
	@echo Creating distribution package...
	mkdir -p $(DIST_DIR)
	cp $(BINARY) $(DIST_DIR)/
	cp plugin.json $(DIST_DIR)/
	@echo Distribution package created in $(DIST_DIR)/
```

The target must:
- Build the binary first (depends on `build`)
- Create `dist/` directory
- Copy the binary and `plugin.json` into `dist/`
- Work on both Linux and Windows

### 3. Add `dist` to `.PHONY` and `clean`

- Add `dist` to the `.PHONY` declaration
- Add `dist/` directory removal to the `clean` target

### 4. Add `dist/` to `.gitignore`

Ensure `dist/` is not tracked in version control.

### 5. Register reserved plugin name

Add `ptx-agent` to the reserved plugin names in `portunix-plugins/registry/plugin-index.json` (separate PR in `portunix-plugins` repo):

```
- `ptx-agent` — separate repo (portunix-agent)
```

## Acceptance Criteria

1. [ ] `plugin.json` exists in project root with correct Portunix plugin schema fields
2. [ ] `plugin.json` declares `type: "helper"`, `runtime: "native"`, `mode: "helper"`
3. [ ] `plugin.json` declares `ai_integration.type: "mcp-server"` with `transport: "stdio"`
4. [ ] `plugin.json` includes `commands` array matching CLI command structure
5. [ ] `make dist` builds the binary and creates `dist/` with binary + manifest
6. [ ] `make clean` removes the `dist/` directory
7. [ ] `portunix plugin install dist/plugin.json` installs the plugin
8. [ ] `portunix plugin enable ptx-agent` enables the plugin
9. [ ] `portunix plugin list` shows ptx-agent with correct version
10. [ ] Plugin dispatch works: `portunix ptx-agent files list /tmp`
11. [ ] Works on Linux and Windows

## Complexity Estimate

Low — mostly build system additions and a static manifest file. No changes to application code.
