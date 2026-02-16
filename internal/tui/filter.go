package tui

import (
	"fmt"
	"strings"

	"github.com/jfmyers9/work/internal/model"
	"github.com/jfmyers9/work/internal/tracker"
)

var (
	statuses = []string{"", "open", "active", "review", "done", "cancelled"}
	types    = []string{"", "feature", "bug", "chore"}
	sorts    = []string{"priority", "created", "updated", "title"}
)

type filterState struct {
	statusIdx  int
	typeIdx    int
	sortIdx    int
	showClosed bool
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
	f.showClosed = false
}

func (f filterState) apply(issues []model.Issue) []model.Issue {
	opts := tracker.FilterOptions{
		Status: statuses[f.statusIdx],
		Type:   types[f.typeIdx],
	}
	if !f.showClosed && f.statusIdx == 0 {
		opts.ExcludeStatuses = []string{"done", "cancelled"}
	}
	filtered := tracker.FilterIssues(issues, opts)
	tracker.SortIssues(filtered, sorts[f.sortIdx])
	return filtered
}

func (f filterState) view() string {
	var parts []string

	if f.statusIdx == 0 && !f.showClosed {
		parts = append(parts, filterTagStyle.Render("open"))
	} else if statuses[f.statusIdx] != "" {
		parts = append(parts, filterTagStyle.Render(statuses[f.statusIdx]))
	} else if f.showClosed {
		parts = append(parts, filterTagStyle.Render("all"))
	}
	if types[f.typeIdx] != "" {
		parts = append(parts, filterTagStyle.Render(types[f.typeIdx]))
	}

	sort := fmt.Sprintf("sort:%s", sorts[f.sortIdx])
	parts = append(parts, filterLabelStyle.Render(sort))

	return "  " + strings.Join(parts, " ")
}
