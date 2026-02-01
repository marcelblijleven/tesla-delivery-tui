package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
)

// KeyMap contains all key bindings
type KeyMap struct {
	Up       key.Binding
	Down     key.Binding
	Left     key.Binding
	Right    key.Binding
	Enter    key.Binding
	Back     key.Binding
	Tab      key.Binding
	ShiftTab key.Binding
	Refresh  key.Binding
	Logout   key.Binding
	Help     key.Binding
	Quit     key.Binding
	Copy     key.Binding
}

// DefaultKeyMap returns the default key bindings
var DefaultKeyMap = KeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "left"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "right"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc", "backspace"),
		key.WithHelp("esc", "back"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next tab"),
	),
	ShiftTab: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "prev tab"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	),
	Logout: key.NewBinding(
		key.WithKeys("L"),
		key.WithHelp("L", "logout"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Copy: key.NewBinding(
		key.WithKeys("y", "c"),
		key.WithHelp("y/c", "copy"),
	),
}

// ShortHelp returns keybindings to be shown in the mini help view
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

// FullHelp returns keybindings for the expanded help view
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.Enter, k.Back, k.Tab, k.ShiftTab},
		{k.Refresh, k.Copy, k.Logout, k.Quit},
	}
}

// LoginKeys returns the help text for login view
func LoginKeys() string {
	return "enter: login • q: quit"
}

// OrdersKeys returns the help text for orders view
func OrdersKeys() string {
	return "↑/↓: navigate • enter: details • y: copy VIN • r: refresh • L: logout • ?: help • q: quit"
}

// DetailKeys returns the help text for detail view, with copy target based on active tab
func DetailKeys(tab Tab) string {
	copyTarget := "VIN"
	if tab == TabJSON {
		copyTarget = "JSON"
	}
	return fmt.Sprintf("tab: tabs • ↑/↓: scroll • y: copy %s • esc: back • r: refresh • ?: help • q: quit", copyTarget)
}
