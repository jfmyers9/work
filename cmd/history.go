package cmd

import (
	"fmt"
	"sort"
	"time"

	"github.com/jfmyers9/work/internal/tracker"
	"github.com/spf13/cobra"
)

var (
	historyLabel string
	historySince string
	historyUntil string
	historyLast  int
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Show all events across issues",
	Long: `Display recent events across all issues (most recent first,
limited to 20).`,
	Example: `  work history --since 2025-01-01
  work history --label backend`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		t, err := loadTracker()
		if err != nil {
			return err
		}

		all, err := t.LoadAllEvents()
		if err != nil {
			return err
		}

		if historyLabel != "" {
			issues, err := t.ListIssues()
			if err != nil {
				return err
			}
			validIDs := make(map[string]bool)
			for _, issue := range issues {
				for _, l := range issue.Labels {
					if l == historyLabel {
						validIDs[issue.ID] = true
						break
					}
				}
			}
			var filtered []tracker.EventWithIssue
			for _, ev := range all {
				if validIDs[ev.IssueID] {
					filtered = append(filtered, ev)
				}
			}
			all = filtered
		}

		var since, until time.Time
		if historySince != "" {
			since, err = parseTimeFlag(historySince)
			if err != nil {
				return err
			}
		}
		if historyUntil != "" {
			until, err = parseTimeFlag(historyUntil)
			if err != nil {
				return err
			}
		}
		all = tracker.FilterEventsWithIssueByTime(all, since, until)

		sort.Slice(all, func(i, j int) bool {
			return all[i].Timestamp.After(all[j].Timestamp)
		})

		if len(all) == 0 {
			fmt.Println("No events")
			return nil
		}

		limit := historyLast
		if limit <= 0 {
			limit = 20
		}
		if len(all) < limit {
			limit = len(all)
		}
		ids := make([]string, len(all))
		for i, ev := range all {
			ids[i] = ev.IssueID
		}
		short := tracker.MinPrefixes(ids)

		for _, ev := range all[:limit] {
			fmt.Printf("%s  %s  %s  (%s)\n",
				ev.Timestamp.Format("2006-01-02 15:04:05"),
				short[ev.IssueID],
				formatEventDetail(ev.Event),
				ev.By)
		}
		return nil
	},
}

func init() {
	historyCmd.Flags().StringVar(&historyLabel, "label", "", "Filter to issues with this label")
	historyCmd.Flags().StringVar(&historySince, "since", "", "Show events after date (YYYY-MM-DD or RFC3339)")
	historyCmd.Flags().StringVar(&historyUntil, "until", "", "Show events before date")
	historyCmd.Flags().IntVar(&historyLast, "last", 0, "Show only the last N events (default 20)")
	rootCmd.AddCommand(historyCmd)
}
