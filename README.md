# issues

A simple, filesystem-based issue tracker CLI written in Go.

Issues are stored as JSON files in a `.issues/` directory, making
them easy to browse, version with git, and merge without conflicts.

## Install

```
go build -o issues .
```

Move the binary somewhere on your `$PATH`:

```
mv issues /usr/local/bin/
```

## Quick Start

```
issues init
issues create "Fix login timeout" --priority 1 --labels bug,auth
issues start a3f
issues comment a3f "Root cause identified"
issues close a3f
```

## Commands

### Tracker Setup

```
issues init                  # Create .issues/ in current directory
```

### Creating and Viewing

```
issues create <title> [flags]
  --description <text>
  --priority <n>             # Lower number = higher priority
  --labels <a,b,c>
  --assignee <name>

issues show <id>             # Full issue details
issues list                  # Table of all issues
issues export                # All issues as JSON array to stdout
```

Issue IDs are 6-character hex strings. All commands accept unique
prefixes — `a3f` resolves to `a3f8b2` if unambiguous.

### Editing

```
issues edit <id> [flags]
  --title <text>
  --description <text>
  --priority <n>
  --labels <a,b,c>
  --assignee <name>
```

### Lifecycle

Default states: `open` → `active` → `done` / `cancelled`

```
issues status <id> <state>   # Explicit transition
issues start <id>            # open → active
issues close <id>            # → done
issues cancel <id>           # → cancelled
issues reopen <id>           # → open
```

Transitions are validated against `.issues/config.json`. Invalid
moves are rejected with a clear error.

### Filtering and Sorting

```
issues list --status=open
issues list --label=bug --assignee=jim
issues list --priority=1 --sort=updated
```

Filters combine with AND logic. Sort options: `priority` (ascending),
`updated` (newest first), `created` (newest first, default).

### History

```
issues log <id>              # Events for one issue
issues history               # Recent events across all issues (last 20)
issues history --since=2026-01-01
issues log <id> --since=2026-02-01 --until=2026-02-15
```

### Comments

```
issues comment <id> "Fixed in commit abc123"
```

### JSON Output

```
issues show <id> --format=json
issues list --format=json
issues export                # Always JSON
```

### Shell Completions

```
eval "$(issues completion bash)"
eval "$(issues completion zsh)"
```

## Storage Layout

```
.issues/
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

Edit `.issues/config.json` to customize states and transitions:

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
