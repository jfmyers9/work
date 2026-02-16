package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/jfmyers9/work/internal/model"
	"github.com/jfmyers9/work/internal/tracker"
)

var (
	statuses = []string{"", "open", "active", "review", "done", "cancelled"}
	types    = []string{"", "feature", "bug", "chore"}
	sorts    = []string{"priority", "created", "updated", "title"}

	filterTagStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Background(lipgloss.Color("57")).
			Foreground(lipgloss.Color("229"))
	filterLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("241"))
)

type filterState struct {
	statusIdx int
	typeIdx   int
	sortIdx   int
}

func (f *filterState) cycleStatus() {
	f.statusIdx = (f.statusIdx + 1) % len(statuses)
}

func (f *filterState) cycleType() {
	f.typeIdx = (f.typeIdx + 1) % len(types)
}

func (f *filterState) cycleSort() {
	f.sortIdx = (f.sortIdx + 1) % len(sorts)
}

func (f *filterState) clear() {
	f.statusIdx = 0
	f.typeIdx = 0
}

func (f filterState) apply(issues []model.Issue) []model.Issue {
	filtered := tracker.FilterIssues(issues, tracker.FilterOptions{
		Status: statuses[f.statusIdx],
		Type:   types[f.typeIdx],
	})
	tracker.SortIssues(filtered, sorts[f.sortIdx])
	return filtered
}

func (f filterState) view() string {
	var parts []string

	if statuses[f.statusIdx] != "" {
		parts = append(parts, filterTagStyle.Render(statuses[f.statusIdx]))
	}
	if types[f.typeIdx] != "" {
		parts = append(parts, filterTagStyle.Render(types[f.typeIdx]))
	}

	sort := fmt.Sprintf("sort:%s", sorts[f.sortIdx])
	parts = append(parts, filterLabelStyle.Render(sort))

	return "  " + strings.Join(parts, " ")
}
