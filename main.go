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

type command struct {
	Name    string
	Summary string
	Usage   string
	Run     func()
}

var commands []command

func init() {
	commands = []command{
		{
			Name:    "init",
			Summary: "Initialize issue tracker in current directory",
			Usage: `usage: work init

Initialize a new work tracker in the current directory.
Creates a .work/ directory to store issues and configuration.`,
			Run: cmdInit,
		},
		{
			Name:    "create",
			Summary: "Create a new issue",
			Usage: `usage: work create <title> [flags]

Create a new issue with the given title.

Flags:
  --description <text>    Issue description
  --priority <N>          Priority level (integer)
  --labels <a,b>          Comma-separated labels
  --assignee <name>       Assignee name
  --type <type>           Issue type (feature|bug|chore)
  --parent <id>           Parent issue ID

Examples:
  work create "Fix login bug" --type bug --priority 1
  work create "Add search" --labels ui,search --assignee alice`,
			Run: cmdCreate,
		},
		{
			Name:    "show",
			Summary: "Show issue details",
			Usage: `usage: work show <id-or-prefix> [flags]

Display full details for a single issue, including comments
and child issues.

Flags:
  --format json    Output as JSON

Examples:
  work show abc123
  work show abc --format=json`,
			Run: cmdShow,
		},
		{
			Name:    "list",
			Summary: "List issues",
			Usage: `usage: work list [flags]

List issues with optional filtering and sorting.

Flags:
  --status <status>      Filter by status (open|active|review|done|cancelled)
  --label <label>        Filter by label
  --assignee <name>      Filter by assignee
  --type <type>          Filter by type
  --priority <N>         Filter by priority
  --parent <id>          Filter by parent issue
  --roots                Show only root issues (no parent)
  --sort <field>         Sort by field (title|priority|status|created|updated)
  --format json          Output as JSON
  --format short         Output compact id+title list

Examples:
  work list --status active
  work list --label backend --sort priority`,
			Run: cmdList,
		},
		{
			Name:    "edit",
			Summary: "Edit an issue",
			Usage: `usage: work edit <id-or-prefix> [flags]

Update fields on an existing issue.

Flags:
  --title <text>          New title
  --description <text>    New description
  --priority <N>          New priority
  --labels <a,b>          Replace labels
  --assignee <name>       New assignee
  --type <type>           New type (feature|bug|chore)

Examples:
  work edit abc123 --title "Updated title"
  work edit abc --priority 2 --labels urgent,backend`,
			Run: cmdEdit,
		},
		{
			Name:    "status",
			Summary: "Change issue status",
			Usage: `usage: work status <id-or-prefix> <state>

Set an issue's status to any valid state.
Valid states: open, active, review, done, cancelled.

Examples:
  work status abc123 active
  work status abc done`,
			Run: cmdStatus,
		},
		{
			Name:    "close",
			Summary: "Close an issue (set status to done)",
			Usage: `usage: work close <id-or-prefix>

Shortcut for: work status <id> done`,
			Run: func() { cmdShortcut("done") },
		},
		{
			Name:    "cancel",
			Summary: "Cancel an issue",
			Usage: `usage: work cancel <id-or-prefix>

Shortcut for: work status <id> cancelled`,
			Run: func() { cmdShortcut("cancelled") },
		},
		{
			Name:    "reopen",
			Summary: "Reopen an issue",
			Usage: `usage: work reopen <id-or-prefix>

Shortcut for: work status <id> open`,
			Run: func() { cmdShortcut("open") },
		},
		{
			Name:    "start",
			Summary: "Start working on an issue (set status to active)",
			Usage: `usage: work start <id-or-prefix>

Shortcut for: work status <id> active`,
			Run: func() { cmdShortcut("active") },
		},
		{
			Name:    "review",
			Summary: "Submit an issue for review (set status to review)",
			Usage: `usage: work review <id-or-prefix>

Shortcut for: work status <id> review`,
			Run: func() { cmdShortcut("review") },
		},
		{
			Name:    "approve",
			Summary: "Approve a reviewed issue (set status to done)",
			Usage: `usage: work approve <id-or-prefix>

Shortcut for: work status <id> done`,
			Run: func() { cmdShortcut("done") },
		},
		{
			Name:    "reject",
			Summary: "Reject a reviewed issue (back to active + reason comment)",
			Usage: `usage: work reject <id-or-prefix> <reason>

Set status back to active and add a rejection comment.

Examples:
  work reject abc123 "Tests are failing"`,
			Run: cmdReject,
		},
		{
			Name:    "comment",
			Summary: "Add a comment to an issue",
			Usage: `usage: work comment <id-or-prefix> <text>

Add a text comment to an issue.

Examples:
  work comment abc123 "Fixed in latest commit"`,
			Run: cmdComment,
		},
		{
			Name:    "link",
			Summary: "Link a child issue to a parent",
			Usage: `usage: work link <child-id> --parent <epic-id>

Set a parent-child relationship between two issues.

Examples:
  work link abc123 --parent def456`,
			Run: cmdLink,
		},
		{
			Name:    "unlink",
			Summary: "Remove parent from a child issue",
			Usage: `usage: work unlink <child-id>

Remove the parent link from a child issue.

Examples:
  work unlink abc123`,
			Run: cmdUnlink,
		},
		{
			Name:    "log",
			Summary: "Show issue event log",
			Usage: `usage: work log <id-or-prefix> [flags]

Display the event history for a single issue.

Flags:
  --since <date>    Show events after date (YYYY-MM-DD or RFC3339)
  --until <date>    Show events before date

Examples:
  work log abc123
  work log abc --since 2025-01-01`,
			Run: cmdLog,
		},
		{
			Name:    "history",
			Summary: "Show all events across issues",
			Usage: `usage: work history [flags]

Display recent events across all issues (most recent first,
limited to 20).

Flags:
  --label <label>   Filter to issues with this label
  --since <date>    Show events after date (YYYY-MM-DD or RFC3339)
  --until <date>    Show events before date

Examples:
  work history --since 2025-01-01
  work history --label backend`,
			Run: cmdHistory,
		},
		{
			Name:    "export",
			Summary: "Export issues as JSON",
			Usage: `usage: work export

Export all issues as a JSON array to stdout.`,
			Run: cmdExport,
		},
		{
			Name:    "completion",
			Summary: "Generate shell completions (bash|zsh)",
			Usage: `usage: work completion <bash|zsh>

Generate shell completion script for bash or zsh.

Examples:
  work completion bash > ~/.bash_completion.d/work
  work completion zsh > ~/.zfunc/_work`,
			Run: cmdCompletion,
		},
	}
}

