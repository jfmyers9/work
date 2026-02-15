package editor

import (
	"testing"
	"time"

	"github.com/jfmyers9/work/internal/model"
)

func TestRoundTrip(t *testing.T) {
	issue := model.Issue{
		ID:          "abc123",
		Title:       "Fix login timeout",
		Status:      "open",
		Type:        "bug",
		Priority:    1,
		Labels:      []string{"backend", "urgent"},
		Assignee:    "jim",
		Created:     time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC),
		Description: "Sessions expire too quickly.",
	}

	text := MarshalIssue(issue)
	title, desc, typ, assignee, prio, labels, err := UnmarshalIssue(text)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if title != issue.Title {
		t.Errorf("title = %q, want %q", title, issue.Title)
	}
	if desc != issue.Description {
		t.Errorf("description = %q, want %q", desc, issue.Description)
	}
	if typ != issue.Type {
		t.Errorf("type = %q, want %q", typ, issue.Type)
	}
	if assignee != issue.Assignee {
		t.Errorf("assignee = %q, want %q", assignee, issue.Assignee)
	}
	if prio != issue.Priority {
		t.Errorf("priority = %d, want %d", prio, issue.Priority)
	}
	if len(labels) != len(issue.Labels) {
		t.Fatalf("labels = %v, want %v", labels, issue.Labels)
	}
	for i, l := range labels {
		if l != issue.Labels[i] {
			t.Errorf("labels[%d] = %q, want %q", i, l, issue.Labels[i])
		}
	}
}

func TestEmptyDescription(t *testing.T) {
	issue := model.Issue{
		ID:      "abc123",
		Title:   "No description",
		Status:  "open",
		Type:    "feature",
		Created: time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC),
	}

	text := MarshalIssue(issue)
	_, desc, _, _, _, _, err := UnmarshalIssue(text)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if desc != "" {
		t.Errorf("description = %q, want empty", desc)
	}
}

func TestMultilineDescription(t *testing.T) {
	issue := model.Issue{
		ID:          "abc123",
		Title:       "Multi-line",
		Status:      "open",
		Type:        "feature",
		Created:     time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC),
		Description: "Line one.\nLine two.\nLine three.",
	}

	text := MarshalIssue(issue)
	_, desc, _, _, _, _, err := UnmarshalIssue(text)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if desc != issue.Description {
		t.Errorf("description = %q, want %q", desc, issue.Description)
	}
}

func TestNoLabels(t *testing.T) {
	issue := model.Issue{
		ID:      "abc123",
		Title:   "No labels",
		Status:  "open",
		Type:    "feature",
		Created: time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC),
	}

	text := MarshalIssue(issue)
	_, _, _, _, _, labels, err := UnmarshalIssue(text)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if labels != nil {
		t.Errorf("labels = %v, want nil", labels)
	}
}

func TestCommentLinesStripped(t *testing.T) {
	text := "Title: Test\nType: bug\nPriority: 0\nLabels: \nAssignee: \n\n# this is a comment\nReal description.\n# another comment\n"
	_, desc, _, _, _, _, err := UnmarshalIssue(text)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if desc != "Real description." {
		t.Errorf("description = %q, want %q", desc, "Real description.")
	}
}

func TestMissingTitleReturnsError(t *testing.T) {
	text := "Type: bug\nPriority: 0\nLabels: \nAssignee: \n\nSome body.\n"
	_, _, _, _, _, _, err := UnmarshalIssue(text)
	if err == nil {
		t.Fatal("expected error for missing title")
	}
}
