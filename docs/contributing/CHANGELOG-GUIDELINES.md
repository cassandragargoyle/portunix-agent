---
title: Changelog Guidelines
description: Rules for keeping CHANGELOG.md current in portunix-agent — every user-visible change adds an entry under [Unreleased] on its feature branch, and the version bump turns [Unreleased] into a dated release section.
category: workflow
ai_load: scoped
status: active
language: en
applies_to:
  - CHANGELOG.md
created: 2026-10-11
last_updated: 2026-10-11
related:
  - ISSUE-DEVELOPMENT-METHODOLOGY.md
  - ISSUE-MANAGEMENT.md
  - GIT-WORKFLOW.md
---

# Changelog Guidelines

## Purpose

`CHANGELOG.md` in the repository root tells users what changed in each version of `ptx-agent`.
It uses the [Keep a Changelog](https://keepachangelog.com/en/1.0.0/) format and the project
follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

The changelog is written by hand. It is not a copy of the git log.

## When to Add an Entry

Add an entry when a change is visible to a user of `ptx-agent` or to a consumer of its public
Go packages (`src/pkg/`):

- A new or changed command, flag, MCP tool, or public Go API
- A bug fix that changes behavior
- A change in the build, packaging, or deployment that a user can notice
- A change in the project workflow that contributors must know

Do not add an entry for internal refactoring without a visible effect, for tests, for acceptance
protocols, or for issue archiving.

## Workflow

### 1. Entry on the feature branch

The developer adds the entry under `## [Unreleased]` **on the feature branch**, in the same commit
as the change or in a separate `docs(N): update changelog` commit. The entry reaches `main` together
with the code.

### 2. Check before merge

The tester verifies that the entry exists and that it describes the delivered behavior. A missing
entry for a user-visible change blocks the merge (see
[Issue Development Methodology](ISSUE-DEVELOPMENT-METHODOLOGY.md)).

### 3. Release

The version bump (skill `/bump-version`) does these steps in this order:

1. Set the new version in `plugin.json` (single source of the version; `make` passes it to the binary)
2. Rename `## [Unreleased]` to `## [X.Y.Z] - YYYY-MM-DD` and its sections to `... in X.Y.Z`
3. Compare the section with `git log <last tag>..HEAD` and add entries that are missing
4. Add a new empty `## [Unreleased]` heading on top
5. Commit `chore: bump version to vX.Y.Z`
6. Create the tag `vX.Y.Z` — only after the commit

## Format

```markdown
## [Unreleased]

### Added in Unreleased

- **Short name of the feature** (Issue #N) — what changed and what it means for the user.

### Fixed in Unreleased

- **Short name of the problem** (Issue #N) — what was wrong and how it behaves now.

## [0.2.0] - 2026-11-03

### Added in 0.2.0
```

### Sections

| Conventional commit | Changelog section |
| ------------------- | ----------------- |
| `feat` | `### Added in X.Y.Z` |
| `fix` | `### Fixed in X.Y.Z` |
| `refactor`, `perf`, `chore`, `build`, `docs` (if user-visible) | `### Changed in X.Y.Z` |
| `BREAKING CHANGE` / `!:` | `### Changed in X.Y.Z` with a **BREAKING** note |
| Feature removal | `### Removed in X.Y.Z` |
| Deprecation | `### Deprecated in X.Y.Z` |
| Security fix | `### Security in X.Y.Z` |

Section headings carry the version (`in X.Y.Z`, or `in Unreleased`), because markdownlint rule
MD024 does not allow two headings with the same text. Do not write empty sections.

### Writing Rules

- English, simple sentences
- Start with the **bold** name of the affected feature or area, then a dash, then the explanation
- Add the issue number `(Issue #N)` and the ADR `(ADR-NNN)` when they exist
- Describe the impact for the user, not the list of changed files
- Newest version on top; version headings without the `v` prefix (`[0.2.0]`), tags with it (`v0.2.0`)
- Date in the `YYYY-MM-DD` format

## Responsibilities

| Role | Responsibility |
| ---- | -------------- |
| Developer | Writes the entry on the feature branch |
| Tester | Checks the entry in the acceptance protocol |
| Release (version bump) | Turns `[Unreleased]` into a dated version section |
