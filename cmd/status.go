package cmd

import (
	"fmt"
	"os"

	"github.com/jfmyers9/work/internal/tracker"
	"github.com/spf13/cobra"
)

var statusNoCompact bool

var statusCmd = &cobra.Command{
	Use:   "status <id> <state>",
	Short: "Change issue status",
	Long: `Set an issue's status to any valid state.
Valid states: open, active, review, done, cancelled.`,
	Example: `  work status abc123 active
  work status abc done`,
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
		newStatus := args[1]

		old, err := t.LoadIssue(id)
		if err != nil {
			return err
		}
		oldStatus := old.Status

		user := tracker.ResolveUser()
		if _, err := t.SetStatus(id, newStatus, user); err != nil {
			return err
		}
		fmt.Printf("%s: %s → %s\n", id, oldStatus, newStatus)

		if newStatus == "done" || newStatus == "cancelled" {
			if !statusNoCompact {
				if err := t.CompactIssue(id); err != nil {
					fmt.Fprintf(os.Stderr, "warning: compact failed: %v\n", err)
				}
			}
		}
		return nil
	},
}

func init() {
	statusCmd.Flags().BoolVar(&statusNoCompact, "no-compact", false, "Skip auto-compaction")
	rootCmd.AddCommand(statusCmd)
}
