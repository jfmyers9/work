package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jfmyers9/work/internal/model"
)

type listModel struct {
	table     table.Model
	allIssues []model.Issue
	filters   filterState
	searching bool
	search    textinput.Model
	query     string
	width     int
}

func newListModel(issues []model.Issue, width int) listModel {
	si := textinput.New()
	si.Placeholder = "search..."
	si.CharLimit = 128
	si.Width = 40

	m := listModel{allIssues: issues, search: si, width: width}
	m.table = newTable(width)
	m.rebuildRows()
	return m
}

func tableColumns(width int) []table.Column {
	const fixed = 8 + 10 + 8 + 4 + 6 // ID + Status + Type + Pri + padding
	titleW := width - fixed
	if titleW < 20 {
		titleW = 20
	}
	return []table.Column{
		{Title: "ID", Width: 8},
		{Title: "Status", Width: 10},
		{Title: "Type", Width: 8},
		{Title: "Pri", Width: 4},
		{Title: "Title", Width: titleW},
	}
}

func newTable(width int) table.Model {
	columns := tableColumns(width)

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(nil),
		table.WithFocused(true),
		table.WithHeight(20),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)
	return t
}

func (m *listModel) rebuildRows() {
	visible := m.filters.apply(m.allIssues)
	if m.query != "" {
		q := strings.ToLower(m.query)
		var filtered []model.Issue
		for _, issue := range visible {
			if strings.Contains(strings.ToLower(issue.Title), q) {
				filtered = append(filtered, issue)
			}
		}
		visible = filtered
	}
	rows := make([]table.Row, len(visible))
	for i, issue := range visible {
		rows[i] = table.Row{
			issue.ID[:min(6, len(issue.ID))],
			issue.Status,
			issue.Type,
			priorityLabel(issue.Priority),
			issue.Title,
		}
	}
	m.table.SetRows(rows)
}

func (m *listModel) resize(width, height int) {
	m.width = width
	m.table.SetColumns(tableColumns(width))
	m.table.SetHeight(height - 5)
}

func (m listModel) Update(msg tea.Msg) (listModel, tea.Cmd) {
	if kmsg, ok := msg.(tea.KeyMsg); ok {
		if m.searching {
			switch kmsg.String() {
			case "esc":
				m.searching = false
				m.query = ""
				m.search.SetValue("")
				m.search.Blur()
				m.rebuildRows()
				return m, nil
			case "enter":
				m.searching = false
				m.search.Blur()
				return m, nil
			default:
				var cmd tea.Cmd
				m.search, cmd = m.search.Update(msg)
				m.query = m.search.Value()
				m.rebuildRows()
				return m, cmd
			}
		}

		switch kmsg.String() {
		case "/":
			m.searching = true
			m.search.SetValue("")
			m.query = ""
			return m, m.search.Focus()
		case "f":
			m.filters.cycleStatus()
			m.rebuildRows()
			return m, nil
		case "t":
			m.filters.cycleType()
			m.rebuildRows()
			return m, nil
		case "o":
			m.filters.cycleSort()
			m.rebuildRows()
			return m, nil
		case "F":
			m.filters.clear()
			m.query = ""
			m.search.SetValue("")
			m.rebuildRows()
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m listModel) View() string {
	filterBar := m.filters.view()
	if m.searching {
		filterBar += "  " + m.search.View()
	} else if m.query != "" {
		filterBar += "  " + filterTagStyle.Render("search:"+m.query)
	}

	rows := m.table.Rows()
	if len(rows) == 0 {
		return filterBar + "\n\n  No issues found.\n"
	}

	styled := make([]table.Row, len(rows))
	for i, row := range rows {
		styled[i] = table.Row{
			row[0],
			styledStatus(row[1]),
			styledType(row[2]),
			styledPriority(priorityFromLabel(row[3])),
			row[4],
		}
	}
	m.table.SetRows(styled)
	out := m.table.View()
	m.table.SetRows(rows)

	return filterBar + "\n" + out
}

func priorityFromLabel(s string) int {
	switch s {
	case "P1":
		return 1
	case "P2":
		return 2
	case "P3":
		return 3
	default:
		return 0
	}
}
