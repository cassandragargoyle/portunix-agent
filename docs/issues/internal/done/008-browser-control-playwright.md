# Issue #008 - Browser Control via Playwright

**Type**: enhancement
**Priority**: high
**Status**: ✅ Implemented
**Branch**: merged to main (df420bb)
**Depends on**: Issue #001, Issue #006
**ADR**: [ADR-002](../../../adr/002-browser-control-playwright.md)

## Feature Description

Extend `ptx-agent` with browser control capabilities using [Playwright](https://playwright.dev/). The implementation draws inspiration from the browser control layer of [browser-use](https://github.com/browser-use/browser-use) — specifically its DOM extraction, interactive element indexing, and action primitives — but does **not** include AI agent communication or autonomous decision-making logic.

The browser module enables two modes of operation:

1. **Background mode** — headless browser automation (scraping, form filling, testing workflows) without user interaction
2. **Interactive mode** — controlling a visible browser window in collaboration with the user (the user sees what the agent does and can intervene)

## Use Case

AI assistants and automation scripts need to interact with web applications — navigate pages, fill forms, click elements, extract data, and manage tabs. Currently `ptx-agent` controls desktop applications via accessibility APIs and keyboard simulation, but has no web browser control capability.

Adding browser control allows:

- Automating web-based workflows (CI dashboards, admin panels, ticketing systems)
- Extracting structured data from web pages for AI assistants
- Filling forms and navigating multi-step web processes
- Collaborative browsing where the user observes and guides the automation
- Testing web applications in both headless and headed modes

## Scope

### In Scope

- Browser lifecycle management (launch, connect, close) in both headless and headed modes
- Page navigation (goto, back, forward, reload)
- DOM extraction with interactive element indexing (inspired by browser-use's element detection)
- Element interaction (click, type, select, hover, drag)
- Keyboard input (send keys, key combinations)
- Screenshot capture (full page and viewport)
- Tab management (open, close, switch, list)
- Content extraction (text, structured data via CSS selectors)
- Scroll control (directional, to element, to coordinates)
- File upload and download handling
- Dialog/popup handling (accept, dismiss, input)
- JavaScript evaluation on page
- Network idle detection before state capture
- MCP tool exposure for all browser actions (extends Issue #006)

### Out of Scope

- AI agent communication or autonomous browsing logic (browser-use's agent layer)
- Browser extension installation or management
- Browser profile/credential management beyond Playwright defaults
- Proxy configuration (may be added in a future issue)
- PDF generation (may be added in a future issue)
- Video recording of browser sessions

## Technical Requirements

### Playwright Integration

- Use [playwright-go](https://github.com/playwright-community/playwright-go) for Go bindings
- Support Chromium as the primary browser engine
- Optional Firefox and WebKit support as secondary targets
- Playwright browser binaries managed via `ptx-agent browser install`

### Command Structure

```bash
# Browser lifecycle
ptx-agent browser install              # Download Playwright browser binaries
ptx-agent browser launch [--headless]  # Launch browser instance
ptx-agent browser connect <ws-url>     # Connect to existing browser (CDP)
ptx-agent browser close                # Close browser instance
ptx-agent browser status               # Show running browser state

# Navigation
ptx-agent browser goto <url>           # Navigate to URL
ptx-agent browser back                 # Go back
ptx-agent browser forward              # Go forward
ptx-agent browser reload               # Reload page

# Page interaction
ptx-agent browser click <index>        # Click element by index
ptx-agent browser type <index> <text>  # Type text into element
ptx-agent browser keys <keys>          # Send key combination
ptx-agent browser scroll <direction>   # Scroll page
ptx-agent browser select <index> <val> # Select dropdown option

# State capture
ptx-agent browser elements             # List interactive elements with indices
ptx-agent browser screenshot [--full]  # Capture screenshot
ptx-agent browser content [--selector] # Extract page text/HTML
ptx-agent browser tabs                 # List open tabs

# Tab management
ptx-agent browser tab switch <id>      # Switch to tab
ptx-agent browser tab close [id]       # Close tab
ptx-agent browser tab new [url]        # Open new tab
```

### MCP Tools to Expose

| Tool | Description | Parameters |
| ---- | ----------- | ---------- |
| `agent_browser_launch` | Launch browser instance | `headless` (bool, optional, default: true) |
| `agent_browser_close` | Close browser instance | — |
| `agent_browser_goto` | Navigate to URL | `url` (string, required), `new_tab` (bool, optional) |
| `agent_browser_back` | Navigate back | — |
| `agent_browser_forward` | Navigate forward | — |
| `agent_browser_reload` | Reload current page | — |
| `agent_browser_click` | Click element by index | `index` (int, required) |
| `agent_browser_type` | Type text into element | `index` (int, required), `text` (string, required), `clear` (bool, optional) |
| `agent_browser_keys` | Send key combination | `keys` (string, required) |
| `agent_browser_scroll` | Scroll page | `direction` (string: up/down/left/right), `amount` (int, optional) |
| `agent_browser_select` | Select dropdown option | `index` (int, required), `value` (string, required) |
| `agent_browser_elements` | List interactive elements | `viewport_only` (bool, optional) |
| `agent_browser_screenshot` | Capture screenshot | `full_page` (bool, optional) |
| `agent_browser_content` | Extract page content | `selector` (string, optional), `format` (string: text/html/markdown) |
| `agent_browser_tabs` | List open tabs | — |
| `agent_browser_tab_switch` | Switch to tab | `tab_id` (int, required) |
| `agent_browser_tab_close` | Close tab | `tab_id` (int, optional) |
| `agent_browser_tab_new` | Open new tab | `url` (string, optional) |
| `agent_browser_evaluate` | Execute JavaScript | `expression` (string, required) |
| `agent_browser_upload` | Upload file to element | `index` (int, required), `path` (string, required) |
| `agent_browser_status` | Get browser and page state | — |

### DOM Extraction and Element Indexing (Inspired by browser-use)

The core of the browser control layer is the element detection and indexing system, adapted from browser-use's approach:

1. **Accessibility tree + DOM fusion** — combine Playwright's `page.Accessibility().Snapshot()` with DOM queries to build a rich element model
2. **Interactive element detection** — identify clickable/interactive elements by analyzing tag names, ARIA roles, event listeners, and CSS cursor properties
3. **Numeric element indexing** — assign unique numeric indices to all interactive elements for unambiguous reference in commands
4. **Visibility classification** — distinguish in-viewport vs. off-screen elements, provide scroll distance hints for hidden elements
5. **Element highlighting** — optional JavaScript injection to visually mark interactive elements (for interactive mode), applied after screenshot capture to keep screenshots clean

### Architecture

```text
┌──────────────────────────────────────────────┐
│  ptx-agent                                   │
│                                              │
│  ┌──────────────┐  ┌─────────────────────┐   │
│  │ CLI commands  │  │ MCP tools           │   │
│  │ (browser *)   │  │ (agent_browser_*)   │   │
│  └──────┬───────┘  └──────────┬──────────┘   │
│         │                     │              │
│         └──────────┬──────────┘              │
│                    ▼                         │
│  ┌─────────────────────────────────────┐     │
│  │  browser package                    │     │
│  │                                     │     │
│  │  ┌───────────┐  ┌───────────────┐   │     │
│  │  │ Session   │  │ DOM Service   │   │     │
│  │  │ Manager   │  │ (element      │   │     │
│  │  │           │  │  detection +  │   │     │
│  │  │           │  │  indexing)    │   │     │
│  │  └─────┬─────┘  └──────┬───────┘   │     │
│  │        │               │           │     │
│  │        └───────┬───────┘           │     │
│  │                ▼                   │     │
│  │  ┌─────────────────────────────┐   │     │
│  │  │  Playwright (playwright-go) │   │     │
│  │  └─────────────────────────────┘   │     │
│  └─────────────────────────────────────┘     │
└──────────────────────────────────────────────┘
```

### Implementation Approach

1. **New package `src/internal/browser/`**
   - `session.go` — browser lifecycle management (launch, connect, close, state)
   - `dom.go` — DOM extraction, accessibility tree processing, element indexing
   - `actions.go` — high-level action primitives (click, type, scroll, etc.)
   - `tabs.go` — tab management
   - `screenshot.go` — screenshot capture and encoding

2. **New CLI commands `src/internal/cli/browser.go`** — cobra commands for `browser *` subcommands

3. **New MCP tools in `src/internal/mcp/browser_handlers.go`** — MCP tool definitions and handlers for browser actions

4. **Browser binary management** — `browser install` command to download Playwright browser binaries to a well-known location

### Platform Considerations

- Linux: Full support (headless and headed)
- Windows: Full support (headless and headed)
- macOS: TBD (follows project platform policy)
- Headless mode requires no display server
- Headed mode requires a running display server (X11/Wayland on Linux, desktop session on Windows)

## Security Considerations

- Browser runs with the same permissions as the `ptx-agent` process
- No automatic credential injection — user must explicitly provide credentials via actions
- JavaScript evaluation (`agent_browser_evaluate`) is powerful and should respect access control (Issue #002)
- Downloaded files should go to a controlled directory, not arbitrary paths
- Domain restriction support may be added for security-sensitive deployments
- Screenshot data may contain sensitive information — handle with care in MCP responses

## Acceptance Criteria

1. [ ] `ptx-agent browser install` downloads Playwright Chromium binaries
2. [ ] `ptx-agent browser launch` starts a browser in headless or headed mode
3. [ ] `ptx-agent browser goto <url>` navigates to a URL and waits for load
4. [ ] `ptx-agent browser elements` returns a numbered list of interactive elements
5. [ ] `ptx-agent browser click <index>` clicks the correct element by its index
6. [ ] `ptx-agent browser type <index> <text>` types text into the correct form field
7. [ ] `ptx-agent browser screenshot` returns a screenshot of the current viewport
8. [ ] `ptx-agent browser content` extracts page text content
9. [ ] Tab management (new, switch, close, list) works correctly
10. [ ] All browser actions are exposed as MCP tools via `ptx-agent mcp serve`
11. [ ] Headless mode works without a display server
12. [ ] Headed mode shows visible browser with element highlights (interactive mode)
13. [ ] Works on both Linux and Windows
14. [ ] Network idle detection before state capture prevents incomplete page snapshots

## Out of Scope

- AI agent communication or autonomous browsing decisions
- Browser extension management
- Proxy/VPN configuration
- PDF generation
- Video recording
- Cookie/session persistence across `ptx-agent` restarts
- Multi-browser-instance management (one instance per `ptx-agent` process)

## Complexity Estimate

High — Playwright integration, DOM extraction with element indexing, dual-mode (headless/headed) support, MCP tool surface, and cross-platform testing require significant implementation effort. The DOM service (element detection and indexing) is the most complex component.

## Portunix Dependency: Browser Binary Installation

`ptx-agent browser install` delegates browser binary installation to the Portunix package system via `portunix install playwright-browsers`.

### Required from Portunix Team (separate issues)

Two changes are needed in the Portunix project:

1. **New package `playwright-browsers`** — a standard package definition in `assets/packages/playwright-browsers.json` that installs Chromium binaries via `npx playwright install chromium` (or equivalent)
2. **Plugin dependency resolution** — extend the plugin manifest `dependencies` section with a `packages` field so that `portunix plugin install ptx-agent` automatically detects and installs the `playwright-browsers` package

### ptx-agent plugin manifest change (future)

Once Portunix supports package dependencies in plugin manifests, the ptx-agent `plugin.json` should declare:

```json
{
  "dependencies": {
    "packages": [
      {
        "name": "playwright-browsers",
        "optional": false,
        "reason": "Required for browser automation"
      }
    ]
  }
}
```

**Current ptx-agent implementation**: `ptx-agent browser install` calls `portunix install playwright-browsers` as a subprocess and passes through stdout/stderr

## Implementation Notes

- Primary inspiration: [browser-use](https://github.com/browser-use/browser-use) — specifically the DOM extraction, element indexing, and action primitives (not the AI agent layer)
- Go bindings: [playwright-go](https://github.com/playwright-community/playwright-go)
- Playwright documentation: <https://playwright.dev/>
- Element indexing approach: assign sequential numeric IDs to interactive elements detected via accessibility tree + DOM analysis, present alongside element type and text for unambiguous reference
- Screenshot strategy (from browser-use): capture screenshot before injecting element highlights to keep visual capture clean
- Network stability: wait for pending requests to settle before capturing page state
