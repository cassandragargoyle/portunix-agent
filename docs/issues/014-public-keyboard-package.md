# Issue #014 - Public keyboard package with a persistent device session

**Type**: enhancement
**Priority**: high
**Status**: Open
**GitHub**: #14
**Branch**: `feature/014-public-keyboard-package`
**Labels**: enhancement, keyboard, library
**Requested by**: portunix-archon issue #1

## Problem Description

`portunix-archon` (`ptx-archon`) controls applications by simulating user input. It must not
implement keyboard simulation again. It will use the keyboard implementation of `portunix-agent`
as a **Go library**.

Three things block this today:

1. The keyboard code is in `src/internal/keyboard/`. Go does not allow another module to import
   a package under `internal/`.
2. `go.mod` is in `src/`, but the module path is `github.com/CassandraGargoyle/portunix-agent`.
   A remote `go get` looks for `go.mod` in the repository root and does not find it.
3. The uinput backend creates a new virtual device for every `Type` and `Key` call and waits
   200 ms until the system registers it (`createDevice` in `backend_uinput.go`). A caller that
   sends many actions pays this delay every time.

## Requirements

### 1. Move the keyboard package to a public location

- Move `src/internal/keyboard/` to `src/pkg/keyboard/` with `git mv` (keeps history).
- Update the three importers: `src/internal/cli/keyboard.go`, `src/internal/cli/scenario.go`,
  `src/internal/mcp/handlers_linux.go`.
- The exported API that exists today stays unchanged: `Backend`, `Controller`, `NewController`,
  `NewControllerWithBackend`, `DefaultTypingDelay`, `NormalizeKeyCombination`, the clipboard
  methods.

### 2. Add a persistent device session

Add an API that opens the device once, sends many events, and closes it at the end.

```go
// Session is an open keyboard device that accepts many actions
type Session interface {
	// Type sends a text string as keyboard input
	Type(text string, delayMs int, noIndent bool) error
	// Key sends a key combination (e.g. "ctrl+c", "alt+Tab")
	Key(combo string) error
	// Close releases the device
	Close() error
}

// Open creates a session on the active backend
func (c *Controller) Open() (Session, error)
```

Behavior:

- **uinput**: `Open` creates the virtual device once (one 200 ms registration delay). `Type`
  and `Key` on the session reuse that device. `Close` destroys it.
- **xdotool**: there is no device. The session delegates every call to the backend. `Close`
  does nothing.
- `Session.Type` and `Session.Key` apply the same validation as `Controller.Type` and
  `Controller.Key` (empty text, empty combination, default delay).
- A call on a closed session returns an error. A second `Close` is safe.
- `Controller.Type` and `Controller.Key` stay as convenience wrappers: open a session, do one
  action, close. The behavior of `ptx-agent keyboard` and of the MCP handlers does not change.
- A custom `Backend` given to `NewControllerWithBackend` does not have to know about sessions.
  If it does not provide one, `Open` returns a session that delegates to the backend. This keeps
  fake backends in tests simple.

### 3. Make the module importable

Align the module path with the location of `go.mod`, so that a remote import works. Two options:

| Option | Change | Import path of the package |
| ------ | ------ | -------------------------- |
| A | Module path becomes `github.com/CassandraGargoyle/portunix-agent/src` | `.../portunix-agent/src/pkg/keyboard` |
| B | Move `go.mod` and `go.sum` to the repository root | `.../portunix-agent/src/pkg/keyboard` |

Both options give the same import path for consumers and both need a rewrite of the import
paths inside this repository. Option A does not touch the build (`Makefile`, release scripts,
GoReleaser run from `src/`). Option B needs changes there.

This requirement can be delivered later than requirements 1 and 2. Until then `portunix-archon`
uses a `replace` directive that points to a local checkout of this repository.

### 4. Optional

- A stub for platforms other than Linux (`//go:build !linux`) where `NewController` returns a
  clear "not supported" error. Today a non-Linux build of an importer fails at compile time.
- A configurable name of the virtual device (today it is always `ptx-agent virtual keyboard`).

## Acceptance Criteria

- [ ] `src/internal/keyboard/` does not exist. The package is in `src/pkg/keyboard/`.
- [ ] Another Go module can import the package (verified with a `replace` directive from
      `portunix-archon`).
- [ ] `Controller.Open()` returns a `Session` with `Type`, `Key`, and `Close`.
- [ ] With the uinput backend, a session with N actions creates exactly one virtual device.
- [ ] `Controller.Type` and `Controller.Key` behave as before.
- [ ] `ptx-agent keyboard type`, `ptx-agent keyboard key`, `ptx-agent keyboard paste`, scenarios,
      and the MCP keyboard tools work as before.
- [ ] Unit tests cover the session with a fake backend: delegation, validation, use after
      `Close`, double `Close`.
- [ ] `make build`, `make test`, and `make lint` pass.
- [ ] Requirement 3: the module path is aligned, or the decision to do it later is recorded here.

## Testing

- **Unit tests**: session behavior with a fake backend. No real input device is needed.
- **Manual test on Linux**: type a text of 100 characters through one session and confirm that
  the device registration delay occurs once. Repeat `ptx-agent keyboard type` and
  `ptx-agent keyboard key` to confirm that there is no regression.
- uinput needs write access to `/dev/uinput`.

## Out of Scope

- Mouse simulation.
- Keyboard simulation on Windows.
- Moving other packages out of `internal/`.

## Decisions

- **Requirement 3 is delivered later** (decided 2026-10-11). The module path stays
  `github.com/CassandraGargoyle/portunix-agent` for now. `portunix-archon` uses a `replace`
  directive that points to `<checkout>/src` and imports
  `github.com/CassandraGargoyle/portunix-agent/pkg/keyboard`. Option A is the preferred way to
  align the module path. The import path for consumers then changes to `.../src/pkg/keyboard`.
- **Requirement 4 (optional) is not part of this delivery.**
- A backend provides its own session through the optional `SessionBackend` interface
  (`OpenSession() (Session, error)`). The uinput backend implements it.

## References

- `src/internal/keyboard/controller.go` — `Backend` interface and `Controller`
- `src/internal/keyboard/backend_uinput.go` — `createDevice`, the 200 ms delay
- portunix-archon issue #1 — <https://github.com/cassandragargoyle/portunix-archon/issues/1>

---

**Created**: 2026-10-10
**Author**: Developer (portunix-archon)
