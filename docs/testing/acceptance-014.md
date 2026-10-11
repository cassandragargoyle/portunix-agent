# Acceptance Protocol - Issue #014

**Issue**: Public keyboard package with a persistent device session
**Branch**: `feature/014-public-keyboard-package` (commit `2de82b5`)
**Tester**: zdendaku (QA/Test Engineer)
**Date**: 2026-10-11
**Testing OS**: Ubuntu 26.04.1 (host, amd64, kernel 7.0.0-34, KDE on X11) — Go 1.26.0, GNU Make 4.4.1

## Test Summary

- Total test scenarios: 22
- Passed: 19
- Failed: 0
- Blocked: 3 (real uinput device — `/dev/uinput` on the test host is `crw------- root root`
  and the tester has no root access)

## Test Plan

All tests ran in a clean clone of the feature branch, not in the developer working tree.
A second clean clone of `main` (commit `ee00a93`) is the reference for the regression tests.

| Layer | Command or method |
| --- | --- |
| Static | `go vet ./...`, `gofmt -s -l .`, `GOOS=windows go build ./...` |
| Unit | `go test -count=1 -race -coverprofile ./...` in `src/` |
| API compatibility | `go doc -all` of the package on `main` and on the branch, then `diff` |
| Black box | A separate Go module with a `replace` directive, tests run with `-race` |
| E2E regression | `ptx-agent` from `main` and from the branch in an isolated `Xvfb` display; `xev` records the key events; the two event logs are compared |
| Real device | Probe program on `/dev/uinput` (blocked, see UI-01 to UI-03) |

The E2E tests ran in `Xvfb :99`, so no key event reached the desktop of the tester. The session
D-Bus address was set to an invalid value and a fake `xclip` was first in `PATH`, so the real
clipboard was not changed. On this host the detected backend is `xdotool`, because `/dev/uinput`
is not accessible.

## Test Cases (Given / When / Then)

### TC-01 — Package location

- **Given** the branch **When** `src/` is listed **Then** `src/internal/keyboard/` does not exist,
  `src/pkg/keyboard/` exists, and `git log --follow` shows the history before the move.

### TC-02 — Import from another module

- **Given** a copy of `go.mod` from `portunix-archon` with a `replace` directive to `<clone>/src`
  **When** a program that imports `.../portunix-agent/pkg/keyboard` is built and run
  **Then** it compiles and `Controller.Open()` works.

### TC-03 — `Controller.Open()` returns a `Session`

- **Given** a controller **When** `Open()` is called **Then** the result has `Type`, `Key`, and `Close`.

### TC-04 — One device for N actions (fake backend)

- **Given** a fake backend that implements `SessionBackend` **When** one session sends 10 actions
  **Then** the backend opens one device and closes one device.

### TC-05 — `Controller.Type` and `Controller.Key` behave as before

- **Given** the API of `main` **When** it is compared with the branch **Then** no exported symbol
  is removed or changed.
- **And** each one-shot call opens and closes one device, and invalid input opens no device.

### TC-06 — CLI regression: `keyboard type`, `key`, `paste`, `read`

- **Given** `ptx-agent` from `main` and from the branch **When** the same commands run in `Xvfb`
  **Then** the recorded key events, the output, and the exit codes are identical.

### TC-07 — Scenario regression

- **Given** a scenario with two commands (one with `sendEnter`) **When** `scenario run` is used
  **Then** the key events are identical to `main`.

### TC-08 — MCP regression

- **Given** `mcp serve` **When** `agent_app_type` and `agent_app_key` are called
  **Then** the responses and the key events are identical to `main`.

### TC-09 — Unit tests cover the session

- **Given** `session_test.go` **Then** it covers delegation, validation, use after `Close`,
  and double `Close` with a fake backend.

### TC-10 — Build, test, lint

- **When** `make build`, `make test`, `make lint` run in the clean clone **Then** all pass.

### TC-11 — Requirement 3 decision

- **Given** the issue file **Then** it records the decision about the module path.

### TC-12 — Windows cross build

- **When** `GOOS=windows go build ./...` runs **Then** the build passes.

### Failure injection

- **FI-1** — **Given** a backend that returns an error **When** `Session.Type`, `Session.Key`,
  and `Controller.Type` are called **Then** the same error reaches the caller and the session
  stays usable.
- **FI-2** — **Given** a backend where `OpenSession` fails **When** `Open`, `Controller.Type`,
  and `Controller.Key` are called **Then** all return that error and `Open` returns a nil session.
- **FI-3** — **Given** a one-shot call where the action fails **Then** the device is still closed.
- **FI-4** — **Given** a device where `Close` fails **Then** the first `Close` returns the error,
  the second `Close` returns nil, and the device is closed once.
- **FI-5** — **Given** 150 goroutines that call `Type`, `Key`, and `Close` on one session
  **When** the test runs with `-race` **Then** there is no data race and the device is closed once.
- **FI-6** — **Given** two sessions from one controller **When** the first is closed
  **Then** the second still works.
- **FI-7** — **Given** no access to `/dev/uinput` **When** a session is opened on `UinputBackend`
  **Then** `Open` returns `cannot open /dev/uinput: ... permission denied` and does not panic.

