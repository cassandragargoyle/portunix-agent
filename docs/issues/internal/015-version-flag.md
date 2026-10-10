# Issue #015 - Implement --version flag

**Type**: enhancement
**Priority**: medium
**Status**: Open
**Branch**: `feature/015-version-flag`
**Labels**: enhancement, cli, build

## Problem Description

`ptx-agent` cannot report its version. Both common forms fail with exit code 1:

```text
$ ptx-agent --version
Error: unknown flag: --version
...
$ ptx-agent version
Error: unknown command "version" for "ptx-agent"
```

Verified on 2026-10-10 (Linux amd64) with three binaries:

| Binary | Build date | `--version` | `version` |
|--------|------------|-------------|-----------|
| `./ptx-agent` (local `make build`) | 2026-06-07 | exit 1, unknown flag | exit 1, unknown command |
| `dist/ptx-agent/ptx-agent` (release 0.1.0) | 2026-02-22 | exit 1, unknown flag | exit 1, unknown command |
| `~/.portunix/plugins/agent/ptx-agent` (deployed plugin) | 2026-02-22 | exit 1, unknown flag | exit 1, unknown command |

Cause:

- `rootCmd` in `src/internal/cli/root.go` does not set the cobra `Version` field and there is
  no `version` subcommand.
- The build does not pass a version to the binary. `Makefile` (`build`) and
  `scripts/make-release.py` (`go build`) do not use `-ldflags`.
- The version exists only in `plugin.json` (`0.1.0`) and as a separate hardcoded string
  `"0.1.0"` in `src/internal/mcp/server.go` (MCP `initialize` response).

Impact: a user or Portunix cannot find which version of the plugin is installed. Bug reports
cannot name the build. Issue #012 (`--help-ai`) also needs the version for its `version` field.

## Requirements

### 1. Single source of the version in Go code

- Add a package-level variable, for example `Version` in `src/internal/cli/` (or a small
  `src/internal/version/` package), with the default value `dev`.
- The build sets the value with `-ldflags "-X <package>.Version=<version>"`.
- Optional: also embed the git commit and the build date.

### 2. `--version` flag

- Set `rootCmd.Version` so that cobra registers `--version`.
- Output goes to stdout, one line, format: `ptx-agent version <version>`.
- Exit code 0. No other output (no usage text, no logs).

### 3. `version` subcommand

- Add `ptx-agent version` with the same output as `--version`.

### 4. Build integration

- `Makefile` target `build`: read the version from `plugin.json` and pass it with `-ldflags`.
  The target must still work on Linux and on Windows (`SHELL := cmd`).
- `scripts/make-release.py`: pass the release version (argument or `plugin.json`) with `-ldflags`
  in the `go build` command.
- A plain `go build` without `-ldflags` must still work and report `dev`.

### 5. Remove the duplicate version string

- `src/internal/mcp/server.go` must use the same variable in the `initialize` response instead
  of the hardcoded `"0.1.0"`.

## Test Plan

Scope: CLI output of the version, build integration, MCP server info. Platforms: Linux, Windows.

### Test Cases

**TC-01 `--version` prints the version**

- Given a binary built with `make build` and `plugin.json` version `0.1.0`
- When the user runs `ptx-agent --version`
- Then stdout is exactly `ptx-agent version 0.1.0`, stderr is empty, exit code is 0

**TC-02 `version` subcommand**

- Given the same binary
- When the user runs `ptx-agent version`
- Then the output and exit code are the same as in TC-01

**TC-03 Build without ldflags**

- Given a binary built with `go build ./cmd/ptx-agent/` in `src/`
- When the user runs `ptx-agent --version`
- Then stdout is `ptx-agent version dev`, exit code is 0

**TC-04 Release build**

- Given archives created by `make dist` for version `X.Y.Z`
- When the user runs `--version` on the Linux binary and on the Windows binary
- Then both report `X.Y.Z`

**TC-05 MCP server reports the same version**

- Given a binary built with version `X.Y.Z`
- When a client sends `initialize` to `ptx-agent mcp serve`
- Then `serverInfo.version` is `X.Y.Z`

**TC-06 Flag does not break subcommands**

- Given the new binary
- When the user runs `ptx-agent files list --version`
- Then the command fails with `unknown flag` (the flag is not persistent) and exit code 1

**TC-07 Help lists the flag**

- Given the new binary
- When the user runs `ptx-agent --help`
- Then the `Flags` section contains `-v, --version` or `--version`, and `Available Commands`
  contains `version`

### Coverage

| Level | What | Target |
|-------|------|--------|
| Unit | `version` command and root `--version` output (cobra `SetOut` + `SetArgs`) | 100 % of the new code |
| Unit | MCP `initialize` uses the shared variable (`server_test.go`) | existing test extended |
| Integration | `make build` then run the binary (TC-01, TC-02, TC-03) | Linux, Windows |
| E2E | `make dist` then run the packaged binaries (TC-04) | Linux, Windows |

### Failure Injection

- `plugin.json` is missing or has no `version` field: the build must fail with a clear message
  or fall back to `dev` (decide and document one behavior).
- Version string with a space or a quote character: `-ldflags` quoting must not break the build.
- Build on Windows with `SHELL := cmd`: the version read from `plugin.json` must not contain
  `\r` or quotes.

### CI Notes

- Add a smoke step after the build: `./ptx-agent --version` must exit with 0 and the output must
  match `^ptx-agent version \S+$`.
- Run the step on Linux and Windows runners.

## Acceptance Criteria

- [ ] `ptx-agent --version` prints `ptx-agent version <version>` to stdout and exits with 0
- [ ] `ptx-agent version` prints the same output and exits with 0
- [ ] `make build` embeds the version from `plugin.json`
- [ ] `make dist` embeds the release version in all platform binaries
- [ ] A build without `-ldflags` reports `dev`
- [ ] MCP `initialize` response uses the same version value (no hardcoded string)
- [ ] Unit tests cover the flag and the subcommand
- [ ] Verified on Linux and Windows

## References

- `src/internal/cli/root.go` — `rootCmd`
- `src/internal/mcp/server.go` — hardcoded `Version: "0.1.0"`
- `Makefile` — `build` target
- `scripts/make-release.py` — `go build` command, `detect_version_from_plugin_json()`
- `plugin.json` — current source of the version
- Issue #012 — `--help-ai` needs the `version` field

---

**Created**: 2026-10-10
**Author**: Tester
