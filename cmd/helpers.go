package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jfmyers9/work/internal/model"
	"github.com/jfmyers9/work/internal/tracker"
	"github.com/spf13/cobra"
)

func loadTracker() (*tracker.Tracker, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	t, err := tracker.Load(wd)
	if err != nil {
		return nil, err
	}
	return t, nil
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
		if ev.Text != "" {
			text := ev.Text
			if len(text) > 60 {
				text = text[:57] + "..."
			}
			return fmt.Sprintf("comment: %s", text)
		}
		return "comment"
	case "link":
		return fmt.Sprintf("link: parent=%s", ev.To)
	case "unlink":
		return fmt.Sprintf("unlink: was parent=%s", ev.From)
	default:
		return ev.Op
	}
}

func resolveID(t *tracker.Tracker, prefix string) (string, error) {
	id, err := t.ResolvePrefix(prefix)
	if err != nil {
		return "", err
	}
	return id, nil
}

// shortID returns the minimum unique prefix for a full issue ID.
func shortID(t *tracker.Tracker, id string) string {
	issues, err := t.ListIssues()
	if err != nil {
		return id
	}
	ids := make([]string, len(issues))
	for i, issue := range issues {
		ids[i] = issue.ID
	}
	return tracker.MinPrefix(id, ids)
}

func completeIssueIDs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	entries, err := os.ReadDir(filepath.Join(wd, ".work", "issues"))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	var ids []string
	for _, e := range entries {
		if e.IsDir() {
			ids = append(ids, e.Name())
		}
	}
	return ids, cobra.ShellCompDirectiveNoFileComp
}
