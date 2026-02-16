package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type issueCreatedMsg struct {
	id    string
	title string
}

const (
	fieldTitle = iota
	fieldType
	fieldPriority
	fieldLabels
	fieldParent
	fieldDescription
	fieldCount
)

var fieldNames = [fieldCount]string{
	"Title", "Type", "Priority", "Labels", "Parent", "Description",
}

type createModel struct {
	inputs [fieldCount]textinput.Model
	focus  int
}

var (
	focusedFieldStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("229"))
	blurredFieldStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

func newCreateModel() createModel {
	var inputs [fieldCount]textinput.Model
	for i := range inputs {
		t := textinput.New()
		t.CharLimit = 256
		t.Width = 60
		inputs[i] = t
	}

	inputs[fieldTitle].Placeholder = "Issue title"
	inputs[fieldType].Placeholder = "feature|bug|chore"
	inputs[fieldType].SetValue("feature")
	inputs[fieldPriority].Placeholder = "1, 2, or 3"
	inputs[fieldPriority].SetValue("2")
	inputs[fieldLabels].Placeholder = "comma-separated labels"
	inputs[fieldParent].Placeholder = "parent issue ID (optional)"
	inputs[fieldDescription].Placeholder = "Short description"

	inputs[fieldTitle].Focus()

	return createModel{inputs: inputs}
}

func (m createModel) Update(msg tea.Msg) (createModel, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "tab", "down":
			m.inputs[m.focus].Blur()
			m.focus = (m.focus + 1) % fieldCount
			return m, m.inputs[m.focus].Focus()
		case "shift+tab", "up":
			m.inputs[m.focus].Blur()
			m.focus = (m.focus - 1 + fieldCount) % fieldCount
			return m, m.inputs[m.focus].Focus()
		}
	}

	var cmd tea.Cmd
	m.inputs[m.focus], cmd = m.inputs[m.focus].Update(msg)
	return m, cmd
}

func (m createModel) View() string {
	var b strings.Builder
	b.WriteString("\n")
	for i, input := range m.inputs {
		style := blurredFieldStyle
		if i == m.focus {
			style = focusedFieldStyle
		}
		b.WriteString(fmt.Sprintf("  %s  %s\n", style.Render(fmt.Sprintf("%-12s", fieldNames[i]+":")), input.View()))
	}
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("  tab/shift-tab: navigate • ctrl+d: submit • esc: cancel"))
	return b.String()
}

func (m createModel) title() string       { return strings.TrimSpace(m.inputs[fieldTitle].Value()) }
func (m createModel) issueType() string    { return strings.TrimSpace(m.inputs[fieldType].Value()) }
func (m createModel) description() string  { return strings.TrimSpace(m.inputs[fieldDescription].Value()) }
func (m createModel) parentID() string     { return strings.TrimSpace(m.inputs[fieldParent].Value()) }

func (m createModel) priority() int {
	p, err := strconv.Atoi(strings.TrimSpace(m.inputs[fieldPriority].Value()))
	if err != nil || p < 1 || p > 3 {
		return 2
	}
	return p
}

func (m createModel) labels() []string {
	raw := strings.TrimSpace(m.inputs[fieldLabels].Value())
	if raw == "" {
		return nil
	}
	var out []string
	for _, l := range strings.Split(raw, ",") {
		l = strings.TrimSpace(l)
		if l != "" {
			out = append(out, l)
		}
	}
	return out
}
