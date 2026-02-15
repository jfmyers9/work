package tracker

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jfmyers9/work/internal/model"
)

func TestConfigRoundTrip(t *testing.T) {
	cfg := model.DefaultConfig()
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got model.Config
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.DefaultState != cfg.DefaultState {
		t.Errorf("default_state: got %q, want %q", got.DefaultState, cfg.DefaultState)
	}
	if got.IDLength != cfg.IDLength {
		t.Errorf("id_length: got %d, want %d", got.IDLength, cfg.IDLength)
	}
	for state, targets := range cfg.Transitions {
		gotTargets, ok := got.Transitions[state]
		if !ok {
			t.Errorf("missing transition for state %q", state)
			continue
		}
		if len(gotTargets) != len(targets) {
			t.Errorf("transition %q: got %d targets, want %d", state, len(gotTargets), len(targets))
		}
	}
}

func TestIssueRoundTrip(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	issue := model.Issue{
		ID:          "a3f8b2",
		Title:       "Fix login timeout",
		Status:      "active",
		Priority:    2,
		Labels:      []string{"bug", "urgent"},
		Assignee:    "jim",
		Created:     now,
		Updated:     now,
		Description: "Sessions expire too quickly",
	}

	data, err := json.Marshal(issue)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got model.Issue
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.ID != issue.ID {
		t.Errorf("id: got %q, want %q", got.ID, issue.ID)
	}
	if got.Title != issue.Title {
		t.Errorf("title: got %q, want %q", got.Title, issue.Title)
	}
	if got.Status != issue.Status {
		t.Errorf("status: got %q, want %q", got.Status, issue.Status)
	}
	if got.Priority != issue.Priority {
		t.Errorf("priority: got %d, want %d", got.Priority, issue.Priority)
	}
	if len(got.Labels) != len(issue.Labels) {
		t.Errorf("labels: got %v, want %v", got.Labels, issue.Labels)
	}
	if got.Assignee != issue.Assignee {
		t.Errorf("assignee: got %q, want %q", got.Assignee, issue.Assignee)
	}
	if !got.Created.Equal(issue.Created) {
		t.Errorf("created: got %v, want %v", got.Created, issue.Created)
	}
	if got.Description != issue.Description {
		t.Errorf("description: got %q, want %q", got.Description, issue.Description)
	}
}

func TestGenerateID_Format(t *testing.T) {
	tr := &Tracker{Config: model.Config{IDLength: 6}}
	id, err := tr.GenerateID()
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if len(id) != 6 {
		t.Errorf("length: got %d, want 6", len(id))
	}
	for _, c := range id {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("non-hex char %q in id %q", c, id)
		}
	}
}

func TestGenerateID_Unique(t *testing.T) {
	tr := &Tracker{Config: model.Config{IDLength: 6}}
	seen := make(map[string]bool, 1000)
	for i := 0; i < 1000; i++ {
		id, err := tr.GenerateID()
		if err != nil {
			t.Fatalf("generate %d: %v", i, err)
		}
		if seen[id] {
			t.Fatalf("duplicate id %q at iteration %d", id, i)
		}
		seen[id] = true
	}
}

func TestInit(t *testing.T) {
	root := t.TempDir()

	tr, err := Init(root)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	// .work/ dir exists
	info, err := os.Stat(filepath.Join(root, ".work"))
	if err != nil {
		t.Fatalf("stat .work: %v", err)
	}
	if !info.IsDir() {
		t.Error(".work is not a directory")
	}

	// .work/issues/ dir exists
	info, err = os.Stat(filepath.Join(root, ".work", "issues"))
	if err != nil {
		t.Fatalf("stat .work/issues: %v", err)
	}
	if !info.IsDir() {
		t.Error(".work/issues is not a directory")
	}

	// config.json is valid
	data, err := os.ReadFile(filepath.Join(root, ".work", "config.json"))
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	var cfg model.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("parse config: %v", err)
	}
	if cfg.DefaultState != "open" {
		t.Errorf("default_state: got %q, want %q", cfg.DefaultState, "open")
	}
	if cfg.IDLength != 6 {
		t.Errorf("id_length: got %d, want 6", cfg.IDLength)
	}

	// Tracker config matches
	if tr.Config.DefaultState != cfg.DefaultState {
		t.Error("tracker config doesn't match written config")
	}
}

