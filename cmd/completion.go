package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:       "completion <shell>",
	Short:     "Generate shell completion script",
	Long:      `Generate shell completion script for bash, zsh, or fish.`,
	Example:   `  work completion bash > ~/.bash_completion.d/work
  work completion zsh > ~/.zfunc/_work
  work completion fish > ~/.config/fish/completions/work.fish`,
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"bash", "zsh", "fish"},
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletionV2(os.Stdout, true)
		case "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			return rootCmd.GenFishCompletion(os.Stdout, true)
		default:
			return fmt.Errorf("unsupported shell: %s (use bash, zsh, or fish)", args[0])
		}
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
