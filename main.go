package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jfmyers9/work/internal/model"
	"github.com/jfmyers9/work/internal/tracker"
)

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "init":
		cmdInit()
	case "create":
		cmdCreate()
	case "show":
		cmdShow()
	case "list":
		cmdList()
	case "edit":
		cmdEdit()
	case "status":
		cmdStatus()
	case "close":
		cmdShortcut("done")
	case "cancel":
		cmdShortcut("cancelled")
	case "reopen":
		cmdShortcut("open")
	case "start":
		cmdShortcut("active")
	case "comment":
		cmdComment()
	case "export":
		cmdExport()
	case "completion":
		cmdCompletion()
	case "log":
		cmdLog()
	case "history":
		cmdHistory()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Fprintln(os.Stderr, `usage: work <command> [args]

Commands:
  init        Initialize issue tracker in current directory
  create      Create a new issue
  show        Show issue details
  list        List issues
  edit        Edit an issue
  status      Change issue status
  close       Close an issue (set status to done)
  cancel      Cancel an issue
  reopen      Reopen an issue
  start       Start working on an issue (set status to active)
  comment     Add a comment to an issue
  log         Show issue event log
  history     Show all events across issues
  export      Export issues as JSON
  completion  Generate shell completions (bash|zsh)`)
}

func cmdInit() {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	_, err = tracker.Init(wd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Initialized work tracker in .work/")
}

func loadTracker() *tracker.Tracker {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	t, err := tracker.Load(wd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	return t
}

// parseFlags extracts --key=value and --key value pairs from args.
func parseFlags(args []string) ([]string, map[string]string) {
	var positional []string
	flags := make(map[string]string)
	for i := 0; i < len(args); i++ {
		if strings.HasPrefix(args[i], "--") {
			raw := strings.TrimPrefix(args[i], "--")
			if k, v, ok := strings.Cut(raw, "="); ok {
				flags[k] = v
			} else if i+1 < len(args) {
				flags[raw] = args[i+1]
				i++
			}
		} else {
			positional = append(positional, args[i])
		}
	}
	return positional, flags
}

func cmdCreate() {
	args := os.Args[2:]
	positional, flags := parseFlags(args)

	if len(positional) == 0 {
		fmt.Fprintln(os.Stderr, "usage: work create <title> [--description ...] [--priority N] [--labels a,b] [--assignee name]")
		os.Exit(1)
	}
	title := positional[0]
	description := flags["description"]
	assignee := flags["assignee"]
	priority := 0
	if p, ok := flags["priority"]; ok {
		n, err := strconv.Atoi(p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid priority: %s\n", p)
			os.Exit(1)
		}
		priority = n
	}
	var labels []string
	if l, ok := flags["labels"]; ok {
		labels = strings.Split(l, ",")
	}

	t := loadTracker()
	issue, err := t.CreateIssue(title, description, assignee, priority, labels)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(issue.ID)
}

func cmdShow() {
	args := os.Args[2:]
	positional, flags := parseFlags(args)

	if len(positional) == 0 {
		fmt.Fprintln(os.Stderr, "usage: work show <id-or-prefix> [--format=json]")
		os.Exit(1)
	}
	prefix := positional[0]
	t := loadTracker()

	id, err := t.ResolvePrefix(prefix)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	issue, err := t.LoadIssue(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if flags["format"] == "json" {
		data, err := json.MarshalIndent(issue, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(data))
		return
	}

	fmt.Printf("ID:          %s\n", issue.ID)
	fmt.Printf("Title:       %s\n", issue.Title)
	fmt.Printf("Status:      %s\n", issue.Status)
	fmt.Printf("Priority:    %d\n", issue.Priority)
	if len(issue.Labels) > 0 {
		fmt.Printf("Labels:      %s\n", strings.Join(issue.Labels, ", "))
	}
	if issue.Assignee != "" {
		fmt.Printf("Assignee:    %s\n", issue.Assignee)
	}
	if issue.Description != "" {
		fmt.Printf("Description: %s\n", issue.Description)
	}
	fmt.Printf("Created:     %s\n", issue.Created.Format(time.RFC3339))
	fmt.Printf("Updated:     %s\n", issue.Updated.Format(time.RFC3339))
}

func cmdList() {
	_, flags := parseFlags(os.Args[2:])

	t := loadTracker()
	issues, err := t.ListIssues()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	opts := tracker.FilterOptions{
		Status:   flags["status"],
		Label:    flags["label"],
		Assignee: flags["assignee"],
	}
	if p, ok := flags["priority"]; ok {
		n, err := strconv.Atoi(p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid priority: %s\n", p)
			os.Exit(1)
		}
		opts.Priority = n
		opts.HasPriority = true
	}
	issues = tracker.FilterIssues(issues, opts)
	tracker.SortIssues(issues, flags["sort"])

	if flags["format"] == "json" {
		data, err := json.MarshalIndent(issues, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(data))
		return
	}

	if len(issues) == 0 {
		fmt.Println("No issues found")
		return
	}

	fmt.Printf("%-8s %-10s %-8s %s\n", "ID", "STATUS", "PRIORITY", "TITLE")
	for _, issue := range issues {
		title := issue.Title
		if len(title) > 50 {
			title = title[:47] + "..."
		}
		fmt.Printf("%-8s %-10s %-8d %s\n", issue.ID, issue.Status, issue.Priority, title)
	}
}

func cmdEdit() {
	args := os.Args[2:]
	positional, flags := parseFlags(args)

	if len(positional) == 0 {
		fmt.Fprintln(os.Stderr, "usage: work edit <id-or-prefix> [--title ...] [--description ...] [--priority N] [--labels a,b] [--assignee name]")
		os.Exit(1)
	}
	prefix := positional[0]
	t := loadTracker()

	id, err := t.ResolvePrefix(prefix)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	issue, err := t.LoadIssue(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	var edited []string
	if v, ok := flags["title"]; ok {
		issue.Title = v
		edited = append(edited, "title")
	}
	if v, ok := flags["description"]; ok {
		issue.Description = v
		edited = append(edited, "description")
	}
	if v, ok := flags["assignee"]; ok {
		issue.Assignee = v
		edited = append(edited, "assignee")
	}
	if p, ok := flags["priority"]; ok {
		n, err := strconv.Atoi(p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid priority: %s\n", p)
			os.Exit(1)
		}
		issue.Priority = n
		edited = append(edited, "priority")
	}
	if l, ok := flags["labels"]; ok {
		issue.Labels = strings.Split(l, ",")
		edited = append(edited, "labels")
	}

	if len(edited) == 0 {
		fmt.Fprintln(os.Stderr, "no fields to update")
		os.Exit(1)
	}

	now := time.Now().UTC()
	issue.Updated = now
	if err := t.SaveIssue(issue); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	event := model.Event{
		Timestamp: now,
		Op:        "edit",
		Fields:    edited,
		By:        "system",
	}
	if err := t.AppendEvent(id, event); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Updated %s\n", id)
}

func cmdStatus() {
	if len(os.Args) < 4 {
		fmt.Fprintln(os.Stderr, "usage: work status <id-or-prefix> <state>")
		os.Exit(1)
	}
	prefix := os.Args[2]
	newStatus := os.Args[3]
	t := loadTracker()

	id, err := t.ResolvePrefix(prefix)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	old, err := t.LoadIssue(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	oldStatus := old.Status

	if _, err := t.SetStatus(id, newStatus); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("%s: %s → %s\n", id, oldStatus, newStatus)
}

func parseTimeFlag(s string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("invalid time %q (use YYYY-MM-DD or RFC3339)", s)
}

func formatEventDetail(ev model.Event) string {
	switch ev.Op {
	case "status":
		return fmt.Sprintf("status: %s → %s", ev.From, ev.To)
	case "edit":
		return fmt.Sprintf("edit: %s", strings.Join(ev.Fields, ", "))
	case "comment":
		text := ev.Text
		if len(text) > 60 {
			text = text[:57] + "..."
		}
		return fmt.Sprintf("comment: %s", text)
	default:
		return ev.Op
	}
}

func cmdLog() {
	args := os.Args[2:]
	positional, flags := parseFlags(args)

	if len(positional) == 0 {
		fmt.Fprintln(os.Stderr, "usage: work log <id-or-prefix> [--since DATE] [--until DATE]")
		os.Exit(1)
	}
	t := loadTracker()

	id, err := t.ResolvePrefix(positional[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	events, err := t.LoadEvents(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	var since, until time.Time
	if s, ok := flags["since"]; ok {
		since, err = parseTimeFlag(s)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	}
	if u, ok := flags["until"]; ok {
		until, err = parseTimeFlag(u)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	}
	events = tracker.FilterEventsByTime(events, since, until)

	if len(events) == 0 {
		fmt.Println("No events")
		return
	}
	for _, ev := range events {
		fmt.Printf("%s  %s  (%s)\n",
			ev.Timestamp.Format("2006-01-02 15:04:05"),
			formatEventDetail(ev),
			ev.By)
	}
}

func cmdHistory() {
	_, flags := parseFlags(os.Args[2:])
	t := loadTracker()

	all, err := t.LoadAllEvents()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	var since, until time.Time
	if s, ok := flags["since"]; ok {
		since, err = parseTimeFlag(s)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	}
	if u, ok := flags["until"]; ok {
		until, err = parseTimeFlag(u)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	}
	all = tracker.FilterEventsWithIssueByTime(all, since, until)

	// Sort by timestamp descending (most recent first)
	sort.Slice(all, func(i, j int) bool {
		return all[i].Timestamp.After(all[j].Timestamp)
	})

	if len(all) == 0 {
		fmt.Println("No events")
		return
	}

	limit := 20
	if len(all) < limit {
		limit = len(all)
	}
	for _, ev := range all[:limit] {
		fmt.Printf("%s  %s  %s  (%s)\n",
			ev.Timestamp.Format("2006-01-02 15:04:05"),
			ev.IssueID,
			formatEventDetail(ev.Event),
			ev.By)
	}
}

func cmdComment() {
	if len(os.Args) < 4 {
		fmt.Fprintln(os.Stderr, "usage: work comment <id-or-prefix> <text>")
		os.Exit(1)
	}
	prefix := os.Args[2]
	text := os.Args[3]
	t := loadTracker()

	id, err := t.ResolvePrefix(prefix)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if _, err := t.AddComment(id, text); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Commented on %s\n", id)
}

func cmdExport() {
	t := loadTracker()
	issues, err := t.ListIssues()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	data, err := json.MarshalIndent(issues, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}

func cmdCompletion() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: work completion <bash|zsh>")
		os.Exit(1)
	}
	switch os.Args[2] {
	case "bash":
		printBashCompletion()
	case "zsh":
		printZshCompletion()
	default:
		fmt.Fprintf(os.Stderr, "unsupported shell: %s (use bash or zsh)\n", os.Args[2])
		os.Exit(1)
	}
}

func issueIDs() []string {
	wd, err := os.Getwd()
	if err != nil {
		return nil
	}
	entries, err := os.ReadDir(filepath.Join(wd, ".work", "issues"))
	if err != nil {
		return nil
	}
	var ids []string
	for _, e := range entries {
		if e.IsDir() {
			ids = append(ids, e.Name())
		}
	}
	return ids
}

func printBashCompletion() {
	fmt.Print(`_work() {
    local cur prev commands
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    commands="init create show list edit status close cancel reopen start comment log history export completion"

    if [ "$COMP_CWORD" -eq 1 ]; then
        COMPREPLY=($(compgen -W "$commands" -- "$cur"))
        return 0
    fi

    case "$prev" in
        show|edit|status|close|cancel|reopen|start|comment|log)
            local ids
            ids=$(ls .work/issues/ 2>/dev/null)
            COMPREPLY=($(compgen -W "$ids" -- "$cur"))
            return 0
            ;;
        status)
            COMPREPLY=($(compgen -W "open active done cancelled" -- "$cur"))
            return 0
            ;;
        completion)
            COMPREPLY=($(compgen -W "bash zsh" -- "$cur"))
            return 0
            ;;
    esac
}
complete -F _work work
`)
}

func printZshCompletion() {
	fmt.Print(`#compdef work

_work() {
    local -a commands
    commands=(
        'init:Initialize work tracker'
        'create:Create a new issue'
        'show:Show issue details'
        'list:List issues'
        'edit:Edit an issue'
        'status:Change issue status'
        'close:Close an issue'
        'cancel:Cancel an issue'
        'reopen:Reopen an issue'
        'start:Start working on an issue'
        'comment:Add a comment to an issue'
        'log:Show issue event log'
        'history:Show all events'
        'export:Export issues as JSON'
        'completion:Generate shell completions'
    )

    if (( CURRENT == 2 )); then
        _describe 'command' commands
        return
    fi

    case "${words[2]}" in
        show|edit|status|close|cancel|reopen|start|comment|log)
            local -a ids
            ids=(${(f)"$(ls .work/issues/ 2>/dev/null)"})
            _describe 'issue' ids
            ;;
        completion)
            _values 'shell' bash zsh
            ;;
    esac
}

_work "$@"
`)
}

func cmdShortcut(targetStatus string) {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "usage: work %s <id-or-prefix>\n", os.Args[1])
		os.Exit(1)
	}
	prefix := os.Args[2]
	t := loadTracker()

	id, err := t.ResolvePrefix(prefix)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	old, err := t.LoadIssue(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	oldStatus := old.Status

	if _, err := t.SetStatus(id, targetStatus); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("%s: %s → %s\n", id, oldStatus, targetStatus)
}
