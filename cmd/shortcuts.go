package cmd

import (
	"fmt"
	"os"

	"github.com/jfmyers9/work/internal/tracker"
	"github.com/spf13/cobra"
)

func newShortcutCmd(use, short, long, example, targetStatus string, withNoCompact bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:               use,
		Short:             short,
		Long:              long,
		Example:           example,
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

			old, err := t.LoadIssue(id)
			if err != nil {
				return err
			}
			oldStatus := old.Status

			user := tracker.ResolveUser()
			if _, err := t.SetStatus(id, targetStatus, user); err != nil {
				return err
			}
			fmt.Printf("%s: %s → %s\n", id, oldStatus, targetStatus)

			if targetStatus == "done" || targetStatus == "cancelled" {
				noCompact, _ := cmd.Flags().GetBool("no-compact")
				if !noCompact {
					if err := t.CompactIssue(id); err != nil {
						fmt.Fprintf(os.Stderr, "warning: compact failed: %v\n", err)
					}
				}
			}
			return nil
		},
	}
	if withNoCompact {
		cmd.Flags().Bool("no-compact", false, "Skip auto-compaction")
	}
	return cmd
}

func init() {
	rootCmd.AddCommand(newShortcutCmd(
		"close <id>",
		"Close an issue (set status to done)",
		`Shortcut for: work status <id> done`,
		"",
		"done",
		true,
	))
	rootCmd.AddCommand(newShortcutCmd(
		"cancel <id>",
		"Cancel an issue",
		`Shortcut for: work status <id> cancelled`,
		"",
		"cancelled",
		true,
	))
	rootCmd.AddCommand(newShortcutCmd(
		"reopen <id>",
		"Reopen an issue",
		`Shortcut for: work status <id> open`,
		"",
		"open",
		false,
	))
	rootCmd.AddCommand(newShortcutCmd(
		"start <id>",
		"Start working on an issue (set status to active)",
		`Shortcut for: work status <id> active`,
		"",
		"active",
		false,
	))
	rootCmd.AddCommand(newShortcutCmd(
		"review <id>",
		"Submit an issue for review (set status to review)",
		`Shortcut for: work status <id> review`,
		"",
		"review",
		false,
	))
	rootCmd.AddCommand(newShortcutCmd(
		"approve <id>",
		"Approve a reviewed issue (set status to done)",
		`Shortcut for: work status <id> done`,
		"",
		"done",
		false,
	))
}
