package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/jfmyers9/work/internal/tracker"
	"github.com/spf13/cobra"
)

var instructionsStatic bool

var instructionsCmd = &cobra.Command{
	Use:   "instructions",
	Short: "Print AI-oriented usage instructions (for Claude Code hooks)",
	Long: `Print usage instructions designed for AI consumption. Output includes
a command reference, behavioral guidance, and optionally a summary
of active issues.

Intended for use as a Claude Code SessionStart hook so that Claude
knows how to use the work CLI at the start of each session.`,
	Example: `  work instructions
  work instructions --static`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(`# work — issue tracker CLI

work is a lightweight, git-friendly issue tracker that stores
all data in a local .work/ directory. Use it for task tracking,
planning, and coordination.

## Command Reference`)
		fmt.Println()
		for _, c := range rootCmd.Commands() {
			if c.Name() == "instructions" || c.Name() == "completion" || c.Name() == "help" {
				continue
			}
			fmt.Printf("- work %s — %s\n", c.Name(), c.Short)
		}

		fmt.Println(`
## Key Workflows

**Creating and tracking issues:**
  work create "Title" --type feature --priority 2 --labels label1,label2
  work create "Title" --description "Details here"

**Lifecycle:** open → active → review → done (or cancelled)
  work start <id>    # open → active
  work review <id>   # active → review
  work approve <id>  # review → done
  work close <id>    # any → done
  work cancel <id>   # any → cancelled

**Viewing issues:**
  work list --status active
  work list --label <label> --format short
  work show <id>

**Editing and commenting:**
  work edit <id> --description "New description"
  work comment <id> "Note text"

## Guidelines

- Use work issues as the single source of truth for plans,
  notes, and state — no separate planning documents.
- Store exploration plans and findings in issue descriptions
  and comments.
- Check work list before starting work to see what is in flight.
- Use labels to group related issues.
- Use --format=json when you need to parse output programmatically.`)

		if instructionsStatic {
			return nil
		}

		wd, err := os.Getwd()
		if err != nil {
			return nil
		}
		t, err := tracker.Load(wd)
		if err != nil {
			return nil
		}
		allIssues, err := t.ListIssues()
		if err != nil {
			return nil
		}
		active := tracker.FilterIssues(allIssues, tracker.FilterOptions{Status: "active"})
		if len(active) == 0 {
			return nil
		}
		tracker.SortIssues(active, "priority")

		fmt.Println()
		fmt.Println("## Active Issues")
		fmt.Println()
		for _, issue := range active {
			title := issue.Title
			if len(title) > 60 {
				title = title[:57] + "..."
			}
			labels := ""
			if len(issue.Labels) > 0 {
				labels = " [" + strings.Join(issue.Labels, ", ") + "]"
			}
			fmt.Printf("- %s: %s (P%d)%s\n", issue.ID, title, issue.Priority, labels)
		}
		return nil
	},
}

func init() {
	instructionsCmd.Flags().BoolVar(&instructionsStatic, "static", false, "Omit the dynamic active-issues section")
	rootCmd.AddCommand(instructionsCmd)
}
