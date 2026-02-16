package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var commentCmd = &cobra.Command{
	Use:   "comment <id> <text>",
	Short: "Add a comment to an issue",
	Long:  `Add a text comment to an issue.`,
	Example: `  work comment abc123 "Fixed in latest commit"`,
	Args:              cobra.ExactArgs(2),
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

		if _, err := t.AddComment(id, args[1], cfg.User); err != nil {
			return err
		}
		fmt.Printf("Commented on %s\n", id)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(commentCmd)
}
