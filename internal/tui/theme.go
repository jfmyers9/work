package tui

import "github.com/charmbracelet/lipgloss"

// Palette — muted, cohesive colors for dark terminals.
var (
	colorBase    = lipgloss.Color("#1e1e2e") // deep background
	colorSurface = lipgloss.Color("#313244") // raised surface
	colorOverlay = lipgloss.Color("#45475a") // overlay bg

	colorText    = lipgloss.Color("#cdd6f4") // primary text
	colorSubtext = lipgloss.Color("#a6adc8") // secondary text
	colorMuted   = lipgloss.Color("#6c7086") // dim text

	colorAccent    = lipgloss.Color("#89b4fa") // primary accent (blue)
	colorAccentDim = lipgloss.Color("#74c7ec") // secondary accent (teal)
	colorGreen     = lipgloss.Color("#a6e3a1")
	colorYellow    = lipgloss.Color("#f9e2af")
	colorOrange    = lipgloss.Color("#fab387")
	colorRed       = lipgloss.Color("#f38ba8")
	colorLavender = lipgloss.Color("#b4befe")
)

// Status colors
var statusColors = map[string]lipgloss.Color{
	"open":      colorMuted,
	"active":    colorAccent,
	"review":    colorYellow,
	"done":      colorGreen,
	"cancelled": colorRed,
}

// Type colors
var typeColors = map[string]lipgloss.Color{
	"feature": colorAccentDim,
	"bug":     colorOrange,
	"chore":   colorSubtext,
}

// Priority colors
var priorityColors = map[int]lipgloss.Color{
	0: colorMuted,
	1: colorRed,
	2: colorYellow,
	3: colorAccent,
}

// Shared styles
var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorBase).
			Background(colorAccent).
			Padding(0, 2)

	headerDimStyle = lipgloss.NewStyle().
			Foreground(colorOverlay).
			Background(colorAccent)

	footerBarStyle = lipgloss.NewStyle().
			Foreground(colorSubtext).
			Background(colorSurface).
			Padding(0, 2)

	statusMsgStyle = lipgloss.NewStyle().
			Foreground(colorGreen).
			Background(colorSurface)

	helpStyle = lipgloss.NewStyle().Foreground(colorMuted)

	labelStyle = lipgloss.NewStyle().Bold(true).Foreground(colorSubtext)

	valueStyle = lipgloss.NewStyle().Foreground(colorText)

	commentMetaStyle = lipgloss.NewStyle().Foreground(colorMuted).Italic(true)

	dividerStyle = lipgloss.NewStyle().Foreground(colorOverlay)

	sectionStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorLavender)

	// Tags/pills
	filterTagStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Background(colorOverlay).
			Foreground(colorAccent).
			Bold(true)

	filterLabelStyle = lipgloss.NewStyle().
				Foreground(colorMuted)

	// Overlays
	overlayStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorAccent).
			Padding(1, 2)

	// Key hint in help
	keyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorAccent)

	descStyle = lipgloss.NewStyle().
			Foreground(colorSubtext)

	// Focused/blurred fields
	focusedFieldStyle = lipgloss.NewStyle().Foreground(colorAccent)
	blurredFieldStyle = lipgloss.NewStyle().Foreground(colorMuted)

	// Picker
	pickerItemStyle = lipgloss.NewStyle().Padding(0, 2)
	pickerSelectedStyle = lipgloss.NewStyle().
				Padding(0, 2).
				Bold(true).
				Foreground(colorBase).
				Background(colorAccent)
)
