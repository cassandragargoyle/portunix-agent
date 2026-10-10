# Acceptance Protocol - Issue #015

**Issue**: Implement --version flag
**Branch**: `feature/015-version-flag` (commit `fd4a389`)
**Tester**: zdendaku (QA/Test Engineer)
**Date**: 2026-10-10
**Testing OS**: Ubuntu 26.04 (host, amd64, kernel 7.0.0-34) — Go 1.26.0, GNU Make 4.4.1

## Test Summary

- Total test scenarios: 17
- Passed: 13
- Failed: 0
- Skipped: 4 (execution on Windows — no Windows host available)

## Test Plan

All tests ran in a clean clone of the feature branch, not in the developer working tree.
Verification combines the automated unit tests, a build with `make build`, a plain
`go build`, the release packaging with `make dist`, and failure injection on `plugin.json`.

| Layer | Command |
| --- | --- |
| Unit | `go test -count=1 -race ./...` in `src/` |
| Static | `go vet ./...`, `gofmt -s -l .` |
| Integration | `make build`, then run `./ptx-agent` |
| E2E | `make dist`, then run the packaged binaries |
| Failure injection | modified `plugin.json`, then `make build` and `make-release.py` |

## Test Cases (Given / When / Then)

Test cases TC-01 to TC-07 are defined in `docs/issues/015-version-flag.md`.
The failure injection cases are:

### FI-1 — `plugin.json` is missing

- **Given** no `plugin.json` **When** `make build` **Then** the build passes and the binary reports `dev`.
- **And When** `make-release.py` runs without a version argument **Then** it fails with a clear message.

### FI-2 — `plugin.json` has no `version` field

- **Given** `plugin.json` without `version` **When** `make build` **Then** the binary reports `dev`.

### FI-3 — Version with a space

- **Given** version `0.1.0 beta` **When** `make build` **Then** the build does not break.

### FI-4 — `plugin.json` with CRLF line endings

- **Given** CRLF line endings **When** `make build` **Then** the output has no `\r` and no quotes.

### FI-5 — Compact JSON (`"version":"0.1.0"`)

- **Given** no space after the colon **When** `make build` **Then** the build does not break.

### FI-6 — Pre-release version

- **Given** version `1.2.3-rc.1+build5` **When** `make build` **Then** the binary reports the full string.

## Test Results

| ID | Platform | Result | Observed |
| --- | --- | --- | --- |
| TC-01 | Linux | PASS | stdout `ptx-agent version 0.1.0\n` (one line), stderr 0 bytes, exit 0 |
| TC-02 | Linux | PASS | output identical to TC-01, stderr 0 bytes, exit 0 |
| TC-03 | Linux | PASS | plain `go build` reports `ptx-agent version dev`, exit 0 |
| TC-04 | Linux | PASS | `make dist`: binary from `ptx-agent-0.1.0-linux-amd64.tar.gz` reports `0.1.0`; release with argument `2.3.4` reports `2.3.4` |
| TC-04 | Windows binary | PARTIAL | `ptx-agent.exe` from the zip is a valid PE32+ x86-64 file and contains the string `0.1.0`; the binary was not run |
| TC-05 | Linux | PASS | `initialize` on `mcp serve` returns `serverInfo.version` = `0.1.0` |
| TC-06 | Linux | PASS | `files list . --version` gives `Error: unknown flag: --version`, exit 1 |
| TC-07 | Linux | PASS | `--help` lists `--version` in `Flags` and `version` in `Available Commands` |
| FI-1 | Linux | PASS | `make build` reports `dev`; release script exits 1 with `Could not detect version from plugin.json` |
| FI-2 | Linux | PASS | `make build` reports `dev`; release script exits 1 |
| FI-3 | Linux | PASS | build passes, version is cut to `0.1.0`; release script rejects it with `Invalid version format` |
| FI-4 | Linux | PASS | output bytes checked with `od -c`: no `\r`, no quotes |
| FI-5 | Linux | PASS | build passes, binary reports `dev` (see observation 2) |
| FI-6 | Linux | PASS | binary reports `1.2.3-rc.1+build5` |
| TC-01 | Windows | SKIPPED | no Windows host |
| TC-02 | Windows | SKIPPED | no Windows host |
| TC-03 | Windows | SKIPPED | no Windows host |
| TC-04 | Windows | SKIPPED | binary not run, `make build` with `SHELL := cmd` not run |

### Functional Tests

- [x] Feature works as specified (Linux)
- [ ] Acceptance criteria met (7 of 8, see mapping)
- [x] Edge cases handled (FI-1 to FI-6)

### Acceptance Criteria Mapping

