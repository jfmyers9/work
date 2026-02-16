package cmd

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jfmyers9/work/internal/editor"
	"github.com/jfmyers9/work/internal/model"
	"github.com/jfmyers9/work/internal/tracker"
	"github.com/spf13/cobra"
)

var (
	editTitle       string
	editDescription string
	editPriority    int
	editLabels      string
	editAssignee    string
	editType        string
)

var editCmd = &cobra.Command{
	Use:   "edit <id>",
	Short: "Edit an issue",
	Long:  `Update fields on an existing issue. If no flags are given, opens the issue in $EDITOR.`,
	Example: `  work edit abc123
  work edit abc123 --title "Updated title"
  work edit abc --priority 2 --labels urgent,backend`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeIssueIDs,
	RunE: func(cmd *cobra.Command, args []string) error {
		t, err := loadTracker()
		if err != nil {
			return err
		}
		id, err := resolveID(t, args[0])
		if err != nil {
			return err
		}

		issue, err := t.LoadIssue(id)
		if err != nil {
			return err
		}

		var edited []string
		if cmd.Flags().Changed("title") {
			issue.Title = editTitle
			edited = append(edited, "title")
		}
		if cmd.Flags().Changed("description") {
			issue.Description = editDescription
			edited = append(edited, "description")
		}
		if cmd.Flags().Changed("assignee") {
			issue.Assignee = editAssignee
			edited = append(edited, "assignee")
		}
		if cmd.Flags().Changed("priority") {
			issue.Priority = editPriority
			edited = append(edited, "priority")
		}
		if cmd.Flags().Changed("labels") {
			issue.Labels = strings.Split(editLabels, ",")
			edited = append(edited, "labels")
		}
		if cmd.Flags().Changed("type") {
			if err := tracker.ValidateType(t.Config, editType); err != nil {
				return err
			}
			issue.Type = editType
			edited = append(edited, "type")
		}

		if len(edited) == 0 {
			return editInEditor(t, issue)
		}

		now := time.Now().UTC()
		issue.Updated = now
		if err := t.SaveIssue(issue); err != nil {
			return err
		}

		event := model.Event{
			Timestamp: now,
			Op:        "edit",
			Fields:    edited,
			By:        cfg.User,
		}
		if err := t.AppendEvent(id, event); err != nil {
			return err
		}
		fmt.Printf("Updated %s\n", id)
		return nil
	},
}

func editInEditor(t *tracker.Tracker, issue model.Issue) error {
	content := editor.MarshalIssue(issue)
	result, err := editor.OpenEditor(content, "work-edit", cfg.Editor)
	if err != nil {
		if errors.Is(err, editor.ErrAborted) {
			fmt.Println("edit cancelled")
			return nil
		}
		return err
	}

	title, description, issueType, assignee, priority, labels, err := editor.UnmarshalIssue(result)
	if err != nil {
		return err
	}

	var edited []string
	if title != issue.Title {
		issue.Title = title
		edited = append(edited, "title")
	}
	if description != issue.Description {
		issue.Description = description
		edited = append(edited, "description")
	}
	if issueType != issue.Type {
		if err := tracker.ValidateType(t.Config, issueType); err != nil {
			return err
		}
		issue.Type = issueType
		edited = append(edited, "type")
	}
	if assignee != issue.Assignee {
		issue.Assignee = assignee
		edited = append(edited, "assignee")
	}
	if priority != issue.Priority {
		issue.Priority = priority
		edited = append(edited, "priority")
	}
	if !labelsEqual(issue.Labels, labels) {
		issue.Labels = labels
		edited = append(edited, "labels")
	}

	if len(edited) == 0 {
		fmt.Println("edit cancelled")
		return nil
	}

	now := time.Now().UTC()
	issue.Updated = now
	if err := t.SaveIssue(issue); err != nil {
		return err
	}

	event := model.Event{
		Timestamp: now,
		Op:        "edit",
		Fields:    edited,
		By:        cfg.User,
	}
	if err := t.AppendEvent(issue.ID, event); err != nil {
		return err
	}
	fmt.Printf("Updated %s\n", issue.ID)
	return nil
}

func labelsEqual(a, b []string) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func init() {
	editCmd.Flags().StringVar(&editTitle, "title", "", "New title")
	editCmd.Flags().StringVar(&editDescription, "description", "", "New description")
	editCmd.Flags().IntVar(&editPriority, "priority", 0, "New priority")
	editCmd.Flags().StringVar(&editLabels, "labels", "", "Replace labels (comma-separated)")
	editCmd.Flags().StringVar(&editAssignee, "assignee", "", "New assignee")
	editCmd.Flags().StringVar(&editType, "type", "", "New type (feature|bug|chore)")
	rootCmd.AddCommand(editCmd)
}
