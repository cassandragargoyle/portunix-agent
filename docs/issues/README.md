# Issue Tracker — portunix-agent

## Overview

Issues are tracked as Markdown files in this directory. For workflow, labels, and templates see [ISSUE-MANAGEMENT.md](../contributing/ISSUE-MANAGEMENT.md).

### Directory Structure

- `internal/` — internal issue specifications (detailed design, architecture decisions)
- `public/` — issues intended for sync with external trackers

### External Repository

- **GitHub**: <https://github.com/CassandraGargoyle/portunix-agent>

## Issue Index

### Active

| # | Title | Type | Priority | Status | Branch | External |
|---|-------|------|----------|--------|--------|----------|
| [001](internal/001-mvp-ptx-agent.md) | MVP: ptx-agent helper for Portunix | enhancement | high | needs-triage | `feature/issue-001-mvp-ptx-agent` | — |
| [002](internal/002-security-access-control.md) | Security: access control for ptx-agent | security | high | needs-triage | TBD | — |
| [012](internal/012-implement-help-ai-flag.md) | Implement --help-ai flag | enhancement | medium | Open | TBD | — |
| [014](internal/014-public-keyboard-package.md) | Public keyboard package with a persistent device session | enhancement | high | Open | `feature/014-public-keyboard-package` | — |
| [015](internal/015-version-flag.md) | Implement --version flag | enhancement | medium | Open | `feature/015-version-flag` | — |

### Done (Archived)

Completed issues archived under `internal/done/`.

| # | Title | Type | Priority | Status | Branch | External |
|---|-------|------|----------|--------|--------|----------|
| [003](internal/done/003-mvp-vscode-accessibility-control.md) | MVP: VS Code control via Chromium Accessibility (AT-SPI2) | enhancement | high | ✅ Implemented | `feature/003-vscode-accessibility-control` | — |
| [004](internal/done/004-windows-vscode-accessibility-uia.md) | Windows: VS Code control via UI Automation (UIA) | enhancement | high | ✅ Implemented | `feature/004-windows-uia-accessibility` | — |
| [005](internal/done/005-vscode-extension-terminal-api.md) | VS Code Extension: Terminal content API for ptx-agent | enhancement | medium | ✅ Implemented | — | — |
| [006](internal/done/006-ptx-agent-mcp-server.md) | Implement MCP Server for ptx-agent | enhancement | high | ✅ Implemented | `feature/006-mcp-server` | — |
| [007](internal/done/007-app-list-verbose.md) | Add --verbose flag to `ptx-agent app list` | enhancement | low | ✅ Implemented | `feature/007-app-list-verbose` | — |
| [008](internal/done/008-browser-control-playwright.md) | Browser control via Playwright | enhancement | high | ✅ Implemented | `feature/008-browser-control-playwright` | — |
| [009](internal/done/009-goal-based-product-search.md) | Goal-Based Product Search on E-Shop | enhancement | medium | ✅ Implemented | `feature/010-mcp-http-transport` | — |
| [010](internal/done/010-mcp-http-transport.md) | MCP Server: HTTP transport with token authentication | enhancement | high | ✅ Implemented | `feature/010-mcp-http-transport` | — |
| [011](internal/done/011-plugin-packaging-deployment.md) | Plugin packaging and deployment | enhancement | high | ✅ Implemented | `feature/011-plugin-packaging` | — |
| [013](internal/done/013-browser-bot-detection-persistent-profile.md) | Browser: bot detection blocks login (persistent profile + stealth) | enhancement | high | ✅ Implemented | `feature/013-browser-persistent-profile` | — |

## Dependencies

```text
#003 depends on #001
#004 depends on #003
#005 depends on #003, #004
#006 depends on #001
#007 depends on #001
#008 depends on #001, #006
#009 depends on #008, #010
#010 depends on #006
#002 relates to #001
#011 depends on #001
#012 relates to #015
```
