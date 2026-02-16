package tui

import "strings"

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

	w := min(width-4, 50)
	border := strings.Repeat("─", w)

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(dividerStyle.Render("  " + border))
	b.WriteString("\n")
	b.WriteString(labelStyle.Render("  Key Bindings"))
	b.WriteString("\n")

	for _, sec := range sections {
		b.WriteString("\n")
		b.WriteString(labelStyle.Render("  " + sec.title))
		b.WriteString("\n")
		for _, k := range sec.keys {
			b.WriteString(helpStyle.Render("    "))
			b.WriteString(titleStyle.Render(padRight(k[0], 10)))
			b.WriteString(helpStyle.Render(k[1]))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(dividerStyle.Render("  " + border))
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("  press ? to close"))
	return b.String()
}

func padRight(s string, n int) string {
	if len(s) >= n {
		return s
	}
	return s + strings.Repeat(" ", n-len(s))
}
