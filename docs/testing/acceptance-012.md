# Acceptance Protocol - Issue #012

**Issue**: Implement --help-ai flag
**Branch**: `feature/issue-12-help-ai-flag` (uncommitted working tree on top of `e63acdf`)
**Tester**: zdendaku (QA/Test Engineer)
**Date**: 2026-10-11
**Testing OS**: Ubuntu 26.04 (host, amd64, kernel 7.0.0-34) — Go 1.26.0, GNU Make 4.4.1

## Test Summary

- Total test scenarios: 24
- Passed: 22
- Failed: 0
- Skipped: 2 (execution on Windows — no Windows host available)

## Test Plan

The tests ran in the developer working tree, because the implementation is not committed yet.
Verification combines the automated unit tests, a build with `make build`, black-box runs of the
binary, a comparison of the JSON output with the `--help` output of every command, and failure
injection on the arguments and on stdout.

| Layer | Command |
| --- | --- |
| Unit | `go test -count=1 -race ./...` in `src/` |
| Static | `go vet ./...`, `gofmt -s -l .`, `GOOS=windows go build ./internal/cli/` |
| Integration | `make build`, then run `./ptx-agent --help-ai` |
| Consistency | script that walks `--help` of every command and compares it with the JSON |
| Failure injection | wrong flag values, flag at other positions, broken stdout |

## Test Cases (Given / When / Then)

### TC-01 — Valid JSON on stdout

- **Given** a binary from `make build` **When** `ptx-agent --help-ai`
  **Then** stdout is one valid JSON document, stderr is empty, exit code is 0.

### TC-02 — Top-level fields

- **Given** the output of TC-01 **Then** it has exactly the fields `tool` (`ptx-agent`), `version`,
  `description`, and `commands` (array).

### TC-03 — Version value

- **Given** `make build` **Then** `version` equals the version in `plugin.json`.
- **Given** a build without `-ldflags` (`go run`) **Then** `version` is `dev`.

### TC-04 — All commands are listed

- **Given** the command tree from `--help` of every command **When** it is compared with the JSON
  **Then** every runnable command is in the JSON and the JSON has no extra command.
- **And** the generated commands `help` and `completion` are not listed.

### TC-05 — Nested commands

- **Then** nested commands are listed by path (`terminal read`, `terminal list`, `tab switch`,
  `tab close`, `tab new`) and a group without an action (`terminal`, `tab`) is not listed.

### TC-06 — Parameters are complete

- **Then** every parameter has `name`, `type`, a boolean `required`, and a non-empty `description`.

### TC-07 — Flags match the help text

- **Given** the `Flags` and `Global Flags` sections of `--help` of every command
  **Then** the JSON lists the same flags, without `--help` and `--help-ai`.

### TC-08 — Types, defaults, required

- **Then** `--headless` is `boolean` with default `true`, `--amount` is `integer` with default `3`,
  `--allowed-dirs` is `array` without a default, `index` is a required `integer`,
  `url` of `tab new` and `id` of `tab close` are optional.

### TC-09 — Flag after a subcommand

- **When** `ptx-agent files list --help-ai`, `ptx-agent browser goto --help-ai`,
  `ptx-agent files list /nonexistent --help-ai`
  **Then** the output is byte-identical to TC-01, exit 0, the subcommand does not run and
  its argument validation does not run.

### TC-10 — Output is deterministic

- **When** the command runs two times **Then** the two outputs are byte-identical.

### TC-11 — Flag is visible in help

- **When** `ptx-agent --help` and `ptx-agent files --help` **Then** both list `--help-ai`.

### TC-12 — Regression

- **Then** `--version`, `version`, `--help`, and the full unit test suite still pass.

### FI-1 — `--help-ai=false`

- **Then** the JSON is not printed and the normal behavior follows (root help, exit 0).

### FI-2 — `--help-ai=maybe`

- **Then** the command fails with a clear error on stderr, exit 1, stdout is empty.

### FI-3 — Flag after the `--` terminator

- **When** `ptx-agent version -- --help-ai` **Then** the value is a positional argument, no JSON.

### FI-4 — Misspelled flag

- **When** `--help-a` or `-help-ai` **Then** `unknown flag` error, exit 1, stdout is empty.

