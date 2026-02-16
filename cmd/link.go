package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var linkCmd = &cobra.Command{
	Use:   "link <child-id>",
	Short: "Link a child issue to a parent",
	Long:  `Set a parent-child relationship between two issues.`,
	Example: `  work link abc123 --parent def456`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeIssueIDs,
	RunE: func(cmd *cobra.Command, args []string) error {
		parentPrefix, _ := cmd.Flags().GetString("parent")

		t, err := loadTracker()
		if err != nil {
			return err
		}
		childID, err := resolveID(t, args[0])
		if err != nil {
			return err
		}
		parentID, err := resolveID(t, parentPrefix)
		if err != nil {
			return err
		}

		if _, err := t.LinkIssue(childID, parentID, cfg.User); err != nil {
			return err
		}
		fmt.Printf("Linked %s → %s\n", childID, parentID)
		return nil
	},
}

func init() {
	linkCmd.Flags().String("parent", "", "Parent issue ID")
	_ = linkCmd.MarkFlagRequired("parent")
	rootCmd.AddCommand(linkCmd)
}
