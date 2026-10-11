# portunix-agent Development Instructions

# CLAUDE.md (root)

@docs/contributing/

## Project Information

- **GitHub Repository**: <https://github.com/CassandraGargoyle/portunix-agent>
- **Project Name**: portunix-agent
- **Primary Purpose**: Plugin for the Portunix system that allows remote control of devices (mouse, keyboard, application switching, file browsing)
- **Primary Language**: Go
- **Platforms**: Linux, Windows (macOS development status TBD)

## Security Guidelines

- Never run destructive commands without confirmation
- Always warn about potential data loss before executing risky operations
- Require explicit confirmation for operations that could lose implemented code

## Project Structure

- `/docs/` - Documentation including contributing guidelines
- `/scripts/` - Installation and setup scripts
- `/templates/` - Project templates
- `/config/` - Configuration files
- `/tests/` - Tests and validation
- `/utils/` - Helper utilities
- `/.claude/` - AI assistant configuration and roles

## Coding Guidelines & Development Instructions

### Strict Rules

- Code, code comments, and user messages must be in English (see [Multilingual Team Communication](docs/contributing/README.md#multilingual-team-communication))
- Communication with team members can be in their preferred language, but documentation remains in English
- When implementing new features, always check existing project structure first
- Prefer editing existing files over creating new ones
- Do not add "Generated with [Claude Code]" signatures to code or files
- NEVER add "Co-Authored-By: Claude" to code files or git commits

See docs/contributing/CODE-STYLE-*.md for language-specific guidelines

### General Principles

- Follow existing conventions and styles in the project
- Always explore existing code before creating new implementations
- Maintain consistency with established patterns
- All program text (messages, labels, constants) in English
- Comments in English, written as phrases (no ending periods)
- English is the official project language per multilingual team guidelines

### TODO Management

- Follow the established TODO format: `TODO:NNN [INITIALS]: description` (see [TODO Guidelines](docs/contributing/TODO-GUIDELINES.md))
- Use per-file sequential numbering (001, 002, 003, etc.)
- Refer to [docs/contributing/TODO-GUIDELINES.md](docs/contributing/TODO-GUIDELINES.md) for complete standards
- Use `TODO:XXX` as temporary placeholder, then request proper numbering

### Issue Tracking & Documentation

- GitHub is the single source of truth for the issue list and assigns the issue number
- Long-form write-ups live in `docs/issues/NNN-name.md` (`NNN` = GitHub number), archived to `docs/issues/done/`
- See [docs/contributing/ISSUE-MANAGEMENT.md](docs/contributing/ISSUE-MANAGEMENT.md); use `/create-issue` to create an issue
- Common issue types: Configuration issues, build problems, compatibility issues

### Changelog

- `CHANGELOG.md` is kept current in the [Keep a Changelog](https://keepachangelog.com/en/1.0.0/) format
- Every user-visible change adds an entry under `## [Unreleased]` on its feature branch, before merge to `main`
- Section headings carry the version: `### Added in X.Y.Z`, `### Fixed in X.Y.Z`, `### Changed in X.Y.Z`
- The version bump (`/bump-version`) turns `[Unreleased]` into `## [X.Y.Z] - YYYY-MM-DD`; the version source is `plugin.json`
- See [docs/contributing/CHANGELOG-GUIDELINES.md](docs/contributing/CHANGELOG-GUIDELINES.md)

### Team Communication

- English is the primary language for all project communication (see [Multilingual Team Communication](docs/contributing/README.md#multilingual-team-communication))
- Use simple, clear English to accommodate non-native speakers
- Avoid idioms, colloquialisms, and culture-specific references
- Team initials assignment follows collision resolution (JS → JSm if JS exists) (see [Team Initials Assignment](docs/contributing/README.md#team-initials-assignment-and-collision-resolution))

### AI Assistant Guidelines

- Claude Code is the preferred AI assistant with pre-configured context
- Alternative AI tools require users to prepare their own context materials
- Follow established AI assistant best practices in [docs/contributing/AI-ASSISTANTS.md](docs/contributing/AI-ASSISTANTS.md)

### Translation Workflow

- Use `.translated/` directory structure for team member translations
- Follow ISO 639-1 language codes (cs, de, fr, etc.)
- Refer to [docs/contributing/TRANSLATION-WORKFLOW.md](docs/contributing/TRANSLATION-WORKFLOW.md) for complete process
- Translations are temporary and can be regenerated as needed

### Repository Management

- Main branch is `main` (not `master`)
- Issues tracked on GitHub, write-ups in `docs/issues/`
- License: MIT (see `LICENSE`)
- GitHub format: https://github.com/CassandraGargoyle/portunix-agent

### System Information Detection

- **Always use** `portunix system info` for OS detection when Portunix is available
- **Only if Portunix is not available**, then use manual detection methods
- Don't write custom OS detection scripts when Portunix already provides this functionality

## Development Setup

### Prerequisites

To be defined based on project requirements

### Initial Setup

```bash
# Clone the repository
git clone https://github.com/CassandraGargoyle/portunix-agent.git
cd portunix-agent

# Add project-specific setup commands here
```

### Build Instructions

- **Build system**: `make` (see `Makefile` in project root)

```bash
make
```

## Testing

### Testing Framework

- Framework: To be determined
- Test command: `# Add test command here`
- Coverage target: 80%

### Running Tests

```bash
# Add test command here
```

## Project-Specific Information

### Dependencies

Dependencies to be documented

### Architecture

Architecture to be documented

### Key Components

Key components to be documented

### Configuration

Configuration to be documented

### Deployment

Deployment process to be documented