### Real uinput device (blocked)

- **UI-01** — **Given** write access to `/dev/uinput` **When** one session sends 20 key actions
  **Then** `/proc/bus/input/devices` shows the same single `ptx-agent virtual keyboard` during
  all actions, the device is gone after `Close`, and there is one 200 ms delay, not 20.
- **UI-02** — **Given** a focused text editor **When** 100 characters are typed through one
  session **Then** all characters arrive and the registration delay occurs once.
- **UI-03** — **Given** the uinput backend is detected **When** `ptx-agent keyboard type`,
  `key`, and `paste` are repeated **Then** they work as on `main`.

## Test Results

| ID | Result | Observed |
| --- | --- | --- |
| TC-01 | PASS | `src/internal/` has no `keyboard`; `src/pkg/keyboard/` has 13 files; `git log --follow` on `controller.go` reaches `eb61b74 Initial public release` |
| TC-02 | PASS | module `github.com/cassandragargoyle/portunix-archon` (copy of `go.mod` at `88cccc0`) builds; output `backend: xdotool open err: <nil>`, `after close: keyboard session is closed` |
| TC-03 | PASS | `var s keyboard.Session = c.Open()` compiles in an external module |
| TC-04 | PASS | `TestSession_OneDeviceForManyActions`: 1 open, 1 close, 10 actions |
| TC-05 | PASS | API diff shows additions only (`Open`, `Session`, `SessionBackend`, `UinputBackend.OpenSession`); `TestController_OpensDevicePerCall` passes |
| TC-06 | PASS | 58 key events on `main`, 58 on the branch, `diff` empty; output and exit codes identical |
| TC-07 | PASS | events `a b Return c d`, identical to `main` |
| TC-08 | PASS | `Typed 3 characters`, `Sent key combination: alt+Tab`, empty text gives `isError: true`; identical to `main` |
| TC-09 | PASS | 6 new tests pass; `session.go` coverage is 100 % for `Type`, `Key`, `Close` and 85.7 % for `Open` |
| TC-10 | PASS | `make build`, `make test` (9 packages `ok`), `make lint` pass; `gofmt -s -l` empty; clone stays clean |
| TC-11 | PASS | section `Decisions` in the issue file: requirement 3 is delivered later, option A is preferred |
| TC-12 | PASS | `GOOS=windows go build ./...` passes |
| FI-1 | PASS | error identity kept (`errors.Is`) |
| FI-2 | PASS | `Open` returns `nil, err` |
| FI-3 | PASS | 2 opens, 2 closes |
| FI-4 | PASS | closed once, later `Key` returns an error |
| FI-5 | PASS | no race report, 1 open, 1 close |
| FI-6 | PASS | 2 opens, 2 closes |
| FI-7 | PASS | `cannot open /dev/uinput: open /dev/uinput: permission denied` |
| UI-01 | BLOCKED | probe program is built, it stops with `BLOCKED: cannot open /dev/uinput` |
| UI-02 | BLOCKED | no access to `/dev/uinput` |
| UI-03 | BLOCKED | no access to `/dev/uinput` |

### Functional Tests

- [x] Feature works as specified (session API, xdotool backend, fake backends)
- [ ] Acceptance criteria met (8 of 9 complete, see mapping)
- [x] Edge cases handled (FI-1 to FI-7)

### Acceptance Criteria Mapping

- [x] `src/internal/keyboard/` does not exist, the package is in `src/pkg/keyboard/` → TC-01
- [x] Another Go module can import the package → TC-02 (copy of the `portunix-archon` module
      definition; the `portunix-archon` repository itself was not changed)
- [x] `Controller.Open()` returns a `Session` with `Type`, `Key`, and `Close` → TC-03
- [ ] With the uinput backend, a session with N actions creates exactly one virtual device →
      logic verified with a fake backend (TC-04); the real device is not verified (UI-01)
- [x] `Controller.Type` and `Controller.Key` behave as before → TC-05, TC-06
- [x] CLI commands, scenarios, and MCP keyboard tools work as before → TC-06, TC-07, TC-08
      (xdotool backend only; uinput backend → UI-03)
- [x] Unit tests cover the session with a fake backend → TC-09
- [x] `make build`, `make test`, and `make lint` pass → TC-10
- [x] Requirement 3: the decision to do it later is recorded → TC-11

### Regression Tests

- [x] Existing functionality unaffected — key events identical to `main` on the xdotool backend
- [ ] uinput backend verified on a real device — blocked
- [x] Cross-platform compatibility — Windows cross build passes; the package is Linux only,
      as before

## Coverage

| Scope | Value |
| --- | --- |
| `pkg/keyboard` (statements) | 44.6 % |
| `session.go` | `Open` 85.7 %, all other functions 100 % |
| `controller.go` | `Type` 87.5 %, `Key` 85.7 %, validation helpers 100 % |
| `backend_uinput.go` device code | 0 % — no automated test can reach a real device |

The uncovered lines in `session.go` and `controller.go` are the `OpenSession` error branches.
The black box test FI-2 covers them from outside the package.

