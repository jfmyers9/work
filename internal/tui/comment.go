package tui

import (
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

type commentAddedMsg struct {
	issueID string
}

type commentModel struct {
	issueID    string
	issueTitle string
	textarea   textarea.Model
}

func newCommentModel(issueID, issueTitle string) commentModel {
	ta := textarea.New()
	ta.Placeholder = "Write a comment..."
	ta.Focus()
	ta.CharLimit = 4096
	ta.SetWidth(78)
	ta.SetHeight(8)
	return commentModel{
		issueID:    issueID,
		issueTitle: issueTitle,
		textarea:   ta,
	}
}

func (m commentModel) Update(msg tea.Msg) (commentModel, tea.Cmd) {
	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func (m commentModel) View() string {
	return "\n  " + labelStyle.Render("Comment on: ") + m.issueTitle +
		"\n\n" + m.textarea.View() +
		"\n\n" + helpStyle.Render("  ctrl+d: submit • esc: cancel")
}
