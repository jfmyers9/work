package tui

import "github.com/charmbracelet/lipgloss"

var (
	statusStyles = map[string]lipgloss.Style{
		"open":      lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
		"active":    lipgloss.NewStyle().Foreground(lipgloss.Color("33")),
		"review":    lipgloss.NewStyle().Foreground(lipgloss.Color("214")),
		"done":      lipgloss.NewStyle().Foreground(lipgloss.Color("34")),
		"cancelled": lipgloss.NewStyle().Foreground(lipgloss.Color("196")),
	}

	typeStyles = map[string]lipgloss.Style{
		"feature": lipgloss.NewStyle().Foreground(lipgloss.Color("117")),
		"bug":     lipgloss.NewStyle().Foreground(lipgloss.Color("203")),
		"chore":   lipgloss.NewStyle().Foreground(lipgloss.Color("248")),
	}

	priorityStyles = map[int]lipgloss.Style{
		0: lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
		1: lipgloss.NewStyle().Foreground(lipgloss.Color("196")),
		2: lipgloss.NewStyle().Foreground(lipgloss.Color("214")),
		3: lipgloss.NewStyle().Foreground(lipgloss.Color("33")),
	}

	titleStyle = lipgloss.NewStyle().Bold(true).Padding(0, 1)
	helpStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
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
