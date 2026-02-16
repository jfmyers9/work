# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic
Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-02-15

### Added

- Filesystem-based issue tracker storing all data in `.work/` directory
- Full issue lifecycle: open, active, review, done, cancelled
- `work create` with `--type`, `--priority`, `--labels`, `--description`
- `work list` with filtering by status, label, type, priority
  - `--format short` for compact output
  - `--last` flag for recently changed issues
  - Title sorting for stable display order
- `work show` for detailed issue view
- `work edit` for modifying issues; opens `$EDITOR` with no flags
- `work comment` for inline comments on issues
- `work start`, `work review`, `work approve`, `work reject`, `work close`,
  `work cancel`, `work reopen` for status transitions
- `work link` / `work unlink` for parent-child issue relationships
- `work history` with label filtering across all issues
- `work log` for per-issue event log
- `work completed` for completion history
- `work compact` to archive completed issues and save space
- `work gc` with `--keep` flag for deleting old completed directories
- `work export` for JSON output
- `work init` with Claude Code hook configuration
- `work instructions` subcommand for SessionStart hook
- User identity support via `WORK_USER` / git config
- Review workflow with approve/reject cycle
- Issue types (task, bug, feature) and priority levels
- Interactive TUI for browsing and managing issues
- Global `--help` and `help` subcommand
- CLI built on Cobra framework
- Build system with Makefile, CI via GitHub Actions, GoReleaser

[unreleased]: https://github.com/jfmyers9/work/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/jfmyers9/work/releases/tag/v0.1.0
