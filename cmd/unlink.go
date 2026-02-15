package cmd

import (
	"fmt"

	"github.com/jfmyers9/work/internal/tracker"
	"github.com/spf13/cobra"
)

var unlinkCmd = &cobra.Command{
	Use:   "unlink <child-id>",
	Short: "Remove parent from a child issue",
	Long:  `Remove the parent link from a child issue.`,
	Example: `  work unlink abc123`,
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

		user := tracker.ResolveUser()
		if _, err := t.UnlinkIssue(id, user); err != nil {
			return err
		}
		fmt.Printf("Unlinked %s\n", id)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(unlinkCmd)
}
