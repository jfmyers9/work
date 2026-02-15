package tracker

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jfmyers9/work/internal/model"
)

// FilterOptions specifies criteria for filtering issues. All non-zero fields
// are combined with AND logic.
type FilterOptions struct {
	Status   string
	Label    string
	Assignee string
	Priority int
	HasPriority bool // distinguishes "filter by priority 0" from "no filter"
}

// FilterIssues returns the subset of issues matching all specified filters.
func FilterIssues(issues []model.Issue, opts FilterOptions) []model.Issue {
	var result []model.Issue
	for _, issue := range issues {
		if opts.Status != "" && issue.Status != opts.Status {
			continue
		}
		if opts.Label != "" && !hasLabel(issue.Labels, opts.Label) {
			continue
		}
		if opts.Assignee != "" && issue.Assignee != opts.Assignee {
			continue
		}
		if opts.HasPriority && issue.Priority != opts.Priority {
			continue
		}
		result = append(result, issue)
	}
	return result
}

func hasLabel(labels []string, target string) bool {
	for _, l := range labels {
		if l == target {
			return true
		}
	}
	return false
}

// SortIssues sorts issues in place by the given field.
// Supported: "priority" (ascending), "updated" (newest first),
// "created" (newest first, default).
func SortIssues(issues []model.Issue, sortBy string) {
	switch sortBy {
	case "priority":
		sort.Slice(issues, func(i, j int) bool {
			return issues[i].Priority < issues[j].Priority
		})
	case "updated":
		sort.Slice(issues, func(i, j int) bool {
			return issues[i].Updated.After(issues[j].Updated)
		})
	default: // "created" or empty
		sort.Slice(issues, func(i, j int) bool {
			return issues[i].Created.After(issues[j].Created)
		})
	}
}

type Tracker struct {
	Root   string
	Config model.Config
}

// GenerateID produces a random hex string of the configured length.
func (t *Tracker) GenerateID() (string, error) {
	// Read enough random bytes, then hex-encode and truncate.
	buf := make([]byte, (t.Config.IDLength+1)/2)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generating id: %w", err)
	}
	return hex.EncodeToString(buf)[:t.Config.IDLength], nil
}

// Init creates the .issues directory structure and writes the default config.
func Init(root string) (*Tracker, error) {
	issuesDir := filepath.Join(root, ".issues", "issues")
	if err := os.MkdirAll(issuesDir, 0o755); err != nil {
		return nil, fmt.Errorf("creating issues dir: %w", err)
	}

	cfg := model.DefaultConfig()
	cfgPath := filepath.Join(root, ".issues", "config.json")

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling config: %w", err)
	}
	if err := os.WriteFile(cfgPath, data, 0o644); err != nil {
		return nil, fmt.Errorf("writing config: %w", err)
	}

	return &Tracker{Root: root, Config: cfg}, nil
}

// Load reads an existing tracker from disk.
func Load(root string) (*Tracker, error) {
	cfgPath := filepath.Join(root, ".issues", "config.json")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}
	var cfg model.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	return &Tracker{Root: root, Config: cfg}, nil
}

// SaveIssue writes an issue to its directory.
func (t *Tracker) SaveIssue(issue model.Issue) error {
	dir := filepath.Join(t.Root, ".issues", "issues", issue.ID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating issue dir: %w", err)
	}
	data, err := json.MarshalIndent(issue, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling issue: %w", err)
	}
	return os.WriteFile(filepath.Join(dir, "issue.json"), data, 0o644)
}

// LoadIssue reads an issue from its directory.
func (t *Tracker) LoadIssue(id string) (model.Issue, error) {
	path := filepath.Join(t.Root, ".issues", "issues", id, "issue.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return model.Issue{}, fmt.Errorf("reading issue: %w", err)
	}
	var issue model.Issue
	if err := json.Unmarshal(data, &issue); err != nil {
		return model.Issue{}, fmt.Errorf("parsing issue: %w", err)
	}
	return issue, nil
}

// AppendEvent writes an event as a JSON line to the issue's history.jsonl.
func (t *Tracker) AppendEvent(id string, event model.Event) error {
	path := filepath.Join(t.Root, ".issues", "issues", id, "history.jsonl")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("opening history: %w", err)
	}
	defer f.Close()
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshaling event: %w", err)
	}
	if _, err := fmt.Fprintf(f, "%s\n", data); err != nil {
		return fmt.Errorf("writing event: %w", err)
	}
	return nil
}

// CreateIssue generates an ID, saves the issue, and records a creation event.
func (t *Tracker) CreateIssue(title, description, assignee string, priority int, labels []string) (model.Issue, error) {
	id, err := t.GenerateID()
	if err != nil {
		return model.Issue{}, err
	}
	now := time.Now().UTC()
	issue := model.Issue{
		ID:          id,
		Title:       title,
		Description: description,
		Status:      t.Config.DefaultState,
		Priority:    priority,
		Labels:      labels,
		Assignee:    assignee,
		Created:     now,
		Updated:     now,
	}
	if err := t.SaveIssue(issue); err != nil {
		return model.Issue{}, err
	}
	event := model.Event{
		Timestamp: now,
		Op:        "create",
		By:        "system",
	}
	if err := t.AppendEvent(id, event); err != nil {
		return model.Issue{}, err
	}
	return issue, nil
}

