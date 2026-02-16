package tui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Up           key.Binding
	Down         key.Binding
	Enter        key.Binding
	Back         key.Binding
	Quit         key.Binding
	Status       key.Binding
	Comment      key.Binding
	Create       key.Binding
	Edit         key.Binding
	Link         key.Binding
	Unlink       key.Binding
	QuickStart   key.Binding
	QuickDone    key.Binding
	QuickReview  key.Binding
	QuickCancel  key.Binding
	FilterStatus key.Binding
	FilterType   key.Binding
	FilterSort   key.Binding
	FilterClear  key.Binding
	Help         key.Binding
	Search       key.Binding
	History      key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("k", "up"),
		key.WithHelp("k/↑", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("j", "down"),
		key.WithHelp("j/↓", "down"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "open"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Status: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "status"),
	),
	Comment: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "comment"),
	),
	Create: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "new issue"),
	),
	Edit: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit in $EDITOR"),
	),
	Link: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "link parent"),
	),
	Unlink: key.NewBinding(
		key.WithKeys("P"),
		key.WithHelp("P", "unlink parent"),
	),
	QuickStart: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "active"),
	),
	QuickDone: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "done"),
	),
	QuickReview: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "review"),
	),
	QuickCancel: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "cancel"),
	),
	FilterStatus: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "cycle status"),
	),
	FilterType: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "cycle type"),
	),
	FilterSort: key.NewBinding(
		key.WithKeys("o"),
		key.WithHelp("o", "cycle sort"),
	),
	FilterClear: key.NewBinding(
		key.WithKeys("F"),
		key.WithHelp("F", "clear filters"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	Search: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "search"),
	),
	History: key.NewBinding(
		key.WithKeys("h"),
		key.WithHelp("h", "history"),
	),
}
