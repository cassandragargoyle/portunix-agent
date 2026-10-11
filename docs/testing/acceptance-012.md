# Acceptance Protocol - Issue #012

**Issue**: Implement --help-ai flag
**Branch**: `feature/issue-12-help-ai-flag` (commit `5d2c99f`)
**Tester**: zdendaku (QA/Test Engineer)
**Date**: 2026-10-11
**Testing OS**: Ubuntu 26.04 (host, amd64, kernel 7.0.0-34) — Go 1.26.0, GNU Make 4.4.1

## Test Summary

- Total test scenarios: 29
- Passed: 29
- Failed: 0
- Skipped: 0

Windows has no local host. The three Windows scenarios passed in CI on `windows-latest`.

This is the third test round. All scenarios ran again on the final commit.

| Round | Commit | Result |
| --- | --- | --- |
| 1 | `12e7ec7` | no failure, seven observations |
| 2 | `b29ddf6` | no failure; CI smoke step, flag names without `--`, descriptions per command |
| 3 | `5d2c99f` | no failure; `allowed-dirs` required, cobra errors win over `--help-ai`, `positional` field replaced by `examples` |

## Test Plan

All tests ran in a clean clone of the feature branch, not in the developer working tree.
Verification combines the automated unit tests, a build with `make build`, black-box runs of the
binary, a comparison of the JSON output with the `--help` output of every command, and failure
injection on the arguments and on stdout.

| Layer | Command |
| --- | --- |
| Unit | `go test -count=1 -race ./...` in `src/`, 20 runs of `internal/cli` with `-shuffle=on` |
| Static | `go vet ./...`, `gofmt -s -l .`, `GOOS=windows go build ./...` |
| Integration | `make build`, then run `./ptx-agent --help-ai` |
| Consistency | script that walks `--help` of every command and compares it with the JSON |
| Failure injection | wrong flag values, flag at other positions, broken stdout |
| CI | smoke step that runs the binary with `--help-ai` on Linux and Windows |

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
- **And** no parameter name starts with `-` and no command has two parameters with the same name.

### TC-07 — Flags match the help text

- **Given** the `Flags` and `Global Flags` sections of `--help` of every command
  **Then** the JSON lists the same flags, by name without the `--` prefix, and without `help`
  and `help-ai`.

### TC-08 — Types, defaults, required

- **Then** `headless` is `boolean` with default `true`, `amount` is `integer` with default `3`,
  `allowed-dirs` is `array` without a default, `index` is a required `integer`,
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

### TC-13 — Positional arguments match the usage line

- **Given** the `Usage` line of `--help` of every command
  **Then** the JSON has the same positional arguments in the same order, before the flags,
  `<name>` is required and `[name]` is optional.
- **And** `examples` of every runnable command has the usage line (`ptx-agent files list <path>`).

### TC-14 — Positional arguments are described per command

- **Then** no positional argument has the generic fallback description.
- **And** the same placeholder has a description that fits its command (`path` of `files list`
  is `Directory path to list`, `path` of `files info` is `File path`).

### TC-15 — Required flag

- **Then** `allowed-dirs` of `files list` and `files info` has `required: true`.
- **When** `ptx-agent files list .` **Then** it fails with a clear error, exit 1.
- **When** `ptx-agent files list . --allowed-dirs .` **Then** it lists the directory.

### TC-16 — Only fields of the ecosystem schema

- **Then** a command has only `name`, `brief`, `description`, `parameters`, `examples`,
  `subcommands`, and a parameter has only `name`, `type`, `required`, `description`, `default`.

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

- **When** `ptx-agent nosuchcmd --help-ai`, `ptx-agent --help-ai --bogus`, or
  `ptx-agent files list --bogus --help-ai`
  **Then** the cobra error is reported, exit 1, stdout is empty.

### FI-7 — stdout is full

- **When** `ptx-agent --help-ai > /dev/full` **Then** the command reports the write error, exit 1.

### FI-8 — Reader closes the pipe early

- **When** `ptx-agent --help-ai | head -1` **Then** no panic and no error text.

### FI-9 — `--help-ai` in the position of a flag value

- **When** `ptx-agent files list . --allowed-dirs --help-ai`
  **Then** `--help-ai` is the value of `allowed-dirs`, no JSON, the command runs.

### FI-10 — Flags are not parsed two times

- **When** `ptx-agent files list . --allowed-dirs --help-ai --allowed-dirs .`
  **Then** the output equals the output of the same command with two normal values.

## Test Results