Proposed matrix for this package:

| Layer | Scope | Where |
| --- | --- | --- |
| Unit | session logic with fake backends | `go test ./...`, CI |
| Integration | real uinput device, one device per session | Linux host with access to `/dev/uinput` |
| E2E | CLI and MCP key events in `Xvfb` (xdotool backend) | Linux host, can run in CI |

## Raw Results

```text
$ go test -count=1 -race ./...        (clean clone, src/)
ok  github.com/CassandraGargoyle/portunix-agent/internal/cli     1.013s
ok  github.com/CassandraGargoyle/portunix-agent/internal/mcp     1.170s
ok  github.com/CassandraGargoyle/portunix-agent/pkg/keyboard     1.021s  coverage: 44.6% of statements
(9 packages ok, 0 failed)

$ go test -count=1 -v -run 'Session|Controller' ./pkg/keyboard/
--- PASS: TestController_DelegatesToBackend (0.00s)
--- PASS: TestSession_DelegatesToBackend (0.00s)
--- PASS: TestSession_Validation (0.00s)
--- PASS: TestSession_UseAfterClose (0.00s)
--- PASS: TestSession_DoubleClose (0.00s)
--- PASS: TestSession_OneDeviceForManyActions (0.00s)
--- PASS: TestController_OpensDevicePerCall (0.00s)

$ go test -count=1 -race -v ./...     (external module, replace => <clone>/src)
--- PASS: TestImportAndAPI (0.00s)
--- PASS: TestBackendErrorPropagates (0.00s)
--- PASS: TestOpenFailure (0.00s)
--- PASS: TestOneShotClosesOnError (0.00s)
--- PASS: TestCloseError (0.00s)
--- PASS: TestConcurrentUse (0.00s)
--- PASS: TestTwoSessions (0.00s)
--- PASS: TestUinputNoAccess (0.00s)
--- PASS: TestDelayDefault (0.00s)

$ diff ev_main.txt ev_branch.txt      (xev key events in Xvfb, 58 lines each)
(no output)

Key events on the branch (+ press, - release):
+Shift_L +H -Shift_L -h +e -e +l -l +l -l +o -o +comma -comma +space -space +1 -1
+Shift_L +exclam -Shift_L -1 +Control_L +a -Control_L -a +Control_L +Shift_L +Return
-Shift_L -Control_L -Return +Control_L +v -Control_L -v +a -a +b -b +Return -Return
+c -c +d -d +m -m +c -c +p -p +Alt_L +Tab -Alt_L -Tab +x -x

$ ./uinput-probe
BLOCKED: cannot open /dev/uinput: open /dev/uinput: permission denied
```

## Observations and Recommendations

None of these is a defect of this branch.

1. **No automated test reaches the uinput device code.** The refactored functions `typeOn`,
   `keyOn`, and `OpenSession` in `backend_uinput.go` run only on a real device. Recommendation:
   add an integration test with a build tag (for example `uinput`) that opens a session, sends
   the Shift key N times, and counts the devices in `/proc/bus/input/devices`.
2. **A session taken directly from `UinputBackend.OpenSession()` has no closed-state check.**
   The check is in the wrapper that `Controller.Open()` returns. A direct caller gets an OS
   error after `Close` instead of `keyboard session is closed`. Recommendation: say in the doc
   comment that consumers use `Controller.Open()`.
3. **`Close` on a uinput session always returns nil.** The result of the destroy `ioctl` and of
   the file close is ignored, as it was before this issue.
4. **xdotool accepts an unknown key name.** `ptx-agent keyboard key nosuchkey+x` exits with 0 and
   sends only `x`. The behavior is the same on `main`. Recommendation: a separate issue.
5. **There is no documentation for library users.** `README.md` does not describe
   `pkg/keyboard`, the `replace` directive, or the need for access to `/dev/uinput`.
6. **The `References` section of the issue file still points to `src/internal/keyboard/`.**
7. **On a host without access to `/dev/uinput` the backend is always xdotool.** The session then
   brings no speed gain. `ptx-agent setup` installs the udev rule that gives access.

## CI Notes

- The branch is not pushed. No CI run exists for commit `2de82b5`.
- The CI workflow runs `go vet`, `gofmt -s`, `go test -race`, and the build on `ubuntu-latest`
  and `windows-latest`. The same commands pass locally.
- CI runners have no `/dev/uinput`, so CI cannot close UI-01 to UI-03.
- The E2E method of this protocol (`Xvfb` + `xev` + `xdotool`) needs no privileges and can be
  added to CI later.

## Final Decision

**STATUS**: CONDITIONAL

**Approval for merge**: NO — pending the verification on a real uinput device

Required before approval:

1. Run UI-01 on a host with write access to `/dev/uinput` (probe program, harmless Shift key only).
2. Run UI-02: type 100 characters through one session into a text editor.
3. Run UI-03: repeat `ptx-agent keyboard type`, `key`, and `paste` with the uinput backend.
4. Push the branch and record the CI result.

**Date**: 2026-10-11
**Tester signature**: zdendaku (QA/Test Engineer)
