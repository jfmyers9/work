package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jfmyers9/work/internal/tracker"
	"github.com/spf13/cobra"
)

var (
	completedSince  string
	completedLabel  string
	completedType   string
	completedFormat string
)

var completedCmd = &cobra.Command{
	Use:   "completed",
	Short: "Show completion history from log",
	Long:  `Show completed issues from the completion log.`,
	Example: `  work completed
  work completed --since 2026-01-01
  work completed --label explore`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		t, err := loadTracker()
		if err != nil {
			return err
		}

		entries, err := t.LoadLog()
		if err != nil {
			return err
		}

		if completedSince != "" {
			since, err := parseTimeFlag(completedSince)
			if err != nil {
				return err
			}
			var filtered []tracker.LogEntry
			for _, e := range entries {
				if !e.Closed.Before(since) {
					filtered = append(filtered, e)
				}
			}
			entries = filtered
		}
		if completedLabel != "" {
			var filtered []tracker.LogEntry
			for _, e := range entries {
				for _, l := range e.Labels {
					if l == completedLabel {
						filtered = append(filtered, e)
						break
					}
				}
			}
			entries = filtered
		}
		if completedType != "" {
			var filtered []tracker.LogEntry
			for _, e := range entries {
				if e.Type == completedType {
					filtered = append(filtered, e)
				}
			}
			entries = filtered
		}

		if completedFormat == "json" {
			data, err := json.MarshalIndent(entries, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		}

		if len(entries) == 0 {
			fmt.Println("No completions")
			return nil
		}

		for _, e := range entries {
			labels := ""
			if len(e.Labels) > 0 {
				labels = " [" + strings.Join(e.Labels, ",") + "]"
			}
			fmt.Printf("%s  %s  %s  %s%s\n", e.Closed.Format("2006-01-02"), e.ID, e.Status, e.Title, labels)
		}
		return nil
	},
}

func init() {
	completedCmd.Flags().StringVar(&completedSince, "since", "", "Show entries after date")
	completedCmd.Flags().StringVar(&completedLabel, "label", "", "Filter by label")
	completedCmd.Flags().StringVar(&completedType, "type", "", "Filter by type")
	completedCmd.Flags().StringVar(&completedFormat, "format", "", "Output format (json)")
	rootCmd.AddCommand(completedCmd)
}
