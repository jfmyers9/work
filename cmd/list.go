package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/jfmyers9/work/internal/tracker"
	"github.com/spf13/cobra"
)

var (
	listStatus   string
	listLabel    string
	listAssignee string
	listType     string
	listPriority int
	listParent   string
	listRoots    bool
	listSort     string
	listFormat   string
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List issues",
	Long:  `List issues with optional filtering and sorting.`,
	Example: `  work list --status active
  work list --label backend --sort priority`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		t, err := loadTracker()
		if err != nil {
			return err
		}
		allIssues, err := t.ListIssues()
		if err != nil {
			return err
		}

		opts := tracker.FilterOptions{
			Status:   listStatus,
			Label:    listLabel,
			Assignee: listAssignee,
			Type:     listType,
		}
		if cmd.Flags().Changed("priority") {
			opts.Priority = listPriority
			opts.HasPriority = true
		}
		if listParent != "" {
			resolved, err := t.ResolvePrefix(listParent)
			if err != nil {
				return err
			}
			opts.ParentID = resolved
		}
		if listRoots {
			opts.RootsOnly = true
		}
		issues := tracker.FilterIssues(allIssues, opts)
		tracker.SortIssues(issues, listSort)

		if listFormat == "json" {
			data, err := json.MarshalIndent(issues, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		}

		if listFormat == "short" {
			for _, issue := range issues {
				fmt.Printf("%s %s\n", issue.ID, issue.Title)
			}
			return nil
		}

		if len(issues) == 0 {
			fmt.Println("No issues found")
			return nil
		}

		childCounts := make(map[string]struct{ done, total int })
		for _, issue := range allIssues {
			if issue.ParentID != "" {
				c := childCounts[issue.ParentID]
				c.total++
				if issue.Status == "done" || issue.Status == "cancelled" {
					c.done++
				}
				childCounts[issue.ParentID] = c
			}
		}

		fmt.Printf("%-8s %-10s %-10s %-8s %-10s %s\n", "ID", "STATUS", "TYPE", "PRIORITY", "CHILDREN", "TITLE")
		for _, issue := range issues {
			title := issue.Title
			if len(title) > 50 {
				title = title[:47] + "..."
			}
			children := ""
			if c, ok := childCounts[issue.ID]; ok {
				children = fmt.Sprintf("%d/%d", c.done, c.total)
			}
			fmt.Printf("%-8s %-10s %-10s %-8d %-10s %s\n", issue.ID, issue.Status, issue.Type, issue.Priority, children, title)
		}
		return nil
	},
}

func init() {
	listCmd.Flags().StringVar(&listStatus, "status", "", "Filter by status (open|active|review|done|cancelled)")
	listCmd.Flags().StringVar(&listLabel, "label", "", "Filter by label")
	listCmd.Flags().StringVar(&listAssignee, "assignee", "", "Filter by assignee")
	listCmd.Flags().StringVar(&listType, "type", "", "Filter by type")
	listCmd.Flags().IntVar(&listPriority, "priority", 0, "Filter by priority")
	listCmd.Flags().StringVar(&listParent, "parent", "", "Filter by parent issue")
	listCmd.Flags().BoolVar(&listRoots, "roots", false, "Show only root issues (no parent)")
	listCmd.Flags().StringVar(&listSort, "sort", "", "Sort by field (title|priority|status|created|updated)")
	listCmd.Flags().StringVar(&listFormat, "format", "", "Output format (json|short)")
	rootCmd.AddCommand(listCmd)
}
