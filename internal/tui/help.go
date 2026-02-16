package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type helpModel struct {
	visible bool
}

func (h helpModel) View(width int) string {
	if !h.visible {
		return ""
	}

	sections := []struct {
		title string
		keys  [][2]string
	}{
		{
			title: "Navigation",
			keys: [][2]string{
				{"j/k", "up / down"},
				{"enter", "open issue"},
				{"esc", "back"},
				{"q", "quit"},
			},
		},
		{
			title: "Actions",
			keys: [][2]string{
				{"s", "change status"},
				{"a/d/r/x", "active / done / review / cancel"},
				{"c", "add comment"},
				{"e", "edit in $EDITOR"},
				{"n", "new issue"},
				{"p/P", "link / unlink parent"},
			},
		},
		{
			title: "Filters (list)",
			keys: [][2]string{
				{"f", "cycle status filter"},
				{"t", "cycle type filter"},
				{"o", "cycle sort order"},
				{"F", "clear filters"},
				{"/", "search by title"},
			},
		},
		{
			title: "Views",
			keys: [][2]string{
				{"h", "event history (detail)"},
				{"?", "toggle this help"},
			},
		},
	}

	keyCol := 12
	var b strings.Builder

	for i, sec := range sections {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(sectionStyle.Render(sec.title))
		b.WriteString("\n")
		for _, k := range sec.keys {
			key := keyStyle.Width(keyCol).Render(k[0])
			desc := descStyle.Render(k[1])
			b.WriteString("  " + key + desc + "\n")
		}
	}

	b.WriteString("\n" + helpStyle.Render("press ? to close"))

	w := min(width-4, 50)
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorAccent).
		Padding(1, 3).
		Width(w).
		Render(b.String())

	return "\n" + box
}
