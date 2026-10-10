# Issue #013 - Browser: bot detection blocks login (persistent profile + stealth support)

**Type**: enhancement
**Priority**: high
**Status**: ✅ Implemented
**Branch**: `feature/013-browser-persistent-profile`
**Depends on**: Issue #008

## Problem Description

`agent_browser_launch` starts a fresh Playwright-bundled Chromium with an empty,
non-persistent profile. Sites protected by anti-bot services (Cloudflare Turnstile,
"Verify you are human" interstitials) detect the automation and refuse to let the
user past the challenge — the verification loops indefinitely even when the human
user solves it manually in the headed window.

Reproduced on `chatgpt.com` (2026-06-07): login page presents a robot check that
never passes in the agent-launched browser, while the same login works in the
user's regular browser. This blocks a whole class of collaborative workflows where
the agent drives a logged-in web app (image generation in the ChatGPT UI, admin
panels behind SSO, etc.).

Likely detection signals:

- `navigator.webdriver = true` (Blink `AutomationControlled` feature enabled by default)
- Playwright-bundled Chromium build instead of a regular installed Chrome/Chromium
- Fresh empty profile — no cookies, no history, no prior Cloudflare clearance token
- Non-persistent context — any clearance obtained is lost on the next launch

## Proposed Solution

Extend the browser module (Issue #008) with launch options that make the headed
collaborative mode usable on protected sites:

1. **Persistent profile**: `agent_browser_launch` option `profile` (or `user_data_dir`)
   → launch via `launch_persistent_context()` with a dedicated user-data directory
   (e.g. `~/.config/ptx-agent/browser-profiles/<name>`). Cookies, logins, and
   anti-bot clearance tokens survive across sessions.
2. **Real browser channel**: option `channel` (`chrome`, `chromium`, `msedge`) to use
   the system-installed browser instead of the Playwright-bundled build.
3. **Reduced automation fingerprint** (headed collaborative mode only):
   `--disable-blink-features=AutomationControlled`, drop obvious automation flags
   so `navigator.webdriver` is not set.
4. **CDP connect mode**: `agent_browser_connect` tool — attach over CDP to a browser
   the user started themselves (`--remote-debugging-port=9222`). The user logs in
   in their own browser; the agent only drives it. This is the most robust fallback
   because the browser is genuinely the user's own.

Items 1–3 are launch-option plumbing; item 4 is a new MCP tool.

## Ethics / Scope Note

The goal is **collaborative automation of the user's own authenticated sessions**
in headed mode — the human is present and solves any challenge themselves. This is
NOT about defeating CAPTCHAs automatically; no challenge-solving logic is in scope.

## Acceptance Criteria

- [x] `agent_browser_launch` accepts `profile` and `channel` options; persistent
      profile directory is created and reused across launches
- [x] With a persistent profile + system Chrome + headed mode, the user can log in
      to `chatgpt.com` once, and subsequent launches reuse the session without a
      new robot check *(persistence mechanism verified by e2e TC3; live
      chatgpt.com login is account-dependent — see acceptance-013.md)*
- [x] `navigator.webdriver` is `false`/undefined in headed collaborative mode
      *(verified by e2e TC4)*
- [x] `agent_browser_connect` attaches to a user-started browser via CDP and all
      existing browser tools work on the attached instance
- [x] Documentation of the recommended flow for anti-bot-protected sites
      (see [docs/BROWSER-ANTI-BOT.md](../../../BROWSER-ANTI-BOT.md))
