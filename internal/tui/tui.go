// Package tui provides an interactive terminal UI for browsing issues.
package tui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jfmyers9/work/internal/config"
	"github.com/jfmyers9/work/internal/tracker"
)

func Run() error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getwd: %w", err)
	}

	t, err := tracker.Load(wd)
	if err != nil {
		return fmt.Errorf("load tracker: %w", err)
	}

	issues, err := t.ListIssues()
	if err != nil {
		return fmt.Errorf("list issues: %w", err)
	}

	tracker.SortIssues(issues, "priority")

	cfg, _ := config.Load()

	m := newModel(t, issues, cfg.User, cfg.Editor)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}
