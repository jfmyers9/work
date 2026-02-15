package tracker

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jfmyers9/work/internal/model"
)

// ResolveUser determines the current user identity.
// Checks WORK_USER env, then git config user.name, then falls back to "system".
func ResolveUser() string {
	if u := os.Getenv("WORK_USER"); u != "" {
		return u
	}
	out, err := exec.Command("git", "config", "user.name").Output()
	if err == nil {
		if name := strings.TrimSpace(string(out)); name != "" {
			return name
		}
	}
	return "system"
}

// FilterOptions specifies criteria for filtering issues. All non-zero fields
// are combined with AND logic.
type FilterOptions struct {
	Status      string
	Label       string
	Assignee    string
	Priority    int
	HasPriority bool // distinguishes "filter by priority 0" from "no filter"
	Type        string
	ParentID    string
	RootsOnly   bool
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
		if opts.Type != "" && issue.Type != opts.Type {
			continue
		}
		if opts.ParentID != "" && issue.ParentID != opts.ParentID {
			continue
		}
		if opts.RootsOnly && issue.ParentID != "" {
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
// "created" (newest first, default), "title" (alphabetically).
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
	case "title":
		sort.Slice(issues, func(i, j int) bool {
			return strings.ToLower(issues[i].Title) < strings.ToLower(issues[j].Title)
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

// Init creates the .work directory structure and writes the default config.
// If config.json already exists, it is loaded and preserved.
func Init(root string) (*Tracker, error) {
	issuesDir := filepath.Join(root, ".work", "issues")
	if err := os.MkdirAll(issuesDir, 0o755); err != nil {
		return nil, fmt.Errorf("creating issues dir: %w", err)
	}

	cfgPath := filepath.Join(root, ".work", "config.json")

	// Preserve existing config if present
	if data, err := os.ReadFile(cfgPath); err == nil {
		var cfg model.Config
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("parsing existing config: %w", err)
		}
		return &Tracker{Root: root, Config: cfg}, nil
	}

	cfg := model.DefaultConfig()
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling config: %w", err)
	}
	if err := os.WriteFile(cfgPath, data, 0o644); err != nil {
		return nil, fmt.Errorf("writing config: %w", err)
	}

	gitattributes := filepath.Join(root, ".work", ".gitattributes")
	if err := os.WriteFile(gitattributes, []byte("* linguist-generated\n"), 0o644); err != nil {
		return nil, fmt.Errorf("writing .gitattributes: %w", err)
	}

	return &Tracker{Root: root, Config: cfg}, nil
}

// Load reads an existing tracker from disk.
func Load(root string) (*Tracker, error) {
	cfgPath := filepath.Join(root, ".work", "config.json")
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
	dir := filepath.Join(t.Root, ".work", "issues", issue.ID)
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
	path := filepath.Join(t.Root, ".work", "issues", id, "issue.json")
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
	path := filepath.Join(t.Root, ".work", "issues", id, "history.jsonl")
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

// ValidateType checks if the given type is allowed by config.
func ValidateType(cfg model.Config, issueType string) error {
	for _, t := range cfg.Types {
		if t == issueType {
			return nil
		}
	}
	return fmt.Errorf("invalid type %q (allowed: %s)", issueType, strings.Join(cfg.Types, ", "))
}

// CreateIssue generates an ID, saves the issue, and records a creation event.
func (t *Tracker) CreateIssue(title, description, assignee string, priority int, labels []string, issueType, parentID, user string) (model.Issue, error) {
	if issueType == "" {
		issueType = t.Config.DefaultType
	}
	if err := ValidateType(t.Config, issueType); err != nil {
		return model.Issue{}, err
	}
	if parentID != "" {
		if err := t.validateParent(parentID); err != nil {
			return model.Issue{}, err
		}
	}
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
		Type:        issueType,
		Priority:    priority,
		Labels:      labels,
		Assignee:    assignee,
		ParentID:    parentID,
		Created:     now,
		Updated:     now,
	}
	if err := t.SaveIssue(issue); err != nil {
		return model.Issue{}, err
	}
	event := model.Event{
		Timestamp: now,
		Op:        "create",
		By:        user,
	}
	if err := t.AppendEvent(id, event); err != nil {
		return model.Issue{}, err
	}
	return issue, nil
}

// ResolvePrefix finds an issue ID matching the given prefix.
// Returns an error if zero or multiple issues match.
func (t *Tracker) ResolvePrefix(prefix string) (string, error) {
	issuesDir := filepath.Join(t.Root, ".work", "issues")
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
func (t *Tracker) SetStatus(id, newStatus, user string) (model.Issue, error) {
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
		By:        user,
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
	path := filepath.Join(t.Root, ".work", "issues", issueID, "history.jsonl")
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
	issuesDir := filepath.Join(t.Root, ".work", "issues")
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
func (t *Tracker) AddComment(id, text, user string) (model.Issue, error) {
	issue, err := t.LoadIssue(id)
	if err != nil {
		return model.Issue{}, err
	}
	now := time.Now().UTC()
	comment := model.Comment{
		Text:    text,
		Created: now,
		By:      user,
	}
	issue.Comments = append(issue.Comments, comment)
	issue.Updated = now
	if err := t.SaveIssue(issue); err != nil {
		return model.Issue{}, err
	}
	event := model.Event{
		Timestamp: now,
		Op:        "comment",
		By:        user,
	}
	if err := t.AppendEvent(id, event); err != nil {
		return model.Issue{}, err
	}
	return issue, nil
}

// validateParent checks that the given parent ID exists and is not itself a child.
func (t *Tracker) validateParent(parentID string) error {
	parent, err := t.LoadIssue(parentID)
	if err != nil {
		return fmt.Errorf("parent issue not found: %s", parentID)
	}
	if parent.ParentID != "" {
		return fmt.Errorf("parent %s is itself a child (no grandchildren allowed)", parentID)
	}
	return nil
}

// LinkIssue sets the parent of a child issue. Validates that the parent exists,
// neither issue creates a grandchild relationship, and no circular ref.
func (t *Tracker) LinkIssue(childID, parentID, user string) (model.Issue, error) {
	if childID == parentID {
		return model.Issue{}, fmt.Errorf("cannot link issue to itself")
	}

	child, err := t.LoadIssue(childID)
	if err != nil {
		return model.Issue{}, err
	}

	if err := t.validateParent(parentID); err != nil {
		return model.Issue{}, err
	}

	// Child must not already be a parent (no grandchildren)
	issues, err := t.ListIssues()
	if err != nil {
		return model.Issue{}, err
	}
	for _, issue := range issues {
		if issue.ParentID == childID {
			return model.Issue{}, fmt.Errorf("issue %s has children and cannot become a child", childID)
		}
	}

	now := time.Now().UTC()
	child.ParentID = parentID
	child.Updated = now
	if err := t.SaveIssue(child); err != nil {
		return model.Issue{}, err
	}

	event := model.Event{
		Timestamp: now,
		Op:        "link",
		To:        parentID,
		By:        user,
	}
	if err := t.AppendEvent(childID, event); err != nil {
		return model.Issue{}, err
	}
	return child, nil
}

// UnlinkIssue removes the parent from a child issue.
func (t *Tracker) UnlinkIssue(childID, user string) (model.Issue, error) {
	child, err := t.LoadIssue(childID)
	if err != nil {
		return model.Issue{}, err
	}
	if child.ParentID == "" {
		return model.Issue{}, fmt.Errorf("issue %s has no parent", childID)
	}

	now := time.Now().UTC()
	oldParent := child.ParentID
	child.ParentID = ""
	child.Updated = now
	if err := t.SaveIssue(child); err != nil {
		return model.Issue{}, err
	}

	event := model.Event{
		Timestamp: now,
		Op:        "unlink",
		From:      oldParent,
		By:        user,
	}
	if err := t.AppendEvent(childID, event); err != nil {
		return model.Issue{}, err
	}
	return child, nil
}

// ListIssues loads all issues from the tracker.
func (t *Tracker) ListIssues() ([]model.Issue, error) {
	issuesDir := filepath.Join(t.Root, ".work", "issues")
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

// LogEntry represents a completed issue in the completion log.
type LogEntry struct {
	ID      string    `json:"id"`
	Title   string    `json:"title"`
	Type    string    `json:"type"`
	Status  string    `json:"status"`
	Labels  []string  `json:"labels,omitempty"`
	Created time.Time `json:"created"`
	Closed  time.Time `json:"closed"`
}

// AppendLog writes a one-line JSON entry to .work/log.jsonl.
func (t *Tracker) AppendLog(issue model.Issue) error {
	entry := LogEntry{
		ID:      issue.ID,
		Title:   issue.Title,
		Type:    issue.Type,
		Status:  issue.Status,
		Labels:  issue.Labels,
		Created: issue.Created,
		Closed:  issue.Updated,
	}
	path := filepath.Join(t.Root, ".work", "log.jsonl")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("opening log: %w", err)
	}
	defer f.Close()
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshaling log entry: %w", err)
	}
	_, err = fmt.Fprintf(f, "%s\n", data)
	return err
}

// LoadLog reads all entries from .work/log.jsonl.
func (t *Tracker) LoadLog() ([]LogEntry, error) {
	path := filepath.Join(t.Root, ".work", "log.jsonl")
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("opening log: %w", err)
	}
	defer f.Close()
	var entries []LogEntry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var entry LogEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			return nil, fmt.Errorf("parsing log entry: %w", err)
		}
		entries = append(entries, entry)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading log: %w", err)
	}
	return entries, nil
}

