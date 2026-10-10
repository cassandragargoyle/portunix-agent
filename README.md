# portunix-agent

Plugin for the [Portunix](https://github.com/CassandraGargoyle/portunix) system that provides remote control of devices where it is installed. It enables keyboard input simulation, window management, application inspection via accessibility APIs, VS Code terminal reading, web browser automation via Playwright, and secure file browsing.

All capabilities are exposed both as CLI commands and as [MCP](https://modelcontextprotocol.io/) tools, allowing AI assistants to control the desktop, interact with applications, and automate web browsers on the target machine. The agent covers two complementary surfaces: **desktop applications** (via platform-native accessibility APIs) and **web applications** (via Playwright browser automation), giving AI assistants a unified way to interact with any software running on the device.

## Features

### Application Control via Accessibility APIs

Inspect any accessible application on the desktop using platform-native accessibility technologies (AT-SPI2 on Linux, UI Automation on Windows).

```bash
ptx-agent app list                      # List desktop windows
ptx-agent app tree <name> [--depth N]   # Dump accessibility tree
ptx-agent app read <name> [--depth N]   # Read text content from application
```

### VS Code Terminal Reading

Read terminal content from VS Code by navigating its accessibility tree to the xterm.js terminal elements.

```bash
ptx-agent vscode status                 # Check VS Code accessibility status
ptx-agent vscode terminal list          # List open terminals
ptx-agent vscode terminal read [--index N]  # Read terminal content
ptx-agent vscode tree                   # Dump VS Code accessibility tree
```

Requires `--force-renderer-accessibility` flag or `editor.accessibilitySupport: "on"` in VS Code settings on Linux. Windows typically auto-enables accessibility when a UIA client connects.

### Keyboard Input Simulation (Linux)

Type text, send key combos, or paste from clipboard with support for both Wayland and X11.

```bash
ptx-agent keyboard type <text> [--delay MS] [--no-indent] [--window NAME]
ptx-agent keyboard key <combo> [--window NAME]
ptx-agent keyboard paste <text> [--window NAME]
ptx-agent keyboard read                 # Read clipboard content
```

Backends: **uinput** (Wayland + X11) with xdotool fallback (X11 only). The `--window` flag focuses the target window before input. The `paste` command uses the clipboard, making it layout-independent (works with Czech, German, and other keyboard layouts).

### Window Management (Linux)

List, search, and focus windows on the desktop.

```bash
ptx-agent windows list                  # List all windows
ptx-agent windows find <query>          # Search windows by name
ptx-agent windows focus <window-id>     # Activate/focus a window
```

Backends: **KWin** (Wayland/KDE via D-Bus scripting API) with xdotool fallback (X11 only).

### Browser Control via Playwright

Automate web browsers in headless (background) or headed (interactive) mode. Navigate pages, interact with elements by numeric index, capture screenshots, extract content, and manage tabs. All actions are also available as MCP tools for AI assistant integration.

```bash
ptx-agent browser install                  # Install browser binaries via Portunix
ptx-agent browser launch [--headless]      # Launch browser instance
ptx-agent browser goto <url>               # Navigate to URL
ptx-agent browser elements                 # List interactive elements with indices
ptx-agent browser click <index>            # Click element by index
ptx-agent browser type <index> <text>      # Type into form field
ptx-agent browser screenshot [--full]      # Capture screenshot
ptx-agent browser content [--selector S]   # Extract page text/HTML
ptx-agent browser tabs                     # List open tabs
ptx-agent browser tab new [url]            # Open new tab
ptx-agent browser close                    # Close browser
```

Requires Playwright browser binaries installed via `ptx-agent browser install`. See [ADR-002](docs/adr/002-browser-control-playwright.md) for architecture details.

### Secure File Browsing

Browse directories with whitelist-based access control and symlink traversal protection.

```bash
ptx-agent files list <path> --allowed-dirs=/home/user,/tmp
ptx-agent files info <path> --allowed-dirs=/home/user,/tmp
```

## Platform Support

| Feature | Linux (X11) | Linux (Wayland/KDE) | Windows |
|---|:---:|:---:|:---:|
| Application accessibility | AT-SPI2 | AT-SPI2 | UI Automation |
| VS Code terminal reading | yes | yes | yes |
| Keyboard input | xdotool | uinput | planned |
| Window management | xdotool | KWin D-Bus | planned |
| Browser control | yes | yes | yes |
| File browsing | yes | yes | yes |

## Building

Requires **Go 1.24+**.

```bash
make build          # Build ptx-agent binary
make test           # Run tests
make lint           # Format + vet
make clean          # Remove build artifacts
```

The compiled binary is `ptx-agent` (or `ptx-agent.exe` on Windows).

Debug tools for accessibility inspection:

```bash
make build-debug    # Build uia-debug (Windows) or atspi-debug (Linux)
```

## Architecture

The agent uses a pluggable backend pattern with compile-time platform separation via Go build tags:

```text
src/
├── cmd/
│   └── ptx-agent/          # Entry point
├── internal/
│   ├── accessibility/       # Cross-platform accessibility abstraction
│   │   ├── backend_atspi_*  # AT-SPI2 via D-Bus (Linux)
│   │   └── backend_uia_*   # UI Automation via COM (Windows)
│   ├── browser/             # Browser control via Playwright
│   ├── cli/                 # Cobra command definitions
│   ├── config/              # Configuration management
│   ├── files/               # Secure file browser
│   ├── ipc/                 # IPC abstraction (D-Bus transport)
│   ├── keyboard/            # Input simulation (uinput, xdotool)
│   ├── mcp/                 # MCP server (JSON-RPC over stdio)
│   ├── scenario/            # Automated scenario execution
│   ├── vscode/              # VS Code control (platform-agnostic)
│   └── windows/             # Window management (KWin, xdotool)
```

Higher-level packages depend only on interfaces, never on platform-specific implementations directly. For example, the `vscode` package uses only `accessibility.Backend` and works on both Linux and Windows without modification.

## Repository

- **GitHub**: <https://github.com/CassandraGargoyle/portunix-agent>

## Documentation

- [Contributing guidelines](docs/contributing/)
- [Architecture Decision Records](docs/adr/)
- [Issue tracker](https://github.com/CassandraGargoyle/portunix-agent/issues) — long-form write-ups in [docs/issues/](docs/issues/)

## License

MIT License - see [LICENSE](LICENSE) for details.