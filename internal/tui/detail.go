package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jfmyers9/work/internal/model"
)

var (
	labelStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("252"))
	commentStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	dividerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
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

	b.WriteString(labelStyle.Render("ID:       "))
	b.WriteString(issue.ID[:min(6, len(issue.ID))])
	b.WriteString("\n")

	b.WriteString(labelStyle.Render("Title:    "))
	b.WriteString(issue.Title)
	b.WriteString("\n")

	b.WriteString(labelStyle.Render("Status:   "))
	b.WriteString(styledStatus(issue.Status))
	b.WriteString("\n")

	b.WriteString(labelStyle.Render("Type:     "))
	b.WriteString(styledType(issue.Type))
	b.WriteString("\n")

	b.WriteString(labelStyle.Render("Priority: "))
	b.WriteString(styledPriority(issue.Priority))
	b.WriteString("\n")

	if len(issue.Labels) > 0 {
		b.WriteString(labelStyle.Render("Labels:   "))
		b.WriteString(strings.Join(issue.Labels, ", "))
		b.WriteString("\n")
	}

	if issue.Assignee != "" {
		b.WriteString(labelStyle.Render("Assignee: "))
		b.WriteString(issue.Assignee)
		b.WriteString("\n")
	}

	if issue.ParentID != "" {
		b.WriteString(labelStyle.Render("Parent:   "))
		b.WriteString(issue.ParentID[:min(6, len(issue.ParentID))])
		b.WriteString("\n")
	}

	b.WriteString(labelStyle.Render("Created:  "))
	b.WriteString(issue.Created.Format("2006-01-02 15:04"))
	b.WriteString("\n")

	b.WriteString(labelStyle.Render("Updated:  "))
	b.WriteString(issue.Updated.Format("2006-01-02 15:04"))
	b.WriteString("\n")

	if issue.Description != "" {
		b.WriteString("\n")
		b.WriteString(dividerStyle.Render(strings.Repeat("─", min(width, 80))))
		b.WriteString("\n")
		b.WriteString(labelStyle.Render("Description"))
		b.WriteString("\n\n")
		b.WriteString(wordWrap(issue.Description, min(width-2, 78)))
		b.WriteString("\n")
	}

	if len(children) > 0 {
		b.WriteString("\n")
		b.WriteString(dividerStyle.Render(strings.Repeat("─", min(width, 80))))
		b.WriteString("\n")
		b.WriteString(labelStyle.Render("Children"))
		b.WriteString("\n\n")
		for _, c := range children {
			b.WriteString(fmt.Sprintf("  %s  %-10s  %s\n",
				c.ID[:min(6, len(c.ID))],
				styledStatus(c.Status),
				c.Title,
			))
		}
	}

	if len(issue.Comments) > 0 {
		b.WriteString("\n")
		b.WriteString(dividerStyle.Render(strings.Repeat("─", min(width, 80))))
		b.WriteString("\n")
		b.WriteString(labelStyle.Render("Comments"))
		b.WriteString("\n")
		for _, c := range issue.Comments {
			b.WriteString("\n")
			meta := c.Created.Format("2006-01-02 15:04")
			if c.By != "" {
				meta = c.By + " • " + meta
			}
			b.WriteString(commentStyle.Render(meta))
			b.WriteString("\n")
			b.WriteString(wordWrap(c.Text, min(width-2, 78)))
			b.WriteString("\n")
		}
	}

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