func findCommand(name string) *command {
	for i := range commands {
		if commands[i].Name == name {
			return &commands[i]
		}
	}
	return nil
}

// printCommandUsage prints a command's usage text to the given writer and exits.
func printCommandUsage(cmd *command, w *os.File, code int) {
	fmt.Fprintln(w, cmd.Usage)
	os.Exit(code)
}

func main() {
	// Handle top-level -h/--help before requiring a subcommand.
	if len(os.Args) >= 2 {
		arg := os.Args[1]
		if arg == "--help" || arg == "-h" {
			printHelp(os.Stdout)
			os.Exit(0)
		}
	}

	if len(os.Args) < 2 {
		printHelp(os.Stderr)
		os.Exit(1)
	}

	// Handle "help" subcommand.
	if os.Args[1] == "help" {
		cmdHelp()
		return
	}

	cmd := findCommand(os.Args[1])
	if cmd == nil {
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}

	// Check for --help/-h on any subcommand before running it.
	for _, arg := range os.Args[2:] {
		if arg == "--help" || arg == "-h" {
			printCommandUsage(cmd, os.Stdout, 0)
		}
	}

	cmd.Run()
}

func cmdHelp() {
	if len(os.Args) < 3 {
		printHelp(os.Stdout)
		os.Exit(0)
	}
	name := os.Args[2]
	cmd := findCommand(name)
	if cmd == nil {
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", name)
		os.Exit(1)
	}
	printCommandUsage(cmd, os.Stdout, 0)
}

