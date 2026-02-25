package tui

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Up      key.Binding
	Down    key.Binding
	Enter   key.Binding
	Back    key.Binding
	Clean   key.Binding
	Archive key.Binding
	Delete  key.Binding
	Docker  key.Binding
	Tab     key.Binding
	Help    key.Binding
	Quit    key.Binding
}

var Keys = KeyMap{
	Up:      key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("up/k", "up")),
	Down:    key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("down/j", "down")),
	Enter:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
	Back:    key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
	Clean:   key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "clean")),
	Archive: key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "archive")),
	Delete:  key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
	Docker:  key.NewBinding(key.WithKeys("D"), key.WithHelp("D", "docker-stop")),
	Tab:     key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next category")),
	Help:    key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	Quit:    key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
}
