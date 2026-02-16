package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type confirmModel struct {
	issueID string
	action  string
}

func newConfirmModel(issueID, action string) confirmModel {
	return confirmModel{
		issueID: issueID,
		action:  action,
	}
}

func (m confirmModel) Update(msg tea.Msg) (confirmModel, bool, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "y", "Y":
			return m, true, nil
		case "n", "N", "esc":
			return m, false, nil
		}
	}
	return m, false, nil
}

func (m confirmModel) View() string {
	prompt := fmt.Sprintf("%s issue %s?", m.action, m.issueID)
	yKey := keyStyle.Render("y")
	nKey := keyStyle.Render("n")
	hint := fmt.Sprintf("  %s yes  %s no", yKey, nKey)

	content := lipgloss.NewStyle().Bold(true).Foreground(colorYellow).Render(prompt) +
		"\n\n" + hint

	return lipgloss.Place(0, 0, lipgloss.Left, lipgloss.Top,
		overlayStyle.Render(content))
}
