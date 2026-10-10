# Wayland Architecture in ptx-agent

## Problem

On X11, tools like `xdotool` can freely list windows, switch focus, and inject keystrokes into any application. On Wayland, this is not possible — Wayland's security model prevents clients from accessing other clients' windows or injecting input globally.

## Solution: Multi-Backend Architecture

ptx-agent uses a **Backend interface pattern** with auto-detection. Each subsystem (windows, keyboard) has multiple backend implementations, and the best available one is selected at runtime.

```text
ptx-agent CLI
    │
    ├── windows.Manager ──→ Backend interface
    │       ├── KWinBackend     (Wayland/KDE)
    │       └── XdotoolBackend  (X11)
    │
    └── keyboard.Controller ──→ Backend interface
            ├── UinputBackend   (Wayland + X11, needs /dev/uinput)
            └── XdotoolBackend  (X11 only)
```

## Window Management on Wayland

### Problem

`xdotool` only sees XWayland windows, not native Wayland windows (Kate, Dolphin, Konsole etc.).

### Solution: KWin Scripting API via DBus

KDE's KWin compositor exposes a **Scripting API** through DBus. We load a JavaScript snippet into KWin that enumerates all windows and prints their info. Steps:

1. Write a KWin JavaScript to a temp file:

   ```javascript
   var clients = workspace.windowList();
   for (var i = 0; i < clients.length; i++) {
       var c = clients[i];
       print(c.internalId + "\t" + c.caption + "\t" + c.resourceClass);
   }
   ```

2. Load it via DBus:

   ```text
   gdbus call --session --dest org.kde.KWin
     --object-path /Scripting
     --method org.kde.kwin.Scripting.loadScript <file>
   ```

3. Run it, read output from `journalctl --user --since <timestamp>`

4. For **focus**, a similar script sets `workspace.activeWindow = client`

### Why journalctl?

KWin's `print()` in scripts outputs to the systemd journal. There is no direct stdout pipe. We use a timestamp to only read entries from our script run.

### Window IDs

- **X11**: numeric IDs (e.g. `14680071`)
- **KWin/Wayland**: UUIDs (e.g. `{4bc2d829-80d0-4664-830d-e8a19ef77da7}`)

The `Window.ID` field is `string` to support both.

## Keyboard Input on Wayland

### Problem

`xdotool type` sends X11 key events — these don't reach native Wayland windows.

### Solution 1: `/dev/uinput` (type command)

The `uinput` kernel module allows creating **virtual input devices**. Key events sent through uinput are handled by the kernel input subsystem and delivered to whatever window has focus, regardless of X11 or Wayland.

```bash
ptx-agent → /dev/uinput → kernel → compositor → focused window
```

**Limitation**: uinput sends raw **scancodes** (hardware key codes). These are mapped through the active keyboard layout. If the user has a Czech layout, scancode 30 (KEY_A on US layout) produces a different character. This means `type` may produce incorrect characters on non-US layouts.

**Requirement**: `/dev/uinput` needs read/write access. Run `ptx-agent setup` to configure udev rules and group membership.

### Solution 2: Clipboard paste (paste command)

To bypass keyboard layout issues, the `paste` command uses a different approach:

1. Copy text to the system clipboard via **Klipper DBus** (KDE's clipboard manager):

   ```text
   gdbus call --session --dest org.kde.klipper
     --object-path /klipper
     --method org.kde.klipper.klipper.setClipboardContents <text>
   ```

2. Send **Ctrl+V** via uinput (only 2 keys — layout-independent)

This works correctly with **any keyboard layout** because the text goes through the clipboard, not through scancode-to-character mapping.

### Clipboard fallback chain

1. **Klipper DBus** (KDE — no extra tools needed)
2. **wl-copy** (generic Wayland)
3. **xclip** (X11)

## Backend Detection

### Windows

```text
KWin Wayland running? → KWinBackend
X11 display available? → XdotoolBackend
Neither? → error
```

### Keyboard

```text
/dev/uinput accessible? → UinputBackend
xdotool in PATH?        → XdotoolBackend
Neither?                → error
```

## Commands Summary

| Command | What it does | Backend used |
| ------- | ------------ | ------------ |
| `windows list` | List all windows | KWin DBus / xdotool |
| `windows find <q>` | Search by name | KWin DBus / xdotool |
| `windows focus <id>` | Activate window | KWin DBus / xdotool |
| `keyboard type <text>` | Type via scancodes | uinput / xdotool |
| `keyboard paste <text>` | Paste via clipboard | Klipper DBus + uinput |
| `keyboard key <combo>` | Send key combination | uinput / xdotool |
| `keyboard --window <w> paste <text>` | Focus window + paste | KWin + Klipper + uinput |

## Extending for Other Platforms

To add a new backend (e.g. Windows, Sway, GNOME/Mutter):

1. Implement the `Backend` interface in a new file (e.g. `backend_sway.go`)
2. Add detection logic to `detect.go`

The CLI and Manager/Controller layers remain unchanged.
