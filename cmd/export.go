package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:     "export",
	Short:   "Export issues as JSON",
	Long:    `Export all issues as a JSON array to stdout.`,
	Example: `  work export`,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		t, err := loadTracker()
		if err != nil {
			return err
		}
		issues, err := t.ListIssues()
		if err != nil {
			return err
		}
		data, err := json.MarshalIndent(issues, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)
}
