package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rejectCmd = &cobra.Command{
	Use:   "reject <id> <reason>",
	Short: "Reject a reviewed issue (back to active + reason comment)",
	Long:  `Set status back to active and add a rejection comment.`,
	Example: `  work reject abc123 "Tests are failing"`,
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
		reason := args[1]

		old, err := t.LoadIssue(id)
		if err != nil {
			return err
		}
		oldStatus := old.Status

		if _, err := t.SetStatus(id, "active", cfg.User); err != nil {
			return err
		}
		if _, err := t.AddComment(id, "Rejected: "+reason, cfg.User); err != nil {
			return err
		}
		fmt.Printf("%s: %s → active (rejected: %s)\n", shortID(t, id), oldStatus, reason)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(rejectCmd)
}
