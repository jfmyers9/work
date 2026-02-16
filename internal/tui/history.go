package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jfmyers9/work/internal/model"
)

type historyModel struct {
	issueID  string
	events   []model.Event
	viewport viewport.Model
}

func newHistoryModel(issueID string, events []model.Event, width, height int) historyModel {
	vp := viewport.New(width, height-4)
	vp.SetContent(renderHistory(issueID, events, width))
	return historyModel{
		issueID:  issueID,
		events:   events,
		viewport: vp,
	}
}

func (m historyModel) Update(msg tea.Msg) (historyModel, tea.Cmd) {
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 4
		m.viewport.SetContent(renderHistory(m.issueID, m.events, msg.Width))
	}
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m historyModel) View() string {
	return m.viewport.View()
}

func renderHistory(issueID string, events []model.Event, width int) string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString("  " + sectionStyle.Render("History for "+issueID))
	b.WriteString("\n")
	contentW := min(width-4, 80)
	b.WriteString("  " + dividerStyle.Render(strings.Repeat("─", contentW)))
	b.WriteString("\n\n")

	if len(events) == 0 {
		b.WriteString("  " + helpStyle.Render("No events recorded.") + "\n")
		return b.String()
	}

	for _, ev := range events {
		ts := ev.Timestamp.Format("2006-01-02 15:04")
		line := formatEvent(ev)
		by := ""
		if ev.By != "" {
			by = " by " + ev.By
		}
		b.WriteString(fmt.Sprintf("  %s  %s %s%s\n",
			commentMetaStyle.Render(ts),
			keyStyle.Width(10).Render(ev.Op),
			line,
			helpStyle.Render(by),
		))
	}

	return b.String()
}

func formatEvent(ev model.Event) string {
	switch ev.Op {
	case "status":
		return styledStatus(ev.From) + " → " + styledStatus(ev.To)
	case "link":
		return "→ parent " + ev.To
	case "unlink":
		return "✕ parent " + ev.From
	case "comment":
		return ""
	case "create":
		return ""
	default:
		if len(ev.Fields) > 0 {
			return strings.Join(ev.Fields, ", ")
		}
		return ""
	}
}
