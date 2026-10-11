# Acceptance Protocol - Issue #016

**Issue**: uinput backend drops characters that need a dead key without an error
**Branch**: `fix/issue-16-uinput-dead-key-characters` (tested as the working tree on top of `08cb2c0`, committed without changes as `cf147d5`)
**Tester**: zdendaku (QA/Test Engineer)
**Date**: 2026-10-11
**Testing OS**: Ubuntu 26.04.1 (host, amd64, kernel 7.0.0-34, KDE on X11) — Go 1.26.0, GNU Make 4.4.1

## Test Summary

- Total test scenarios: 24
- Passed: 24
- Failed: 0
- Blocked: 0

## Test Plan

The tested change is the uncommitted working tree of the branch (5 modified files, no untracked
file). The tests ran on a copy of `src/` and on binaries built from it. A temporary worktree of
`main` (commit `08cb2c0`) is the reference for the reproduction and for the regression tests.

| Layer | Command or method |
| --- | --- |
| Static | `go vet ./...`, `gofmt -s -l .`, `GOOS=windows go build ./...` |
| Unit | `go test -count=1 -race -coverprofile ./...` in `src/`; new tests repeated 20 times |
| Keymap regression | The reverse keymap of the real `cz(winkeys)` layout is dumped 20 times on `main` and 20 times on the branch; the value sets are compared for each character |
| Failure injection | Temporary tests in the copy of the package (fake device, no reverse map, real `us` layout) |
| Black box | A separate Go module with a `replace` directive calls `Controller.Type` and `Session.Type` |
| E2E on a real device | `ptx-agent keyboard type` and the probe program on `/dev/uinput`; a temporary window receives the text and the received text is compared byte by byte |

Test host: X11 session, layout `cz`, variant `winkeys`, input method `ibus`, write access to
`/dev/uinput`. This is the configuration of the issue report. The detected backend is `uinput`.

E2E method: a temporary Tk 8.6.17 window (`wish`) with a text field takes the focus and writes its
content to a file. The script checks that the receiver has the focus before it sends a key, and
stops if it has not. One run repeats the main case with a Qt receiver (`kdialog --inputbox`,
kdialog 25.12.3). Each run takes the focus of the desktop for about 2 seconds. No test text
contains a newline.

## Test Cases (Given / When / Then)

### TC-01 — Reproduction on `main`

- **Given** `ptx-agent` from `main` **When** `keyboard type "Příliš žluťoučký kůň úpěl ďábelské ódy"`
  runs **Then** the characters `ť`, `ň`, `ď`, `ó` are lost and the exit code is 0 (the reported bug).

### TC-02 — The issue text is typed completely (criterion 2)

- **Given** `ptx-agent` from the branch **When** the same command runs with `--delay 20`
  **Then** the receiver has exactly the input text and the exit code is 0.

### TC-03 — No delay between the keys

- **Given** the branch **When** the command runs with `--delay 0` **Then** the text is complete.

### TC-04 — Extended character set

- **Given** the branch **When** `ŤŇĎÓ ťňďó Ěščřžýáíéúů ĚŠČŘŽÝÁÍÉÚŮ äöü ôâ ľĺŕ` is typed
  **Then** the text is complete (upper case, and the dead keys caron, acute, ring, diaeresis,
  circumflex).

### TC-05 — Qt receiver

- **Given** a Qt input box **When** the issue text is typed **Then** the text is complete.

### TC-06 — Unmapped character gives an error, no key is sent (criterion 1)

- **Given** the branch **When** `ab🙂cd` is typed **Then** the exit code is 1, the error names
  the character and its position, and the receiver is empty.
- **Reference**: `main` types `abcd` and exits with 0.

### TC-07 — Unmapped character at the end of a long text

- **Given** the branch **When** the issue text followed by ` ж` is typed **Then** the error names
  `ж` at position 40 and the receiver is empty (all or nothing).

### TC-08 — `Controller.Type` from another module (criterion 3)

- **Given** the probe module **When** `Controller.Type` gets the issue text **Then** it returns
  `nil` and the text is complete.
- **And When** it gets `ťň中` **Then** it returns the error and the receiver is empty.

### TC-09 — `Session.Type` from another module (criterion 3)

- **Given** one open session **When** `Type("ťó")`, `Type("x🙂y")`, `Type("ďň")` are called
  **Then** the results are `nil`, error, `nil` and the receiver has `ťóďň`.

### TC-10 — Unit tests (criterion 4)

- **Given** `xkb_test.go` **Then** it covers a dead key sequence with the `cz` fixture, the whole
  issue text, and three unmapped characters (emoji, other alphabet, first position).

### TC-11 — Direct mappings do not change (criterion 5)

