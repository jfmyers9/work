package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	prompt := fmt.Sprintf("Change %s from %s to:", p.issueID, styledStatus(p.current))

	var items string
	for i, opt := range p.options {
		if i == p.cursor {
			items += pickerSelectedStyle.Render("▸ " + opt) + "\n"
		} else {
			items += pickerItemStyle.Render("  " + styledStatus(opt)) + "\n"
		}
	}

	content := labelStyle.Render(prompt) + "\n\n" + items +
		"\n" + helpStyle.Render("j/k: navigate  enter: select  esc: cancel")

	return lipgloss.Place(0, 0, lipgloss.Left, lipgloss.Top,
		overlayStyle.Render(content))
}

func (p statusPicker) selected() string {
	if len(p.options) == 0 {
		return ""
	}
	return p.options[p.cursor]
}