| ID | Platform | Result | Observed |
| --- | --- | --- | --- |
| TC-01 | Linux | PASS | stdout 24111 bytes, valid JSON, stderr 0 bytes, exit 0 |
| TC-02 | Linux | PASS | keys `commands`, `description`, `tool`, `version`; unit test decodes with `DisallowUnknownFields` |
| TC-03 | Linux | PASS | `make build` reports `0.1.0`, `go run` reports `dev` |
| TC-04 | Linux | PASS | 44 runnable commands in `--help`, 44 in the JSON, 0 missing, 0 extra |
| TC-05 | Linux | PASS | paths present, groups `terminal` and `tab` not listed |
| TC-06 | Linux | PASS | 48 parameters, 0 incomplete, 0 with a `-` prefix, 0 duplicates |
| TC-07 | Linux | PASS | 0 mismatches over 44 commands |
| TC-08 | Linux | PASS | values as expected (`TestHelpAIParameters` and manual check) |
| TC-09 | Linux | PASS | output identical (`cmp`), exit 0, also for `files --help-ai` and `keyboard type --help-ai` |
| TC-10 | Linux | PASS | outputs identical (`cmp`) |
| TC-11 | Linux | PASS | flag is in `Flags` of the root and in `Global Flags` of `files` |
| TC-12 | Linux | PASS | `go test -count=1 -race ./...` green (9 packages), `go vet` clean, `gofmt -s -l` empty (see observation 1 for `-shuffle=on`) |
| TC-13 | Linux | PASS | 21 positional arguments, 0 mismatches with the usage lines; 44 usage lines in `examples`, 0 mismatches |
| TC-14 | Linux | PASS | 0 fallback descriptions |
| TC-15 | Linux | PASS | `required: true`; `required flag(s) "allowed-dirs" not set`, exit 1; listing works with the flag |
| TC-16 | Linux | PASS | command keys `brief`, `description`, `examples`, `name`, `parameters`; no extra parameter field |
| FI-1 | Linux | PASS | root help on stdout, exit 0 |
| FI-2 | Linux | PASS | `invalid argument "maybe" for "--help-ai" flag`, exit 1, stdout 0 bytes |
| FI-3 | Linux | PASS | `unknown command "--help-ai" for "ptx-agent version"`, exit 1, no JSON |
| FI-4 | Linux | PASS | `unknown flag: --help-a`, exit 1, stdout 0 bytes |
| FI-5 | Linux | PASS | JSON only, identical to TC-01, exit 0 |
| FI-6 | Linux | PASS | `unknown command "nosuchcmd"`, `unknown flag: --bogus`; exit 1, stdout 0 bytes |
| FI-7 | Linux | PASS | `write /dev/stdout: no space left on device`, exit 1 |
| FI-8 | Linux | PASS | first line `{`, no panic, no error text; the process ends with SIGPIPE (status 141) |
| FI-9 | Linux | PASS | no JSON; the command runs with the allowed directory `--help-ai` and gives `access denied`, exit 1 |
| FI-10 | Linux | PASS | 842 bytes in both runs, exit 0 |
| TC-01 | Windows | PASS | CI smoke step: `go run ./cmd/ptx-agent --help-ai` gives JSON that `jq -e` accepts |
| TC-04 | Windows | PASS | CI: `TestHelpAIListsAllCommands` passed on `windows-latest` (unit level, smaller command set) |
| TC-09 | Windows | PASS | CI smoke step: `files list --help-ai` gives the JSON, exit 0 |

### Functional Tests

- [x] Feature works as specified (Linux, Windows through CI)
- [x] Acceptance criteria met (6 of 6, see mapping)
- [x] Edge cases handled (FI-1 to FI-10)

### Acceptance Criteria Mapping

- [x] `ptx-agent --help-ai` outputs valid JSON to stdout → TC-01
- [x] JSON contains `tool`, `version`, `description`, `commands` fields → TC-02, TC-03
- [x] All subcommands (including nested: files, vscode, app, mcp, browser) are listed → TC-04, TC-05
- [x] Parameters include names, types, required flag, and descriptions → TC-06, TC-07, TC-08, TC-13, TC-14, TC-15
- [x] Exit code is 0 → TC-01, TC-09
- [x] No other output (logs, banners) is mixed into the JSON → TC-01 (stderr 0 bytes, one JSON document)

### Regression Tests

- [x] Existing functionality unaffected — `go test -count=1 -race ./...` green (9 packages)
- [x] Cross-platform compatibility verified — build, unit tests, and the `--help-ai` smoke step
  green on `ubuntu-latest` and `windows-latest`

## Coverage

| Function | Coverage |
| --- | --- |
| `helpAIRequested` | 100.0% |
| `helpAIConfirmed` | 100.0% |
| `resetFlags` | 100.0% |
| `printHelpAI` | 90.9% |
| `describable` | 100.0% |
| `describeSubcommands` | 88.9% |
| `describeCommand` | 100.0% |
| `positionalParameters` | 86.7% |
| `flagParameter` | 100.0% |
| `flagType` | 100.0% |
| `Execute` (`root.go`) | 0.0% in unit tests, covered by the CI smoke step |

The new file `help_ai.go` is above the 80% project target. The uncovered lines are the fallback
branches that the current command tree does not reach (hidden command, unknown placeholder).

## Raw Results

