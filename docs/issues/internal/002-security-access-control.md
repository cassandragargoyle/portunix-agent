# Issue #002 - Security: access control for ptx-agent

**Type**: security
**Priority**: high
**Status**: needs-triage
**Branch**: TBD

## Problem

Once `ptx-agent setup` is run, any process running under the same user can use ptx-agent to simulate keyboard input, read clipboard, and control windows. There is no authentication or authorization mechanism — the only gate is the `input` group membership, which is a per-user (not per-process) permission.

This means any script, application, or malware running in the user session can invoke ptx-agent to:

- Type arbitrary text into any focused application
- Read clipboard contents (potentially sensitive data)
- Switch and control application windows
- Execute key combinations (Ctrl+A, Delete, etc.)

## Expected Behavior

ptx-agent should have an access control mechanism that limits which processes or contexts can invoke its capabilities.

## Possible Approaches

1. **Socket with authentication** — ptx-agent runs as a daemon, clients must authenticate (token, socket permissions)
2. **Per-invocation authorization** — prompt user for confirmation before executing sensitive operations
3. **Allowlist of callers** — restrict execution to specific parent processes (e.g. only Portunix dispatcher)
4. **Scoped tokens** — short-lived tokens issued by Portunix that authorize specific operations
5. **File permission restriction** — limit binary execution to specific group/user beyond just `input`

## Impact

Without access control, the security boundary is effectively the user session — anything running as the user has full desktop control capabilities through ptx-agent.

## Related

- Issue #001 — MVP ptx-agent (current implementation)
- `ptx-agent setup` — grants `/dev/uinput` access via udev rule + input group