// CompactIssue strips a completed issue to minimal metadata.
// Logs to .work/log.jsonl, truncates description, clears comments,
// and compacts history to create + close events only.
func (t *Tracker) CompactIssue(id string) error {
	issue, err := t.LoadIssue(id)
	if err != nil {
		return err
	}
	if issue.Status != "done" && issue.Status != "cancelled" {
		return fmt.Errorf("can only compact done/cancelled issues (current: %s)", issue.Status)
	}

	if err := t.AppendLog(issue); err != nil {
		return fmt.Errorf("appending to log: %w", err)
	}

	if desc := issue.Description; desc != "" {
		if idx := strings.IndexByte(desc, '\n'); idx >= 0 {
			issue.Description = desc[:idx]
		}
		if len(issue.Description) > 120 {
			issue.Description = issue.Description[:120]
		}
	}

	issue.Comments = nil
	if err := t.SaveIssue(issue); err != nil {
		return err
	}

	return t.compactHistory(id)
}

func (t *Tracker) compactHistory(id string) error {
	events, err := t.LoadEvents(id)
	if err != nil {
		return err
	}
	if len(events) == 0 {
		return nil
	}

	var compacted []model.Event
	compacted = append(compacted, events[0])
	for i := len(events) - 1; i > 0; i-- {
		if events[i].Op == "status" && (events[i].To == "done" || events[i].To == "cancelled") {
			compacted = append(compacted, events[i])
			break
		}
	}

	path := filepath.Join(t.Root, ".work", "issues", id, "history.jsonl")
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("rewriting history: %w", err)
	}
	defer f.Close()
	for _, ev := range compacted {
		data, err := json.Marshal(ev)
		if err != nil {
			return fmt.Errorf("marshaling event: %w", err)
		}
		fmt.Fprintf(f, "%s\n", data)
	}
	return nil
}

