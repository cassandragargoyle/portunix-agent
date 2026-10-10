# Issue #004 - Windows: VS Code control via UI Automation (UIA)

**Type**: enhancement
**Priority**: high
**Status**: ✅ Implemented
**Depends on**: Issue #003 (MVP VS Code accessibility control — Linux)

## Feature Description

Implement the Windows accessibility backend for `ptx-agent` VS Code control, using Microsoft UI Automation (UIA) via COM. This extends the cross-platform accessibility architecture established in Issue #003 to work on Windows.

Windows UI Automation is the native accessibility framework used by all major screen readers (Narrator, NVDA, JAWS). Chromium/Electron applications — including VS Code — expose their UI as a UIA tree that can be navigated and queried programmatically. This is the same technology that enables blind users to interact with VS Code on Windows.

## Use Case

Same as Issue #003 — observe and interact with VS Code from outside the process:

- **Automated workflows**: Read terminal output to determine command results
- **Remote assistance**: See terminal content on the target machine
- **Intelligent automation**: Combine terminal reading with keyboard input for feedback loops
- **Non-intrusive integration**: No VS Code extension required — uses the standard Windows accessibility API

## Background: How It Works on Windows

### Windows UI Automation vs Linux AT-SPI2

On Linux (Issue #003), accessibility uses a two-layer architecture: AT-SPI2 protocol running over D-Bus transport. On Windows, UI Automation is a **single unified API** accessed via COM — there is no separate transport layer.

```text
Linux (Issue #003):                    Windows (this issue):
┌──────────────────────┐               ┌──────────────────────┐
│  vscode package      │               │  vscode package      │
│  (unchanged)         │               │  (unchanged)         │
├──────────────────────┤               ├──────────────────────┤
│  accessibility iface │               │  accessibility iface │
│  ┌────────────────┐  │               │  ┌────────────────┐  │
│  │ AT-SPI2 backend│  │               │  │  UIA backend   │  │
│  │ (uses IPC)     │  │               │  │  (uses COM     │  │
│  └───────┬────────┘  │               │  │   directly)    │  │
├──────────┼───────────┤               │  └────────────────┘  │
│  ipc interface       │               │  No IPC layer needed │
│  ┌───────┴────────┐  │               └──────────────────────┘
│  │ D-Bus backend  │  │
│  └────────────────┘  │
└──────────────────────┘
```

**Key insight**: The IPC abstraction layer (`ipc` package) was designed for AT-SPI2's architecture where accessibility protocol rides on top of D-Bus. On Windows, UIA is a direct COM API — the UIA backend implements `accessibility.Backend` directly without an intermediate IPC layer.

### Chromium UIA Provider

Since Chrome 126 (default from Chrome 138), Chromium implements native UIA support, mapping its internal `AXNode` tree directly to UIA control types, properties, and patterns. This eliminates the older MSAA-to-UIA proxy and provides full-fidelity accessibility data.

VS Code (Electron) inherits this Chromium UIA provider. The accessibility tree is exposed through standard COM interfaces:

```text
ptx-agent (Go)
    |
    | COM / go-ole
    |
    v
IUIAutomation (entry point)
    |
    v
Desktop (root element)
    |
    +-- Window (Chrome_WidgetWin_1, "... - Visual Studio Code")
         |
         +-- Document (Chromium web content root)
              |
              +-- ... (VS Code layout panels, ~30 levels deep)
                   |
                   +-- List (xterm-accessibility-tree)
                        |
                        +-- ListItem (terminal row 1)
                        +-- ListItem (terminal row 2)
                        +-- ...
```

### Key UIA COM Interfaces

| Interface | Purpose |
| --------- | ------- |
| `IUIAutomation` | Entry point — root element, tree walkers, conditions, events |
| `IUIAutomationElement` | Single node — properties (Name, ControlType, AutomationId), patterns |
| `IUIAutomationTreeWalker` | Navigate parent/child/sibling relationships |
| `IUIAutomationCondition` | Filter elements during search |
| `IUIAutomationCacheRequest` | Batch-fetch multiple properties in a single cross-process call |
| `IUIAutomationTextPattern` | Read text content (DocumentRange, GetText) |

### UIA Connection Flow

```text
1. CoInitializeEx(COINIT_MULTITHREADED)
2. CoCreateInstance(CLSID_CUIAutomation, IID_IUIAutomation) → pAutomation
3. pAutomation.GetRootElement() → pDesktop
4. Find VS Code window by Name/ClassName condition
5. Navigate tree → find terminal List element
6. Read ListItem names or use TextPattern → terminal content
7. Release COM objects, CoUninitialize()
```

### VS Code Accessibility Tree on Windows (UIA)

```text
Window (Chrome_WidgetWin_1)                    UIA_WindowControlTypeId
  └── Document (web content root)              UIA_DocumentControlTypeId
       ├── Group (sidebar)                     UIA_GroupControlTypeId
       ├── Group (editor area)
       │    └── Document (editor content)      UIA_DocumentControlTypeId
       ├── Group (terminal area)
       │    └── List (xterm-accessibility-tree) UIA_ListControlTypeId
       │         ├── ListItem (row 1)          UIA_ListItemControlTypeId
       │         ├── ListItem (row 2)          UIA_ListItemControlTypeId
       │         └── ...
       └── Group (status bar)
```

**Role mapping from Linux AT-SPI2 to Windows UIA:**

| AT-SPI2 role (string) | UIA ControlType (int) | UIA ControlType name |
| ---------------------- | --------------------- | -------------------- |
| `"application"` | 50004 | `UIA_WindowControlTypeId` |
| `"frame"` | 50032 | `UIA_PaneControlTypeId` |
| `"document web"` | 50030 | `UIA_DocumentControlTypeId` |
| `"panel"` | 50026 | `UIA_GroupControlTypeId` |
| `"list"` | 50008 | `UIA_ListControlTypeId` |
| `"list item"` | 50009 | `UIA_ListItemControlTypeId` |

### Terminal Content Access

xterm.js creates the same accessibility DOM structure on all platforms. On Windows, each terminal row (ListItem) exposes its text content via:

1. **`Name` property** — `IUIAutomationElement::get_CurrentName()` — contains the row text
2. **`TextPattern`** — `IUIAutomationTextPattern::DocumentRange::GetText()` — structured text access (if available)

The `Name` property approach is simpler and likely sufficient for the MVP.

### Prerequisite: Enabling Accessibility

**Windows is more favorable than Linux here.** On Linux, `--force-renderer-accessibility` is required. On Windows:

1. **Automatic activation**: Chromium on Windows automatically enables accessibility when a UIA client connects. Simply calling `CoCreateInstance(CLSID_CUIAutomation)` and querying the tree triggers activation.
2. **VS Code setting**: `"editor.accessibilitySupport": "on"` in `settings.json` — the most reliable explicit method.
3. **System screen reader**: Having Narrator, NVDA, or JAWS running triggers automatic activation.
4. **No `--force-renderer-accessibility` needed** — VS Code on Windows does not require this flag.

**Important**: `--force-renderer-accessibility` is NOT reliably supported in VS Code on Windows (VS Code intercepts Electron's command-line parsing). The recommended approach is `editor.accessibilitySupport: "on"` or relying on automatic UIA client detection.

## Scope

### Phase 0: Platform Build Separation (prerequisite)

The existing Issue #003 code has no build tags — all AT-SPI2 and D-Bus files compile on every platform, which breaks Windows builds (missing D-Bus dependencies). Before adding Windows code, existing Linux-specific files must be tagged and renamed.

**Rename and tag existing files:**

| Current file | Rename to | Build tag to add |
| ------------ | --------- | ---------------- |
| `ipc/backend_dbus.go` | `ipc/backend_dbus_linux.go` | `//go:build linux` |
| `ipc/detect.go` | `ipc/detect_linux.go` | `//go:build linux` |
| `ipc/ipc_test.go` | (keep) | no tag (uses mocks only) |
| `accessibility/backend_atspi.go` | `accessibility/backend_atspi_linux.go` | `//go:build linux` |
| `accessibility/backend_atspi_text.go` | `accessibility/backend_atspi_text_linux.go` | `//go:build linux` |
| `accessibility/backend_atspi_tree.go` | `accessibility/backend_atspi_tree_linux.go` | `//go:build linux` |
| `accessibility/detect.go` | `accessibility/detect_linux.go` | `//go:build linux` |
| `accessibility/accessibility.go` | (keep) | no tag (interface, all platforms) |
| `accessibility/accessibility_test.go` | (keep) | no tag (uses mocks only) |
| `vscode/vscode.go` | (keep) | no tag (platform-agnostic) |
| `vscode/terminal.go` | (keep) | no tag (platform-agnostic) |
| `vscode/vscode_test.go` | (keep) | no tag (uses mocks only) |

**Rules:**

- Every file that imports a platform-specific package (`godbus/dbus`, `go-ole`) MUST have a build tag
- Every file that references platform-specific types (e.g. `dbus.Conn`) MUST have a build tag
- Interface definitions and test files using mocks remain untagged (compile everywhere)
- Use both file suffix (`_linux.go`) AND explicit build tag (`//go:build linux`) — belt and suspenders
- After this phase: `go build ./...` must succeed on both Linux and Windows (Windows build produces no accessibility backend yet, but compiles cleanly)

### Phase 1: AccessibleRef Generalization (prerequisite refactoring)

The current `AccessibleRef` struct is AT-SPI2-specific (`BusName`, `Path`). For cross-platform support, it must be generalized to support both AT-SPI2 and UIA node references.

**Design options:**

**Option A — Interface-based ref (recommended)**:

```go
// AccessibleRef is an opaque reference to a node in the accessibility tree
type AccessibleRef interface {
    // String returns a human-readable representation for debugging
    String() string
}

// ATSPIRef identifies a node on AT-SPI2/D-Bus
type ATSPIRef struct {
    BusName string
    Path    string
}

// UIARef identifies a node via Windows UI Automation
type UIARef struct {
    // Holds an IUIAutomationElement pointer (as uintptr for Go)
    Element uintptr
}
```

**Option B — Union struct with platform tag**:
claude 
```go
type AccessibleRef struct {
    // AT-SPI2 (Linux)
    BusName string
    Path    string
    // UIA (Windows)
    Element uintptr
}
```

Option A is cleaner but requires updating all `Backend` method signatures. Option B is simpler but pollutes the struct with unused fields per platform.

### Phase 2: UIA Backend Implementation

- Implement `accessibility.Backend` interface using Windows UI Automation via COM
- Use `github.com/go-ole/go-ole` for COM interop
- Use build tags (`//go:build windows`) for platform-specific compilation
- No IPC layer needed — UIA backend uses COM directly

### Phase 3: IPC Detect / Accessibility Detect Update

- Update `ipc.Detect()` to return no-op or skip on Windows (UIA doesn't use the IPC layer)
- Update `accessibility.Detect()` to select UIA backend on Windows
- Maintain the existing AT-SPI2 path on Linux unchanged

### Phase 4: Fix Leaky Abstractions in vscode Package

The current `vscode` package has AT-SPI2-specific assumptions that must be fixed:

1. **`hasTextInterface()` in `terminal.go:122`** — checks for `"org.a11y.atspi.Text"` string. This should use a platform-agnostic interface name or the `Backend` should expose a `HasTextCapability(ref)` method.
2. **Error messages** — references to `--force-renderer-accessibility` should be platform-aware.

### Phase 5: Functional Terminal Reading on Windows

- VS Code discovery via UIA (find window by Name/ClassName)
- Tree navigation to terminal List element
- Text content reading from ListItem elements
- Same CLI commands work on both platforms: `portunix agent vscode terminal read/list/status`

## Technical Requirements

- **Language**: Go (consistent with existing ptx-agent)
- **COM library**: `github.com/go-ole/go-ole` (pure Go COM interop)
- **Optional**: `github.com/auuunya/go-element` — Go wrapper for UIA (built on go-ole). May be used as reference or dependency if it covers needed functionality. Evaluate whether it supports `IUIAutomationTextPattern` (likely not — may need direct go-ole syscalls for that).
- **Platform**: Windows 10/11 (UIA is available on all supported Windows versions)
- **Build tags**: `//go:build windows` for all Windows-specific files
- **Architecture**: UIA backend implements `accessibility.Backend` directly (no IPC layer)

### Key COM GUIDs

```text
CLSID_CUIAutomation:  {FF48DBA4-60EF-4201-AA87-54103EEF594E}
IID_IUIAutomation:    {30CBE57D-D9D0-452A-AB13-7AC5AC4825EE}
```

### Backend Architecture

```text
src/internal/
  ├── ipc/                                           # Linux-only IPC layer
  │    ├── ipc.go                                    # interface — no build tag (all platforms)
  │    ├── backend_dbus_linux.go                     # //go:build linux — D-Bus impl (renamed from backend_dbus.go)
  │    ├── detect_linux.go                           # //go:build linux (renamed from detect.go)
  │    ├── detect_windows.go                         # //go:build windows — returns "not used" error
  │    └── ipc_test.go                               # no tag (mock-based tests)
  │
  ├── accessibility/
  │    ├── accessibility.go                          # Backend interface — no build tag (all platforms)
  │    ├── ref.go                                    # AccessibleRef interface — no build tag — NEW
  │    ├── ref_linux.go                              # //go:build linux — ATSPIRef — NEW
  │    ├── ref_windows.go                            # //go:build windows — UIARef — NEW
  │    ├── backend_atspi_linux.go                    # //go:build linux (renamed from backend_atspi.go)
  │    ├── backend_atspi_text_linux.go               # //go:build linux (renamed from backend_atspi_text.go)
  │    ├── backend_atspi_tree_linux.go               # //go:build linux (renamed from backend_atspi_tree.go)
  │    ├── backend_uia_windows.go                    # //go:build windows — NEW
  │    ├── backend_uia_tree_windows.go               # //go:build windows — NEW (tree walking)
  │    ├── backend_uia_text_windows.go               # //go:build windows — NEW (text reading)
  │    ├── detect_linux.go                           # //go:build linux (renamed from detect.go)
  │    ├── detect_windows.go                         # //go:build windows — selects UIA — NEW
  │    └── accessibility_test.go                     # no tag (mock-based tests)
  │
  └── vscode/                                        # Platform-agnostic (no build tags)
       ├── vscode.go                                 # Fix leaky abstractions
       ├── terminal.go                               # Fix hasTextInterface, error messages
       └── vscode_test.go
```

### Dependency Direction (unchanged principle)

```text
cli/vscode.go → vscode.Controller → accessibility.Backend (interface)
                                      ↑
                    backend_atspi.go (Linux)  |  backend_uia.go (Windows)
                    uses ipc.Connector       |  uses COM directly
```

## Acceptance Criteria

### Platform Build Separation (Phase 0)

- [ ] All AT-SPI2/D-Bus files renamed with `_linux.go` suffix
- [ ] All renamed files have explicit `//go:build linux` tag on first line
- [ ] `ipc/ipc.go` (interface) has NO build tag — compiles on all platforms
- [ ] `accessibility/accessibility.go` (interface) has NO build tag — compiles on all platforms
- [ ] Test files using mocks have NO build tag — compile on all platforms
- [ ] `go build ./...` succeeds on Linux with no behavior change
- [ ] `go build ./...` succeeds on Windows (no accessibility backend available yet, but compiles)
- [ ] All existing unit tests pass unchanged on Linux

### AccessibleRef Generalization

- [ ] `AccessibleRef` supports both AT-SPI2 (BusName/Path) and UIA (Element pointer) references
- [ ] All existing AT-SPI2 code compiles and works unchanged after refactoring
- [ ] Existing unit tests pass without modification

### UIA Backend

- [ ] `accessibility.UIABackend` implements `accessibility.Backend` interface
- [ ] Uses `go-ole` for COM interop (no CGo)
- [ ] Properly initializes and releases COM resources (CoInitialize/CoUninitialize)
- [ ] Build-tagged `//go:build windows`

### Functional

- [ ] Can discover running VS Code instances via UIA tree
- [ ] Can navigate VS Code accessibility tree to find terminal panel
- [ ] Can read text content from VS Code integrated terminal
- [ ] Can list multiple open terminal instances
- [ ] Can read content from a specific terminal (by index)
- [ ] Reports clear error when no VS Code instance is found
- [ ] Reports clear error when accessibility is not available
- [ ] Same CLI commands work on both Linux and Windows

### Cross-Platform

- [ ] `accessibility.Detect()` selects UIA on Windows, AT-SPI2 on Linux
- [ ] `vscode` package has no platform-specific imports or string literals
- [ ] Linux build is not affected (no regressions)
- [ ] Windows build excludes AT-SPI2/D-Bus code, Linux build excludes UIA/COM code

### Testing

- [ ] Unit tests for UIA backend with mock COM interfaces
- [ ] Unit tests for refactored `AccessibleRef` interface
- [ ] Integration test: `portunix agent vscode status` on Windows
- [ ] Existing Linux tests pass unchanged

## Known Limitations and Risks

### UIA Specific

- **COM threading model**: Go's goroutine scheduler may interact unpredictably with COM apartment threading. All COM calls should happen on a single OS thread (`runtime.LockOSThread()`).
- **Element lifetime**: `IUIAutomationElement` pointers are valid only while the element exists in the UI. Caching element references across calls may lead to stale pointers.
- **Tree depth**: VS Code's accessibility tree is deeply nested (~30 levels). UIA `CacheRequest` should be used to batch property fetches and reduce cross-process round-trips.
- **Native UIA availability**: Chrome 126+ / Electron 31+ for native UIA. Older versions use MSAA-to-UIA proxy which may behave differently.

### Go + COM

- **`go-ole` maturity**: go-ole is widely used (e.g., by WMI libraries) but UIA-specific wrappers are less battle-tested. May need to define vtable layouts manually for some interfaces.
- **`IUIAutomationTextPattern`**: The `go-element` library does not wrap TextPattern. Direct go-ole syscalls will be needed for structured text reading.
- **Memory management**: COM objects must be explicitly `Release()`d. Failing to do so causes memory leaks. Consider a helper that wraps COM pointers with cleanup.

### VS Code Specific

- **Window class name**: VS Code uses Chromium's `Chrome_WidgetWin_1` window class — same as Chrome browser. Must also check window title to distinguish VS Code from Chrome.
- **Multiple VS Code windows**: Each VS Code window is a separate top-level UIA element. Need to handle multi-window scenarios.
- **Accessibility activation timing**: After connecting a UIA client, Chromium may take a moment to build the full accessibility tree. May need a brief wait or retry logic.

## Out of Scope

- Writing/editing content in VS Code editor
- Clicking buttons or invoking VS Code commands via UIA actions
- Real-time terminal output streaming (UIA event subscriptions)
- Scrollback buffer beyond currently visible rows
- Reading editor content (future enhancement)
- macOS Accessibility API backend (separate future issue)
- Modifications to the IPC layer (not used by UIA backend)

## Future Extensions

1. **UIA event subscriptions**: Monitor `LiveRegionChanged` events for real-time terminal output
2. **TextPattern support**: Full structured text access with character/word/line granularity
3. **CacheRequest optimization**: Batch-fetch properties for entire subtrees in a single call
4. **Editor content reading**: Navigate to editor Document element, read via TextPattern
5. **VS Code actions**: Use `InvokePattern`, `ExpandCollapsePattern` for UI interaction
6. **MSAA fallback**: Support older Electron versions via IAccessible2 (if needed)

## Complexity Estimate

Medium — The accessibility.Backend interface is already defined and tested. The main work is implementing the UIA backend using COM, which is well-documented but requires careful COM lifecycle management in Go. The `AccessibleRef` generalization is the highest-risk refactoring as it touches all existing code.

## Research References

- [UI Automation Overview — Microsoft Learn](https://learn.microsoft.com/en-us/windows/win32/winauto/uiauto-uiautomationoverview)
- [Native UIA for Windows in Chromium](https://developer.chrome.com/blog/windows-uia-support-update)
- [Edge UIA Explainer](https://github.com/MicrosoftEdge/MSEdgeExplainers/blob/main/Accessibility/UIA/explainer.md)
- [xterm.js Screen Reader Mode](https://github.com/xtermjs/xterm.js/wiki/Design-Document:-Screen-Reader-Mode)
- [VS Code Accessibility Docs](https://code.visualstudio.com/docs/configure/accessibility/accessibility)
- [VS Code --force-renderer-accessibility](https://github.com/microsoft/vscode/issues/84833)
- [go-ole (Go COM interop)](https://github.com/go-ole/go-ole)
- [go-element (Go UIA wrapper)](https://github.com/auuunya/go-element)
- [w32uiautomation (archived reference)](https://github.com/hnakamur/w32uiautomation)
- [Electron Accessibility Documentation](https://www.electronjs.org/docs/latest/tutorial/accessibility)

---

**Created**: 2026-02-18
**Author**: Architect
