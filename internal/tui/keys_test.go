package tui

import (
	"strings"
	"testing"
)

func TestDefaultKeyMap(t *testing.T) {
	km := DefaultKeyMap

	// Test that all bindings have keys defined
	bindings := []struct {
		name    string
		binding interface{ Keys() []string }
	}{
		{"Up", km.Up},
		{"Down", km.Down},
		{"Left", km.Left},
		{"Right", km.Right},
		{"Enter", km.Enter},
		{"Back", km.Back},
		{"Tab", km.Tab},
		{"ShiftTab", km.ShiftTab},
		{"Refresh", km.Refresh},
		{"Logout", km.Logout},
		{"Help", km.Help},
		{"Quit", km.Quit},
	}

	for _, b := range bindings {
		t.Run(b.name, func(t *testing.T) {
			keys := b.binding.Keys()
			if len(keys) == 0 {
				t.Errorf("%s has no keys bound", b.name)
			}
		})
	}
}

func TestDefaultKeyMap_SpecificKeys(t *testing.T) {
	km := DefaultKeyMap

	tests := []struct {
		name         string
		binding      interface{ Keys() []string }
		expectedKeys []string
	}{
		{"Up", km.Up, []string{"up", "k"}},
		{"Down", km.Down, []string{"down", "j"}},
		{"Left", km.Left, []string{"left", "h"}},
		{"Right", km.Right, []string{"right", "l"}},
		{"Enter", km.Enter, []string{"enter"}},
		{"Back", km.Back, []string{"esc", "backspace"}},
		{"Tab", km.Tab, []string{"tab"}},
		{"ShiftTab", km.ShiftTab, []string{"shift+tab"}},
		{"Refresh", km.Refresh, []string{"r"}},
		{"Logout", km.Logout, []string{"L"}},
		{"Help", km.Help, []string{"?"}},
		{"Quit", km.Quit, []string{"q", "ctrl+c"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keys := tt.binding.Keys()
			for _, expected := range tt.expectedKeys {
				found := false
				for _, k := range keys {
					if k == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("%s missing expected key %q", tt.name, expected)
				}
			}
		})
	}
}

func TestKeyMap_ShortHelp(t *testing.T) {
	km := DefaultKeyMap
	help := km.ShortHelp()

	if len(help) == 0 {
		t.Error("ShortHelp() returned empty slice")
	}

	// Should contain Help and Quit at minimum
	if len(help) < 2 {
		t.Error("ShortHelp() should contain at least 2 bindings")
	}
}

func TestKeyMap_FullHelp(t *testing.T) {
	km := DefaultKeyMap
	help := km.FullHelp()

	if len(help) == 0 {
		t.Error("FullHelp() returned empty slice")
	}

	// Should have multiple groups
	if len(help) < 3 {
		t.Errorf("FullHelp() returned %d groups, want at least 3", len(help))
	}

	// Each group should have bindings
	for i, group := range help {
		if len(group) == 0 {
			t.Errorf("FullHelp() group %d is empty", i)
		}
	}
}

func TestLoginKeys(t *testing.T) {
	keys := LoginKeys()

	if keys == "" {
		t.Error("LoginKeys() returned empty string")
	}

	// Should contain relevant keys
	expectedParts := []string{"enter", "quit"}
	for _, part := range expectedParts {
		if !strings.Contains(strings.ToLower(keys), part) {
			t.Errorf("LoginKeys() missing %q", part)
		}
	}
}

func TestOrdersKeys(t *testing.T) {
	keys := OrdersKeys()

	if keys == "" {
		t.Error("OrdersKeys() returned empty string")
	}

	// Should contain relevant keys
	expectedParts := []string{"navigate", "enter", "refresh", "logout", "quit"}
	for _, part := range expectedParts {
		if !strings.Contains(strings.ToLower(keys), part) {
			t.Errorf("OrdersKeys() missing %q", part)
		}
	}
}

func TestDetailKeys(t *testing.T) {
	keys := DetailKeys(TabDetails)

	if keys == "" {
		t.Error("DetailKeys() returned empty string")
	}

	// Should contain relevant keys
	expectedParts := []string{"tab", "scroll", "back", "refresh", "quit", "copy vin"}
	for _, part := range expectedParts {
		if !strings.Contains(strings.ToLower(keys), part) {
			t.Errorf("DetailKeys(TabDetails) missing %q", part)
		}
	}

	// JSON tab should say "copy JSON"
	jsonKeys := DetailKeys(TabJSON)
	if !strings.Contains(strings.ToLower(jsonKeys), "copy json") {
		t.Error("DetailKeys(TabJSON) should contain 'copy JSON'")
	}
}

func TestVimKeybindings(t *testing.T) {
	km := DefaultKeyMap

	// h/j/k/l should be bound for vim users
	vimBindings := map[string]interface{ Keys() []string }{
		"h": km.Left,
		"j": km.Down,
		"k": km.Up,
		"l": km.Right,
	}

	for key, binding := range vimBindings {
		t.Run("vim key "+key, func(t *testing.T) {
			found := false
			for _, k := range binding.Keys() {
				if k == key {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Vim key %q not bound", key)
			}
		})
	}
}

func TestArrowKeybindings(t *testing.T) {
	km := DefaultKeyMap

	// Arrow keys should be bound for non-vim users
	arrowBindings := map[string]interface{ Keys() []string }{
		"up":    km.Up,
		"down":  km.Down,
		"left":  km.Left,
		"right": km.Right,
	}

	for key, binding := range arrowBindings {
		t.Run("arrow key "+key, func(t *testing.T) {
			found := false
			for _, k := range binding.Keys() {
				if k == key {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Arrow key %q not bound", key)
			}
		})
	}
}
