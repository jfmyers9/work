package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var (
	createDescription string
	createPriority    int
	createLabels      string
	createAssignee    string
	createType        string
	createParent      string
)

var createCmd = &cobra.Command{
	Use:   "create <title>",
	Short: "Create a new issue",
	Long:  `Create a new issue with the given title.`,
	Example: `  work create "Fix login bug" --type bug --priority 1
  work create "Add search" --labels ui,search --assignee alice`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		title := args[0]

		var labels []string
		if createLabels != "" {
			labels = strings.Split(createLabels, ",")
		}

		t, err := loadTracker()
		if err != nil {
			return err
		}

		parentID := ""
		if createParent != "" {
			resolved, err := t.ResolvePrefix(createParent)
			if err != nil {
				return err
			}
			parentID = resolved
		}

		issue, err := t.CreateIssue(title, createDescription, createAssignee, createPriority, labels, createType, parentID, cfg.User)
		if err != nil {
			return err
		}
		fmt.Println(shortID(t, issue.ID))
		return nil
	},
}

func init() {
	createCmd.Flags().StringVar(&createDescription, "description", "", "Issue description")
	createCmd.Flags().IntVar(&createPriority, "priority", 0, "Priority level")
	createCmd.Flags().StringVar(&createLabels, "labels", "", "Comma-separated labels")
	createCmd.Flags().StringVar(&createAssignee, "assignee", "", "Assignee name")
	createCmd.Flags().StringVar(&createType, "type", "", "Issue type (feature|bug|chore)")
	createCmd.Flags().StringVar(&createParent, "parent", "", "Parent issue ID")
	rootCmd.AddCommand(createCmd)
}
