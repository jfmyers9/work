package cmd

import (
	"fmt"

	"github.com/jfmyers9/work/internal/tracker"
	"github.com/spf13/cobra"
)

var rehashCmd = &cobra.Command{
	Use:   "rehash",
	Short: "Upgrade old hex IDs to Crockford Base32",
	Long: `Re-generate IDs for issues that use the old hex encoding.
New Crockford Base32 IDs are shorter to type and have better
prefix diversity. All references (parent links, log entries)
are updated automatically.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		t, err := loadTracker()
		if err != nil {
			return err
		}

		issues, err := t.ListIssues()
		if err != nil {
			return err
		}

		var rehashed int
		for _, issue := range issues {
			if !tracker.IsHexID(issue.ID) {
				continue
			}
			newID, err := t.RehashIssue(issue.ID)
			if err != nil {
				return fmt.Errorf("rehashing %s: %w", issue.ID, err)
			}
			fmt.Printf("%s → %s  %s\n", issue.ID, newID, issue.Title)
			rehashed++
		}

		if rehashed == 0 {
			fmt.Println("No hex IDs to rehash")
		} else {
			fmt.Printf("\nRehashed %d issues\n", rehashed)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(rehashCmd)
}
