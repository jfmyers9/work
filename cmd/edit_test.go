package cmd

import (
	"strings"
	"testing"

	"github.com/jfmyers9/work/internal/editor"
	"github.com/jfmyers9/work/internal/tracker"
)

func setupEditTest(t *testing.T) (*tracker.Tracker, string) {
	t.Helper()
	root := t.TempDir()
	tr, err := tracker.Init(root)
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	issue, err := tr.CreateIssue("Original title", "Original description", "jim", 2, []string{"backend"}, "bug", "", "testuser")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	return tr, issue.ID
}

func TestEditInEditor_ChangedFields(t *testing.T) {
	tr, id := setupEditTest(t)
	original := editor.OpenEditor
	defer func() { editor.OpenEditor = original }()

	editor.OpenEditor = func(content, prefix, editorBin string) (string, error) {
		return strings.Replace(content, "Original title", "Updated title", 1), nil
	}

	issue, err := tr.LoadIssue(id)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if err := editInEditor(tr, issue); err != nil {
		t.Fatalf("editInEditor: %v", err)
	}

	updated, err := tr.LoadIssue(id)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if updated.Title != "Updated title" {
		t.Errorf("title = %q, want %q", updated.Title, "Updated title")
	}
	if updated.Description != "Original description" {
		t.Errorf("description changed unexpectedly to %q", updated.Description)
	}

	events, err := tr.LoadEvents(id)
	if err != nil {
		t.Fatalf("load events: %v", err)
	}
	last := events[len(events)-1]
	if last.Op != "edit" {
		t.Fatalf("last event op = %q, want edit", last.Op)
	}
	if len(last.Fields) != 1 || last.Fields[0] != "title" {
		t.Errorf("edited fields = %v, want [title]", last.Fields)
	}
}

func TestEditInEditor_NoChanges(t *testing.T) {
	tr, id := setupEditTest(t)
	original := editor.OpenEditor
	defer func() { editor.OpenEditor = original }()

	editor.OpenEditor = func(content, prefix, editorBin string) (string, error) {
		return content, nil
	}

	issue, err := tr.LoadIssue(id)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if err := editInEditor(tr, issue); err != nil {
		t.Fatalf("editInEditor: %v", err)
	}

	events, err := tr.LoadEvents(id)
	if err != nil {
		t.Fatalf("load events: %v", err)
	}
	for _, ev := range events {
		if ev.Op == "edit" {
			t.Error("no edit event should be recorded when nothing changed")
		}
	}
}

func TestEditInEditor_Abort(t *testing.T) {
	tr, id := setupEditTest(t)
	original := editor.OpenEditor
	defer func() { editor.OpenEditor = original }()

	editor.OpenEditor = func(content, prefix, editorBin string) (string, error) {
		return "", editor.ErrAborted
	}

	issue, err := tr.LoadIssue(id)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if err := editInEditor(tr, issue); err != nil {
		t.Fatalf("editInEditor should not error on abort: %v", err)
	}
}

func TestEditInEditor_InvalidType(t *testing.T) {
	tr, id := setupEditTest(t)
	original := editor.OpenEditor
	defer func() { editor.OpenEditor = original }()

	editor.OpenEditor = func(content, prefix, editorBin string) (string, error) {
		return strings.Replace(content, "Type: bug", "Type: invalid", 1), nil
	}

	issue, err := tr.LoadIssue(id)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	err = editInEditor(tr, issue)
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
	if !strings.Contains(err.Error(), "invalid type") {
		t.Errorf("error = %q, want it to mention invalid type", err.Error())
	}
}

func TestEditInEditor_OnlyChangedFieldsInHistory(t *testing.T) {
	tr, id := setupEditTest(t)
	original := editor.OpenEditor
	defer func() { editor.OpenEditor = original }()

	editor.OpenEditor = func(content, prefix, editorBin string) (string, error) {
		content = strings.Replace(content, "Priority: 2", "Priority: 1", 1)
		content = strings.Replace(content, "Assignee: jim", "Assignee: alice", 1)
		return content, nil
	}

	issue, err := tr.LoadIssue(id)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if err := editInEditor(tr, issue); err != nil {
		t.Fatalf("editInEditor: %v", err)
	}

	events, err := tr.LoadEvents(id)
	if err != nil {
		t.Fatalf("load events: %v", err)
	}
	last := events[len(events)-1]
	if last.Op != "edit" {
		t.Fatalf("last event op = %q, want edit", last.Op)
	}
	fields := make(map[string]bool)
	for _, f := range last.Fields {
		fields[f] = true
	}
	if !fields["priority"] || !fields["assignee"] {
		t.Errorf("fields = %v, want priority and assignee", last.Fields)
	}
	if fields["title"] || fields["description"] || fields["type"] || fields["labels"] {
		t.Errorf("fields = %v, should not include unchanged fields", last.Fields)
	}
}
