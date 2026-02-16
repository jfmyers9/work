package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/jfmyers9/work/internal/model"
)

// Table cell styles for custom rendering. We bypass the bubbles
// table's renderRow (which uses ANSI-unaware runewidth.Truncate)
// and render cells ourselves with ansi.Truncate.
var (
	listHeaderCellStyle = lipgloss.NewStyle().
		Padding(0, 1).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colorOverlay).
		BorderBottom(true).
		Bold(true).
		Foreground(colorSubtext)

	listCellStyle = lipgloss.NewStyle().Padding(0, 1)

	listSelectedStyle = lipgloss.NewStyle().
				Foreground(colorText).
				Background(colorOverlay).
				Bold(true)
)

type listModel struct {
	table        table.Model
	allIssues    []model.Issue
	filters      filterState
	searching    bool
	search       textinput.Model
	query        string
	width        int
	tableHeight  int
	scrollOffset int
}

func newListModel(issues []model.Issue, width int) listModel {
	si := textinput.New()
	si.Placeholder = "search..."
	si.CharLimit = 128
	si.Width = 40

	m := listModel{allIssues: issues, search: si, width: width, tableHeight: 20}
	m.table = newTable(width)
	m.rebuildRows()
	return m
}

// tableColumns computes column content widths. cellPad accounts for
// Padding(0, 1) on each cell (1 left + 1 right = 2 per column).
func tableColumns(width int) []table.Column {
	const (
		idW     = 8
		statusW = 10
		typeW   = 8
		priW    = 4
		cellPad = 2
		numCols = 5
	)
	fixed := idW + statusW + typeW + priW + numCols*cellPad
	titleW := width - fixed
	if titleW < 20 {
		titleW = 20
	}
	return []table.Column{
		{Title: "ID", Width: idW},
		{Title: "Status", Width: statusW},
		{Title: "Type", Width: typeW},
		{Title: "Pri", Width: priW},
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
		BorderForeground(colorOverlay).
		BorderBottom(true).
		Bold(true).
		Foreground(colorSubtext)
	s.Selected = s.Selected.
		Foreground(colorText).
		Background(colorOverlay).
		Bold(true)
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
	m.clampScroll()
}

func (m *listModel) resize(width, height int) {
	m.width = width
	m.table.SetColumns(tableColumns(width))
	m.tableHeight = height - 5
	m.table.SetHeight(m.tableHeight)
	m.clampScroll()
}

func (m *listModel) clampScroll() {
	cursor := m.table.Cursor()
	if cursor < m.scrollOffset {
		m.scrollOffset = cursor
	}
	if m.tableHeight > 0 && cursor >= m.scrollOffset+m.tableHeight {
		m.scrollOffset = cursor - m.tableHeight + 1
	}
	maxOffset := len(m.table.Rows()) - m.tableHeight
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.scrollOffset > maxOffset {
		m.scrollOffset = maxOffset
	}
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
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
	m.clampScroll()
	return m, cmd
}

// renderTable builds table output with ANSI-aware truncation.
// Styling is applied AFTER truncation so ANSI escape codes don't
// consume the column width budget.
func (m listModel) renderTable() string {
	rows := m.table.Rows()
	cols := tableColumns(m.width)
	cursor := m.table.Cursor()

	hCells := make([]string, len(cols))
	for i, col := range cols {
		truncated := ansi.Truncate(col.Title, col.Width, "…")
		inner := lipgloss.NewStyle().
			Width(col.Width).MaxWidth(col.Width).Inline(true).
			Render(truncated)
		hCells[i] = listHeaderCellStyle.Render(inner)
	}
	header := lipgloss.JoinHorizontal(lipgloss.Top, hCells...)

	start := m.scrollOffset
	end := start + m.tableHeight
	if end > len(rows) {
		end = len(rows)
	}

	dataLines := make([]string, 0, end-start)
	for r := start; r < end; r++ {
		cells := make([]string, len(cols))
		for i, value := range rows[r] {
			styled := styleCellValue(i, value)
			truncated := ansi.Truncate(styled, cols[i].Width, "…")
			inner := lipgloss.NewStyle().
				Width(cols[i].Width).MaxWidth(cols[i].Width).Inline(true).
				Render(truncated)
			cells[i] = listCellStyle.Render(inner)
		}
		row := lipgloss.JoinHorizontal(lipgloss.Top, cells...)
		if r == cursor {
			row = listSelectedStyle.Render(row)
		}
		dataLines = append(dataLines, row)
	}

	return header + "\n" + strings.Join(dataLines, "\n")
}

func styleCellValue(col int, value string) string {
	switch col {
	case 1:
		return styledStatus(value)
	case 2:
		return styledType(value)
	case 3:
		return styledPriority(priorityFromLabel(value))
	default:
		return value
	}
}

func (m listModel) View() string {
	rows := m.table.Rows()

	filterBar := m.filters.view()
	if m.searching {
		filterBar += "  " + m.search.View()
	} else if m.query != "" {
		filterBar += "  " + filterTagStyle.Render("search:"+m.query)
	}

	count := helpStyle.Render(fmt.Sprintf("%d issues", len(rows)))
	filterBar += "  " + count

	if len(rows) == 0 {
		empty := lipgloss.NewStyle().Foreground(colorMuted).Italic(true).Render("No issues found.")
		return filterBar + "\n\n  " + empty + "\n"
	}

	return filterBar + "\n" + m.renderTable()
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
