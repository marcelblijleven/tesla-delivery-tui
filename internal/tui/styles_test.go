package tui

import (
	"testing"
)

func TestGetStatusBadgeStyle(t *testing.T) {
	tests := []struct {
		name   string
		status string
	}{
		{"booked status", "BOOKED"},
		{"booked lowercase", "booked"},
		{"book partial", "order_book"},
		{"progress status", "IN_PROGRESS"},
		{"pending status", "PENDING"},
		{"processing status", "PROCESSING"},
		{"delivered status", "DELIVERED"},
		{"complete status", "COMPLETE"},
		{"cancelled status", "CANCELLED"},
		{"cancel partial", "order_cancel"},
		{"unknown status", "UNKNOWN_STATUS"},
		{"empty status", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style := GetStatusBadgeStyle(tt.status)
			// Just verify it returns a style without panicking
			_ = style.Render("test")
		})
	}
}

func TestContainsAny(t *testing.T) {
	tests := []struct {
		name    string
		s       string
		substrs []string
		want    bool
	}{
		{"contains first", "hello world", []string{"hello", "foo"}, true},
		{"contains second", "hello world", []string{"foo", "world"}, true},
		{"contains none", "hello world", []string{"foo", "bar"}, false},
		{"empty string", "", []string{"hello"}, false},
		{"empty substrs", "hello", []string{}, false},
		{"case insensitive", "HELLO", []string{"hello"}, true},
		{"partial match", "processing", []string{"process"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := containsAny(tt.s, tt.substrs...); got != tt.want {
				t.Errorf("containsAny() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToLower(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{"lowercase", "hello", "hello"},
		{"uppercase", "HELLO", "hello"},
		{"mixed", "HeLLo WoRLd", "hello world"},
		{"empty", "", ""},
		{"numbers", "ABC123", "abc123"},
		{"special chars", "Hello-World_Test", "hello-world_test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toLower(tt.s); got != tt.want {
				t.Errorf("toLower() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		substr string
		want   bool
	}{
		{"contains at start", "hello world", "hello", true},
		{"contains at end", "hello world", "world", true},
		{"contains in middle", "hello world", "lo wo", true},
		{"does not contain", "hello world", "foo", false},
		{"empty string", "", "hello", false},
		{"empty substr", "hello", "", true},
		{"exact match", "hello", "hello", true},
		{"longer substr", "hi", "hello", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := contains(tt.s, tt.substr); got != tt.want {
				t.Errorf("contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStylesCanRender(t *testing.T) {
	// Test that all styles can render without panicking
	// lipgloss.Style.Render always returns a string (may be empty for empty input)

	t.Run("AppStyle", func(t *testing.T) {
		_ = AppStyle.Render("test")
	})
	t.Run("TitleStyle", func(t *testing.T) {
		_ = TitleStyle.Render("test")
	})
	t.Run("SubtitleStyle", func(t *testing.T) {
		_ = SubtitleStyle.Render("test")
	})
	t.Run("HeaderStyle", func(t *testing.T) {
		_ = HeaderStyle.Render("test")
	})
	t.Run("StatusBadgeBase", func(t *testing.T) {
		_ = StatusBadgeBase.Render("test")
	})
	t.Run("StatusBooked", func(t *testing.T) {
		_ = StatusBooked.Render("test")
	})
	t.Run("StatusInProgress", func(t *testing.T) {
		_ = StatusInProgress.Render("test")
	})
	t.Run("StatusDelivered", func(t *testing.T) {
		_ = StatusDelivered.Render("test")
	})
	t.Run("StatusCancelled", func(t *testing.T) {
		_ = StatusCancelled.Render("test")
	})
	t.Run("TableHeaderStyle", func(t *testing.T) {
		_ = TableHeaderStyle.Render("test")
	})
	t.Run("TableRowStyle", func(t *testing.T) {
		_ = TableRowStyle.Render("test")
	})
	t.Run("TableSelectedStyle", func(t *testing.T) {
		_ = TableSelectedStyle.Render("test")
	})
	t.Run("TabStyle", func(t *testing.T) {
		_ = TabStyle.Render("test")
	})
	t.Run("ActiveTabStyle", func(t *testing.T) {
		_ = ActiveTabStyle.Render("test")
	})
	t.Run("TabBarStyle", func(t *testing.T) {
		_ = TabBarStyle.Render("test")
	})
	t.Run("LabelStyle", func(t *testing.T) {
		_ = LabelStyle.Render("test")
	})
	t.Run("SubheadingStyle", func(t *testing.T) {
		_ = SubheadingStyle.Render("test")
	})
	t.Run("ValueStyle", func(t *testing.T) {
		_ = ValueStyle.Render("test")
	})
	t.Run("ChangedValueStyle", func(t *testing.T) {
		_ = ChangedValueStyle.Render("test")
	})
	t.Run("OldValueStyle", func(t *testing.T) {
		_ = OldValueStyle.Render("test")
	})
	t.Run("HelpStyle", func(t *testing.T) {
		_ = HelpStyle.Render("test")
	})
	t.Run("ErrorStyle", func(t *testing.T) {
		_ = ErrorStyle.Render("test")
	})
	t.Run("SuccessStyle", func(t *testing.T) {
		_ = SuccessStyle.Render("test")
	})
	t.Run("CardStyle", func(t *testing.T) {
		_ = CardStyle.Render("test")
	})
	t.Run("SpinnerStyle", func(t *testing.T) {
		_ = SpinnerStyle.Render("test")
	})
	t.Run("TaskCompleteStyle", func(t *testing.T) {
		_ = TaskCompleteStyle.Render("test")
	})
	t.Run("TaskIncompleteStyle", func(t *testing.T) {
		_ = TaskIncompleteStyle.Render("test")
	})
	t.Run("JSONKeyStyle", func(t *testing.T) {
		_ = JSONKeyStyle.Render("test")
	})
	t.Run("JSONStringStyle", func(t *testing.T) {
		_ = JSONStringStyle.Render("test")
	})
	t.Run("JSONNumberStyle", func(t *testing.T) {
		_ = JSONNumberStyle.Render("test")
	})
	t.Run("JSONBoolStyle", func(t *testing.T) {
		_ = JSONBoolStyle.Render("test")
	})
	t.Run("DiffAddedStyle", func(t *testing.T) {
		_ = DiffAddedStyle.Render("test")
	})
	t.Run("DiffRemovedStyle", func(t *testing.T) {
		_ = DiffRemovedStyle.Render("test")
	})
}

func TestColorsNotEmpty(t *testing.T) {
	colors := []struct {
		name  string
		color interface{}
	}{
		{"TeslaRed", TeslaRed},
		{"TeslaGray", TeslaGray},
		{"TeslaWhite", TeslaWhite},
		{"StatusBlue", StatusBlue},
		{"StatusYellow", StatusYellow},
		{"StatusGreen", StatusGreen},
		{"StatusRed", StatusRed},
		{"Muted", Muted},
		{"Highlight", Highlight},
	}

	for _, tt := range colors {
		t.Run(tt.name, func(t *testing.T) {
			if tt.color == nil {
				t.Errorf("%s is nil", tt.name)
			}
		})
	}
}