func printHelp(w *os.File) {
	fmt.Fprintln(w, "usage: work <command> [args]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Commands:")
	for _, cmd := range commands {
		fmt.Fprintf(w, "  %-12s%s\n", cmd.Name, cmd.Summary)
	}
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

// booleanFlags lists flags that take no value.
var booleanFlags = map[string]bool{
	"roots": true,
	"help":  true,
}

// parseFlags extracts --key=value, --key value, and -h pairs from args.
// Flags listed in booleanFlags are treated as present/absent with no value.
func parseFlags(args []string) ([]string, map[string]string) {
	var positional []string
	flags := make(map[string]string)
	for i := 0; i < len(args); i++ {
		if args[i] == "-h" {
			flags["help"] = ""
		} else if strings.HasPrefix(args[i], "--") {
			raw := strings.TrimPrefix(args[i], "--")
			if k, v, ok := strings.Cut(raw, "="); ok {
				flags[k] = v
			} else if booleanFlags[raw] {
				flags[raw] = ""
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

// commandUsage prints the registry usage for the named command to stderr and exits 1.
// Used by command functions when required args are missing.
func commandUsage(name string) {
	cmd := findCommand(name)
	if cmd != nil {
		printCommandUsage(cmd, os.Stderr, 1)
	}
	fmt.Fprintf(os.Stderr, "usage: work %s\n", name)
	os.Exit(1)
}

func cmdCreate() {
	args := os.Args[2:]
	positional, flags := parseFlags(args)

	if len(positional) == 0 {
		commandUsage("create")
	}
	title := positional[0]
	description := flags["description"]
	assignee := flags["assignee"]
	issueType := flags["type"]
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

	parentID := ""
	if p, ok := flags["parent"]; ok {
		resolved, err := t.ResolvePrefix(p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		parentID = resolved
	}

	user := tracker.ResolveUser()
	issue, err := t.CreateIssue(title, description, assignee, priority, labels, issueType, parentID, user)
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
		commandUsage("show")
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
	fmt.Printf("Type:        %s\n", issue.Type)
	fmt.Printf("Priority:    %d\n", issue.Priority)
	if len(issue.Labels) > 0 {
		fmt.Printf("Labels:      %s\n", strings.Join(issue.Labels, ", "))
	}
	if issue.Assignee != "" {
		fmt.Printf("Assignee:    %s\n", issue.Assignee)
	}
	if issue.ParentID != "" {
		fmt.Printf("Parent:      %s\n", issue.ParentID)
	}
	if issue.Description != "" {
		fmt.Printf("Description: %s\n", issue.Description)
	}
	fmt.Printf("Created:     %s\n", issue.Created.Format(time.RFC3339))
	fmt.Printf("Updated:     %s\n", issue.Updated.Format(time.RFC3339))

	// Show children if this issue is a parent
	allIssues, err := t.ListIssues()
	if err == nil {
		var children []model.Issue
		for _, i := range allIssues {
			if i.ParentID == issue.ID {
				children = append(children, i)
			}
		}
		if len(children) > 0 {
			done := 0
			for _, c := range children {
				if c.Status == "done" || c.Status == "cancelled" {
					done++
				}
			}
			fmt.Printf("\nChildren: %d/%d done\n", done, len(children))
			for _, c := range children {
				fmt.Printf("  %-8s %-10s %s\n", c.ID, c.Status, c.Title)
			}
		}
	}

	if len(issue.Comments) > 0 {
		fmt.Printf("\nComments:\n")
		for _, comment := range issue.Comments {
			fmt.Printf("  [%s] (%s): %s\n",
				comment.Created.Format("2006-01-02 15:04:05"),
				comment.By,
				comment.Text)
		}
	}
}

func cmdList() {
	_, flags := parseFlags(os.Args[2:])

	t := loadTracker()
	allIssues, err := t.ListIssues()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	opts := tracker.FilterOptions{
		Status:   flags["status"],
		Label:    flags["label"],
		Assignee: flags["assignee"],
		Type:     flags["type"],
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
	if p, ok := flags["parent"]; ok {
		resolved, err := t.ResolvePrefix(p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		opts.ParentID = resolved
	}
	if _, ok := flags["roots"]; ok {
		opts.RootsOnly = true
	}
	issues := tracker.FilterIssues(allIssues, opts)
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

	if flags["format"] == "short" {
		for _, issue := range issues {
			fmt.Printf("%s %s\n", issue.ID, issue.Title)
		}
		return
	}

	if len(issues) == 0 {
		fmt.Println("No issues found")
		return
	}

	childCounts := make(map[string]struct{ done, total int })
	for _, issue := range allIssues {
		if issue.ParentID != "" {
			c := childCounts[issue.ParentID]
			c.total++
			if issue.Status == "done" || issue.Status == "cancelled" {
				c.done++
			}
			childCounts[issue.ParentID] = c
		}
	}

	fmt.Printf("%-8s %-10s %-10s %-8s %-10s %s\n", "ID", "STATUS", "TYPE", "PRIORITY", "CHILDREN", "TITLE")
	for _, issue := range issues {
		title := issue.Title
		if len(title) > 50 {
			title = title[:47] + "..."
		}
		children := ""
		if c, ok := childCounts[issue.ID]; ok {
			children = fmt.Sprintf("%d/%d", c.done, c.total)
		}
		fmt.Printf("%-8s %-10s %-10s %-8d %-10s %s\n", issue.ID, issue.Status, issue.Type, issue.Priority, children, title)
	}
}

func cmdEdit() {
	args := os.Args[2:]
	positional, flags := parseFlags(args)

	if len(positional) == 0 {
		commandUsage("edit")
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
	if v, ok := flags["type"]; ok {
		if err := tracker.ValidateType(t.Config, v); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		issue.Type = v
		edited = append(edited, "type")
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

	user := tracker.ResolveUser()
	event := model.Event{
		Timestamp: now,
		Op:        "edit",
		Fields:    edited,
		By:        user,
	}
	if err := t.AppendEvent(id, event); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Updated %s\n", id)
}

func cmdStatus() {
	if len(os.Args) < 4 {
		commandUsage("status")
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

	user := tracker.ResolveUser()
	if _, err := t.SetStatus(id, newStatus, user); err != nil {
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
	case "link":
		return fmt.Sprintf("link: parent=%s", ev.To)
	case "unlink":
		return fmt.Sprintf("unlink: was parent=%s", ev.From)
	default:
		return ev.Op
	}
}

func cmdLog() {
	args := os.Args[2:]
	positional, flags := parseFlags(args)

	if len(positional) == 0 {
		commandUsage("log")
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

	if label, ok := flags["label"]; ok {
		issues, err := t.ListIssues()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		validIDs := make(map[string]bool)
		for _, issue := range issues {
			for _, l := range issue.Labels {
				if l == label {
					validIDs[issue.ID] = true
					break
				}
			}
		}
		var filtered []tracker.EventWithIssue
		for _, ev := range all {
			if validIDs[ev.IssueID] {
				filtered = append(filtered, ev)
			}
		}
		all = filtered
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
		commandUsage("comment")
	}
	prefix := os.Args[2]
	text := os.Args[3]
	t := loadTracker()

	id, err := t.ResolvePrefix(prefix)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	user := tracker.ResolveUser()
	if _, err := t.AddComment(id, text, user); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Commented on %s\n", id)
}

func cmdLink() {
	args := os.Args[2:]
	positional, flags := parseFlags(args)

	if len(positional) == 0 {
		commandUsage("link")
	}
	parentPrefix, ok := flags["parent"]
	if !ok {
		commandUsage("link")
	}

	t := loadTracker()

	childID, err := t.ResolvePrefix(positional[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	parentID, err := t.ResolvePrefix(parentPrefix)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	user := tracker.ResolveUser()
	if _, err := t.LinkIssue(childID, parentID, user); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Linked %s → %s\n", childID, parentID)
}

func cmdUnlink() {
	if len(os.Args) < 3 {
		commandUsage("unlink")
	}
	prefix := os.Args[2]
	t := loadTracker()

	id, err := t.ResolvePrefix(prefix)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	user := tracker.ResolveUser()
	if _, err := t.UnlinkIssue(id, user); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Unlinked %s\n", id)
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
		commandUsage("completion")
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

func commandNames() string {
	var names []string
	for _, cmd := range commands {
		names = append(names, cmd.Name)
	}
	names = append(names, "help")
	return strings.Join(names, " ")
}

func commandNamesWithID() string {
	needsID := map[string]bool{
		"show": true, "edit": true, "status": true,
		"close": true, "cancel": true, "reopen": true,
		"start": true, "review": true, "approve": true,
		"reject": true, "comment": true, "link": true,
		"unlink": true, "log": true,
	}
	var names []string
	for _, cmd := range commands {
		if needsID[cmd.Name] {
			names = append(names, cmd.Name)
		}
	}
	return strings.Join(names, "|")
}

func printBashCompletion() {
	fmt.Printf(`_work() {
    local cur prev commands
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    commands="%s"

    if [ "$COMP_CWORD" -eq 1 ]; then
        COMPREPLY=($(compgen -W "$commands" -- "$cur"))
        return 0
    fi

    case "$prev" in
        %s)
            local ids
            ids=$(ls .work/issues/ 2>/dev/null)
            COMPREPLY=($(compgen -W "$ids" -- "$cur"))
            return 0
            ;;
        status)
            COMPREPLY=($(compgen -W "open active review done cancelled" -- "$cur"))
            return 0
            ;;
        completion)
            COMPREPLY=($(compgen -W "bash zsh" -- "$cur"))
            return 0
            ;;
    esac
}
complete -F _work work
`, commandNames(), commandNamesWithID())
}

func printZshCompletion() {
	fmt.Print(`#compdef work

_work() {
    local -a commands
    commands=(
`)
	for _, cmd := range commands {
		fmt.Printf("        '%s:%s'\n", cmd.Name, cmd.Summary)
	}
	fmt.Print(`        'help:Show help for a command'
    )

    if (( CURRENT == 2 )); then
        _describe 'command' commands
        return
    fi

    case "${words[2]}" in
        `)
	fmt.Print(commandNamesWithID())
	fmt.Print(`)
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
		commandUsage(os.Args[1])
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

	user := tracker.ResolveUser()
	if _, err := t.SetStatus(id, targetStatus, user); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("%s: %s → %s\n", id, oldStatus, targetStatus)
}

func cmdReject() {
	if len(os.Args) < 4 {
		commandUsage("reject")
	}
	prefix := os.Args[2]
	reason := os.Args[3]
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

	user := tracker.ResolveUser()
	if _, err := t.SetStatus(id, "active", user); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if _, err := t.AddComment(id, "Rejected: "+reason, user); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("%s: %s → active (rejected: %s)\n", id, oldStatus, reason)
}
