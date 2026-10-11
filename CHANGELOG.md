# Changelog

All notable changes to portunix-agent will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added in Unreleased

- **Public keyboard package** (Issue #14) — the keyboard implementation moved from
  `src/internal/keyboard/` to `src/pkg/keyboard/`, so that another Go module (for example
  `portunix-archon`) can import it as a library. The exported API is unchanged.
- **Persistent keyboard session** (Issue #14) — `Controller.Open()` returns a `Session` with `Type`,
  `Key`, and `Close`. With the uinput backend a session creates the virtual device once, so a caller
  that sends many actions pays the 200 ms device registration delay only one time.
- **`--version` flag and `version` subcommand** (Issue #15) — `ptx-agent --version` and
  `ptx-agent version` print `ptx-agent version <version>`. The build takes the version from
  `plugin.json`; a build without it reports `dev`.

### Changed in Unreleased

- **GitHub-first issue tracking** — GitHub issues are the list of record and assign the issue numbers;
  each issue links to its long-form write-up in `docs/issues/NNN-*.md` (archived in
  `docs/issues/done/`).
- **Changelog** — this file was added, with the rules for keeping it current in
  `docs/contributing/CHANGELOG-GUIDELINES.md`.

## [0.1.0] - 2026-10-10

### Added in 0.1.0

- **Initial public release** of the `ptx-agent` helper plugin for Portunix.
- **Application control via accessibility APIs** — list windows, dump the accessibility tree, and read
  text from applications (AT-SPI2 on Linux, UI Automation on Windows).
- **VS Code terminal reading** — status check, terminal list, and terminal content through the
  accessibility tree.
- **Keyboard input simulation (Linux)** — type text, send key combinations, paste through the
  clipboard; uinput backend (Wayland and X11) with xdotool fallback (X11).
- **Window management (Linux)** — list, find, and focus windows; KWin D-Bus backend (Wayland/KDE) with
  xdotool fallback (X11).
- **Browser control via Playwright** — headless and headed mode, navigation, element interaction by
  index, screenshots, content extraction, tabs, persistent profile.
- **Secure file browsing** — directory listing and file info with whitelist-based access control and
  symlink traversal protection.
- **MCP server** — all capabilities are available as MCP tools over stdio and HTTP transport.
- **Plugin packaging** — `plugin.json` manifest and release packaging for deployment as a Portunix
  helper plugin.
