# #012 - Implement --help-ai Flag

**Status**: ✅ Implemented
**GitHub**: #12
**Priority**: Medium
**Type**: Enhancement
**Labels**: enhancement, cli, ai-integration

## Context

All Portunix ecosystem CLIs should provide a `--help-ai` flag that outputs
machine-readable JSON describing the tool's commands, parameters, and capabilities.
This enables AI assistants (Claude Code, MCP clients) to discover and use the CLI
programmatically without parsing human-readable help text.

The pattern is already implemented in:
- **portunix** (Go/Cobra) — `src/cmd/help_registry.go` → `GenerateAIHelp()`
- **portunix-reco** (Python/argparse) — `cli/help_ai.py` → `print_help_ai()`
- **portunix-plugins/docgen** (Go/Cobra) — `cmd/docgen/main.go` → `printHelpAI()`

## Requirements

### 1. Add `--help-ai` Flag to Root Command

Register a persistent boolean flag on the root command in `src/internal/cli/root.go`:

```go
rootCmd.PersistentFlags().BoolVar(&helpAI, "help-ai", false,
    "Show machine-readable help in JSON format")
```

### 2. Create `help_ai.go` in `src/internal/cli/`

Add a `printHelpAI()` function that generates the JSON output. The root command
should check the flag and call this function before subcommand dispatch.

### 3. JSON Output Structure

Output must follow the ecosystem-standard schema. ptx-agent has a rich command
tree with nested subcommands:

```json
{
  "tool": "ptx-agent",
  "version": "<version>",
  "description": "Portunix agent for remote device control",
  "commands": [
    {
      "name": "files",
      "brief": "Browse files and directories",
      "subcommands": [
        {"name": "list", "brief": "List directory contents", "parameters": [
          {"name": "path", "type": "string", "required": true, "description": "Directory path to list"}
        ]},
        {"name": "info", "brief": "Show file metadata", "parameters": [
          {"name": "path", "type": "string", "required": true, "description": "File path"}
        ]}
      ]
    },
    {
      "name": "vscode",
      "brief": "Control VS Code via accessibility",
      "subcommands": [
        {"name": "terminal read", "brief": "Read terminal content"},
        {"name": "terminal list", "brief": "List open terminals"},
        {"name": "status", "brief": "Show VS Code accessibility status"},
        {"name": "tree", "brief": "Dump accessibility tree (debug)"}
      ]
    },
    {
      "name": "app",
      "brief": "Interact with desktop applications via accessibility",
      "subcommands": [
        {"name": "list", "brief": "List desktop windows"},
        {"name": "tree", "brief": "Dump accessibility tree of an application"},
        {"name": "read", "brief": "Read text content from an application"}
      ]
    },
    {
      "name": "mcp",
      "brief": "MCP server for AI assistant integration",
      "subcommands": [
        {"name": "serve", "brief": "Start MCP server"},
        {"name": "configure", "brief": "Register with Claude Code"},
        {"name": "remove", "brief": "Unregister from Claude Code"},
        {"name": "status", "brief": "Check MCP registration status"}
      ]
    },
    {
      "name": "browser",
      "brief": "Control a web browser via Playwright",
      "subcommands": [
        {"name": "install", "brief": "Install Playwright browser binaries"},
        {"name": "launch", "brief": "Launch a browser instance"},
        {"name": "goto", "brief": "Navigate to a URL"},
        {"name": "click", "brief": "Click an element by its index"},
        {"name": "type", "brief": "Type text into an element"},
        {"name": "elements", "brief": "List interactive elements with indices"},
        {"name": "screenshot", "brief": "Capture a screenshot"},
        {"name": "content", "brief": "Extract page text content"},
        {"name": "tabs", "brief": "List open tabs"}
      ]
    }
  ]
}
```

Include full parameter details for all subcommands (flags, types, defaults, descriptions).

### 4. Framework

- **Language**: Go
- **CLI framework**: Cobra
- **CLI package**: `src/internal/cli/`
- **Entry point**: `src/cmd/ptx-agent/main.go` → `cli.Execute()`

## Acceptance Criteria

- [ ] `ptx-agent --help-ai` outputs valid JSON to stdout
- [ ] JSON contains `tool`, `version`, `description`, `commands` fields
- [ ] All subcommands (including nested: files, vscode, app, mcp, browser) are listed
- [ ] Parameters include names, types, required flag, and descriptions
- [ ] Exit code is 0
- [ ] No other output (logs, banners) is mixed into the JSON

## References

- portunix `src/cmd/help_registry.go` — Go/Cobra reference with nested commands
- portunix-plugins/docgen `cmd/docgen/main.go` — lightweight Go pattern

---

**Created**: 2026-03-26
**Closed**: 2026-10-11 (implemented in `5d2c99f`)
**Author**: Architect