### FI-5 — Combination with other flags

- **When** `--help-ai --version` or `--help --help-ai` **Then** only the JSON is printed, exit 0.

### FI-6 — Unknown command or unknown flag together with `--help-ai`

- **When** `ptx-agent nosuchcmd --help-ai` or `ptx-agent --help-ai --bogus`
  **Then** the behavior is defined and stdout stays clean.

### FI-7 — stdout is full

- **When** `ptx-agent --help-ai > /dev/full` **Then** the command reports the write error, exit 1.

### FI-8 — Reader closes the pipe early

- **When** `ptx-agent --help-ai | head -1` **Then** no panic and no error text.

## Test Results

| ID | Platform | Result | Observed |
| --- | --- | --- | --- |
| TC-01 | Linux | PASS | stdout 20591 bytes, valid JSON, stderr 0 bytes, exit 0 |
| TC-02 | Linux | PASS | keys `tool`, `version`, `description`, `commands`; unit test decodes with `DisallowUnknownFields` |
| TC-03 | Linux | PASS | `make build` reports `0.1.0`, `go run` reports `dev` |
| TC-04 | Linux | PASS | 44 runnable commands in `--help`, 44 in the JSON, 0 missing, 0 extra |
| TC-05 | Linux | PASS | paths present, groups `terminal` and `tab` not listed |
| TC-06 | Linux | PASS | 48 parameters, 0 incomplete |
| TC-07 | Linux | PASS | 0 mismatches over 44 commands |
| TC-08 | Linux | PASS | values as expected (`TestHelpAIParameters` and manual check) |
| TC-09 | Linux | PASS | output identical (`cmp`), exit 0, no `accepts 1 arg(s)` error |
| TC-10 | Linux | PASS | outputs identical (`cmp`) |
| TC-11 | Linux | PASS | flag is in `Flags` of the root and in `Global Flags` of `files` |
| TC-12 | Linux | PASS | `go test -count=1 -race ./...` green (9 packages), `go vet` clean, `gofmt -s -l` empty |
| FI-1 | Linux | PASS | root help on stdout, exit 0 |
| FI-2 | Linux | PASS | `invalid argument "maybe" for "--help-ai" flag`, exit 1, stdout 0 bytes |
| FI-3 | Linux | PASS | `unknown command "--help-ai" for "ptx-agent version"`, exit 1, no JSON |
| FI-4 | Linux | PASS | `unknown flag: --help-a`, exit 1, stdout 0 bytes |
| FI-5 | Linux | PASS | JSON only, exit 0 |
| FI-6 | Linux | PASS | JSON, exit 0, stderr 0 bytes (see observation 2) |
| FI-7 | Linux | PASS | `write /dev/stdout: no space left on device`, exit 1 |
| FI-8 | Linux | PASS | first line `{`, exit status 0, no error text |
| TC-01 | Windows | SKIPPED | no Windows host; cross-compile of `internal/cli` passes |
| TC-04 | Windows | SKIPPED | no Windows host; the Windows command set is smaller (no `keyboard`, `windows`, `scenario`) |

### Functional Tests

- [x] Feature works as specified (Linux)
- [x] Acceptance criteria met (6 of 6, see mapping)
- [x] Edge cases handled (FI-1 to FI-8)

### Acceptance Criteria Mapping

- [x] `ptx-agent --help-ai` outputs valid JSON to stdout → TC-01
- [x] JSON contains `tool`, `version`, `description`, `commands` fields → TC-02, TC-03
- [x] All subcommands (including nested: files, vscode, app, mcp, browser) are listed → TC-04, TC-05
- [x] Parameters include names, types, required flag, and descriptions → TC-06, TC-07, TC-08
- [x] Exit code is 0 → TC-01, TC-09
- [x] No other output (logs, banners) is mixed into the JSON → TC-01 (stderr 0 bytes, one JSON document)

### Regression Tests

- [x] Existing functionality unaffected — `go test -count=1 -race ./...` green (9 packages)
- [ ] Cross-platform compatibility verified — Windows cross-compile passes, Windows execution not verified

## Coverage