```text
$ go test -count=1 -v -run 'HelpAI' ./internal/cli/
--- PASS: TestHelpAIRequested (0.00s)
--- PASS: TestHelpAIConfirmed (0.00s)
--- PASS: TestHelpAIConfirmedKeepsFlagValues (0.00s)
--- PASS: TestHelpAITopLevelFields (0.00s)
--- PASS: TestHelpAIListsAllCommands (0.00s)
--- PASS: TestHelpAINestedSubcommands (0.00s)
--- PASS: TestHelpAIParameters (0.00s)
--- PASS: TestHelpAIParametersAreComplete (0.00s)
--- PASS: TestHelpAIPositionalArgsAreUsed (0.00s)
--- PASS: TestHelpAIFlagIsPersistent (0.00s)

$ ./ptx-agent --help-ai | jq -c '.commands[] | select(.name=="files").subcommands[] | select(.name=="info")'
{"name":"info","brief":"Show file metadata","description":"...","parameters":[
  {"name":"path","type":"string","required":true,"description":"File path"},
  {"name":"allowed-dirs","type":"array","required":true,"description":"comma-separated list of allowed directory paths (required)"}],
  "examples":["ptx-agent files info <path>"]}

$ ./ptx-agent nosuchcmd --help-ai; echo "exit=$?"
Error: unknown command "nosuchcmd" for "ptx-agent"
exit=1

$ ./ptx-agent files list --help-ai | cmp - ref.json && echo same
same

consistency script:
leaves 44 44 missing set() extra set()
params 48 positional 21 bad [] flag mismatches [] positional mismatches [] dups [] example mismatches []
required flags [('files info', 'allowed-dirs'), ('files list', 'allowed-dirs')]

shuffled runs of internal/cli (20 each):
main (e63acdf): failed 11 of 20
branch (5d2c99f): failed 5 of 20, always TestVersionOutput/--version
```

## Observations and Recommendations

None of these blocks the merge.

Fixed during the three rounds:

- **The wiring in `Execute` had no automated test** — CI step `Smoke test help-ai output` (`b29ddf6`).
- **Flag names carried the `--` prefix** — listed without the prefix, the same as in `portunix` (`b29ddf6`).
- **Descriptions of positional arguments were generic** — one description per command (`b29ddf6`).
- **`allowed-dirs` reported `required: false`** — the flag is required in cobra (`ac30c87`).
- **No test found unused entries in `positionalArgs`** — `TestHelpAIPositionalArgsAreUsed` (`ac30c87`).
- **`--help-ai` won over every other error** — an unknown command or flag is an error again, and
  `--help-ai` in the position of a flag value is a value (`5d2c99f`).
- **The `positional` field was outside the ecosystem schema** — removed; the usage line in
  `examples` shows the positional arguments (`5d2c99f`).

Open:

1. **The tests of `internal/cli` depend on their order.** With `-shuffle=on` the test
   `TestVersionOutput/--version` fails when `TestHelpListsVersion` runs before it, because the
   `help` flag stays set on the shared root command. The defect comes from issue #15 and is on
   `main` (11 of 20 shuffled runs fail there, 5 of 20 on this branch). CI runs the tests in the
   fixed order and is green. Recommendation: a separate issue; `resetFlags` from this branch can
   be used in the cleanup of `runRoot`.
2. **`examples` holds a usage line, not a real example.** The value `ptx-agent files list <path>`
   uses placeholders. It fits the schema, but a client that runs examples as they are would fail.
3. **The error text of `files` without `--allowed-dirs` changed** from
   `at least one allowed directory must be configured` to
   `required flag(s) "allowed-dirs" not set`. The exit code is the same. `CHANGELOG.md` does not
   mention it.
4. **A reader that closes the pipe early ends the process with SIGPIPE** (status 141 for
   `ptx-agent --help-ai | head -1`). This is the normal behavior of a command-line tool.
5. **The output depends on the platform.** The Linux build lists `keyboard`, `windows`, and
   `scenario`; the Windows build does not. This is correct (the JSON describes the binary that
   runs), but a client must not cache the output across platforms.
6. **The output has more than the issue example**: the commands `version`, `setup`, `keyboard`,
   `windows`, `scenario`, more `browser` subcommands, and the fields `description` and
   `examples`. The owner approved the wider command list before the implementation.

## CI Notes

- CI run 38125555294 on commit `5d2c99f` passed in all jobs (2026-10-11): Lint, Unit Tests,
  Cross Platform Tests on `ubuntu-latest` and `windows-latest`, Quality Gate.
- The step `Smoke test help-ai output` passed on both systems. It runs
  `go run ./cmd/ptx-agent --help-ai` and `files list --help-ai` and checks the result with `jq`.
- The `cross-platform` job ran `go test ./...` on `windows-latest`, including the ten
  `TestHelpAI*` tests.
- CI does not run the tests with `-shuffle=on` (see observation 1).
- The consistency script (comparison with `--help` of every command) ran on Linux only.
- The earlier runs 38118576180 (`3167683`), 38118824087 (`b29ddf6`), and 38120028809 (`ac30c87`)
  were also green.

## Final Decision

**STATUS**: PASS

**Approval for merge**: YES

The tested commit is `5d2c99f`. A later commit on the branch that changes only this protocol does
not need a new test round.

**Date**: 2026-10-11
**Tester signature**: zdendaku (QA/Test Engineer)
