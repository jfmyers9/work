package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jfmyers9/work/internal/model"
)

type detailModel struct {
	issue    model.Issue
	children []model.Issue
	viewport viewport.Model
	ready    bool
}

func newDetailModel(issue model.Issue, children []model.Issue, width, height int) detailModel {
	vp := viewport.New(width, height-4)
	vp.SetContent(renderDetail(issue, children, width))
	return detailModel{
		issue:    issue,
		children: children,
		viewport: vp,
		ready:    true,
	}
}

func (m detailModel) Update(msg tea.Msg) (detailModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 4
		m.viewport.SetContent(renderDetail(m.issue, m.children, msg.Width))
	}
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m detailModel) View() string {
	return m.viewport.View()
}

func renderDetail(issue model.Issue, children []model.Issue, width int) string {
	var b strings.Builder
	contentW := min(width-4, 80)
	labelW := 10

	field := func(name, value string) {
		b.WriteString("  " + labelStyle.Width(labelW).Render(name) + " " + valueStyle.Render(value) + "\n")
	}

	b.WriteString("\n")
	field("ID", issue.ID[:min(6, len(issue.ID))])
	field("Title", issue.Title)
	field("Status", styledStatus(issue.Status))
	field("Type", styledType(issue.Type))
	field("Priority", styledPriority(issue.Priority))

	if len(issue.Labels) > 0 {
		tags := make([]string, len(issue.Labels))
		for i, l := range issue.Labels {
			tags[i] = filterTagStyle.Render(l)
		}
		b.WriteString("  " + labelStyle.Width(labelW).Render("Labels") + " " + strings.Join(tags, " ") + "\n")
	}

	if issue.Assignee != "" {
		field("Assignee", issue.Assignee)
	}
	if issue.ParentID != "" {
		field("Parent", issue.ParentID[:min(6, len(issue.ParentID))])
	}
	field("Created", issue.Created.Format("2006-01-02 15:04"))
	field("Updated", issue.Updated.Format("2006-01-02 15:04"))

	if issue.Description != "" {
		b.WriteString("\n")
		b.WriteString("  " + sectionStyle.Render("Description") + "\n")
		b.WriteString("  " + dividerStyle.Render(strings.Repeat("─", contentW)) + "\n")
		b.WriteString("\n")
		wrapped := wordWrap(issue.Description, contentW-2)
		for _, line := range strings.Split(wrapped, "\n") {
			b.WriteString("  " + line + "\n")
		}
	}

	if len(children) > 0 {
		b.WriteString("\n")
		b.WriteString("  " + sectionStyle.Render("Children") + "\n")
		b.WriteString("  " + dividerStyle.Render(strings.Repeat("─", contentW)) + "\n")
		b.WriteString("\n")
		for _, c := range children {
			id := lipgloss.NewStyle().Foreground(colorMuted).Render(c.ID[:min(6, len(c.ID))])
			b.WriteString(fmt.Sprintf("  %s  %-10s  %s\n",
				id,
				styledStatus(c.Status),
				valueStyle.Render(c.Title),
			))
		}
	}

	if len(issue.Comments) > 0 {
		b.WriteString("\n")
		b.WriteString("  " + sectionStyle.Render("Comments") + "\n")
		b.WriteString("  " + dividerStyle.Render(strings.Repeat("─", contentW)) + "\n")
		for _, c := range issue.Comments {
			b.WriteString("\n")
			meta := c.Created.Format("2006-01-02 15:04")
			if c.By != "" {
				meta = c.By + " • " + meta
			}
			b.WriteString("  " + commentMetaStyle.Render(meta) + "\n")
			wrapped := wordWrap(c.Text, contentW-4)
			for _, line := range strings.Split(wrapped, "\n") {
				b.WriteString("  │ " + line + "\n")
			}
		}
	}

	b.WriteString("\n")
	return b.String()
}

func wordWrap(s string, width int) string {
	if width <= 0 {
		return s
	}
	var result strings.Builder
	for _, line := range strings.Split(s, "\n") {
		if len(line) <= width {
			result.WriteString(line)
			result.WriteString("\n")
			continue
		}
		words := strings.Fields(line)
		if len(words) == 0 {
			result.WriteString("\n")
			continue
		}
		col := 0
		for i, w := range words {
			if i > 0 && col+1+len(w) > width {
				result.WriteString("\n")
				col = 0
			} else if i > 0 {
				result.WriteString(" ")
				col++
			}
			result.WriteString(w)
			col += len(w)
		}
		result.WriteString("\n")
	}
	return strings.TrimRight(result.String(), "\n")
}
