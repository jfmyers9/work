package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var gcDays int

var gcCmd = &cobra.Command{
	Use:   "gc",
	Short: "Delete old completed issue directories",
	Long: `Delete issue directories for issues completed more than N
days ago (default: 30). Metadata is preserved in
.work/log.jsonl.`,
	Example: `  work gc
  work gc --days 7`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		t, err := loadTracker()
		if err != nil {
			return err
		}
		purged, err := t.GarbageCollect(gcDays)
		if err != nil {
			return err
		}

		if len(purged) == 0 {
			fmt.Println("No issues to purge")
			return nil
		}

		fmt.Printf("Purged %d issues\n", len(purged))
		for _, id := range purged {
			fmt.Printf("  %s\n", id)
		}
		fmt.Println("Use 'work completed' to view completion history")
		return nil
	},
}

func init() {
	gcCmd.Flags().IntVar(&gcDays, "days", 30, "Age threshold in days")
	rootCmd.AddCommand(gcCmd)
}
