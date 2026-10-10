# Acceptance Protocol - Issue #013

**Issue**: Browser: bot detection blocks login (persistent profile + stealth support)
**Branch**: `feature/013-browser-persistent-profile` (merged to `main` at `a9c9a5e`)
**Tester**: zdendaku (QA/Test Engineer)
**Date**: 2026-06-08
**Testing OS**: Ubuntu 25.10 (host) — Go 1.24.4, Google Chrome 145.0.7632.75, Playwright Chromium 1208

## Test Summary

- Total test scenarios: 11
- Passed: 11
- Failed: 0
- Skipped: 0

## Test Plan

Verification combines automated unit tests (logic, no browser), automated e2e
tests (real browsers, `-tags e2e`), and CLI smoke tests. The e2e suite is
excluded from the default `make test` run so CI without a display/Chrome stays
green; it is run explicitly here.

| Layer | Command |
|---|---|
| Unit | `make test` |
| E2E | `go -C src test -tags e2e -run TestE2E ./internal/browser/ -timeout 240s` |
| CLI smoke | `./ptx-agent browser launch/connect --help` |

## Test Cases (Given / When / Then)

### TC1 — Profile dir resolution (unit)
- **Given** a plain profile name **When** `ProfileDir("default")` **Then** returns absolute path under `ptx-agent/browser-profiles/default`.
- **Given** traversal/abs names (`..`, `a/b`, `/abs`, `.hidden`) **When** `ProfileDir` **Then** error (path-traversal guard).

### TC2 — Fingerprint args (unit)
- **Given** headed mode **When** `fingerprintArgs(false)` **Then** returns `--disable-blink-features=AutomationControlled` + ignores `--enable-automation`.
- **Given** headless mode **When** `fingerprintArgs(true)` **Then** returns `(nil, nil)` — no masking.

### TC3 — Persistent profile survives relaunch (e2e, acceptance criterion 1)
- **Given** a launch with `Profile: "e2e-test"` and a cookie set **When** the session is closed and re-launched with the same profile **Then** the cookie is still present.

### TC4 — navigator.webdriver (e2e, acceptance criterion 3)
- **Given** headed launch **When** evaluating `navigator.webdriver` **Then** value is `false`/undefined (masked).
- **Given** headless launch **When** evaluating `navigator.webdriver` **Then** value is `true` (transparent — control case).

### TC5 — Browser channel (e2e)
- **Given** `Channel: "chrome"` **When** launched **Then** runs the system Google Chrome (UA reports `HeadlessChrome/145`) and `Status.Channel == "chrome"`.

### TC6 — CDP connect (e2e, acceptance criterion 4)
- **Given** a user-started `google-chrome --remote-debugging-port=9223` **When** `Connect(endpoint)` **Then** `Status.Connected == true`, tab count ≥ 1, existing tools (goto/title) drive the attached browser.
- **And When** the session is closed **Then** the browser remains reachable on the CDP endpoint (Close only disconnects).

### TC7 — Failure injection (e2e)
- Traversal profile name → launch error.
- Second launch while one is active → "already running" error.
- Connect to invalid endpoint → error.

### TC8 — CLI surface (smoke)
- `browser launch --help` exposes `--profile` and `--channel`.
- `browser connect --help` exposes `--endpoint` (default `http://localhost:9222`).

## Test Results

### Functional Tests
- [x] Feature works as specified
- [x] Acceptance criteria met (see mapping below)
- [x] Edge cases handled (failure injection TC7)

### Acceptance Criteria Mapping
- [x] AC1 — `agent_browser_launch` accepts `profile`/`channel`; persistent dir created and reused → TC1, TC3, TC5
- [x] AC2 — persistent profile + system Chrome + headed: session reused across launches → TC3 (cookie persistence; covers the reuse mechanism. Live chatgpt.com login is environment/account-dependent and not part of the automated suite — mechanism verified)
- [x] AC3 — `navigator.webdriver` false/undefined in headed mode → TC4
- [x] AC4 — `agent_browser_connect` attaches via CDP, existing tools work → TC6
- [x] AC5 — documentation of recommended anti-bot flow → `docs/BROWSER-ANTI-BOT.md`

### Regression Tests
- [x] Existing functionality unaffected — full `make test` green (9 packages), default browser tests pass unchanged
- [x] e2e suite excluded from default run (build tag `e2e` verified — default run cached/green)

## Raw Results

```
make test:                 ok (all packages)
e2e suite:                 ok  internal/browser  4.081s
  TestE2EPersistentProfileSurvivesRelaunch   PASS
  TestE2EWebdriverFingerprint/headed-masked  PASS
  TestE2EWebdriverFingerprint/headless-...   PASS
  TestE2EChannelChrome                       PASS (UA: HeadlessChrome/145.0.0.0)
  TestE2ECDPConnect                          PASS
  TestE2EFailures/{invalid-profile,double-launch,dead-cdp-endpoint}  PASS
gofmt / go vet:            clean
```

## CI Notes

- Default pipeline (`make test`) needs no browser/display — e2e tests are gated behind `-tags e2e`.
- To run e2e in CI, provision Playwright Chromium (`ptx-agent browser install`) and `google-chrome`; the channel/CDP cases self-skip if `google-chrome` is absent.
- The CDP test binds port 9223; ensure it is free in the CI runner.

## Final Decision

**STATUS**: PASS

**Approval for merge**: YES (note: merge to `main` already performed by developer at `a9c9a5e`; this protocol retroactively confirms acceptance)

**Date**: 2026-06-08
**Tester signature**: zdendaku

### Notes / Recommendations
- AC2's live verification against `chatgpt.com` depends on a real account and the
  site's current anti-bot posture; the persistence *mechanism* is proven by TC3.
  Recommend a one-time manual confirmation against the target site before relying
  on it in production workflows.
- Consider adding the e2e suite to a nightly CI job (not per-commit) given its
  runtime and external dependencies.
