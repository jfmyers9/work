package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jfmyers9/work/internal/tracker"
	"github.com/spf13/cobra"
)

var initLocal bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize issue tracker in current directory",
	Long: `Initialize a new work tracker in the current directory.
Creates a .work/ directory to store issues and configuration.
Also creates a Claude Code SessionStart hook so that Claude
auto-discovers the work CLI.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		_, err = tracker.Init(wd)
		if err != nil {
			return err
		}
		fmt.Println("Initialized work tracker in .work/")
		fmt.Println("")
		fmt.Println("For compact git diffs, run:")
		fmt.Println("  git config diff.work.textconv 'jq -c .'")

		settingsFile := "settings.json"
		if initLocal {
			settingsFile = "settings.local.json"
		}
		claudeDir := filepath.Join(wd, ".claude")
		if err := os.MkdirAll(claudeDir, 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not create .claude/: %v\n", err)
			return nil
		}
		settingsPath := filepath.Join(claudeDir, settingsFile)

		existing := make(map[string]any)
		if data, err := os.ReadFile(settingsPath); err == nil {
			_ = json.Unmarshal(data, &existing)
		}
		existing["hooks"] = map[string]any{
			"SessionStart": []any{
				map[string]any{
					"matcher": "",
					"hooks": []any{
						map[string]any{
							"type":    "command",
							"command": "work instructions 2>/dev/null || true",
						},
					},
				},
			},
			"PreToolUse": []any{
				map[string]any{
					"matcher": "Bash",
					"hooks": []any{
						map[string]any{
							"type":    "command",
							"command": `sh -c 'input=$(cat); case "$input" in *"git commit"*) changes=$(git diff --name-only -- .work/ 2>/dev/null; git ls-files --others --exclude-standard -- .work/ 2>/dev/null); [ -n "$changes" ] && echo "Warning: unstaged .work/ changes — run: git add .work/";; esac; true'`,
						},
					},
				},
			},
		}
		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		enc.SetIndent("", "  ")
		enc.SetEscapeHTML(false)
		if err := enc.Encode(existing); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not marshal settings: %v\n", err)
			return nil
		}
		if err := os.WriteFile(settingsPath, buf.Bytes(), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not write %s: %v\n", settingsFile, err)
			return nil
		}
		fmt.Printf("Created Claude Code hook in .claude/%s\n", settingsFile)
		return nil
	},
}

func init() {
	initCmd.Flags().BoolVar(&initLocal, "local", false, "Write hook to settings.local.json instead of settings.json")
	rootCmd.AddCommand(initCmd)
}