- **Given** the real `cz(winkeys)` layout **When** the reverse keymap of `main` and of the branch
  are compared **Then** each character that `main` maps has the same single key stroke on the
  branch, and each new entry is a two stroke sequence.

### TC-12 — The choice of the dead key is stable

- **Given** 20 builds of the reverse keymap **Then** each dead key sequence is the same in all
  builds.

### TC-13 — Build, test, lint

- **When** `go vet`, `gofmt -s -l`, `go test -race ./...` run on the copy **Then** all pass.

### TC-14 — Windows cross build

- **When** `GOOS=windows go build ./...` runs **Then** the build passes.

### TC-15 — Changelog

- **Given** `CHANGELOG.md` **Then** `### Fixed in Unreleased` has an English entry with `(Issue #16)`.

### Failure injection

- **FI-1** — **Given** a backend without a reverse map (layout detection failed) **When** `abě`
  is typed **Then** the error names position 3 and no key is sent; ASCII text is still typed
  with the US fallback.
- **FI-2** — **Given** the real `us` layout **Then** the reverse keymap has no dead key sequence
  and `žluťoučký` gives an error for `ž` at position 1.
- **FI-3** — **Given** a device that fails on the third stroke **When** `aťb` is typed
  **Then** the device error reaches the caller and the message names `ť`.
- **FI-4** — **Given** an empty text **Then** no key is sent and there is no error.
  **Given** `ť` and a newline with `noIndent` **Then** the strokes are dead key, `t`, Enter, Home.
- **FI-5** — **Given** 50 repeated calls with dead key characters **Then** the stored mapping
  does not change (the dead key mapping is shared by pointer).
- **FI-6** — **Given** the real `cz(winkeys)` layout **Then** each dead key sequence has a dead
  key and a base key, and no dead key is nested.
- **FI-7** — **Given** a text with an invalid UTF-8 byte **When** it is typed
  **Then** the error names `U+FFFD` and no key is sent.
- **FI-8** — **Given** a session where one `Type` call failed with an unmapped character
  **Then** the session stays usable (see TC-09).
- **FI-9** — **When** the unit tests run with `-race` **Then** there is no data race report.

## Test Results

| ID | Result | Observed |
| --- | --- | --- |
| TC-01 | PASS | received `Příliš žluoučký ků úpěl ábelské dy`, exit 0, reverse map with 113 entries |
| TC-02 | PASS | received text is equal to the input, exit 0, reverse map with 226 entries |
| TC-03 | PASS | received text is equal to the input |
| TC-04 | PASS | received text is equal to the input |
| TC-05 | PASS | `kdialog` returned the complete text |
| TC-06 | PASS | `cannot type character "🙂" (U+1F642) at position 3: no key mapping in the active keyboard layout`, exit 1, receiver empty |
| TC-07 | PASS | `cannot type character "ж" (U+0436) at position 40: ...`, exit 1, receiver empty |
| TC-08 | PASS | `Controller.Type: <nil>` with the complete text; `"中" (U+4E2D) at position 3` with an empty receiver |
| TC-09 | PASS | `<nil>`, `"🙂" (U+1F642) at position 2`, `<nil>`, `Close: <nil>`; received `ťóďň` |
| TC-10 | PASS | `TestKeysForChar_DeadKey`, `TestTypeOn_DeadKeyText`, `TestTypeOn_UnmappedCharacter`, `TestDeadKeyCompositions_Unique` pass, also with `-count=20` |
| TC-11 | PASS | 113 characters on `main`, 226 on the branch; no character is missing; no branch value that `main` never produced; 113 new entries, all with two strokes |
| TC-12 | PASS | all 113 dead key sequences are identical in 20 builds (`ť` = Shift+AE12, `t`; `ó` = AE12, `o`) |
| TC-13 | PASS | `go vet` clean, `gofmt -s -l` empty, 9 packages `ok` with `-race` |
| TC-14 | PASS | build passes |
| TC-15 | PASS | entry is present |
| FI-1 | PASS | error at position 3, 0 strokes; `Hello, World!` gives 13 strokes |
| FI-2 | PASS | `cannot type character "ž" (U+017E) at position 1: ...` |
| FI-3 | PASS | `failed to type character "ť": device write failed` |
| FI-4 | PASS | 0 strokes; strokes Shift+AE12, `t`, Enter, Home |
| FI-5 | PASS | mapping unchanged |
| FI-6 | PASS | 113 sequences checked |
| FI-7 | PASS | `cannot type character "�" (U+FFFD) at position 3: ...`, exit 1, receiver empty |
| FI-8 | PASS | third call typed `ďň` |
| FI-9 | PASS | no race report |

### Functional Tests

- [x] Feature works as specified
- [x] Acceptance criteria met (5 of 5)
- [x] Edge cases handled (FI-1 to FI-9)

### Acceptance Criteria Mapping