// CompactAllDone compacts all done/cancelled issues.
func (t *Tracker) CompactAllDone() ([]string, error) {
	issues, err := t.ListIssues()
	if err != nil {
		return nil, err
	}
	var compacted []string
	for _, issue := range issues {
		if issue.Status == "done" || issue.Status == "cancelled" {
			if err := t.CompactIssue(issue.ID); err != nil {
				return compacted, fmt.Errorf("compacting %s: %w", issue.ID, err)
			}
			compacted = append(compacted, issue.ID)
		}
	}
	return compacted, nil
}

// GarbageCollect removes issue directories for issues completed
// more than maxAgeDays ago. Logs each issue before deletion.
func (t *Tracker) GarbageCollect(maxAgeDays int) ([]string, error) {
	issues, err := t.ListIssues()
	if err != nil {
		return nil, err
	}
	cutoff := time.Now().UTC().AddDate(0, 0, -maxAgeDays)
	var purged []string
	for _, issue := range issues {
		if (issue.Status == "done" || issue.Status == "cancelled") && issue.Updated.Before(cutoff) {
			if err := t.AppendLog(issue); err != nil {
				return purged, fmt.Errorf("logging %s: %w", issue.ID, err)
			}
			dir := filepath.Join(t.Root, ".work", "issues", issue.ID)
			if err := os.RemoveAll(dir); err != nil {
				return purged, fmt.Errorf("removing %s: %w", issue.ID, err)
			}
			purged = append(purged, issue.ID)
		}
	}
	return purged, nil
}
