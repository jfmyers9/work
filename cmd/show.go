package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jfmyers9/work/internal/model"
	"github.com/spf13/cobra"
)

var showFormat string

var showCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show issue details",
	Long: `Display full details for a single issue, including comments
and child issues.`,
	Example: `  work show abc123
  work show abc --format=json`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeIssueIDs,
	RunE: func(cmd *cobra.Command, args []string) error {
		prefix := args[0]
		t, err := loadTracker()
		if err != nil {
			return err
		}

		id, err := t.ResolvePrefix(prefix)
		if err != nil {
			entries, logErr := t.LoadLog()
			if logErr == nil {
				for _, e := range entries {
					if strings.HasPrefix(e.ID, prefix) {
						return fmt.Errorf("issue %s was purged (completed %s)\nUse 'work completed' to view completion history", e.ID, e.Closed.Format("2006-01-02"))
					}
				}
			}
			return err
		}
		issue, err := t.LoadIssue(id)
		if err != nil {
			return err
		}

		if showFormat == "json" {
			data, err := json.MarshalIndent(issue, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
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
		return nil
	},
}

func init() {
	showCmd.Flags().StringVar(&showFormat, "format", "", "Output format (json)")
	rootCmd.AddCommand(showCmd)
}
