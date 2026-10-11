# Acceptance Protocol - Issue #017

**Issue**: xdotool backend reports success for an unknown key name
**Branch**: `fix/issue-17-xdotool-unknown-key-name` (commit `7779b4b`, clean working tree)
**Tester**: zdendaku (QA/Test Engineer)
**Date**: 2026-10-11
**Testing OS**: Ubuntu 26.04.1 (host, amd64, kernel 7.0.0-34, KDE on X11) — Go 1.26.0, GNU Make 4.4.1, xdotool 3.20160805.1

## Test Summary

- Total test scenarios: 17
- Passed: 17
- Failed: 0
- Blocked: 0

## Test Plan

The tested change is commit `7779b4b` (3 files). An export of `main` (commit `d9aa6a4`) is the
reference for the reproduction and for the regression tests. Both binaries are built with
`go build ./cmd/ptx-agent/`.

| Layer | Command or method |
| --- | --- |
| Static | `go vet ./...`, `gofmt -s -l .`, `GOOS=windows go build ./...` |
| Unit | `go test -count=1 -race -coverprofile ./...` in `src/`; new tests repeated 20 times |
| Failure injection | A fake `xdotool` in `PATH` (records its arguments, or exits with 3), no X server, no backend |
| Black box | A separate Go module with a `replace` directive calls `Controller.Key` and `Session.Key` |
| E2E with the real xdotool | `ptx-agent keyboard key` in a virtual X server; a recorder lists the received key events |

E2E method: the tests do not run on the desktop session. A virtual X server (`Xvfb :99`, layout
`us`) receives the keys, and `xev -root -event keyboard` records each key press with its keysym
and modifier state. The host gives write access to `/dev/uinput`, so the detected backend is
`uinput`. To get the xdotool backend, `ptx-agent` runs in a `bwrap` sandbox (bubblewrap 0.11.1)
that replaces `/dev/uinput` with a file without access rights. This is the condition of the
manual test in the issue ("without access to `/dev/uinput`").

## Test Cases (Given / When / Then)

### TC-01 — Reproduction on `main`

- **Given** `ptx-agent` from `main` with the xdotool backend **When** `keyboard key ctrl+nosuchkey`
  runs **Then** the exit code is 0, the message is `Sent key combination: ctrl+nosuchkey`, and
  only `Control_L` is received (the reported bug).

### TC-02 — Unknown key name gives an error (criterion 1)

- **Given** `ptx-agent` from the branch with the xdotool backend **When** `keyboard key
  ctrl+nosuchkey` runs **Then** the exit code is 1 and the error names the key and the combination.

### TC-03 — No key is sent for an invalid combination (criterion 2)

- **Given** the branch **When** 25 invalid combinations are sent (unknown name at the first, the
  middle, and the last position, empty parts, X keysym names that are not in the key table,
  a non-ASCII letter, wrong separators) **Then** each one gives an error and the recorder has no
  key event.
- **Reference**: `main` reports success for 18 of them. It sends only the valid part for 9, it
  sends nothing for 3, and 6 are X keysym names that `main` passed to xdotool (finding 3).

### TC-04 — Same error message for both backends (criterion 3)

- **Given** the branch **When** 8 invalid combinations are sent once with the uinput backend and
  once with the xdotool backend **Then** the `Error:` lines are equal byte by byte.

### TC-05 — Valid combinations work as before (criterion 4)

- **Given** 22 valid combinations, with `ctrl+s`, `alt+Tab`, `Return` among them **When** they are
  sent on `main` and on the branch **Then** the received key events are the same for each
  combination that `main` sent completely.

### TC-06 — Each accepted key name reaches xdotool with a name it knows

- **Given** the 72 key names of `comboKeyMap` **When** each one is sent alone on the branch
  **Then** the exit code is 0 and the expected keysym is received.
- **Reference**: the same names on `main`.

### TC-07 — Unit tests (criterion 5)

- **Given** `keyboard_test.go` **Then** it covers the validation of valid and invalid
  combinations without an X server, and the completeness of the xdotool key name table.

### TC-08 — `Controller.Key` and `Session.Key` from another module

- **Given** the probe module with `NewXdotoolBackend` **When** `Controller.Key` and `Session.Key`
  get a valid and an invalid combination **Then** the valid ones return `nil` and are received,
  the invalid ones return the error, and the session stays usable after an error.

### TC-09 — Build, test, lint

- **When** `go vet`, `gofmt -s -l`, `go test -race ./...` run on the branch **Then** all pass.

### TC-10 — Windows cross build

- **When** `GOOS=windows go build ./...` runs **Then** the build passes.

