package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	pickerItemStyle    = lipgloss.NewStyle().Padding(0, 2)
	pickerSelectedItem = lipgloss.NewStyle().Padding(0, 2).
				Bold(true).
				Foreground(lipgloss.Color("229")).
				Background(lipgloss.Color("57"))
)

type statusChangedMsg struct {
	issueID string
	status  string
}

type statusPicker struct {
	issueID string
	current string
	options []string
	cursor  int
}

func newStatusPicker(issueID, current string, options []string) statusPicker {
	return statusPicker{
		issueID: issueID,
		current: current,
		options: options,
	}
}

func (p statusPicker) Update(msg tea.Msg) (statusPicker, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if p.cursor < len(p.options)-1 {
				p.cursor++
			}
		case "k", "up":
			if p.cursor > 0 {
				p.cursor--
			}
		}
	}
	return p, nil
}

func (p statusPicker) View() string {
	s := fmt.Sprintf("\n  Change status of %s from %s to:\n\n", p.issueID, styledStatus(p.current))
	for i, opt := range p.options {
		styled := styledStatus(opt)
		if i == p.cursor {
			s += pickerSelectedItem.Render("▸ "+opt) + "\n"
		} else {
			s += pickerItemStyle.Render("  "+styled) + "\n"
		}
	}
	return s
}

func (p statusPicker) selected() string {
	if len(p.options) == 0 {
		return ""
	}
	return p.options[p.cursor]
}
