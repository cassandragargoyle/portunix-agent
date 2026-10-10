---
title: Issue Management
description: Workflow for creating, tracking, and archiving issues in portunix-agent — GitHub-first, where GitHub assigns the number and each issue keeps a long-form write-up in docs/issues/NNN-*.md (archived to docs/issues/done/ on completion). No overview tables.
category: workflow
ai_load: scoped
status: active
language: en
applies_to:
  - docs/issues/**
created: 2026-02-16
last_updated: 2026-10-10
related:
  - MARKDOWN-FRONTMATTER.md
---

# Issue Management Guidelines

> **AI assistant load scope (Claude Code, Gemini):**
>
> **Load when:**
>
> - Creating a new issue (GitHub issue plus the write-up under `docs/issues/`)
> - Updating the status of an issue
> - Closing an issue and archiving it to `docs/issues/done/`
>
> **Do NOT load for:** general code changes that merely reference an issue number, or single-question Q&A unrelated to issue workflow.
>
> **Authoritative scope:** the GitHub-first model, numbering, and the `done/` archive procedure. Body formatting rules live in [MARKDOWN-STYLE.md](MARKDOWN-STYLE.md).

## Purpose

This document defines how issues are created, tracked, and archived in portunix-agent.

> **GitHub is the single source of truth for the issue list.** Status, labels,
> and assignee live on GitHub. The repository keeps only the long-form write-up
> for each issue. There are **no overview tables** and no mapping files.

- Open issues: <https://github.com/cassandragargoyle/portunix-agent/issues>
- Closed issues: <https://github.com/cassandragargoyle/portunix-agent/issues?q=is%3Aissue+is%3Aclosed>

## GitHub-First Workflow

The issue is **created on GitHub first**, so **GitHub assigns the number `N`**.
The GitHub issue carries only the basics (title, a short description, and a link
to the detail file). The full write-up lives in `docs/issues/NNN-name.md`, where
`NNN` is the GitHub issue number zero-padded to three digits.

The `create-issue` skill (`/create-issue`) automates this flow.

### 1. Create the GitHub issue

```bash
gh issue create --repo cassandragargoyle/portunix-agent \
  --title "Short descriptive title" \
  --body "One-line summary.

Full details: docs/issues/NNN-name.md" \
  --label enhancement
```

Take `N` from the URL that `gh` prints. Never invent a number.

### 2. Write the detail file

Create `docs/issues/NNN-name.md` with the `# Issue #N: Title` heading, problem
description, acceptance criteria, and testing notes (see [Issue Template](#issue-template)).

### 3. Link the detail file from GitHub

```bash
gh issue edit N --repo cassandragargoyle/portunix-agent --body "One-line summary.

Full details: https://github.com/cassandragargoyle/portunix-agent/blob/main/docs/issues/NNN-name.md"
```

The link starts working once the file reaches `main` on GitHub.

### 4. Reference the number

- Branch: `feature/issue-N-short-name` or `fix/issue-N-short-name`
- Commits: `feat(N): …`, `fix(N): …`, `refs #N`, `closes #N`

### 5. Archive on completion

When the issue is implemented or closed:

```bash
git mv docs/issues/NNN-name.md docs/issues/done/NNN-name.md
gh issue close N --repo cassandragargoyle/portunix-agent --comment "Implemented in <version/commit>."
```

Update the GitHub body link to the `done/` path when archiving.

## Publication and Sensitive Content

`docs/issues/` is **published to GitHub** together with the rest of the repository.
Write every issue as public content:

- No customer names, business relationships, deals, or priorities tied to a customer
- No third-party or employer project names and internal identifiers
- No local paths such as `/home/<user>/…` — use repository-relative paths
- No internal hostnames, IP addresses, credentials, or tokens (placeholders like
  `<token>` or `example.com` are fine)

## File Structure

```text
docs/issues/
├── NNN-name.md          # Active issue write-up (NNN = GitHub issue number)
└── done/                # Archived issues (implemented or closed)
    └── NNN-name.md
```

Acceptance protocols stay in `docs/testing/`.

## Numbering

- GitHub issues and pull requests share one counter, so numbers skip where a pull
  request consumed one.
- Files use three-digit zero padding (`009-…`, `150-…`); GitHub references do not
  (`#9`, `#150`).

### Legacy numbering

Until 2026-10 issues were numbered internally and stored in `docs/issues/internal/`,
with an index table in `docs/issues/README.md`. Issues `001`–`015` come from that
period. During the migration they were created on GitHub with identical numbers
(`#1`–`#15`), so no renumbering was needed. Their files keep the old header style
(`# Issue #NNN - Title`).

## Labels

Every issue gets **one type label**:

- `bug` — something does not work as documented
- `enhancement` — new feature or improvement
- `documentation` — documentation only

Add other labels (for example `accessibility`) when they help filtering. Create a
new label only when several issues need it:

```bash
gh label create <name> --repo cassandragargoyle/portunix-agent \
  --color 1d76db --description "<what the label groups>"
```

## Issue Template

```markdown
# Issue #N: Title

**Status**: Open
**Priority**: Low | Medium | High | Critical
**Type**: Bug | Enhancement | Documentation
**Created**: YYYY-MM-DD
**GitHub**: #N
**Related**: #...

## Summary

## Problem Description

## Acceptance Criteria

1. ...

## Testing
```

### Formatting Rules

- **No icons or emoji** in issue files: not in headings and not in the `Status`
  field (`Open`, not `📋 Open`)
- Older issues using emoji are legacy style — do not copy it into new issues
- Exception: quoted literal program output stays verbatim

## Commit and Pull Request References

```bash
git commit -m "fix(42): keyboard device timeout (refs #42)"
git commit -m "feat(42): retry logic (closes #42)"
```

Pull requests on GitHub reference the issue with `Fixes #N` / `Refs #N`.
