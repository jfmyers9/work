package tracker

import (
	"encoding/json"
	"testing"

	"github.com/jfmyers9/work/internal/model"
)

func TestFullWorkflow(t *testing.T) {
	root := t.TempDir()

	// Init tracker
	tr, err := Init(root)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	// Create an issue with all fields
	issue, err := tr.CreateIssue("Login timeout bug", "Sessions expire too quickly", "jim", 1, []string{"bug", "urgent"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if issue.Status != "open" {
		t.Fatalf("initial status: got %q, want open", issue.Status)
	}
	id := issue.ID

	// Start the issue (open → active)
	issue, err = tr.SetStatus(id, "active")
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	if issue.Status != "active" {
		t.Fatalf("after start: got %q, want active", issue.Status)
	}

	// Add a comment
	issue, err = tr.AddComment(id, "Investigating the root cause")
	if err != nil {
		t.Fatalf("comment: %v", err)
	}
	if len(issue.Comments) != 1 {
		t.Fatalf("comments count: got %d, want 1", len(issue.Comments))
	}
	if issue.Comments[0].Text != "Investigating the root cause" {
		t.Errorf("comment text: got %q", issue.Comments[0].Text)
	}
	if issue.Comments[0].By != "system" {
		t.Errorf("comment by: got %q", issue.Comments[0].By)
	}

	// Close the issue (active → done)
	issue, err = tr.SetStatus(id, "done")
	if err != nil {
		t.Fatalf("close: %v", err)
	}
	if issue.Status != "done" {
		t.Fatalf("after close: got %q, want done", issue.Status)
	}

	// Reopen (done → open)
	issue, err = tr.SetStatus(id, "open")
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	if issue.Status != "open" {
		t.Fatalf("after reopen: got %q, want open", issue.Status)
	}

	// List and verify
	issues, err := tr.ListIssues()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("list count: got %d, want 1", len(issues))
	}
	if issues[0].ID != id {
		t.Errorf("list id: got %q, want %q", issues[0].ID, id)
	}
	if issues[0].Status != "open" {
		t.Errorf("list status: got %q, want open", issues[0].Status)
	}

	// Show (reload) and verify all fields
	loaded, err := tr.LoadIssue(id)
	if err != nil {
		t.Fatalf("show: %v", err)
	}
	if loaded.Title != "Login timeout bug" {
		t.Errorf("title: got %q", loaded.Title)
	}
	if loaded.Description != "Sessions expire too quickly" {
		t.Errorf("description: got %q", loaded.Description)
	}
	if loaded.Priority != 1 {
		t.Errorf("priority: got %d", loaded.Priority)
	}
	if loaded.Assignee != "jim" {
		t.Errorf("assignee: got %q", loaded.Assignee)
	}
	if len(loaded.Labels) != 2 || loaded.Labels[0] != "bug" || loaded.Labels[1] != "urgent" {
		t.Errorf("labels: got %v", loaded.Labels)
	}
	if len(loaded.Comments) != 1 {
		t.Errorf("comments: got %d, want 1", len(loaded.Comments))
	}

	// Log: verify all events in order
	events, err := tr.LoadEvents(id)
	if err != nil {
		t.Fatalf("log: %v", err)
	}
	expectedOps := []string{"create", "status", "comment", "status", "status"}
	if len(events) != len(expectedOps) {
		t.Fatalf("events count: got %d, want %d", len(events), len(expectedOps))
	}
	for i, expected := range expectedOps {
		if events[i].Op != expected {
			t.Errorf("event %d: got op %q, want %q", i, events[i].Op, expected)
		}
	}
	// Verify specific event details
	if events[1].From != "open" || events[1].To != "active" {
		t.Errorf("start event: from=%q to=%q", events[1].From, events[1].To)
	}
	if events[2].Text != "Investigating the root cause" {
		t.Errorf("comment event text: got %q", events[2].Text)
	}
	if events[3].From != "active" || events[3].To != "done" {
		t.Errorf("close event: from=%q to=%q", events[3].From, events[3].To)
	}
	if events[4].From != "done" || events[4].To != "open" {
		t.Errorf("reopen event: from=%q to=%q", events[4].From, events[4].To)
	}

	// Export: verify JSON output
	data, err := json.MarshalIndent(issues, "", "  ")
	if err != nil {
		t.Fatalf("export marshal: %v", err)
	}
	var exported []model.Issue
	if err := json.Unmarshal(data, &exported); err != nil {
		t.Fatalf("export unmarshal: %v", err)
	}
	if len(exported) != 1 {
		t.Fatalf("exported count: got %d, want 1", len(exported))
	}
	if exported[0].ID != id {
		t.Errorf("exported id: got %q", exported[0].ID)
	}

	// --format=json on show: marshal single issue
	showData, err := json.MarshalIndent(loaded, "", "  ")
	if err != nil {
		t.Fatalf("show json marshal: %v", err)
	}
	var showIssue model.Issue
	if err := json.Unmarshal(showData, &showIssue); err != nil {
		t.Fatalf("show json unmarshal: %v", err)
	}
	if showIssue.ID != id || showIssue.Title != "Login timeout bug" {
		t.Errorf("show json: id=%q title=%q", showIssue.ID, showIssue.Title)
	}

	// --format=json on list: marshal issue array
	listData, err := json.MarshalIndent(issues, "", "  ")
	if err != nil {
		t.Fatalf("list json marshal: %v", err)
	}
	var listIssues []model.Issue
	if err := json.Unmarshal(listData, &listIssues); err != nil {
		t.Fatalf("list json unmarshal: %v", err)
	}
	if len(listIssues) != 1 || listIssues[0].Status != "open" {
		t.Errorf("list json: count=%d", len(listIssues))
	}
}

func TestAddComment(t *testing.T) {
	root := t.TempDir()
	tr, err := Init(root)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	issue, err := tr.CreateIssue("Comment test", "", "", 0, nil)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// Add multiple comments
	updated, err := tr.AddComment(issue.ID, "First comment")
	if err != nil {
		t.Fatalf("comment 1: %v", err)
	}
	if len(updated.Comments) != 1 {
		t.Fatalf("after 1st: got %d comments", len(updated.Comments))
	}

	updated, err = tr.AddComment(issue.ID, "Second comment")
	if err != nil {
		t.Fatalf("comment 2: %v", err)
	}
	if len(updated.Comments) != 2 {
		t.Fatalf("after 2nd: got %d comments", len(updated.Comments))
	}

	// Reload and verify persistence
	loaded, err := tr.LoadIssue(issue.ID)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(loaded.Comments) != 2 {
		t.Fatalf("persisted comments: got %d, want 2", len(loaded.Comments))
	}
	if loaded.Comments[0].Text != "First comment" {
		t.Errorf("comment 0: got %q", loaded.Comments[0].Text)
	}
	if loaded.Comments[1].Text != "Second comment" {
		t.Errorf("comment 1: got %q", loaded.Comments[1].Text)
	}

	// Verify history events
	events, err := tr.LoadEvents(issue.ID)
	if err != nil {
		t.Fatalf("load events: %v", err)
	}
	// create + 2 comments = 3 events
	if len(events) != 3 {
		t.Fatalf("events: got %d, want 3", len(events))
	}
	if events[1].Op != "comment" || events[1].Text != "First comment" {
		t.Errorf("event 1: op=%q text=%q", events[1].Op, events[1].Text)
	}
	if events[2].Op != "comment" || events[2].Text != "Second comment" {
		t.Errorf("event 2: op=%q text=%q", events[2].Op, events[2].Text)
	}

	// Updated timestamp should have advanced
	if !loaded.Updated.After(loaded.Created) {
		t.Error("updated timestamp should be after created")
	}
}

func TestFilterIssues_WithComments(t *testing.T) {
	root := t.TempDir()
	tr, err := Init(root)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	issue, err := tr.CreateIssue("With comments", "", "", 0, nil)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if _, err := tr.AddComment(issue.ID, "A comment"); err != nil {
		t.Fatalf("comment: %v", err)
	}

	// List should include the issue with its comment
	issues, err := tr.ListIssues()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("count: got %d", len(issues))
	}
	if len(issues[0].Comments) != 1 {
		t.Errorf("comments on listed issue: got %d, want 1", len(issues[0].Comments))
	}
}
