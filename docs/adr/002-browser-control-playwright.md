# ADR-002: Browser Control via Playwright for ptx-agent

## Status

Accepted

## Context

### Current Situation

`ptx-agent` currently provides remote device control capabilities focused on desktop applications:

- **File browsing** — list and navigate files on the target machine
- **VS Code control** — read terminal content, navigate UI via accessibility APIs (AT-SPI2 on Linux, UIA on Windows)
- **Application control** — keyboard input simulation, window management, application switching
- **MCP server** — all above capabilities exposed to AI assistants via MCP protocol (ADR-001)

However, `ptx-agent` has no capability to interact with web browsers or web applications. Modern workflows increasingly involve web-based tools (CI dashboards, admin panels, SaaS applications, ticketing systems), and AI assistants need a way to automate interactions with these tools.

### Motivation

Browser automation is a natural extension of `ptx-agent`'s device control mission. Two distinct use cases drive this decision:

1. **Background automation** — headless browser execution for tasks that do not require user visibility: data extraction, form submission, health checks, API-less web interactions. The agent operates autonomously and returns results.

2. **Interactive collaboration** — headed browser execution where the user observes the browser and the agent's actions in real-time. The user can intervene, guide, or correct the agent. This mode is essential for workflows that require human judgment (CAPTCHA, complex navigation, visual verification).

### Why Playwright

Several browser automation options were evaluated:

| Option | Pros | Cons |
| ------ | ---- | ---- |
| **Playwright** | Multi-browser support (Chromium, Firefox, WebKit), modern API, active development, Go bindings available, built-in wait mechanisms, network interception | Requires browser binary download |
| Chrome DevTools Protocol (CDP) direct | No extra dependencies, fine-grained control | Chromium-only, low-level, requires manual wait/retry logic |
| Selenium/WebDriver | Wide browser support, mature ecosystem | Heavier, slower, Go bindings less maintained |
| Rod (Go CDP library) | Pure Go, no extra binaries | Chromium-only, smaller community |

Playwright was selected because:

- **Multi-browser support** — Chromium as primary, Firefox and WebKit as secondary targets
- **Go bindings** — [playwright-go](https://github.com/playwright-community/playwright-go) provides idiomatic Go API
- **Built-in reliability** — auto-wait for elements, network idle detection, retry mechanisms
- **Consistent API** — same API across headless and headed modes (background and interactive)
- **Active ecosystem** — large community, regular updates, extensive documentation

### Inspiration from browser-use

The [browser-use](https://github.com/browser-use/browser-use) project demonstrates effective patterns for browser control that we adopt at the browser interaction layer:

- **DOM + accessibility tree fusion** — combining DOM snapshots with accessibility tree data to build a rich element model
- **Interactive element indexing** — assigning unique numeric IDs to interactive elements for unambiguous reference
- **Clean screenshot capture** — taking screenshots before element highlight injection
- **Network settling** — waiting for network requests to complete before capturing page state
- **Viewport-aware visibility** — tracking which elements are visible vs. off-screen

We adopt these patterns for the browser control primitives only. The AI agent communication layer, autonomous decision-making, and LLM integration from browser-use are **explicitly excluded** — `ptx-agent` provides the tools, and the AI assistant (via MCP) or the user decides what actions to take.

## Decision

`ptx-agent` will be extended with a browser control module using Playwright, supporting two modes of operation:

### Background Mode (Headless)

- Browser runs without a visible window
- Suitable for automated data extraction, form filling, and web testing
- No user interaction required during execution
- Results returned via CLI output or MCP tool responses

### Interactive Mode (Headed)

- Browser runs with a visible window that the user can observe
- Interactive elements are visually highlighted with numeric indices
- The user sees what the agent does and can intervene at any time
- Suitable for collaborative workflows, guided automation, and visual verification

### Architecture

The browser module is structured as a new internal package (`src/internal/browser/`) with clear separation of concerns:

- **Session Manager** — browser lifecycle (launch, connect, close), mode selection (headless/headed)
- **DOM Service** — page state extraction, accessibility tree processing, interactive element detection and indexing
- **Action Primitives** — click, type, scroll, select, keyboard input, file upload, JavaScript evaluation
- **Tab Manager** — tab lifecycle, switching, listing

All browser actions are exposed through:

1. **CLI commands** — `ptx-agent browser *` subcommands for direct use
2. **MCP tools** — `agent_browser_*` tools for AI assistant integration (extends the MCP server from ADR-001)

### Element Indexing Model

Interactive elements on the page are assigned sequential numeric indices. Each element entry includes:

- **Index** — unique numeric ID (e.g., `[1]`, `[2]`, `[3]`)
- **Type** — element type (button, input, link, select, etc.)
- **Text** — visible text or label
- **Visibility** — in-viewport or off-screen (with scroll distance hint)

This allows unambiguous element reference in both CLI commands and MCP tool calls:

```bash
# Click the element with index 5
ptx-agent browser click 5

# Type into the element with index 3
ptx-agent browser type 3 "hello world"
```

## Consequences

### Positive

1. `ptx-agent` gains web browser control, completing the device control surface (desktop apps + web apps)
2. AI assistants can automate web-based workflows via MCP tools without requiring a separate browser automation setup
3. Two operation modes (background and interactive) cover both autonomous and collaborative use cases
4. Element indexing provides a simple, unambiguous interface for referring to page elements
5. Playwright's cross-browser support allows targeting Chromium, Firefox, and WebKit
6. Consistent architecture — browser control follows the same CLI + MCP pattern as existing capabilities

### Negative

1. **Dependency on Playwright binaries** — requires downloading browser binaries (~200-400 MB per browser), increasing disk footprint
2. **Complexity increase** — DOM extraction with element indexing is a non-trivial subsystem
3. **Platform-specific display requirements** — headed mode requires a display server (X11/Wayland on Linux)
4. **playwright-go maturity** — Go bindings are community-maintained, not official Playwright team

### Mitigation

- Browser binary download is explicit (`ptx-agent browser install`) and only needed when browser control is used — not a mandatory dependency
- DOM service complexity is isolated in a dedicated package with clear interfaces
- Headless mode works without a display server for CI/automation use cases
- playwright-go is actively maintained and tracks upstream Playwright releases closely

## Related Decisions

- **ADR-001**: ptx-agent Provides Its Own MCP Server (browser tools extend the MCP surface)
- **Portunix ADR-038**: Distributed MCP Server Ecosystem for Plugins

## Related Issues

- **portunix-agent Issue #001**: MVP ptx-agent helper for Portunix
- **portunix-agent Issue #006**: Implement MCP server for ptx-agent
- **portunix-agent Issue #008**: Browser control via Playwright (implementation issue)

## Date

2026-02-21

## Author

ZK