// ResolvePrefix finds an issue ID matching the given prefix.
// Returns an error if zero or multiple issues match.
func (t *Tracker) ResolvePrefix(prefix string) (string, error) {
	issuesDir := filepath.Join(t.Root, ".issues", "issues")
	entries, err := os.ReadDir(issuesDir)
	if err != nil {
		return "", fmt.Errorf("reading issues dir: %w", err)
	}
	var matches []string
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), prefix) {
			matches = append(matches, e.Name())
		}
	}
	switch len(matches) {
	case 0:
		return "", fmt.Errorf("no issue found with prefix %q", prefix)
	case 1:
		return matches[0], nil
	default:
		return "", fmt.Errorf("ambiguous prefix %q, matches: %s", prefix, strings.Join(matches, ", "))
	}
}

// ValidateTransition checks if moving from one state to another is allowed by config.
func ValidateTransition(cfg model.Config, from, to string) error {
	if from == to {
		return fmt.Errorf("invalid transition: already in state %q", from)
	}
	allowed, ok := cfg.Transitions[from]
	if !ok {
		return fmt.Errorf("invalid transition: unknown state %q", from)
	}
	for _, s := range allowed {
		if s == to {
			return nil
		}
	}
	return fmt.Errorf("invalid transition: cannot move from %q to %q (allowed: %s)", from, to, strings.Join(allowed, ", "))
}

// SetStatus validates the transition and updates the issue's status.
func (t *Tracker) SetStatus(id, newStatus string) (model.Issue, error) {
	issue, err := t.LoadIssue(id)
	if err != nil {
		return model.Issue{}, err
	}
	if err := ValidateTransition(t.Config, issue.Status, newStatus); err != nil {
		return model.Issue{}, err
	}
	oldStatus := issue.Status
	now := time.Now().UTC()
	issue.Status = newStatus
	issue.Updated = now
	if err := t.SaveIssue(issue); err != nil {
		return model.Issue{}, err
	}
	event := model.Event{
		Timestamp: now,
		Op:        "status",
		From:      oldStatus,
		To:        newStatus,
		By:        "system",
	}
	if err := t.AppendEvent(id, event); err != nil {
		return model.Issue{}, err
	}
	return issue, nil
}

// EventWithIssue pairs an event with the issue ID it belongs to.
type EventWithIssue struct {
	model.Event
	IssueID string
}

// LoadEvents reads all events from an issue's history.jsonl.
// Returns empty slice (not error) if the file doesn't exist.
func (t *Tracker) LoadEvents(issueID string) ([]model.Event, error) {
	path := filepath.Join(t.Root, ".issues", "issues", issueID, "history.jsonl")
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("opening history: %w", err)
	}
	defer f.Close()

	var events []model.Event
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var ev model.Event
		if err := json.Unmarshal(scanner.Bytes(), &ev); err != nil {
			return nil, fmt.Errorf("parsing event: %w", err)
		}
		events = append(events, ev)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading history: %w", err)
	}
	return events, nil
}

// LoadAllEvents reads events from every issue, annotated with issue ID.
func (t *Tracker) LoadAllEvents() ([]EventWithIssue, error) {
	issuesDir := filepath.Join(t.Root, ".issues", "issues")
	entries, err := os.ReadDir(issuesDir)
	if err != nil {
		return nil, fmt.Errorf("reading issues dir: %w", err)
	}
	var all []EventWithIssue
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		events, err := t.LoadEvents(e.Name())
		if err != nil {
			return nil, err
		}
		for _, ev := range events {
			all = append(all, EventWithIssue{Event: ev, IssueID: e.Name()})
		}
	}
	return all, nil
}

// FilterEventsByTime returns events within the given time window.
// Zero-value since/until means no bound on that side.
func FilterEventsByTime(events []model.Event, since, until time.Time) []model.Event {
	var result []model.Event
	for _, ev := range events {
		if !since.IsZero() && ev.Timestamp.Before(since) {
			continue
		}
		if !until.IsZero() && !ev.Timestamp.Before(until) {
			continue
		}
		result = append(result, ev)
	}
	return result
}

// FilterEventsWithIssueByTime is like FilterEventsByTime but for EventWithIssue slices.
func FilterEventsWithIssueByTime(events []EventWithIssue, since, until time.Time) []EventWithIssue {
	var result []EventWithIssue
	for _, ev := range events {
		if !since.IsZero() && ev.Timestamp.Before(since) {
			continue
		}
		if !until.IsZero() && !ev.Timestamp.Before(until) {
			continue
		}
		result = append(result, ev)
	}
	return result
}

// AddComment appends a comment to the issue and records a history event.
func (t *Tracker) AddComment(id, text string) (model.Issue, error) {
	issue, err := t.LoadIssue(id)
	if err != nil {
		return model.Issue{}, err
	}
	now := time.Now().UTC()
	comment := model.Comment{
		Text:    text,
		Created: now,
		By:      "system",
	}
	issue.Comments = append(issue.Comments, comment)
	issue.Updated = now
	if err := t.SaveIssue(issue); err != nil {
		return model.Issue{}, err
	}
	event := model.Event{
		Timestamp: now,
		Op:        "comment",
		Text:      text,
		By:        "system",
	}
	if err := t.AppendEvent(id, event); err != nil {
		return model.Issue{}, err
	}
	return issue, nil
}

// ListIssues loads all issues from the tracker.
func (t *Tracker) ListIssues() ([]model.Issue, error) {
	issuesDir := filepath.Join(t.Root, ".issues", "issues")
	entries, err := os.ReadDir(issuesDir)
	if err != nil {
		return nil, fmt.Errorf("reading issues dir: %w", err)
	}
	var issues []model.Issue
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		issue, err := t.LoadIssue(e.Name())
		if err != nil {
			return nil, err
		}
		issues = append(issues, issue)
	}
	return issues, nil
}
