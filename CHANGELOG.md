# Changelog

All notable changes to portunix-agent will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.1] - 2026-10-11

### Added in 0.1.1

- **Public keyboard package** (Issue #14) — the keyboard implementation moved from
  `src/internal/keyboard/` to `src/pkg/keyboard/`, so that another Go module (for example
  `portunix-archon`) can import it as a library. The exported API is unchanged.
- **Persistent keyboard session** (Issue #14) — `Controller.Open()` returns a `Session` with `Type`,
  `Key`, and `Close`. With the uinput backend a session creates the virtual device once, so a caller
  that sends many actions pays the 200 ms device registration delay only one time.
- **`--version` flag and `version` subcommand** (Issue #15) — `ptx-agent --version` and
  `ptx-agent version` print `ptx-agent version <version>`. The build takes the version from
  `plugin.json`; a build without it reports `dev`.
- **`--help-ai` flag** (Issue #12) — `ptx-agent --help-ai` prints a JSON description of all commands,
  their arguments, and their flags (name, type, required, default, description), so that AI
  assistants can discover the CLI without parsing the help text. The output follows the Portunix
  ecosystem schema and is generated from the command tree.

### Fixed in 0.1.1

- **uinput backend lost characters without an error** (Issue #16) — `Type` skipped a character that
  the active keyboard layout has no key for, and still reported success. Characters that the layout
  produces with a dead key and a base letter (for example `ť`, `ň`, `ď`, `ó` on the `cz` layout) are
  now typed as a dead key sequence. A character that the layout cannot produce at all makes `Type`
  return an error that names the character and its position, and no key is sent. This applies to
  `Controller.Type` and to `Session.Type`.
- **xdotool backend reported success for an unknown key name** (Issue #17) — `Key` with a
  combination like `ctrl+nosuchkey` sent only the valid part and still reported success. The
  combination is now validated before xdotool runs: an unknown key name returns the same error as
  the uinput backend (`unknown key "nosuchkey" in combination "ctrl+nosuchkey"`), and no key is
  sent. Both backends now accept the same key names, without regard to case (`pageup`, `tab`,
  `f4`, `enter`). X keysym names outside this set (for example `Print`, `XF86AudioMute`) are no
  longer passed to xdotool and return an error.

### Changed in 0.1.1

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