func TestInit_PreservesExistingConfig(t *testing.T) {
	root := t.TempDir()

	// First init — creates default config
	tr1, err := Init(root)
	if err != nil {
		t.Fatalf("first init: %v", err)
	}

	// Modify the config on disk (change IDLength to distinguish it)
	cfgPath := filepath.Join(root, ".work", "config.json")
	tr1.Config.IDLength = 10
	data, err := json.MarshalIndent(tr1.Config, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(cfgPath, data, 0o644); err != nil {
		t.Fatalf("write modified config: %v", err)
	}

	// Second init — should load existing config, not overwrite
	tr2, err := Init(root)
	if err != nil {
		t.Fatalf("second init: %v", err)
	}
	if tr2.Config.IDLength != 10 {
		t.Errorf("id_length: got %d, want 10 (config was overwritten)", tr2.Config.IDLength)
	}

	// Verify the file on disk is unchanged
	diskData, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	var diskCfg model.Config
	if err := json.Unmarshal(diskData, &diskCfg); err != nil {
		t.Fatalf("parse config: %v", err)
	}
	if diskCfg.IDLength != 10 {
		t.Errorf("disk id_length: got %d, want 10", diskCfg.IDLength)
	}
}

func TestIssueSaveLoad(t *testing.T) {
	root := t.TempDir()
	tr, err := Init(root)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	now := time.Now().UTC().Truncate(time.Second)
	issue := model.Issue{
		ID:       "abc123",
		Title:    "Test issue",
		Status:   "open",
		Priority: 1,
		Labels:   []string{"test"},
		Created:  now,
		Updated:  now,
	}

	if err := tr.SaveIssue(issue); err != nil {
		t.Fatalf("save: %v", err)
	}

	got, err := tr.LoadIssue("abc123")
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if got.ID != issue.ID || got.Title != issue.Title || got.Status != issue.Status {
		t.Errorf("loaded issue doesn't match saved: got %+v", got)
	}
}

func TestCreateIssue(t *testing.T) {
	root := t.TempDir()
	tr, err := Init(root)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	issue, err := tr.CreateIssue("Bug report", "something broke", "jim", 2, []string{"bug", "ux"}, "", "", "testuser")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if len(issue.ID) != 6 {
		t.Errorf("id length: got %d, want 6", len(issue.ID))
	}
	if issue.Title != "Bug report" {
		t.Errorf("title: got %q", issue.Title)
	}
	if issue.Status != "open" {
		t.Errorf("status: got %q, want %q", issue.Status, "open")
	}
	if issue.Priority != 2 {
		t.Errorf("priority: got %d, want 2", issue.Priority)
	}
	if issue.Assignee != "jim" {
		t.Errorf("assignee: got %q", issue.Assignee)
	}
	if issue.Created.IsZero() {
		t.Error("created is zero")
	}

	// Verify issue.json was written
	loaded, err := tr.LoadIssue(issue.ID)
	if err != nil {
		t.Fatalf("load created issue: %v", err)
	}
	if loaded.Title != "Bug report" {
		t.Errorf("loaded title: got %q", loaded.Title)
	}

	// Verify history.jsonl has the create event
	histPath := filepath.Join(root, ".work", "issues", issue.ID, "history.jsonl")
	events := readEvents(t, histPath)
	if len(events) != 1 {
		t.Fatalf("events: got %d, want 1", len(events))
	}
	if events[0].Op != "create" {
		t.Errorf("event op: got %q, want %q", events[0].Op, "create")
	}
	if events[0].By != "testuser" {
		t.Errorf("event by: got %q", events[0].By)
	}
}

func TestResolvePrefix_ExactMatch(t *testing.T) {
	root := t.TempDir()
	tr, err := Init(root)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	// Create issue with known ID
	issue := model.Issue{ID: "a1b2c3", Title: "test", Status: "open", Created: time.Now(), Updated: time.Now()}
	if err := tr.SaveIssue(issue); err != nil {
		t.Fatalf("save: %v", err)
	}

	id, err := tr.ResolvePrefix("a1b2c3")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if id != "a1b2c3" {
		t.Errorf("got %q, want %q", id, "a1b2c3")
	}
}

func TestResolvePrefix_PrefixMatch(t *testing.T) {
	root := t.TempDir()
	tr, err := Init(root)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	issue := model.Issue{ID: "a1b2c3", Title: "test", Status: "open", Created: time.Now(), Updated: time.Now()}
	if err := tr.SaveIssue(issue); err != nil {
		t.Fatalf("save: %v", err)
	}

	id, err := tr.ResolvePrefix("a1b")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if id != "a1b2c3" {
		t.Errorf("got %q, want %q", id, "a1b2c3")
	}
}

func TestResolvePrefix_NoMatch(t *testing.T) {
	root := t.TempDir()
	tr, err := Init(root)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	_, err = tr.ResolvePrefix("zzz")
	if err == nil {
		t.Fatal("expected error for no match")
	}
	if !strings.Contains(err.Error(), "no issue found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestResolvePrefix_Ambiguous(t *testing.T) {
	root := t.TempDir()
	tr, err := Init(root)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	now := time.Now()
	for _, id := range []string{"aa1111", "aa2222"} {
		issue := model.Issue{ID: id, Title: "test", Status: "open", Created: now, Updated: now}
		if err := tr.SaveIssue(issue); err != nil {
			t.Fatalf("save %s: %v", id, err)
		}
	}

	_, err = tr.ResolvePrefix("aa")
	if err == nil {
		t.Fatal("expected error for ambiguous prefix")
	}
	if !strings.Contains(err.Error(), "ambiguous") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestListIssues_Empty(t *testing.T) {
	root := t.TempDir()
	tr, err := Init(root)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	issues, err := tr.ListIssues()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("expected empty list, got %d", len(issues))
	}
}

func TestListIssues_Multiple(t *testing.T) {
	root := t.TempDir()
	tr, err := Init(root)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	now := time.Now()
	for _, id := range []string{"aaa111", "bbb222", "ccc333"} {
		issue := model.Issue{ID: id, Title: "Issue " + id, Status: "open", Created: now, Updated: now}
		if err := tr.SaveIssue(issue); err != nil {
			t.Fatalf("save: %v", err)
		}
	}

	issues, err := tr.ListIssues()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(issues) != 3 {
		t.Errorf("count: got %d, want 3", len(issues))
	}
}

func TestEditIssue(t *testing.T) {
	root := t.TempDir()
	tr, err := Init(root)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	issue, err := tr.CreateIssue("Original title", "desc", "", 0, nil, "", "", "testuser")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	origUpdated := issue.Updated

	// Simulate edit: load, modify, save, append event
	time.Sleep(10 * time.Millisecond) // ensure timestamp differs
	loaded, err := tr.LoadIssue(issue.ID)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	loaded.Title = "New title"
	loaded.Priority = 3
	loaded.Updated = time.Now().UTC()
	if err := tr.SaveIssue(loaded); err != nil {
		t.Fatalf("save: %v", err)
	}
	event := model.Event{
		Timestamp: loaded.Updated,
		Op:        "edit",
		Fields:    []string{"title", "priority"},
		By:        "system",
	}
	if err := tr.AppendEvent(issue.ID, event); err != nil {
		t.Fatalf("append event: %v", err)
	}

	// Verify changes persisted
	final, err := tr.LoadIssue(issue.ID)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if final.Title != "New title" {
		t.Errorf("title: got %q", final.Title)
	}
	if final.Priority != 3 {
		t.Errorf("priority: got %d", final.Priority)
	}
	if final.Description != "desc" {
		t.Errorf("description should be preserved: got %q", final.Description)
	}
	if !final.Updated.After(origUpdated) {
		t.Error("updated timestamp should have advanced")
	}

	// Verify history has both create and edit events
	histPath := filepath.Join(root, ".work", "issues", issue.ID, "history.jsonl")
	events := readEvents(t, histPath)
	if len(events) != 2 {
		t.Fatalf("events: got %d, want 2", len(events))
	}
	if events[0].Op != "create" {
		t.Errorf("first event: got %q, want create", events[0].Op)
	}
	if events[1].Op != "edit" {
		t.Errorf("second event: got %q, want edit", events[1].Op)
	}
	if len(events[1].Fields) != 2 {
		t.Errorf("edit fields count: got %d, want 2", len(events[1].Fields))
	}
}

func TestAppendEvent_MultipleEvents(t *testing.T) {
	root := t.TempDir()
	tr, err := Init(root)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	issue := model.Issue{ID: "evt123", Title: "test", Status: "open", Created: time.Now(), Updated: time.Now()}
	if err := tr.SaveIssue(issue); err != nil {
		t.Fatalf("save: %v", err)
	}

	for i := 0; i < 3; i++ {
		ev := model.Event{Timestamp: time.Now().UTC(), Op: "test", By: "system"}
		if err := tr.AppendEvent("evt123", ev); err != nil {
			t.Fatalf("append %d: %v", i, err)
		}
	}

	histPath := filepath.Join(root, ".work", "issues", "evt123", "history.jsonl")
	events := readEvents(t, histPath)
	if len(events) != 3 {
		t.Errorf("events: got %d, want 3", len(events))
	}
}

func TestValidateTransition_Valid(t *testing.T) {
	cfg := model.DefaultConfig()
	valid := [][2]string{
		{"open", "active"},
		{"open", "done"},
		{"open", "cancelled"},
		{"active", "done"},
		{"active", "cancelled"},
		{"active", "open"},
		{"active", "review"},
		{"review", "done"},
		{"review", "active"},
		{"done", "open"},
		{"cancelled", "open"},
	}
	for _, pair := range valid {
		if err := ValidateTransition(cfg, pair[0], pair[1]); err != nil {
			t.Errorf("%s→%s: unexpected error: %v", pair[0], pair[1], err)
		}
	}
}

func TestValidateTransition_Invalid(t *testing.T) {
	cfg := model.DefaultConfig()
	invalid := [][2]string{
		{"done", "active"},
		{"done", "cancelled"},
		{"cancelled", "active"},
		{"cancelled", "done"},
		{"open", "open"},
		{"active", "active"},
	}
	for _, pair := range invalid {
		err := ValidateTransition(cfg, pair[0], pair[1])
		if err == nil {
			t.Errorf("%s→%s: expected error, got nil", pair[0], pair[1])
		}
		if !strings.Contains(err.Error(), "invalid transition") {
			t.Errorf("%s→%s: unexpected error text: %v", pair[0], pair[1], err)
		}
	}
}

func TestValidateTransition_UnknownState(t *testing.T) {
	cfg := model.DefaultConfig()
	err := ValidateTransition(cfg, "nonexistent", "open")
	if err == nil {
		t.Fatal("expected error for unknown state")
	}
	if !strings.Contains(err.Error(), "unknown state") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSetStatus(t *testing.T) {
	root := t.TempDir()
	tr, err := Init(root)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	issue, err := tr.CreateIssue("Status test", "", "", 0, nil, "", "", "testuser")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// open → active
	updated, err := tr.SetStatus(issue.ID, "active", "testuser")
	if err != nil {
		t.Fatalf("set status open→active: %v", err)
	}
	if updated.Status != "active" {
		t.Errorf("status: got %q, want active", updated.Status)
	}

	// active → done
	updated, err = tr.SetStatus(issue.ID, "done", "testuser")
	if err != nil {
		t.Fatalf("set status active→done: %v", err)
	}
	if updated.Status != "done" {
		t.Errorf("status: got %q, want done", updated.Status)
	}

	// done → open (reopen)
	updated, err = tr.SetStatus(issue.ID, "open", "testuser")
	if err != nil {
		t.Fatalf("set status done→open: %v", err)
	}
	if updated.Status != "open" {
		t.Errorf("status: got %q, want open", updated.Status)
	}
}

func TestSetStatus_InvalidTransition(t *testing.T) {
	root := t.TempDir()
	tr, err := Init(root)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	issue, err := tr.CreateIssue("Invalid test", "", "", 0, nil, "", "", "testuser")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// open → active → done, then try done → active (invalid)
	if _, err := tr.SetStatus(issue.ID, "active", "testuser"); err != nil {
		t.Fatalf("open→active: %v", err)
	}
	if _, err := tr.SetStatus(issue.ID, "done", "testuser"); err != nil {
		t.Fatalf("active→done: %v", err)
	}
	_, err = tr.SetStatus(issue.ID, "active", "testuser")
	if err == nil {
		t.Fatal("expected error for done→active")
	}
	if !strings.Contains(err.Error(), "invalid transition") {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify status didn't change
	loaded, err := tr.LoadIssue(issue.ID)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if loaded.Status != "done" {
		t.Errorf("status should still be done, got %q", loaded.Status)
	}
}

func TestSetStatus_SameState(t *testing.T) {
	root := t.TempDir()
	tr, err := Init(root)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	issue, err := tr.CreateIssue("Same state", "", "", 0, nil, "", "", "testuser")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	_, err = tr.SetStatus(issue.ID, "open", "testuser")
	if err == nil {
		t.Fatal("expected error for open→open")
	}
	if !strings.Contains(err.Error(), "already in state") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSetStatus_Persists(t *testing.T) {
	root := t.TempDir()
	tr, err := Init(root)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	issue, err := tr.CreateIssue("Persist test", "", "", 0, nil, "", "", "testuser")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if _, err := tr.SetStatus(issue.ID, "active", "testuser"); err != nil {
		t.Fatalf("set status: %v", err)
	}

	// Reload from disk
	tr2, err := Load(root)
	if err != nil {
		t.Fatalf("reload tracker: %v", err)
	}
	loaded, err := tr2.LoadIssue(issue.ID)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if loaded.Status != "active" {
		t.Errorf("persisted status: got %q, want active", loaded.Status)
	}
}

func TestSetStatus_HistoryEvent(t *testing.T) {
	root := t.TempDir()
	tr, err := Init(root)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	issue, err := tr.CreateIssue("History test", "", "", 0, nil, "", "", "testuser")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if _, err := tr.SetStatus(issue.ID, "active", "testuser"); err != nil {
		t.Fatalf("set status: %v", err)
	}

	histPath := filepath.Join(root, ".work", "issues", issue.ID, "history.jsonl")
	events := readEvents(t, histPath)
	if len(events) != 2 {
		t.Fatalf("events: got %d, want 2", len(events))
	}
	ev := events[1]
	if ev.Op != "status" {
		t.Errorf("op: got %q, want status", ev.Op)
	}
	if ev.From != "open" {
		t.Errorf("from: got %q, want open", ev.From)
	}
	if ev.To != "active" {
		t.Errorf("to: got %q, want active", ev.To)
	}
	if ev.By != "testuser" {
		t.Errorf("by: got %q, want testuser", ev.By)
	}
}

func TestValidateTransition_CustomConfig(t *testing.T) {
	cfg := model.Config{
		Transitions: map[string][]string{
			"todo":   {"doing"},
			"doing":  {"review", "todo"},
			"review": {"done", "doing"},
			"done":   {"todo"},
		},
	}

	// Valid custom transitions
	if err := ValidateTransition(cfg, "todo", "doing"); err != nil {
		t.Errorf("todo→doing: %v", err)
	}
	if err := ValidateTransition(cfg, "review", "done"); err != nil {
		t.Errorf("review→done: %v", err)
	}

	// Invalid in custom config
	err := ValidateTransition(cfg, "todo", "review")
	if err == nil {
		t.Error("expected error for todo→review")
	}
	err = ValidateTransition(cfg, "todo", "done")
	if err == nil {
		t.Error("expected error for todo→done")
	}
}

// --- Filter and Sort tests ---

func makeTestIssues() []model.Issue {
	now := time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)
	return []model.Issue{
		{ID: "aaa111", Title: "Open bug", Status: "open", Priority: 1, Labels: []string{"bug"}, Assignee: "jim", Created: now, Updated: now.Add(3 * time.Hour)},
		{ID: "bbb222", Title: "Active feature", Status: "active", Priority: 2, Labels: []string{"feature"}, Assignee: "alice", Created: now.Add(time.Hour), Updated: now.Add(time.Hour)},
		{ID: "ccc333", Title: "Open feature", Status: "open", Priority: 0, Labels: []string{"feature", "urgent"}, Assignee: "jim", Created: now.Add(2 * time.Hour), Updated: now.Add(2 * time.Hour)},
		{ID: "ddd444", Title: "Done bug", Status: "done", Priority: 3, Labels: []string{"bug"}, Assignee: "alice", Created: now.Add(3 * time.Hour), Updated: now.Add(4 * time.Hour)},
	}
}

func TestFilterIssues_ByStatus(t *testing.T) {
	issues := makeTestIssues()
	got := FilterIssues(issues, FilterOptions{Status: "open"})
	if len(got) != 2 {
		t.Fatalf("count: got %d, want 2", len(got))
	}
	for _, i := range got {
		if i.Status != "open" {
			t.Errorf("unexpected status %q", i.Status)
		}
	}
}

func TestFilterIssues_ByLabel(t *testing.T) {
	issues := makeTestIssues()
	got := FilterIssues(issues, FilterOptions{Label: "bug"})
	if len(got) != 2 {
		t.Fatalf("count: got %d, want 2", len(got))
	}
	for _, i := range got {
		if !hasLabel(i.Labels, "bug") {
			t.Errorf("issue %s missing label bug", i.ID)
		}
	}
}

func TestFilterIssues_ByAssignee(t *testing.T) {
	issues := makeTestIssues()
	got := FilterIssues(issues, FilterOptions{Assignee: "alice"})
	if len(got) != 2 {
		t.Fatalf("count: got %d, want 2", len(got))
	}
	for _, i := range got {
		if i.Assignee != "alice" {
			t.Errorf("unexpected assignee %q", i.Assignee)
		}
	}
}

func TestFilterIssues_ByPriority(t *testing.T) {
	issues := makeTestIssues()
	got := FilterIssues(issues, FilterOptions{Priority: 0, HasPriority: true})
	if len(got) != 1 {
		t.Fatalf("count: got %d, want 1", len(got))
	}
	if got[0].ID != "ccc333" {
		t.Errorf("expected ccc333, got %s", got[0].ID)
	}
}

func TestFilterIssues_Combined(t *testing.T) {
	issues := makeTestIssues()
	got := FilterIssues(issues, FilterOptions{Status: "open", Label: "bug"})
	if len(got) != 1 {
		t.Fatalf("count: got %d, want 1", len(got))
	}
	if got[0].ID != "aaa111" {
		t.Errorf("expected aaa111, got %s", got[0].ID)
	}
}

func TestFilterIssues_NoMatch(t *testing.T) {
	issues := makeTestIssues()
	got := FilterIssues(issues, FilterOptions{Status: "cancelled"})
	if len(got) != 0 {
		t.Errorf("expected empty, got %d", len(got))
	}
}

func TestFilterIssues_NoFilter(t *testing.T) {
	issues := makeTestIssues()
	got := FilterIssues(issues, FilterOptions{})
	if len(got) != len(issues) {
		t.Errorf("count: got %d, want %d", len(got), len(issues))
	}
}

func TestSortIssues_ByPriority(t *testing.T) {
	issues := makeTestIssues()
	SortIssues(issues, "priority")
	for i := 1; i < len(issues); i++ {
		if issues[i].Priority < issues[i-1].Priority {
			t.Errorf("not sorted by priority: %d before %d", issues[i-1].Priority, issues[i].Priority)
		}
	}
	if issues[0].Priority != 0 {
		t.Errorf("first should be priority 0, got %d", issues[0].Priority)
	}
}

func TestSortIssues_ByUpdated(t *testing.T) {
	issues := makeTestIssues()
	SortIssues(issues, "updated")
	for i := 1; i < len(issues); i++ {
		if issues[i].Updated.After(issues[i-1].Updated) {
			t.Errorf("not sorted by updated desc: %v after %v", issues[i-1].Updated, issues[i].Updated)
		}
	}
}

func TestSortIssues_ByCreated(t *testing.T) {
	issues := makeTestIssues()
	SortIssues(issues, "created")
	for i := 1; i < len(issues); i++ {
		if issues[i].Created.After(issues[i-1].Created) {
			t.Errorf("not sorted by created desc: %v after %v", issues[i-1].Created, issues[i].Created)
		}
	}
}

func TestSortIssues_Default(t *testing.T) {
	issues := makeTestIssues()
	SortIssues(issues, "")
	// Default is created descending — same as "created"
	for i := 1; i < len(issues); i++ {
		if issues[i].Created.After(issues[i-1].Created) {
			t.Errorf("default sort not by created desc")
		}
	}
}

// --- LoadEvents / LoadAllEvents / FilterEventsByTime tests ---

func TestLoadEvents_SingleIssue(t *testing.T) {
	root := t.TempDir()
	tr, err := Init(root)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	issue, err := tr.CreateIssue("Test", "", "", 0, nil, "", "", "testuser") // creates 1 event
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := tr.SetStatus(issue.ID, "active", "testuser"); err != nil { // creates 2nd event
		t.Fatalf("set status: %v", err)
	}

	events, err := tr.LoadEvents(issue.ID)
	if err != nil {
		t.Fatalf("load events: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("count: got %d, want 2", len(events))
	}
	if events[0].Op != "create" {
		t.Errorf("first op: got %q, want create", events[0].Op)
	}
	if events[1].Op != "status" || events[1].From != "open" || events[1].To != "active" {
		t.Errorf("second event: got op=%q from=%q to=%q", events[1].Op, events[1].From, events[1].To)
	}
}

func TestLoadEvents_NoHistory(t *testing.T) {
	root := t.TempDir()
	tr, err := Init(root)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	// Create issue dir without history.jsonl
	issue := model.Issue{ID: "nohist", Title: "test", Status: "open", Created: time.Now(), Updated: time.Now()}
	if err := tr.SaveIssue(issue); err != nil {
		t.Fatalf("save: %v", err)
	}

	events, err := tr.LoadEvents("nohist")
	if err != nil {
		t.Fatalf("load events: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected empty, got %d", len(events))
	}
}

func TestLoadAllEvents_MultipleIssues(t *testing.T) {
	root := t.TempDir()
	tr, err := Init(root)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	i1, err := tr.CreateIssue("First", "", "", 0, nil, "", "", "testuser")
	if err != nil {
		t.Fatalf("create 1: %v", err)
	}
	i2, err := tr.CreateIssue("Second", "", "", 0, nil, "", "", "testuser")
	if err != nil {
		t.Fatalf("create 2: %v", err)
	}
	if _, err := tr.SetStatus(i1.ID, "active", "testuser"); err != nil {
		t.Fatalf("set status: %v", err)
	}

	all, err := tr.LoadAllEvents()
	if err != nil {
		t.Fatalf("load all: %v", err)
	}
	// i1: create + status = 2, i2: create = 1 → total 3
	if len(all) != 3 {
		t.Fatalf("count: got %d, want 3", len(all))
	}

	// Verify issue IDs are annotated
	ids := make(map[string]int)
	for _, ev := range all {
		ids[ev.IssueID]++
	}
	if ids[i1.ID] != 2 {
		t.Errorf("issue1 events: got %d, want 2", ids[i1.ID])
	}
	if ids[i2.ID] != 1 {
		t.Errorf("issue2 events: got %d, want 1", ids[i2.ID])
	}
}

func TestFilterEventsByTime_Since(t *testing.T) {
	base := time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC)
	events := []model.Event{
		{Timestamp: base, Op: "create"},
		{Timestamp: base.Add(24 * time.Hour), Op: "status"},
		{Timestamp: base.Add(48 * time.Hour), Op: "edit"},
	}

	since := base.Add(24 * time.Hour)
	got := FilterEventsByTime(events, since, time.Time{})
	if len(got) != 2 {
		t.Fatalf("count: got %d, want 2", len(got))
	}
	if got[0].Op != "status" {
		t.Errorf("first: got %q, want status", got[0].Op)
	}
}

func TestFilterEventsByTime_Until(t *testing.T) {
	base := time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC)
	events := []model.Event{
		{Timestamp: base, Op: "create"},
		{Timestamp: base.Add(24 * time.Hour), Op: "status"},
		{Timestamp: base.Add(48 * time.Hour), Op: "edit"},
	}

	until := base.Add(24 * time.Hour)
	got := FilterEventsByTime(events, time.Time{}, until)
	if len(got) != 1 {
		t.Fatalf("count: got %d, want 1", len(got))
	}
	if got[0].Op != "create" {
		t.Errorf("first: got %q, want create", got[0].Op)
	}
}

func TestFilterEventsByTime_SinceAndUntil(t *testing.T) {
	base := time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC)
	events := []model.Event{
		{Timestamp: base, Op: "create"},
		{Timestamp: base.Add(24 * time.Hour), Op: "status"},
		{Timestamp: base.Add(48 * time.Hour), Op: "edit"},
		{Timestamp: base.Add(72 * time.Hour), Op: "status"},
	}

	since := base.Add(12 * time.Hour)
	until := base.Add(60 * time.Hour)
	got := FilterEventsByTime(events, since, until)
	if len(got) != 2 {
		t.Fatalf("count: got %d, want 2", len(got))
	}
	if got[0].Op != "status" || got[1].Op != "edit" {
		t.Errorf("got ops %q %q, want status edit", got[0].Op, got[1].Op)
	}
}

func TestResolveUser_EnvVar(t *testing.T) {
	t.Setenv("WORK_USER", "envuser")
	got := ResolveUser()
	if got != "envuser" {
		t.Errorf("got %q, want envuser", got)
	}
}

func TestResolveUser_Fallback(t *testing.T) {
	t.Setenv("WORK_USER", "")
	got := ResolveUser()
	// Falls back to git config user.name or "system" — either is acceptable
	if got == "" {
		t.Error("ResolveUser returned empty string")
	}
}

func TestReviewWorkflow(t *testing.T) {
	root := t.TempDir()
	tr, err := Init(root)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	issue, err := tr.CreateIssue("Review test", "", "", 0, nil, "", "", "alice")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// open → active
	issue, err = tr.SetStatus(issue.ID, "active", "alice")
	if err != nil {
		t.Fatalf("open→active: %v", err)
	}

	// active → review
	issue, err = tr.SetStatus(issue.ID, "review", "alice")
	if err != nil {
		t.Fatalf("active→review: %v", err)
	}
	if issue.Status != "review" {
		t.Errorf("status: got %q, want review", issue.Status)
	}

	// review → done (approve)
	issue, err = tr.SetStatus(issue.ID, "done", "bob")
	if err != nil {
		t.Fatalf("review→done: %v", err)
	}
	if issue.Status != "done" {
		t.Errorf("status: got %q, want done", issue.Status)
	}

	// Verify events record correct users
	events, err := tr.LoadEvents(issue.ID)
	if err != nil {
		t.Fatalf("load events: %v", err)
	}
	if events[0].By != "alice" {
		t.Errorf("create by: got %q, want alice", events[0].By)
	}
	if events[3].By != "bob" {
		t.Errorf("approve by: got %q, want bob", events[3].By)
	}
}

func TestRejectWorkflow(t *testing.T) {
	root := t.TempDir()
	tr, err := Init(root)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	issue, err := tr.CreateIssue("Reject test", "", "", 0, nil, "", "", "alice")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if _, err = tr.SetStatus(issue.ID, "active", "alice"); err != nil {
		t.Fatalf("open→active: %v", err)
	}
	if _, err = tr.SetStatus(issue.ID, "review", "alice"); err != nil {
		t.Fatalf("active→review: %v", err)
	}

	// review → active (reject)
	issue, err = tr.SetStatus(issue.ID, "active", "bob")
	if err != nil {
		t.Fatalf("review→active: %v", err)
	}
	if issue.Status != "active" {
		t.Errorf("status: got %q, want active", issue.Status)
	}

	// Add rejection comment
	issue, err = tr.AddComment(issue.ID, "Rejected: needs error handling", "bob")
	if err != nil {
		t.Fatalf("rejection comment: %v", err)
	}
	if len(issue.Comments) != 1 {
		t.Fatalf("comments: got %d, want 1", len(issue.Comments))
	}
	if issue.Comments[0].By != "bob" {
		t.Errorf("comment by: got %q, want bob", issue.Comments[0].By)
	}
	if issue.Comments[0].Text != "Rejected: needs error handling" {
		t.Errorf("comment text: got %q", issue.Comments[0].Text)
	}
}

func TestReviewTransition_InvalidFromOpen(t *testing.T) {
	cfg := model.DefaultConfig()
	err := ValidateTransition(cfg, "open", "review")
	if err == nil {
		t.Error("expected error for open→review")
	}
}

func TestReviewTransition_InvalidFromDone(t *testing.T) {
	cfg := model.DefaultConfig()
	err := ValidateTransition(cfg, "done", "review")
	if err == nil {
		t.Error("expected error for done→review")
	}
}

func TestUserIdentityOnComment(t *testing.T) {
	root := t.TempDir()
	tr, err := Init(root)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	issue, err := tr.CreateIssue("Identity test", "", "", 0, nil, "", "", "creator")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	issue, err = tr.AddComment(issue.ID, "hello", "commenter")
	if err != nil {
		t.Fatalf("comment: %v", err)
	}
	if issue.Comments[0].By != "commenter" {
		t.Errorf("comment by: got %q, want commenter", issue.Comments[0].By)
	}

	events, err := tr.LoadEvents(issue.ID)
	if err != nil {
		t.Fatalf("load events: %v", err)
	}
	if events[0].By != "creator" {
		t.Errorf("create event by: got %q, want creator", events[0].By)
	}
	if events[1].By != "commenter" {
		t.Errorf("comment event by: got %q, want commenter", events[1].By)
	}
}

// --- Type tests ---

func TestCreateIssue_DefaultType(t *testing.T) {
	root := t.TempDir()
	tr, err := Init(root)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	issue, err := tr.CreateIssue("No type specified", "", "", 0, nil, "", "", "testuser")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if issue.Type != "feature" {
		t.Errorf("type: got %q, want feature", issue.Type)
	}
}

func TestCreateIssue_ExplicitType(t *testing.T) {
	root := t.TempDir()
	tr, err := Init(root)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	issue, err := tr.CreateIssue("A bug", "", "", 0, nil, "bug", "", "testuser")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if issue.Type != "bug" {
		t.Errorf("type: got %q, want bug", issue.Type)
	}

	loaded, err := tr.LoadIssue(issue.ID)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if loaded.Type != "bug" {
		t.Errorf("persisted type: got %q, want bug", loaded.Type)
	}
}

func TestCreateIssue_InvalidType(t *testing.T) {
	root := t.TempDir()
	tr, err := Init(root)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	_, err = tr.CreateIssue("Bad type", "", "", 0, nil, "epic", "", "testuser")
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
	if !strings.Contains(err.Error(), "invalid type") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateType(t *testing.T) {
	cfg := model.DefaultConfig()

	for _, valid := range []string{"feature", "bug", "chore"} {
		if err := ValidateType(cfg, valid); err != nil {
			t.Errorf("type %q should be valid: %v", valid, err)
		}
	}

	err := ValidateType(cfg, "epic")
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
	if !strings.Contains(err.Error(), "invalid type") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestFilterIssues_ByType(t *testing.T) {
	now := time.Now()
	issues := []model.Issue{
		{ID: "aaa111", Title: "Feature A", Status: "open", Type: "feature", Created: now, Updated: now},
		{ID: "bbb222", Title: "Bug B", Status: "open", Type: "bug", Created: now, Updated: now},
		{ID: "ccc333", Title: "Chore C", Status: "open", Type: "chore", Created: now, Updated: now},
		{ID: "ddd444", Title: "Feature D", Status: "active", Type: "feature", Created: now, Updated: now},
	}

	got := FilterIssues(issues, FilterOptions{Type: "bug"})
	if len(got) != 1 {
		t.Fatalf("count: got %d, want 1", len(got))
	}
	if got[0].ID != "bbb222" {
		t.Errorf("expected bbb222, got %s", got[0].ID)
	}

	got = FilterIssues(issues, FilterOptions{Type: "feature"})
	if len(got) != 2 {
		t.Fatalf("count: got %d, want 2", len(got))
	}

	got = FilterIssues(issues, FilterOptions{Type: "feature", Status: "active"})
	if len(got) != 1 {
		t.Fatalf("count: got %d, want 1", len(got))
	}
	if got[0].ID != "ddd444" {
		t.Errorf("expected ddd444, got %s", got[0].ID)
	}
}

func TestIssueRoundTrip_WithType(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	issue := model.Issue{
		ID:      "abc123",
		Title:   "Typed issue",
		Status:  "open",
		Type:    "bug",
		Created: now,
		Updated: now,
	}

	data, err := json.Marshal(issue)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got model.Issue
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Type != "bug" {
		t.Errorf("type: got %q, want bug", got.Type)
	}
}

func TestConfigRoundTrip_WithTypes(t *testing.T) {
	cfg := model.DefaultConfig()
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got model.Config
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.DefaultType != "feature" {
		t.Errorf("default_type: got %q, want feature", got.DefaultType)
	}
	if len(got.Types) != 3 {
		t.Errorf("types count: got %d, want 3", len(got.Types))
	}
}

// --- Parent-child (epic) tests ---

func TestLinkIssue(t *testing.T) {
	root := t.TempDir()
	tr, err := Init(root)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	parent, err := tr.CreateIssue("Epic", "", "", 0, nil, "", "", "testuser")
	if err != nil {
		t.Fatalf("create parent: %v", err)
	}
	child, err := tr.CreateIssue("Task", "", "", 0, nil, "", "", "testuser")
	if err != nil {
		t.Fatalf("create child: %v", err)
	}

	linked, err := tr.LinkIssue(child.ID, parent.ID, "testuser")
	if err != nil {
		t.Fatalf("link: %v", err)
	}
	if linked.ParentID != parent.ID {
		t.Errorf("parent_id: got %q, want %q", linked.ParentID, parent.ID)
	}

	// Verify persisted
	loaded, err := tr.LoadIssue(child.ID)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if loaded.ParentID != parent.ID {
		t.Errorf("persisted parent_id: got %q, want %q", loaded.ParentID, parent.ID)
	}

	// Verify link event
	events, err := tr.LoadEvents(child.ID)
	if err != nil {
		t.Fatalf("load events: %v", err)
	}
	last := events[len(events)-1]
	if last.Op != "link" || last.To != parent.ID {
		t.Errorf("link event: op=%q to=%q", last.Op, last.To)
	}
}

func TestUnlinkIssue(t *testing.T) {
	root := t.TempDir()
	tr, err := Init(root)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	parent, err := tr.CreateIssue("Epic", "", "", 0, nil, "", "", "testuser")
	if err != nil {
		t.Fatalf("create parent: %v", err)
	}
	child, err := tr.CreateIssue("Task", "", "", 0, nil, "", parent.ID, "testuser")
	if err != nil {
		t.Fatalf("create child: %v", err)
	}

	unlinked, err := tr.UnlinkIssue(child.ID, "testuser")
	if err != nil {
		t.Fatalf("unlink: %v", err)
	}
	if unlinked.ParentID != "" {
		t.Errorf("parent_id should be empty, got %q", unlinked.ParentID)
	}

	// Verify unlink event
	events, err := tr.LoadEvents(child.ID)
	if err != nil {
		t.Fatalf("load events: %v", err)
	}
	last := events[len(events)-1]
	if last.Op != "unlink" || last.From != parent.ID {
		t.Errorf("unlink event: op=%q from=%q", last.Op, last.From)
	}
}

func TestUnlinkIssue_NoParent(t *testing.T) {
	root := t.TempDir()
	tr, err := Init(root)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	issue, err := tr.CreateIssue("Orphan", "", "", 0, nil, "", "", "testuser")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	_, err = tr.UnlinkIssue(issue.ID, "testuser")
	if err == nil {
		t.Fatal("expected error unlinking issue with no parent")
	}
	if !strings.Contains(err.Error(), "has no parent") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLinkIssue_SelfLink(t *testing.T) {
	root := t.TempDir()
	tr, err := Init(root)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	issue, err := tr.CreateIssue("Self", "", "", 0, nil, "", "", "testuser")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	_, err = tr.LinkIssue(issue.ID, issue.ID, "testuser")
	if err == nil {
		t.Fatal("expected error for self-link")
	}
	if !strings.Contains(err.Error(), "cannot link issue to itself") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLinkIssue_NoGrandchildren(t *testing.T) {
	root := t.TempDir()
	tr, err := Init(root)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	grandparent, err := tr.CreateIssue("Grandparent", "", "", 0, nil, "", "", "testuser")
	if err != nil {
		t.Fatalf("create grandparent: %v", err)
	}
	parent, err := tr.CreateIssue("Parent", "", "", 0, nil, "", "", "testuser")
	if err != nil {
		t.Fatalf("create parent: %v", err)
	}
	child, err := tr.CreateIssue("Child", "", "", 0, nil, "", "", "testuser")
	if err != nil {
		t.Fatalf("create child: %v", err)
	}

	// Link parent under grandparent
	if _, err := tr.LinkIssue(parent.ID, grandparent.ID, "testuser"); err != nil {
		t.Fatalf("link parent: %v", err)
	}

	// Try to make parent a parent of child — should fail because parent has a parent
	_, err = tr.LinkIssue(child.ID, parent.ID, "testuser")
	if err == nil {
		t.Fatal("expected error for grandchild")
	}
	if !strings.Contains(err.Error(), "is itself a child") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLinkIssue_ChildWithChildrenCannotBecomeChild(t *testing.T) {
	root := t.TempDir()
	tr, err := Init(root)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	epic, err := tr.CreateIssue("Epic", "", "", 0, nil, "", "", "testuser")
	if err != nil {
		t.Fatalf("create epic: %v", err)
	}
	mid, err := tr.CreateIssue("Mid", "", "", 0, nil, "", "", "testuser")
	if err != nil {
		t.Fatalf("create mid: %v", err)
	}
	leaf, err := tr.CreateIssue("Leaf", "", "", 0, nil, "", "", "testuser")
	if err != nil {
		t.Fatalf("create leaf: %v", err)
	}

	// Link leaf under mid
	if _, err := tr.LinkIssue(leaf.ID, mid.ID, "testuser"); err != nil {
		t.Fatalf("link leaf: %v", err)
	}

	// Try to link mid under epic — should fail because mid already has children
	_, err = tr.LinkIssue(mid.ID, epic.ID, "testuser")
	if err == nil {
		t.Fatal("expected error: mid has children")
	}
	if !strings.Contains(err.Error(), "has children") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLinkIssue_ParentNotFound(t *testing.T) {
	root := t.TempDir()
	tr, err := Init(root)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	child, err := tr.CreateIssue("Child", "", "", 0, nil, "", "", "testuser")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	_, err = tr.LinkIssue(child.ID, "nonexistent", "testuser")
	if err == nil {
		t.Fatal("expected error for missing parent")
	}
	if !strings.Contains(err.Error(), "parent issue not found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCreateIssue_WithParent(t *testing.T) {
	root := t.TempDir()
	tr, err := Init(root)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	parent, err := tr.CreateIssue("Epic", "", "", 0, nil, "", "", "testuser")
	if err != nil {
		t.Fatalf("create parent: %v", err)
	}

	child, err := tr.CreateIssue("Task", "", "", 0, nil, "", parent.ID, "testuser")
	if err != nil {
		t.Fatalf("create child: %v", err)
	}
	if child.ParentID != parent.ID {
		t.Errorf("parent_id: got %q, want %q", child.ParentID, parent.ID)
	}

	// Verify persisted
	loaded, err := tr.LoadIssue(child.ID)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if loaded.ParentID != parent.ID {
		t.Errorf("persisted parent_id: got %q, want %q", loaded.ParentID, parent.ID)
	}
}

func TestCreateIssue_WithInvalidParent(t *testing.T) {
	root := t.TempDir()
	tr, err := Init(root)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	_, err = tr.CreateIssue("Task", "", "", 0, nil, "", "nonexistent", "testuser")
	if err == nil {
		t.Fatal("expected error for missing parent")
	}
}

func TestFilterIssues_ByParent(t *testing.T) {
	now := time.Now()
	issues := []model.Issue{
		{ID: "aaa111", Title: "Epic", Status: "open", Created: now, Updated: now},
		{ID: "bbb222", Title: "Task 1", Status: "open", ParentID: "aaa111", Created: now, Updated: now},
		{ID: "ccc333", Title: "Task 2", Status: "done", ParentID: "aaa111", Created: now, Updated: now},
		{ID: "ddd444", Title: "Standalone", Status: "open", Created: now, Updated: now},
	}

	got := FilterIssues(issues, FilterOptions{ParentID: "aaa111"})
	if len(got) != 2 {
		t.Fatalf("count: got %d, want 2", len(got))
	}
	for _, i := range got {
		if i.ParentID != "aaa111" {
			t.Errorf("issue %s has wrong parent %q", i.ID, i.ParentID)
		}
	}
}

func TestFilterIssues_RootsOnly(t *testing.T) {
	now := time.Now()
	issues := []model.Issue{
		{ID: "aaa111", Title: "Epic", Status: "open", Created: now, Updated: now},
		{ID: "bbb222", Title: "Task 1", Status: "open", ParentID: "aaa111", Created: now, Updated: now},
		{ID: "ccc333", Title: "Standalone", Status: "open", Created: now, Updated: now},
	}

	got := FilterIssues(issues, FilterOptions{RootsOnly: true})
	if len(got) != 2 {
		t.Fatalf("count: got %d, want 2", len(got))
	}
	for _, i := range got {
		if i.ParentID != "" {
			t.Errorf("issue %s has parent %q, expected root", i.ID, i.ParentID)
		}
	}
}

func TestFilterIssues_ParentAndStatus(t *testing.T) {
	now := time.Now()
	issues := []model.Issue{
		{ID: "aaa111", Title: "Epic", Status: "open", Created: now, Updated: now},
		{ID: "bbb222", Title: "Task 1", Status: "open", ParentID: "aaa111", Created: now, Updated: now},
		{ID: "ccc333", Title: "Task 2", Status: "done", ParentID: "aaa111", Created: now, Updated: now},
	}

	got := FilterIssues(issues, FilterOptions{ParentID: "aaa111", Status: "open"})
	if len(got) != 1 {
		t.Fatalf("count: got %d, want 1", len(got))
	}
	if got[0].ID != "bbb222" {
		t.Errorf("expected bbb222, got %s", got[0].ID)
	}
}

func TestIssueRoundTrip_WithParentID(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	issue := model.Issue{
		ID:       "abc123",
		Title:    "Child issue",
		Status:   "open",
		ParentID: "parent1",
		Created:  now,
		Updated:  now,
	}

	data, err := json.Marshal(issue)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got model.Issue
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.ParentID != "parent1" {
		t.Errorf("parent_id: got %q, want parent1", got.ParentID)
	}
}

func TestIssueRoundTrip_ParentIDOmittedWhenEmpty(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	issue := model.Issue{
		ID:      "abc123",
		Title:   "No parent",
		Status:  "open",
		Created: now,
		Updated: now,
	}

	data, err := json.Marshal(issue)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(data), "parent_id") {
		t.Error("parent_id should be omitted when empty")
	}
}

func TestCompactIssue(t *testing.T) {
	root := t.TempDir()
	tr, err := Init(root)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	issue, err := tr.CreateIssue("Explore: big analysis", "# Full Analysis\n\nLots of detailed content here.\nLine 3.\nLine 4.", "", 0, []string{"explore"}, "", "", "testuser")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := tr.AddComment(issue.ID, "a comment", "testuser"); err != nil {
		t.Fatalf("comment: %v", err)
	}
	if _, err := tr.SetStatus(issue.ID, "active", "testuser"); err != nil {
		t.Fatalf("start: %v", err)
	}
	if _, err := tr.SetStatus(issue.ID, "done", "testuser"); err != nil {
		t.Fatalf("close: %v", err)
	}

	if err := tr.CompactIssue(issue.ID); err != nil {
		t.Fatalf("compact: %v", err)
	}

	loaded, err := tr.LoadIssue(issue.ID)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	// Description truncated to first line
	if loaded.Description != "# Full Analysis" {
		t.Errorf("description: got %q, want %q", loaded.Description, "# Full Analysis")
	}

	// Comments cleared
	if len(loaded.Comments) != 0 {
		t.Errorf("comments: got %d, want 0", len(loaded.Comments))
	}

	// History compacted to create + close
	events, err := tr.LoadEvents(issue.ID)
	if err != nil {
		t.Fatalf("load events: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("events: got %d, want 2", len(events))
	}
	if events[0].Op != "create" {
		t.Errorf("first event: got %q, want create", events[0].Op)
	}
	if events[1].Op != "status" || events[1].To != "done" {
		t.Errorf("last event: op=%q to=%q, want status/done", events[1].Op, events[1].To)
	}

	// Log entry written
	entries, err := tr.LoadLog()
	if err != nil {
		t.Fatalf("load log: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("log entries: got %d, want 1", len(entries))
	}
	if entries[0].ID != issue.ID {
		t.Errorf("log id: got %q, want %q", entries[0].ID, issue.ID)
	}
	if entries[0].Title != "Explore: big analysis" {
		t.Errorf("log title: got %q", entries[0].Title)
	}
}

func TestCompactIssue_OnlyDone(t *testing.T) {
	root := t.TempDir()
	tr, err := Init(root)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	issue, err := tr.CreateIssue("Active issue", "", "", 0, nil, "", "", "testuser")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := tr.SetStatus(issue.ID, "active", "testuser"); err != nil {
		t.Fatalf("start: %v", err)
	}

	err = tr.CompactIssue(issue.ID)
	if err == nil {
		t.Fatal("expected error compacting active issue")
	}
	if !strings.Contains(err.Error(), "can only compact done/cancelled") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCompactIssue_LongDescription(t *testing.T) {
	root := t.TempDir()
	tr, err := Init(root)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	longLine := strings.Repeat("x", 200)
	issue, err := tr.CreateIssue("Long desc", longLine, "", 0, nil, "", "", "testuser")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := tr.SetStatus(issue.ID, "active", "testuser"); err != nil {
		t.Fatalf("start: %v", err)
	}
	if _, err := tr.SetStatus(issue.ID, "done", "testuser"); err != nil {
		t.Fatalf("close: %v", err)
	}

	if err := tr.CompactIssue(issue.ID); err != nil {
		t.Fatalf("compact: %v", err)
	}

	loaded, err := tr.LoadIssue(issue.ID)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(loaded.Description) != 120 {
		t.Errorf("description length: got %d, want 120", len(loaded.Description))
	}
}

func TestAppendAndLoadLog(t *testing.T) {
	root := t.TempDir()
	tr, err := Init(root)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	now := time.Now().UTC()
	issue := model.Issue{
		ID:      "abc123",
		Title:   "Test issue",
		Type:    "bug",
		Status:  "done",
		Labels:  []string{"backend"},
		Created: now,
		Updated: now,
	}

	if err := tr.AppendLog(issue); err != nil {
		t.Fatalf("append: %v", err)
	}

	entries, err := tr.LoadLog()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("count: got %d, want 1", len(entries))
	}
	if entries[0].ID != "abc123" {
		t.Errorf("id: got %q", entries[0].ID)
	}
	if entries[0].Type != "bug" {
		t.Errorf("type: got %q", entries[0].Type)
	}
	if entries[0].Labels[0] != "backend" {
		t.Errorf("label: got %q", entries[0].Labels[0])
	}
}

func TestLoadLog_Empty(t *testing.T) {
	root := t.TempDir()
	tr, err := Init(root)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	entries, err := tr.LoadLog()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected empty, got %d", len(entries))
	}
}

func TestGarbageCollect(t *testing.T) {
	root := t.TempDir()
	tr, err := Init(root)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	// Create an issue and close it with a backdated Updated time
	issue, err := tr.CreateIssue("Old issue", "old content", "", 0, nil, "", "", "testuser")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := tr.SetStatus(issue.ID, "active", "testuser"); err != nil {
		t.Fatalf("start: %v", err)
	}
	if _, err := tr.SetStatus(issue.ID, "done", "testuser"); err != nil {
		t.Fatalf("close: %v", err)
	}

	// Backdate the issue to 60 days ago
	loaded, err := tr.LoadIssue(issue.ID)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	loaded.Updated = time.Now().UTC().AddDate(0, 0, -60)
	if err := tr.SaveIssue(loaded); err != nil {
		t.Fatalf("save: %v", err)
	}

	// Create a recent done issue (should NOT be purged)
	recent, err := tr.CreateIssue("Recent issue", "", "", 0, nil, "", "", "testuser")
	if err != nil {
		t.Fatalf("create recent: %v", err)
	}
	if _, err := tr.SetStatus(recent.ID, "active", "testuser"); err != nil {
		t.Fatalf("start recent: %v", err)
	}
	if _, err := tr.SetStatus(recent.ID, "done", "testuser"); err != nil {
		t.Fatalf("close recent: %v", err)
	}

	purged, err := tr.GarbageCollect(30)
	if err != nil {
		t.Fatalf("gc: %v", err)
	}

	if len(purged) != 1 {
		t.Fatalf("purged count: got %d, want 1", len(purged))
	}
	if purged[0] != issue.ID {
		t.Errorf("purged id: got %q, want %q", purged[0], issue.ID)
	}

	// Old issue directory should be gone
	_, err = tr.LoadIssue(issue.ID)
	if err == nil {
		t.Error("expected error loading purged issue")
	}

	// Recent issue should still exist
	_, err = tr.LoadIssue(recent.ID)
	if err != nil {
		t.Errorf("recent issue should still exist: %v", err)
	}

	// Log should have an entry for the purged issue
	entries, err := tr.LoadLog()
	if err != nil {
		t.Fatalf("load log: %v", err)
	}
	if len(entries) < 1 {
		t.Fatal("expected at least 1 log entry")
	}
	found := false
	for _, e := range entries {
		if e.ID == issue.ID {
			found = true
			break
		}
	}
	if !found {
		t.Error("purged issue not found in log")
	}
}

func TestCompactAllDone(t *testing.T) {
	root := t.TempDir()
	tr, err := Init(root)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	// Create 2 done issues and 1 active
	for _, title := range []string{"Done 1", "Done 2"} {
		issue, err := tr.CreateIssue(title, "content", "", 0, nil, "", "", "testuser")
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		if _, err := tr.SetStatus(issue.ID, "active", "testuser"); err != nil {
			t.Fatalf("start: %v", err)
		}
		if _, err := tr.SetStatus(issue.ID, "done", "testuser"); err != nil {
			t.Fatalf("close: %v", err)
		}
	}
	active, err := tr.CreateIssue("Still active", "important content", "", 0, nil, "", "", "testuser")
	if err != nil {
		t.Fatalf("create active: %v", err)
	}
	if _, err := tr.SetStatus(active.ID, "active", "testuser"); err != nil {
		t.Fatalf("start active: %v", err)
	}

	compacted, err := tr.CompactAllDone()
	if err != nil {
		t.Fatalf("compact all: %v", err)
	}
	if len(compacted) != 2 {
		t.Errorf("compacted count: got %d, want 2", len(compacted))
	}

	// Active issue should still have its description
	loaded, err := tr.LoadIssue(active.ID)
	if err != nil {
		t.Fatalf("load active: %v", err)
	}
	if loaded.Description != "important content" {
		t.Errorf("active description should be preserved: got %q", loaded.Description)
	}
}

func TestAddComment_NoTextInEvent(t *testing.T) {
	root := t.TempDir()
	tr, err := Init(root)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	issue, err := tr.CreateIssue("Comment test", "", "", 0, nil, "", "", "testuser")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if _, err := tr.AddComment(issue.ID, "hello world", "testuser"); err != nil {
		t.Fatalf("comment: %v", err)
	}

	// Comment text should be in issue.json
	loaded, err := tr.LoadIssue(issue.ID)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if loaded.Comments[0].Text != "hello world" {
		t.Errorf("comment text: got %q", loaded.Comments[0].Text)
	}

	// But NOT in history.jsonl event
	events, err := tr.LoadEvents(issue.ID)
	if err != nil {
		t.Fatalf("load events: %v", err)
	}
	commentEvent := events[1]
	if commentEvent.Text != "" {
		t.Errorf("event should not have Text, got %q", commentEvent.Text)
	}
	if commentEvent.By != "testuser" {
		t.Errorf("event By: got %q", commentEvent.By)
	}
}

// readEvents reads all events from a history.jsonl file.
func readEvents(t *testing.T, path string) []model.Event {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open history: %v", err)
	}
	defer f.Close()
	var events []model.Event
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var ev model.Event
		if err := json.Unmarshal(scanner.Bytes(), &ev); err != nil {
			t.Fatalf("parse event: %v", err)
		}
		events = append(events, ev)
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan: %v", err)
	}
	return events
}