### TC-11 — Changelog

- **Given** `CHANGELOG.md` **Then** `### Fixed in Unreleased` has an English entry with
  `(Issue #17)` that also describes the smaller set of key names.

### Failure injection

- **FI-1** — **Given** a fake `xdotool` that records its arguments **When** 5 invalid combinations
  are sent **Then** the fake is never called. **When** `ctrl+s`, `Alt + F4`, `PAGEUP`, `-` are sent
  **Then** the recorded arguments are `key -- ctrl+s`, `key -- alt+F4`, `key -- Prior`,
  `key -- minus`.
- **FI-2** — **Given** an `xdotool` that exits with 3 **When** `ctrl+s` is sent **Then** the error
  `failed to send key combination` reaches the caller. **When** `ctrl+nosuchkey` is sent
  **Then** the validation error comes first.
- **FI-3** — **Given** no X server (`DISPLAY=:77`) **When** `ctrl+s` is sent **Then** the exit code
  is 1. **When** `ctrl+nosuchkey` is sent **Then** the validation error is reported.
- **FI-4** — **Given** no access to `/dev/uinput` and no `xdotool` in `PATH` **Then** the error is
  `no supported keyboard backend found`.
- **FI-5** — **Given** combinations that look like options or shell input (`--help`,
  `ctrl+s;xterm`, `ctrl+s` + newline + `Return`) **Then** they are rejected by the validation and
  never reach xdotool.
- **FI-6** — **When** the new unit tests run 20 times with `-race` **Then** all pass and there is
  no data race report.

## Test Results

| ID | Result | Observed |
| --- | --- | --- |
| TC-01 | PASS | exit 0, `Sent key combination: ctrl+nosuchkey`, received `Control_L` |
| TC-02 | PASS | `Error: unknown key "nosuchkey" in combination "ctrl+nosuchkey"`, exit 1 |
| TC-03 | PASS | 25 of 25 give exit 1 and no key event (for example `nosuchkey+ctrl`, `ctrl+`, `+s`, `ctrl+ab`, `Print`, `ctrl+ž`, `ctrl-s`) |
| TC-04 | PASS | 8 of 8 equal (for example `unknown key "print" in combination "Print"`, `unknown key "" in combination "ctrl+"`) |
| TC-05 | PASS | 18 combinations have the same events on both builds (`ctrl+s` → `Control_L`, `s`; `alt+Tab` → `Alt_L`, `Tab`; `Return` → `Return`); 4 that were incomplete on `main` are complete now (`ALT+TAB`, `PageUp`, `alt+f4`, `ctrl + s`) |
| TC-06 | PASS | 72 of 72 give exit 0 and the expected keysym; 41 names have the same events as on `main`; 31 names that `main` did not send are sent now (`return`, `tab`, `pageup` → `Prior`, `f1`–`f12`, `-` → `minus`, `=` → `equal`, ...) |
| TC-07 | PASS | `TestNormalizeKeyCombination` (17 cases), `TestNormalizeKeyCombination_CoversComboKeyMap`, `TestXdotoolKey_Validation` pass |
| TC-08 | PASS | `unknown key "nosuchkey" in combination "ctrl+nosuchkey"`, `<nil>`, `<nil>`, `unknown key "nosuchkey" in combination "shift+nosuchkey"`, `<nil>`, `Close: <nil>`; received `Control_L s Alt_L Tab Next` |
| TC-09 | PASS | `go vet` clean, `gofmt -s -l` empty, 9 packages `ok` with `-race` |
| TC-10 | PASS | build passes |
| TC-11 | PASS | entry is present |
| FI-1 | PASS | 0 calls for the invalid combinations; the 4 argument lines are as expected |
| FI-2 | PASS | `failed to send key combination "ctrl+s": exit status 3`; validation error for `ctrl+nosuchkey` |
| FI-3 | PASS | `failed to send key combination "ctrl+s": exit status 1`, exit 1; validation error for `ctrl+nosuchkey` |
| FI-4 | PASS | `no supported keyboard backend found (tried: uinput, xdotool)`, exit 1 |
| FI-5 | PASS | `unknown key "--help" ...`, `unknown key "s;xterm" ...`, `unknown key "s\nreturn" ...`, no call of xdotool |
| FI-6 | PASS | `ok` with `-count=20 -race` |

### Functional Tests

- [x] Feature works as specified
- [x] Acceptance criteria met (5 of 5)
- [x] Edge cases handled (FI-1 to FI-6)

### Acceptance Criteria Mapping

