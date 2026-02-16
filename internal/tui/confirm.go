package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
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
	return fmt.Sprintf(
		"\n  %s issue %s? (y/n)\n",
		m.action,
		m.issueID,
	)
}
