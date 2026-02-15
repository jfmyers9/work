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

work show <id>             # Full issue details
work list                  # Table of all issues
work export                # All issues as JSON array to stdout
```

Issue IDs are 6-character hex strings. All commands accept unique
prefixes — `a3f` resolves to `a3f8b2` if unambiguous.

### Editing

```
work edit <id> [flags]
  --title <text>
  --description <text>
  --priority <n>
  --labels <a,b,c>
  --assignee <name>
```

### Lifecycle

Default states: `open` → `active` → `done` / `cancelled`

```
work status <id> <state>   # Explicit transition
work start <id>            # open → active
work close <id>            # → done
work cancel <id>           # → cancelled
work reopen <id>           # → open
```

Transitions are validated against `.work/config.json`. Invalid
moves are rejected with a clear error.

### Filtering and Sorting

```
work list --status=open
work list --label=bug --assignee=jim
work list --priority=1 --sort=updated
```

Filters combine with AND logic. Sort options: `priority` (ascending),
`updated` (newest first), `created` (newest first, default).

### History

```
work log <id>              # Events for one issue
work history               # Recent events across all issues (last 20)
work history --since=2026-01-01
work log <id> --since=2026-02-01 --until=2026-02-15
```

### Comments

```
work comment <id> "Fixed in commit abc123"
```

### JSON Output

```
work show <id> --format=json
work list --format=json
work export                # Always JSON
```

### Shell Completions

```
eval "$(work completion bash)"
eval "$(work completion zsh)"
```

## Storage Layout

```
.work/
  config.json                # States, transitions, defaults
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
  "states": ["open", "active", "done", "cancelled"],
  "transitions": {
    "open": ["active", "done", "cancelled"],
    "active": ["done", "cancelled", "open"],
    "done": ["open"],
    "cancelled": ["open"]
  },
  "default_state": "open",
  "id_length": 6
}
```

## Running Tests

```
go test ./...
```