- [x] 1. `Key` with an unknown key name returns an error with the xdotool backend → TC-02, TC-08
- [x] 2. No key is sent when the combination is not valid → TC-03, FI-1, FI-5
- [x] 3. The error message has the same form for both backends → TC-04
- [x] 4. Valid combinations (`ctrl+s`, `alt+Tab`, `Return`) work as before → TC-05, TC-06
- [x] 5. Unit tests cover the validation → TC-07

## Coverage

| Scope | `main` | Branch |
| --- | --- | --- |
| Package `pkg/keyboard` | 53.6 % | 56.2 % |
| `XdotoolBackend.Key` | 0 % | 83.3 % (the branch for a failed xdotool call is covered only by FI-2 and FI-3) |
| `comboToKeys` | 0 % | 100 % |
| `NormalizeKeyCombination`, `normalizeKeyName` | 100 % | 100 % |

`NewXdotoolBackend` and `XdotoolBackend.Type` stay at 0 % in the unit tests. The E2E tests of this
protocol cover `NewXdotoolBackend` and `Key` with the real xdotool.

Test matrix:

| Area | Unit | Black box | E2E real xdotool |
| --- | --- | --- | --- |
| Unknown key name | TC-07 | TC-08 | TC-02, TC-03, TC-04 |
| Valid combination regression | TC-07 | TC-08 | TC-05, TC-06 |
| Key name table | TC-07 | — | TC-06 |
| xdotool failure | — | — | FI-2, FI-3, FI-4 |

## Findings

No finding blocks the merge.

1. **The E2E tests ran in a virtual X server, not on the desktop session.** The layout was `us`,
   and there was no window manager. The manual test of the issue on the real KDE session
   (`cz` layout) was not repeated. The validation does not depend on the X server, but the key
   name table was checked only against the `us` keymap of Xvfb.
2. **xdotool adds `Alt_L` to the function keys in the test server.** `f1` to `f12` are received
   as `Alt_L` + `F1`...`F12`. The direct call `xdotool key F1` gives the same result, and so does
   `main` with `F1`. The change does not cause it. It was not checked on the desktop session.
   Recommendation: check `ptx-agent keyboard key f5` on the desktop, and create a separate issue
   if `Alt` is added there too.
3. **The set of key names is smaller for xdotool callers.** X keysym names that `main` passed to
   xdotool now give an error: `Print`, `F13`, `XF86AudioMute`, `period`, and also `Page_Up` and
   `Prior` (the accepted names are `pageup` and `pagedown`). This was agreed before the
   implementation and the changelog entry describes it. Consumers such as `portunix-archon`
   must use the names of the key table.
4. **Upper case letters differ between the backends.** With xdotool `Ctrl+C` sends `Control_L`,
   `Shift_L`, `C`. The uinput backend changes the combination to lower case and sends no `Shift`.
   The behavior is the same on `main`. Recommendation: a separate issue.
5. **The order and the count of keys are not checked.** `s+ctrl`, `ctrl+ctrl`, and `ctrl+a+b`
   are accepted. The behavior is the same on `main` and for the uinput backend.
6. **Spaces around a key name are now accepted.** `ctrl + s` and `ctrl+s ` gave an error on
   `main` with xdotool and are sent as `ctrl+s` now. The uinput backend already did this.
7. **The CLI prints the usage text after the unknown key error.** The error is a runtime error,
   not a usage error. This is the general behavior of the CLI for errors from `RunE` (see
   finding 4 of the protocol for issue #16).
8. **The failed xdotool call has no test in the repository.** FI-2 ran with a temporary fake.
   Recommendation: add a case with a failing fake to `TestXdotoolKey_Validation`.
9. **The MCP tool `agent_app_key` was not tested separately.** It calls `Controller.Key`, which
   TC-08 covers.

## CI Notes

- The change is not pushed, so there is no CI run for it yet. The commands of the CI workflow
  (`go vet`, `gofmt -s`, `go test -race`, the build for Linux and Windows) pass locally.
- The new unit tests use a fake `xdotool` shell script and need no X server, so CI can run them.
- The E2E tests need `Xvfb`, `xev`, `xdotool`, and `bwrap`. They are a candidate for a CI job on
  Linux, because they need no real device and no desktop session.

## Final Decision

**STATUS**: PASS

**Approval for merge**: YES — the merge decision belongs to the project owner

Recommended, not required:

1. Check one function key on the desktop session (finding 2).
2. Add the test for a failed xdotool call (finding 8).
3. Create an issue for finding 4.

**Date**: 2026-10-11
**Tester signature**: zdendaku (QA/Test Engineer)
