package cmd

import (
	"fmt"
	"strings"
	"time"

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
	Long:  `Update fields on an existing issue.`,
	Example: `  work edit abc123 --title "Updated title"
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
			return fmt.Errorf("no fields to update")
		}

		now := time.Now().UTC()
		issue.Updated = now
		if err := t.SaveIssue(issue); err != nil {
			return err
		}

		user := tracker.ResolveUser()
		event := model.Event{
			Timestamp: now,
			Op:        "edit",
			Fields:    edited,
			By:        user,
		}
		if err := t.AppendEvent(id, event); err != nil {
			return err
		}
		fmt.Printf("Updated %s\n", id)
		return nil
	},
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
