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
	width      int
}

func newCommentModel(issueID, issueTitle string, width int) commentModel {
	ta := textarea.New()
	ta.Placeholder = "Write a comment..."
	ta.Focus()
	ta.CharLimit = 4096
	w := min(width-8, 78)
	if w < 40 {
		w = 40
	}
	ta.SetWidth(w)
	ta.SetHeight(8)
	return commentModel{
		issueID:    issueID,
		issueTitle: issueTitle,
		textarea:   ta,
		width:      width,
	}
}

func (m commentModel) Update(msg tea.Msg) (commentModel, tea.Cmd) {
	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func (m commentModel) View() string {
	content := labelStyle.Render("Comment on: ") + valueStyle.Render(m.issueTitle) +
		"\n\n" + m.textarea.View() +
		"\n\n" + helpStyle.Render("ctrl+d: submit • esc: cancel")

	return overlayStyle.Render(content)
}