- [x] `ptx-agent --version` prints `ptx-agent version <version>` to stdout, exit 0 → TC-01
- [x] `ptx-agent version` prints the same output, exit 0 → TC-02
- [x] `make build` embeds the version from `plugin.json` → TC-01
- [x] `make dist` embeds the release version in all platform binaries → TC-04 (Windows binary checked by content only)
- [x] A build without `-ldflags` reports `dev` → TC-03
- [x] MCP `initialize` uses the same version value → TC-05, `TestHandleInitialize`
- [x] Unit tests cover the flag and the subcommand → `version_test.go` (5 tests)
- [ ] Verified on Linux and Windows → Linux only

### Regression Tests

- [x] Existing functionality unaffected — `go test -count=1 ./...` green (9 packages), `go vet` clean, `gofmt -s -l` empty
- [ ] Cross-platform compatibility verified — Windows cross-compile passes, Windows execution not verified

## Raw Results

```text
$ go test -count=1 -v -run 'Version|HelpLists|HandleInitialize' ./internal/cli/ ./internal/mcp/
--- PASS: TestVersionOutput (0.00s)
    --- PASS: TestVersionOutput/--version (0.00s)
    --- PASS: TestVersionOutput/version (0.00s)
--- PASS: TestVersionDefaultIsDev (0.00s)
--- PASS: TestVersionSubcommandRejectsArgs (0.00s)
--- PASS: TestVersionFlagIsNotPersistent (0.00s)
--- PASS: TestHelpListsVersion (0.00s)
--- PASS: TestHandleInitialize (0.00s)

$ make build
Building ptx-agent version 0.1.0...
cd src && go build -ldflags "-X github.com/CassandraGargoyle/portunix-agent/internal/version.Version=0.1.0" -o ../ptx-agent ./cmd/ptx-agent/

$ ./ptx-agent --version | od -c
0000000   p   t   x   -   a   g   e   n   t       v   e   r   s   i   o
0000020   n       0   .   1   .   0  \n

$ echo '{"jsonrpc":"2.0","id":1,"method":"initialize",...}' | ./ptx-agent mcp serve
{"jsonrpc":"2.0","id":1,"result":{"protocolVersion":"2024-11-05","capabilities":{"tools":{}},"serverInfo":{"name":"ptx-agent","version":"0.1.0"}}}
```

## Observations and Recommendations

None of these blocks the merge.

1. **Short flag `-v` is not available.** `ptx-agent -v` fails with `unknown shorthand flag`.
   This is a design decision (the letter stays free for a future `--verbose`) and TC-07 allows it.
2. **The Makefile needs the exact form `"version": "X.Y.Z"`.** With compact JSON
   (`"version":"0.1.0"`) the build silently falls back to `dev` (FI-5). The only signal is the
   line `Building ptx-agent version dev...`. Recommendation: print a warning when the fallback
   is used.
3. **A version with a space is cut without a warning** in `make build` (FI-3: `0.1.0 beta`
   becomes `0.1.0`). The release script rejects such a version, so releases are not affected.
4. **The fallback to `dev` is documented only in a Makefile comment.** The issue asks to decide
   and document one behavior. Recommendation: add one sentence to `README.md`.
5. **Release with a version argument can differ from `plugin.json`.** `make-release.py 2.3.4`
   builds binaries that report `2.3.4`, but the packaged `plugin.json` still says `0.1.0`.
   This behavior existed before this issue. Recommendation: a separate issue.
6. **`make build` needs GNU Make 4.2 or newer** because of `$(file < plugin.json)`.
   The version of Make on the Windows build machines must be checked.

## CI Notes

- The branch adds a smoke step to the `cross-platform` job (`ubuntu-latest`, `windows-latest`).
  The step runs `go run ./cmd/ptx-agent --version` and `version` and checks the output format.
- CI run 38081028664 on commit `d255697` passed in all jobs (2026-10-10). The smoke step passed
  on `windows-latest`, which covers TC-03 on Windows.
- The CI step does not cover `make build` on Windows (`SHELL := cmd`) and does not run the
  packaged `ptx-agent.exe`. These two checks need a Windows host.

## Final Decision

**STATUS**: CONDITIONAL

**Approval for merge**: NO — pending the Windows verification

Required before approval:

1. Run TC-01, TC-02 and TC-03 on Windows (`make build` with `SHELL := cmd`, then `ptx-agent.exe --version`).
2. Run `ptx-agent.exe --version` from the `make dist` zip on Windows (TC-04).

**Merge decision**: The project owner (zdendaku) approved the merge to `main` on 2026-10-10 as
an exception to the CONDITIONAL status. The two checks above stay open as follow-up work and
the tester status does not change to PASS until they are done.

**Date**: 2026-10-10
**Tester signature**: zdendaku (QA/Test Engineer)
