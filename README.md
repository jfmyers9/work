# work

A simple, filesystem-based issue tracker CLI written in Go.

Issues are stored as JSON files in a `.work/` directory, making
them easy to browse, version with git, and merge without conflicts.

## Install

```
go build -o work .
```

Move the binary somewhere on your `$PATH`:

```
mv work /usr/local/bin/
```

## Quick Start

```
work init
work create "Fix login timeout" --priority 1 --labels bug,auth
work start a3f
work comment a3f "Root cause identified"
work close a3f
```

## Commands

### Tracker Setup

```
work init                  # Create .work/ in current directory
```

### Creating and Viewing

```
work create <title> [flags]
  --description <text>
  --priority <n>             # Lower number = higher priority
  --labels <a,b,c>
  --assignee <name>
  --type <feature|bug|chore> # Default: feature
  --parent <id>              # Link as child of parent issue

work show <id>             # Full issue details
work list                  # Table of all issues
work export                # All issues as JSON array to stdout
```

Issue IDs are 6-character hex strings. All commands accept unique
prefixes — `a3f` resolves to `a3f8b2` if unambiguous.

If an issue has been purged by `gc`, `work show` will tell you
and point to `work completed` for history.

### Editing

```
work edit <id> [flags]
  --title <text>
  --description <text>
  --priority <n>
  --labels <a,b,c>
  --assignee <name>
  --type <feature|bug|chore>
```

### Lifecycle

Default states: `open` → `active` → `review` → `done` / `cancelled`

```
work status <id> <state>   # Explicit transition
work start <id>            # open → active
work review <id>           # active → review
work approve <id>          # review → done
work reject <id> <reason>  # review → active (adds comment)
work close <id>            # → done
work cancel <id>           # → cancelled
work reopen <id>           # → open
```

Transitions are validated against `.work/config.json`. Invalid
moves are rejected with a clear error.

### Linking (Parent/Child)

```
work link <child-id> --parent <parent-id>
work unlink <child-id>
```

Link issues into parent/child hierarchies. `work show` on a
parent displays its children with completion progress.
No grandchildren — a parent cannot itself be a child.

### Filtering and Sorting

```
work list --status=open
work list --label=bug --assignee=jim
work list --priority=1 --sort=updated
work list --type=bug
work list --parent=a3f              # Children of a specific issue
work list --roots                   # Only top-level issues
```

Filters combine with AND logic. Sort options: `priority` (ascending),
`updated` (newest first), `created` (newest first, default), `title`
(alphabetically).

### History

```
work log <id>              # Events for one issue
work history               # Recent events across all issues (last 20)
work history --since=2026-01-01
work history --label=bug   # Filter to issues with label
work log <id> --since=2026-02-01 --until=2026-02-15
```

### Comments

```
work comment <id> "Fixed in commit abc123"
```

### Output Formats

```
work show <id> --format=json
work list --format=json
work list --format=short   # ID and title only
work export                # Always JSON
```

### Maintenance

```
work compact <id>          # Strip description/comments/history
work compact --all-done    # Compact all done/cancelled issues
work completed             # Show completion history from log
work completed --since=2026-01-01
work completed --label=bug --type=feature --format=json
work gc                    # Purge issues completed 30+ days ago
work gc --days 7           # Custom age threshold
```

Closing or cancelling an issue auto-compacts it. Use
`--no-compact` to preserve full history:

```
work close <id> --no-compact
work cancel <id> --no-compact
```

### Help

```
work --help                # List all commands
work help <command>        # Detailed help for a command
work <command> --help      # Same as above
```

### Shell Completions

```
eval "$(work completion bash)"
eval "$(work completion zsh)"
```

## Issue Types

Issues have a type: `feature` (default), `bug`, or `chore`.
Set on creation with `--type` or change later with `work edit`.
Types are configured in `.work/config.json`.

## User Identity

All events and comments record who performed the action.
Identity is resolved in order:

1. `WORK_USER` environment variable
2. `git config user.name`
3. `"system"` fallback

## Storage Layout

```
.work/
  config.json                # States, transitions, defaults
  log.jsonl                  # Completion log (compacted/purged issues)
  issues/
    <6-char-hex>/
      issue.json             # Current issue state (mutable)
      history.jsonl           # Append-only event log
```

Each issue lives in its own directory. This means git merges
almost never conflict — two people creating or editing different
issues touch different files.

## Configuration

Edit `.work/config.json` to customize states and transitions:

```json
{
  "transitions": {
    "open": ["active", "done", "cancelled"],
    "active": ["done", "cancelled", "open", "review"],
    "review": ["done", "active"],
    "done": ["open"],
    "cancelled": ["open"]
  },
  "default_state": "open",
  "types": ["feature", "bug", "chore"],
  "default_type": "feature",
  "id_length": 6
}
```

## Running Tests

```
go test ./...
```
