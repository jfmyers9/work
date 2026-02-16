package tui

import "github.com/charmbracelet/lipgloss"

var (
	statusStyles = map[string]lipgloss.Style{
		"open":      lipgloss.NewStyle().Foreground(statusColors["open"]),
		"active":    lipgloss.NewStyle().Foreground(statusColors["active"]),
		"review":    lipgloss.NewStyle().Foreground(statusColors["review"]),
		"done":      lipgloss.NewStyle().Foreground(statusColors["done"]),
		"cancelled": lipgloss.NewStyle().Foreground(statusColors["cancelled"]),
	}

	typeStyles = map[string]lipgloss.Style{
		"feature": lipgloss.NewStyle().Foreground(typeColors["feature"]),
		"bug":     lipgloss.NewStyle().Foreground(typeColors["bug"]),
		"chore":   lipgloss.NewStyle().Foreground(typeColors["chore"]),
	}

	priorityStyles = map[int]lipgloss.Style{
		0: lipgloss.NewStyle().Foreground(priorityColors[0]),
		1: lipgloss.NewStyle().Foreground(priorityColors[1]),
		2: lipgloss.NewStyle().Foreground(priorityColors[2]),
		3: lipgloss.NewStyle().Foreground(priorityColors[3]),
	}
)

func styledStatus(s string) string {
	if st, ok := statusStyles[s]; ok {
		return st.Render(s)
	}
	return s
}

func styledType(t string) string {
	if st, ok := typeStyles[t]; ok {
		return st.Render(t)
	}
	return t
}

func styledPriority(p int) string {
	if st, ok := priorityStyles[p]; ok {
		return st.Render(priorityLabel(p))
	}
	return priorityLabel(p)
}

func priorityLabel(p int) string {
	switch p {
	case 1:
		return "P1"
	case 2:
		return "P2"
	case 3:
		return "P3"
	default:
		return "--"
	}
}
