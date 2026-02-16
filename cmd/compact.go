package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var compactAllDone bool
var compactRewrite bool

var compactCmd = &cobra.Command{
	Use:   "compact [id]",
	Short: "Compact completed issues to save space",
	Long: `Compact a completed issue by stripping its description,
comments, and history to minimal metadata.`,
	Example: `  work compact abc123
  work compact --all-done`,
	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: completeIssueIDs,
	RunE: func(cmd *cobra.Command, args []string) error {
		t, err := loadTracker()
		if err != nil {
			return err
		}

		if compactRewrite {
			n, err := t.RewriteAllIssues()
			if err != nil {
				return err
			}
			fmt.Printf("Rewrote %d issues\n", n)
			return nil
		}

		if compactAllDone {
			compacted, err := t.CompactAllDone()
			if err != nil {
				return err
			}
			if len(compacted) == 0 {
				fmt.Println("No done/cancelled issues to compact")
				return nil
			}
			fmt.Printf("Compacted %d issues\n", len(compacted))
			return nil
		}

		if len(args) == 0 {
			return fmt.Errorf("usage: work compact <id-or-prefix> or work compact --all-done")
		}

		id, err := resolveID(t, args[0])
		if err != nil {
			return err
		}
		if err := t.CompactIssue(id); err != nil {
			return err
		}
		fmt.Printf("Compacted %s\n", shortID(t, id))
		return nil
	},
}

func init() {
	compactCmd.Flags().BoolVar(&compactAllDone, "all-done", false, "Compact all done/cancelled issues")
	compactCmd.Flags().BoolVar(&compactRewrite, "rewrite", false, "Rewrite all issues to current on-disk format")
	rootCmd.AddCommand(compactCmd)
}