| Function | Coverage |
| --- | --- |
| `helpAIRequested` | 100.0% |
| `printHelpAI` | 90.9% |
| `describable` | 100.0% |
| `describeSubcommands` | 88.9% |
| `describeCommand` | 100.0% |
| `positionalParameters` | 86.7% |
| `flagParameter` | 100.0% |
| `flagType` | 100.0% |
| `Execute` (`root.go`) | 0.0% |

The new file `help_ai.go` is above the 80% project target. The uncovered lines are the fallback
branches that the current command tree does not reach (hidden command, unknown placeholder).
`Execute` has no unit test; the black-box runs TC-01 and TC-09 cover it by hand only.

## Raw Results

```text
$ go test -count=1 -v -run 'HelpAI' ./internal/cli/
--- PASS: TestHelpAIRequested (0.00s)
--- PASS: TestHelpAITopLevelFields (0.00s)
--- PASS: TestHelpAIListsAllCommands (0.00s)
--- PASS: TestHelpAINestedSubcommands (0.00s)
--- PASS: TestHelpAIParameters (0.00s)
--- PASS: TestHelpAIParametersAreComplete (0.00s)
--- PASS: TestHelpAIFlagIsPersistent (0.00s)

$ ./ptx-agent --help-ai | head -5
{
  "tool": "ptx-agent",
  "version": "0.1.0",
  "description": "Portunix agent for remote device control",
  "commands": [

$ ./ptx-agent files list --help-ai | cmp - ref.json && echo same
same

$ ./ptx-agent --help-ai > /dev/full; echo "exit=$?"
write /dev/stdout: no space left on device
exit=1

consistency script:
help leaves 44 json leaves 44 missing set() extra set()
params 48 bad []
flag mismatches []
```

## Observations and Recommendations

None of these blocks the merge.

1. **The wiring in `Execute` has no automated test.** If the check in `Execute` is removed, all
   unit tests still pass. Recommendation: add a smoke step to the `cross-platform` CI job, next to
   the version smoke step, for example
   `go run ./cmd/ptx-agent --help-ai | python3 -c "import json,sys; json.load(sys.stdin)"`.
   This also gives the missing execution on Windows.
2. **`--help-ai` wins over every other error.** `ptx-agent nosuchcmd --help-ai` and
   `ptx-agent --help-ai --bogus` print the JSON with exit 0 (FI-6). This is a result of the check
   before cobra. It is acceptable for a discovery flag, but it is not written in the issue.
3. **The flag is also found in the position of a flag value.**
   `ptx-agent keyboard type --window --help-ai text` prints the JSON and does not use `--help-ai`
   as the window name. The case is not realistic.
4. **Flag names carry the `--` prefix** (`--headless`). The reference implementation in `portunix`
   lists flags without the prefix (`dry-run`). A client that reads both tools must know the
   difference. Recommendation: decide one form for the ecosystem.
5. **Descriptions of positional arguments are generic.** The issue example has
   `Directory path to list` for `files list`; the output has `File or directory path`, because the
   description belongs to the placeholder name and not to the command.
6. **The output depends on the platform.** The Linux build lists `keyboard`, `windows`, and
   `scenario`; the Windows build does not. This is correct (the JSON describes the binary that
   runs), but a client must not cache the output across platforms.
7. **The output has more than the issue example**: the commands `version`, `setup`, `keyboard`,
   `windows`, `scenario`, more `browser` subcommands, and the optional `description` field.
   The owner approved this before the implementation.

## CI Notes

- The implementation is not committed and not pushed, so no CI run exists for it yet.
- The `cross-platform` job runs `go test ./...` on `ubuntu-latest` and `windows-latest`. This will
  run the seven `TestHelpAI*` tests on Windows, including the comparison with the cobra tree.
- No CI step runs the binary with `--help-ai` (see observation 1).

## Final Decision

**STATUS**: CONDITIONAL

**Approval for merge**: NO — pending the commit and a green CI run

Required before approval:

1. Commit the implementation and push the branch.
2. The CI run for that commit is green in all jobs, including `windows-latest`.

The Linux verification is complete and has no failure. After the two steps above the status
changes to PASS without a new test round, if the committed code equals the tested working tree.

**Date**: 2026-10-11
**Tester signature**: zdendaku (QA/Test Engineer)
