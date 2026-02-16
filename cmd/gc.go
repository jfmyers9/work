package cmd

import (
	"fmt"
	"sort"
	"time"

	"github.com/jfmyers9/work/internal/model"
	"github.com/spf13/cobra"
)

var (
	gcDays int
	gcKeep int
)

var gcCmd = &cobra.Command{
	Use:   "gc",
	Short: "Delete old completed issue directories",
	Long: `Delete issue directories for completed issues. Uses --days
(age threshold, default 30) and/or --keep N (keep only the
last N completed issues). When both are given, an issue is
only purged if it exceeds BOTH thresholds. Metadata is
preserved in .work/log.jsonl.`,
	Example: `  work gc
  work gc --days 7
  work gc --keep 10`,
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

		// Collect done/cancelled issues sorted by updated desc (newest first)
		var completed []model.Issue
		for _, issue := range allIssues {
			if issue.Status == "done" || issue.Status == "cancelled" {
				completed = append(completed, issue)
			}
		}
		sort.Slice(completed, func(i, j int) bool {
			return completed[i].Updated.After(completed[j].Updated)
		})

		// Determine which issues to purge
		cutoff := time.Now().UTC().AddDate(0, 0, -gcDays)
		keepByCount := make(map[string]bool)
		if gcKeep > 0 {
			for i, issue := range completed {
				if i < gcKeep {
					keepByCount[issue.ID] = true
				}
			}
		}

		daysSet := cmd.Flags().Changed("days")
		keepSet := gcKeep > 0

		var toPurge []model.Issue
		for _, issue := range completed {
			oldEnough := issue.Updated.Before(cutoff)
			beyondKeep := keepSet && !keepByCount[issue.ID]

			switch {
			case keepSet && daysSet:
				// Both flags: purge only if exceeds both thresholds
				if beyondKeep && oldEnough {
					toPurge = append(toPurge, issue)
				}
			case keepSet:
				// --keep only: purge everything beyond the limit
				if beyondKeep {
					toPurge = append(toPurge, issue)
				}
			default:
				// --days only (or neither, using default 30)
				if oldEnough {
					toPurge = append(toPurge, issue)
				}
			}
		}

		if len(toPurge) == 0 {
			fmt.Println("No issues to purge")
			return nil
		}

		var purged []string
		for _, issue := range toPurge {
			if err := t.PurgeIssue(issue); err != nil {
				return err
			}
			purged = append(purged, issue.ID)
		}

		deduped, err := t.DeduplicateLog()
		if err != nil {
			return fmt.Errorf("deduplicating log: %w", err)
		}

		fmt.Printf("Purged %d issues\n", len(purged))
		for _, id := range purged {
			fmt.Printf("  %s\n", id)
		}
		if deduped > 0 {
			fmt.Printf("Removed %d duplicate log entries\n", deduped)
		}
		fmt.Println("Use 'work completed' to view completion history")
		return nil
	},
}

func init() {
	gcCmd.Flags().IntVar(&gcDays, "days", 30, "Age threshold in days")
	gcCmd.Flags().IntVar(&gcKeep, "keep", 0, "Keep only the last N completed issues")
	rootCmd.AddCommand(gcCmd)
}
