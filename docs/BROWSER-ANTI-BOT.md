# Browser: Working with Anti-Bot-Protected Sites

Sites protected by anti-bot services (Cloudflare Turnstile, "Verify you are
human" interstitials) detect automated browsers and may refuse login even when
a human solves the challenge manually. This document describes the recommended
flow for collaborative automation of the user's own authenticated sessions.

> **Scope note:** The goal is collaborative automation in headed mode — the
> human is present and solves any challenge themselves. Automatic
> CAPTCHA-solving is out of scope and not supported.

## Why a default launch gets blocked

A plain `agent_browser_launch` starts a Playwright-bundled Chromium with:

- `navigator.webdriver = true` (Blink `AutomationControlled` feature)
- a fresh empty profile — no cookies, no prior anti-bot clearance token
- a non-persistent context — any clearance obtained is lost on close

## Recommended flow

### Step 1: Persistent profile + system browser (headed)

Launch with a named persistent profile and the system-installed Chrome:

- MCP: `agent_browser_launch` with `{"headless": false, "profile": "default", "channel": "chrome"}`
- CLI: `ptx-agent browser launch --headless=false --profile default --channel chrome`

Profile data is stored under `~/.config/ptx-agent/browser-profiles/<name>`
(Linux) or the platform equivalent (`os.UserConfigDir()`), so cookies, logins,
and anti-bot clearance tokens survive across sessions. In headed mode the
launch also drops `--enable-automation` and disables the Blink
`AutomationControlled` feature, so `navigator.webdriver` is not set.

The user logs in once and solves any challenge manually; subsequent launches
with the same profile reuse the session.

### Step 2: CDP connect (fallback)

If a site still refuses the agent-launched browser, attach to a browser the
user started themselves:

1. The user starts their browser with remote debugging enabled:

   ```bash
   google-chrome --remote-debugging-port=9222
   ```

2. The agent attaches:
   - MCP: `agent_browser_connect` with `{"endpoint": "http://localhost:9222"}`
   - CLI: `ptx-agent browser connect --endpoint http://localhost:9222`

All browser tools (navigate, click, type, screenshot, tabs, ...) work on the
attached instance. This is the most robust option because the browser is
genuinely the user's own — its binary, profile, and startup flags carry no
automation fingerprint.

Closing the session (`agent_browser_close`) only disconnects from a
CDP-attached browser; it does not terminate it.

## Option reference

| Option | Tool / flag | Values | Effect |
|---|---|---|---|
| `profile` | `agent_browser_launch` / `--profile` | any plain name | persistent user-data dir under `ptx-agent/browser-profiles/<name>` |
| `channel` | `agent_browser_launch` / `--channel` | `chrome`, `chromium`, `msedge`, ... | use a system-installed browser instead of the bundled build |
| `endpoint` | `agent_browser_connect` / `--endpoint` | URL | CDP endpoint of a user-started browser (default `http://localhost:9222`) |

Related: [Issue #013](issues/done/013-browser-bot-detection-persistent-profile.md),
[ADR-002](adr/002-browser-control-playwright.md)