- [x] 1. `Type` with an unmapped character returns an error that names the character, no key is
      sent → TC-06, TC-07, FI-1, FI-7
- [x] 2. On the `cz` layout the issue text is typed completely with the uinput backend → TC-02,
      TC-03, TC-05 (real device, Tk and Qt receivers)
- [x] 3. Same behavior for `Controller.Type` and `Session.Type` → TC-08, TC-09
- [x] 4. Unit tests cover an unmapped character and a dead key sequence with a `cz` reverse
      keymap → TC-10
- [x] 5. Typing of characters with a direct mapping does not change → TC-11

## Coverage

| Scope | `main` | Branch |
| --- | --- | --- |
| Package `pkg/keyboard` | 48.7 % | 53.6 % |
| `strokesForText`, `keysForChar`, `mappingKeys`, `cheaperMapping` | — | 100 % |
| `addDeadKeySequences` | — | 91.7 % |
| `buildReverseMap` | — | 90.3 % |
| `typeOn` | 0 % | 73.3 % (the `noIndent` branch and the device error branch are covered only by FI-3 and FI-4, which are not in the repository) |

The functions that need a real device (`UinputBackend.Type`, `createDevice`, `emit`, `typeKeys`)
stay at 0 % in the unit tests. The E2E tests of this protocol cover them.

Test matrix:

| Area | Unit | Black box | E2E real device |
| --- | --- | --- | --- |
| Dead key sequence | TC-10 | TC-08, TC-09 | TC-02 to TC-05 |
| Unmapped character | TC-10, FI-1, FI-2 | TC-08, TC-09 | TC-06, TC-07, FI-7 |
| Direct mapping regression | existing tests | — | TC-11 (keymap comparison) |
| Device failure | FI-3 | — | — |

## Findings

No finding blocks the merge.

1. **Three failure cases have no test in the repository.** FI-1 (no reverse map), FI-3 (device
   error), and FI-4 (`noIndent` with a dead key character) ran as temporary tests. Recommendation:
   add them to `xkb_test.go`, the fake device (`recordingTyper`) is already there.
2. **`keyboard type` reports the number of bytes.** The command prints `Typed 53 characters.` for
   the issue text, which has 38 characters (`len(text)` in `src/internal/cli/keyboard.go`). The
   behavior is the same on `main`. Recommendation: a separate issue.
3. **The reverse keymap is not stable for 14 ASCII symbols.** For the characters
   `! # % & ' / @ [ ] ^ { | }` and the backtick the layout has two keys with the same number of modifiers, and the choice depends on the map
   order. Both choices type the right character. The behavior is the same on `main` and the dead
   key sequences are not affected (TC-12). Recommendation: a separate issue, `cheaperMapping` can
   be reused for the direct mappings.
4. **The CLI prints the usage text after the unmapped character error.** The error is a runtime
   error, not a usage error. This is the general behavior of the CLI for errors from `RunE`.
5. **Three dead keys are not supported.** `dead_belowdot`, `dead_hook`, and `dead_horn` have no
   entry in the composition table. Their characters give the unmapped character error, so no
   input is lost without an error.
6. **A layout that needs a dead key for a character is the only source of sequences.** A character
   that the layout cannot produce at all (emoji, other alphabet) gives an error. The clipboard
   fallback from step 2 of the proposed solution is not implemented. This was agreed before the
   implementation and it is not an acceptance criterion.
7. **The composition is done by the receiver.** The tests cover Tk with `ibus` and Qt on X11.
   Wayland sessions and receivers without compose support were not tested.
8. **The error behavior changes for callers.** A caller that sent text with unmapped
   characters got success before and gets an error now. The changelog entry describes this.
   Consumers such as `portunix-archon` should handle the error.

## CI Notes

- The change is not pushed, so there is no CI run for it yet. The commands of
  the CI workflow (`go vet`, `gofmt -s`, `go test -race`, the build for Linux and Windows) pass
  locally.
- CI runners have no `/dev/uinput`, so CI cannot repeat TC-02 to TC-09. The unit tests with the
  fake device cover the key stroke sequences.
- The keymap comparison of TC-11 needs the system XKB files and a reference build of `main`. It
  is not a candidate for CI in this form.

## Final Decision

**STATUS**: PASS

**Approval for merge**: YES — after the tested working tree is committed without further changes
and the CI run of that commit passes

Recommended, not required:

1. Add the temporary failure injection tests (finding 1) to the repository.
2. Create issues for findings 2 and 3.

**Merge decision**: The project owner (zdendaku) approved the merge to `main` on 2026-10-11
without waiting for the CI run, as an exception to the condition above. The CI run of the
commit stays open as follow-up work.

**Date**: 2026-10-11
**Tester signature**: zdendaku (QA/Test Engineer)
