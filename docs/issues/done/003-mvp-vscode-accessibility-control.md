# Issue #003 - MVP: VS Code control via Chromium Accessibility layer (AT-SPI2)

**Type**: enhancement
**Priority**: high
**Status**: ✅ Implemented
**GitHub**: #3
**Branch**: `feature/003-vscode-accessibility-control` (merged to main)
**Closed**: 2026-02-18
**Depends on**: Issue #001 (MVP ptx-agent)

## Feature Description

Extend `ptx-agent` with the ability to read and interact with Visual Studio Code through the Chromium Accessibility layer (AT-SPI2). The MVP focuses on **reading VS Code integrated terminal content** — the first step toward full programmatic control of VS Code without relying on VS Code extensions or its Extension API.

VS Code is an Electron application built on Chromium, which exposes its entire UI as an accessibility tree via AT-SPI2 (Assistive Technology Service Provider Interface) over D-Bus on Linux. By connecting to this tree, ptx-agent can read text content, identify UI elements, and eventually invoke actions — all from outside the VS Code process.

## Use Case

Portunix needs the ability to observe what is happening inside VS Code, particularly in the integrated terminal. This enables:

- **Automated workflows**: Read terminal output to determine command results (success/failure, output values)
- **Remote assistance**: An operator can see terminal content on the target machine without screen sharing
- **Intelligent automation**: Combine terminal content reading with existing keyboard input (Issue #001) to create feedback loops — type a command, read the result, decide next action
- **Non-intrusive integration**: No VS Code extension required — works with any VS Code installation via the standard accessibility protocol

## Background: How It Works

### Layered Architecture with Replaceable Modules

The architecture follows the same **pluggable backend pattern** established in Issue #001 (windows: xdotool/kwin, keyboard: uinput/xdotool). Two independent abstraction layers ensure each technology can be replaced without affecting the rest of the system:

```text
┌─────────────────────────────────────────────────────────────┐
│  CLI Layer (cobra commands)                                 │
│  portunix agent vscode terminal read/list/status            │
├─────────────────────────────────────────────────────────────┤
│  VS Code Layer (vscode package)                             │
│  Application-specific tree navigation and terminal reading  │
│  Depends on: Accessibility Interface (not on AT-SPI2)       │
├─────────────────────────────────────────────────────────────┤
│  Accessibility Interface   ← replaceable module             │
│  Defines: tree navigation, text reading, app discovery      │
│  ┌──────────────────────┐  ┌─────────────────────────────┐  │
│  │  AT-SPI2 Backend     │  │  UI Automation Backend      │  │
│  │  (Linux)             │  │  (Windows, future)          │  │
│  │  Depends on: IPC     │  │  Depends on: COM/Win API    │  │
│  └──────────┬───────────┘  └─────────────────────────────┘  │
├─────────────┼───────────────────────────────────────────────┤
│  IPC Interface             ← replaceable module             │
│  Defines: connect, call method, get property, subscribe     │
│  ┌──────────┴───────────┐  ┌─────────────────────────────┐  │
│  │  D-Bus Backend       │  │  Other IPC Backend          │  │
│  │  (godbus/dbus/v5)    │  │  (COM, named pipes, future) │  │
│  └──────────────────────┘  └─────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

**Key design principle**: The `vscode` package never imports `dbus` or `atspi` directly. It depends only on the `Accessibility` interface. The AT-SPI2 backend never imports `godbus/dbus` directly — it depends only on the `IPC` interface. This makes each layer independently testable and replaceable.

### AT-SPI2 over D-Bus (Linux Implementation Detail)

```text
ptx-agent (Go)
    |
    | IPC interface → D-Bus backend (godbus/dbus/v5)
    |
    v
Session D-Bus  --[org.a11y.Bus.GetAddress()]--> Accessibility Bus (dedicated D-Bus)
                                                      |
                                                      v
                                              AT-SPI2 Registry
                                                      |
                                         +------------+------------+
                                         |                         |
                                   VS Code (Electron)       Other apps...
                                         |
                                   Chromium ATK bridge
                                         |
                                   Accessibility Tree
                                         |
                              +----------+-----------+
                              |          |           |
                           Editor    Terminal     Sidebar
                              |          |
                          AtkText    AtkText
                         interface   interface
```

### VS Code Accessibility Tree Structure

VS Code (Electron/Chromium) exposes its UI as an AT-SPI2 tree:

```text
Application ("Code")                          ROLE_APPLICATION
  └── Frame ("filename - project - VS Code")  ROLE_FRAME
       └── Panel                              ROLE_PANEL
            └── Document Web                  ROLE_DOCUMENT_WEB
                 ├── Panel (sidebar)
                 ├── Panel (editor area)
                 │    └── Document (editor content)  -- org.a11y.atspi.Text
                 ├── Panel (terminal area)
                 │    └── List (xterm-accessibility-tree)
                 │         ├── ListItem (row 1)      -- org.a11y.atspi.Text
                 │         ├── ListItem (row 2)      -- org.a11y.atspi.Text
                 │         └── ...
                 └── Panel (status bar)
```

**Important**: VS Code terminal does NOT use `ATSPI_ROLE_TERMINAL`. Instead, xterm.js creates a parallel DOM with `role="list"` containing `role="listitem"` children — these are exposed through Chromium's ATK bridge as web-oriented AT-SPI2 roles.

### Terminal Content Access

xterm.js maintains an invisible parallel DOM (`.xterm-accessibility-tree`) alongside the visual canvas renderer. Each terminal row is a list item implementing `org.a11y.atspi.Text` interface. Text content is retrieved via `GetText(startOffset, endOffset)`.

### Prerequisite: Enabling Accessibility

Chromium does **not** build the accessibility tree by default (performance optimization). It must be explicitly enabled:

1. Launch VS Code with `--force-renderer-accessibility` flag, **or**
2. Set environment variable `ACCESSIBILITY_ENABLED=1`, **or**
3. Set VS Code setting `"editor.accessibilitySupport": "on"`

On Linux, persistent flags can be added to `~/.config/code-flags.conf` (one flag per line).

## Scope (MVP)

### 1. AT-SPI2 Connection and Tree Discovery

- Connect to the AT-SPI2 accessibility bus via D-Bus
- Enumerate accessible applications (find VS Code)
- Navigate the VS Code accessibility tree to locate the terminal panel

### 2. Terminal Content Reading

- Read text content from all visible terminal rows via `org.a11y.atspi.Text.GetText()`
- Aggregate rows into a single terminal output string
- Support reading a specific terminal instance (when multiple terminals are open)

### 3. VS Code Detection and Validation

- Detect running VS Code instances via AT-SPI2
- Verify that accessibility tree is available (detect when `--force-renderer-accessibility` is missing)
- Report clear error when accessibility is not enabled

## Technical Requirements

- **Language**: Go (consistent with existing ptx-agent)
- **D-Bus library**: `github.com/godbus/dbus/v5` (pure Go, no CGo) — used only inside the D-Bus IPC backend
- **Protocol**: AT-SPI2 D-Bus interfaces (direct D-Bus calls, no ATK/libatspi dependency)
- **Platforms**: Linux (X11 and Wayland — AT-SPI2 works identically on both)
- **Architecture**: Pluggable backend pattern with Go interfaces (same as Issue #001)

### Modular Design Requirement

The implementation MUST use two independent abstraction layers defined as Go interfaces:

**1. IPC Interface** (`ipc` package) — abstracts inter-process communication transport

- MVP backend: D-Bus (`godbus/dbus/v5`)
- Future backends: COM (Windows), named pipes, direct socket
- Defines operations: `Connect()`, `Call()`, `GetProperty()`, `Subscribe()`
- The AT-SPI2 backend depends ONLY on this interface, never on `godbus/dbus` directly

**2. Accessibility Interface** (`accessibility` package) — abstracts accessibility technology

- MVP backend: AT-SPI2 (Linux)
- Future backends: UI Automation (Windows), macOS Accessibility API
- Defines operations: `FindApplication()`, `GetChildren()`, `GetRole()`, `GetText()`, `GetName()`
- The `vscode` package depends ONLY on this interface, never on AT-SPI2 directly

This separation ensures:

- Each layer is independently testable with mock implementations
- Windows support can be added by implementing `UIAutomationBackend` + `COMBackend` without modifying existing code
- Backend auto-detection follows the established pattern from `windows.Detect()` and `keyboard.Detect()`

### Key D-Bus Interfaces Used

| Interface | Purpose |
| --------- | ------- |
| `org.a11y.Bus` | Get accessibility bus address (on session bus) |
| `org.a11y.atspi.Registry` | Discover accessible applications |
| `org.a11y.atspi.Accessible` | Navigate tree: name, role, children, state |
| `org.a11y.atspi.Text` | Read text content: `GetText()`, `CharacterCount` |
| `org.a11y.atspi.Cache` | Bulk tree snapshot via `GetItems()` |
| `org.a11y.atspi.Collection` | Search elements by role/state (optimization) |

### AT-SPI2 Connection Flow

```text
1. dbus.SessionBus() → connect to session D-Bus
2. Call org.a11y.Bus.GetAddress() → get accessibility bus address
3. dbus.Connect(address) → connect to accessibility bus
4. Registry.GetDesktop(0) → get desktop root
5. Enumerate children → find "Code" application
6. Navigate tree → find terminal panel
7. Query org.a11y.atspi.Text → read terminal content
```

### Backend Architecture (extends Issue #001 pattern)

```text
src/internal/
  ├── ipc/                                # IPC abstraction layer
  │    ├── ipc.go                        # IPC interface definition (Connect, Call, GetProperty, Subscribe)
  │    ├── backend_dbus.go               # D-Bus backend (godbus/dbus/v5) — Linux
  │    ├── detect.go                     # IPC backend auto-detection
  │    └── ipc_test.go                   # Unit tests with mock IPC
  │
  ├── accessibility/                      # Accessibility abstraction layer
  │    ├── accessibility.go              # Accessibility interface definition (FindApplication, GetChildren, GetRole, GetText, GetName)
  │    ├── backend_atspi.go              # AT-SPI2 backend — Linux (depends on ipc.IPC interface)
  │    ├── backend_atspi_text.go         # AT-SPI2 Text interface operations
  │    ├── backend_atspi_tree.go         # AT-SPI2 tree walking and element search
  │    ├── detect.go                     # Accessibility backend auto-detection
  │    └── accessibility_test.go         # Unit tests with mock accessibility
  │
  └── vscode/                             # VS Code specific logic
       ├── vscode.go                     # VS Code discovery (depends on accessibility.Accessibility interface)
       ├── terminal.go                   # Terminal content extraction
       └── vscode_test.go               # Unit tests with mock accessibility
```

**Dependency direction** (no circular imports, each layer depends only on the interface above):

```text
cli/vscode.go → vscode.VSCode → accessibility.Accessibility (interface) → ipc.IPC (interface)
                                  ↑                                        ↑
                                  backend_atspi.go (implements)            backend_dbus.go (implements)
```

## Proposed Commands

```bash
# Read terminal content from VS Code
portunix agent vscode terminal read              # read active terminal content
portunix agent vscode terminal read --index 2    # read specific terminal (by index)
portunix agent vscode terminal list              # list open terminal instances

# VS Code discovery and diagnostics
portunix agent vscode status                     # show VS Code accessibility status
portunix agent vscode tree                       # dump accessibility tree (debug)
```

### Example Output

```bash
$ portunix agent vscode terminal read
zdenek@dev:~/project$ make build
go build -o ptx-agent ./src/cmd/ptx-agent/
zdenek@dev:~/project$ echo $?
0
zdenek@dev:~/project$ _

$ portunix agent vscode terminal list
Terminal 1: "bash" (active)
Terminal 2: "node server.js"
Terminal 3: "go test ./..."

$ portunix agent vscode status
VS Code PID: 12345
Accessibility: enabled
Terminal count: 3
Active terminal: 1
```

## Known Limitations and Risks

### AT-SPI2 Specific

- **Accessibility must be explicitly enabled** in VS Code — without `--force-renderer-accessibility`, the tree is empty
- **Web-oriented roles**: Terminal uses `ROLE_LIST` + `ROLE_LIST_ITEM`, not `ROLE_TERMINAL` — navigation is more complex than with native terminals
- **Fragmented text**: Terminal content is split across individual row elements — must be aggregated
- **Performance**: AT-SPI2 is "chatty" — each tree node query is a D-Bus round-trip; use `Cache.GetItems()` and `Collection.GetMatches()` for efficiency
- **Live region limitation**: xterm.js announces max 20 rows at a time — scrollback buffer access may be limited
- **Tree update latency**: Chromium batches accessibility updates asynchronously from the renderer process

### VS Code Specific

- **Electron detection regression**: Electron v37+ has a regression where `accessibilitySupportEnabled` may return false even when a screen reader is active — `--force-renderer-accessibility` is the reliable workaround
- **No standard terminal role**: Different from native terminal emulators (GNOME Terminal/VTE) which expose `ATSPI_ROLE_TERMINAL` with clean text access

## Acceptance Criteria

### Modular Architecture

- [ ] `ipc.IPC` interface is defined in its own package, independent of D-Bus
- [ ] `ipc.DBusBackend` implements `ipc.IPC` using `godbus/dbus/v5` — only place where godbus is imported
- [ ] `accessibility.Accessibility` interface is defined in its own package, independent of AT-SPI2
- [ ] `accessibility.ATSPIBackend` implements `accessibility.Accessibility` — depends only on `ipc.IPC` interface
- [ ] `vscode` package depends only on `accessibility.Accessibility` interface — no AT-SPI2 or D-Bus imports
- [ ] Backend auto-detection follows the established pattern (`ipc.Detect()`, `accessibility.Detect()`)

### Functional

- [ ] Can connect to AT-SPI2 accessibility bus via D-Bus
- [ ] Can discover running VS Code instances in the accessibility tree
- [ ] Can navigate VS Code accessibility tree to find terminal panel
- [ ] Can read text content from VS Code integrated terminal
- [ ] Can list multiple open terminal instances
- [ ] Can read content from a specific terminal (by index)
- [ ] Reports clear error when VS Code accessibility is not enabled
- [ ] Reports clear error when no VS Code instance is found
- [ ] Works on both X11 and Wayland (AT-SPI2 is display-server agnostic)

### Testing

- [ ] Unit tests for `ipc` package with mock IPC backend
- [ ] Unit tests for `accessibility` package with mock IPC
- [ ] Unit tests for `vscode` package with mock accessibility
- [ ] Command help text for all subcommands
- [ ] `portunix agent vscode status` shows diagnostics

## Out of Scope (MVP)

- Writing/editing content in VS Code editor
- Clicking buttons or invoking VS Code commands via accessibility actions
- Real-time terminal output streaming (event-based `text-changed:insert` listeners)
- Scrollback buffer beyond currently visible/accessible rows
- Reading editor content (future enhancement)
- Reading sidebar/explorer content (future enhancement)
- VS Code Extension API integration (different approach entirely)
- Windows platform backends (UI Automation + COM) — but the interface design MUST allow adding them later
- Code generation from AT-SPI2 XML definitions (`dbus-codegen-go`)
- Mouse interaction via accessibility `Component` interface

## Future Extensions (post-MVP)

1. **Event-based terminal monitoring**: Subscribe to `text-changed:insert` for real-time output tracking
2. **Editor content reading**: Navigate to editor panel, read document text
3. **VS Code actions**: Use `org.a11y.atspi.Action` to invoke UI element actions (click buttons, open menus)
4. **Collection-based search**: Use `org.a11y.atspi.Collection.GetMatches()` for efficient element discovery
5. **Full scrollback**: Access terminal buffer beyond visible area through xterm.js accessible view
6. **Cross-application**: Generalize accessibility package to control any accessible application (browsers, LibreOffice, etc.)
7. **Windows support**: Implement `UIAutomationBackend` (accessibility) + `COMBackend` (ipc) — no changes to vscode package needed
8. **macOS support**: Implement `AXBackend` (accessibility) using macOS Accessibility API

## Complexity Estimate

Medium-High — requires understanding of AT-SPI2 D-Bus protocol and VS Code's accessibility tree structure. No existing Go AT-SPI2 library exists, so the `accessibility` package is built from scratch using `godbus/dbus/v5`. The accessibility tree structure is well-documented but VS Code's specific tree layout must be discovered empirically (using Accerciser).

## Research References

- [AT-SPI2 Architecture](https://gnome.pages.gitlab.gnome.org/at-spi2-core/devel-docs/architecture.html)
- [AT-SPI2 D-Bus Interfaces](https://documentation.ubuntu.com/desktop/en/latest/reference/accessibility/dbus/)
- [AT-SPI2 XML Definitions (at-spi2-core)](https://gitlab.gnome.org/GNOME/at-spi2-core/-/tree/main/xml)
- [Chromium Accessibility Design](https://www.chromium.org/developers/design-documents/accessibility/)
- [xterm.js Screen Reader Mode](https://github.com/xtermjs/xterm.js/wiki/Design-Document:-Screen-Reader-Mode)
- [VS Code Accessibility Docs](https://code.visualstudio.com/docs/configure/accessibility/accessibility)
- [VS Code --force-renderer-accessibility](https://github.com/microsoft/vscode/issues/84833)
- [godbus/dbus v5](https://pkg.go.dev/github.com/godbus/dbus/v5)
- [Accerciser (AT-SPI2 inspector)](https://github.com/GNOME/accerciser)

---

**Created**: 2026-02-17
**Author**: Architect
