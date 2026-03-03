package ui

import "github.com/charmbracelet/bubbles/key"

// GlobalKeyMap defines keys available in all views.
type GlobalKeyMap struct {
	Quit    key.Binding
	Tab1    key.Binding
	Tab2    key.Binding
	Tab3    key.Binding
	Help    key.Binding
	Back    key.Binding
}

var GlobalKeys = GlobalKeyMap{
	Quit:    key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit")),
	Tab1:    key.NewBinding(key.WithKeys("1"), key.WithHelp("1", "sessions")),
	Tab2:    key.NewBinding(key.WithKeys("2"), key.WithHelp("2", "search")),
	Tab3:    key.NewBinding(key.WithKeys("3"), key.WithHelp("3", "stats")),
	Help:    key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	Back:    key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
}

// ListKeyMap defines keys for the session list view.
type ListKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Enter  key.Binding
	Filter key.Binding
}

var ListKeys = ListKeyMap{
	Up:     key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
	Down:   key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
	Enter:  key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "open")),
	Filter: key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter")),
}

// DetailKeyMap defines keys for the session detail view.
type DetailKeyMap struct {
	Up       key.Binding
	Down     key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Toggle   key.Binding
	Export   key.Binding
	Home     key.Binding
	End      key.Binding
}

var DetailKeys = DetailKeyMap{
	Up:       key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
	Down:     key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
	PageUp:   key.NewBinding(key.WithKeys("pgup", "ctrl+u"), key.WithHelp("pgup", "page up")),
	PageDown: key.NewBinding(key.WithKeys("pgdown", "ctrl+d"), key.WithHelp("pgdn", "page dn")),
	Toggle:   key.NewBinding(key.WithKeys("tab", "t"), key.WithHelp("tab/t", "toggle tool")),
	Export:   key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "export")),
	Home:     key.NewBinding(key.WithKeys("home", "g"), key.WithHelp("home", "top")),
	End:      key.NewBinding(key.WithKeys("end", "G"), key.WithHelp("end", "bottom")),
}

// SearchKeyMap defines keys for the search view.
type SearchKeyMap struct {
	Execute key.Binding
	Clear   key.Binding
}

var SearchKeys = SearchKeyMap{
	Execute: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "search")),
	Clear:   key.NewBinding(key.WithKeys("ctrl+u"), key.WithHelp("ctrl+u", "clear")),
}

// StatsKeyMap defines keys for the stats view.
type StatsKeyMap struct {
	Daily   key.Binding
	Weekly  key.Binding
	Monthly key.Binding
}

var StatsKeys = StatsKeyMap{
	Daily:   key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "daily")),
	Weekly:  key.NewBinding(key.WithKeys("w"), key.WithHelp("w", "weekly")),
	Monthly: key.NewBinding(key.WithKeys("m"), key.WithHelp("m", "monthly")),
}
