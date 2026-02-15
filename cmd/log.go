package cmd

import (
	"fmt"
	"time"

	"github.com/jfmyers9/work/internal/tracker"
	"github.com/spf13/cobra"
)

var (
	logSince string
	logUntil string
)

var logCmd = &cobra.Command{
	Use:   "log <id>",
	Short: "Show issue event log",
	Long:  `Display the event history for a single issue.`,
	Example: `  work log abc123
  work log abc --since 2025-01-01`,
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

		events, err := t.LoadEvents(id)
		if err != nil {
			return err
		}

		var since, until time.Time
		if logSince != "" {
			since, err = parseTimeFlag(logSince)
			if err != nil {
				return err
			}
		}
		if logUntil != "" {
			until, err = parseTimeFlag(logUntil)
			if err != nil {
				return err
			}
		}
		events = tracker.FilterEventsByTime(events, since, until)

		if len(events) == 0 {
			fmt.Println("No events")
			return nil
		}
		for _, ev := range events {
			fmt.Printf("%s  %s  (%s)\n",
				ev.Timestamp.Format("2006-01-02 15:04:05"),
				formatEventDetail(ev),
				ev.By)
		}
		return nil
	},
}

func init() {
	logCmd.Flags().StringVar(&logSince, "since", "", "Show events after date (YYYY-MM-DD or RFC3339)")
	logCmd.Flags().StringVar(&logUntil, "until", "", "Show events before date")
	rootCmd.AddCommand(logCmd)
}
