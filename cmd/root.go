package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "work",
	Short:         "A lightweight, git-friendly issue tracker",
	Long:          "usage: work <command> [args]",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		printHelp(os.Stderr)
		return fmt.Errorf("unknown command")
	}

	defaultHelp := rootCmd.HelpFunc()
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		if cmd == rootCmd {
			printHelp(os.Stdout)
			return
		}
		defaultHelp(cmd, args)
	})

	rootCmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		return fmt.Errorf("unknown command: %s", os.Args[1])
	})
}

func printHelp(w *os.File) {
	fmt.Fprintln(w, "usage: work <command> [args]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Commands:")
	for _, c := range rootCmd.Commands() {
		if c.Name() == "completion" || c.Name() == "help" {
			continue
		}
		fmt.Fprintf(w, "  %-14s%s\n", c.Name(), c.Short)
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
