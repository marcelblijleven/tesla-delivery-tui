package tui

import "github.com/charmbracelet/lipgloss"

// Colors
var (
	TeslaRed     = lipgloss.Color("#E31937")
	TeslaGray    = lipgloss.Color("#393C41")
	TeslaWhite   = lipgloss.Color("#FFFFFF")
	StatusBlue   = lipgloss.Color("#3B82F6")
	StatusYellow = lipgloss.Color("#EAB308")
	StatusGreen  = lipgloss.Color("#22C55E")
	StatusRed    = lipgloss.Color("#EF4444")
	Muted        = lipgloss.Color("#9CA3AF")
	Highlight    = lipgloss.Color("#FBBF24")
	SubtleBg     = lipgloss.Color("#1A1A2E")
)

// Styles
var (
	// App
	AppStyle = lipgloss.NewStyle().
			Padding(1, 2)

	// Title
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(TeslaRed).
			MarginBottom(1)

	// Subtitle
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(Muted).
			MarginBottom(1)

	// Header
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(TeslaWhite).
			Background(TeslaGray).
			Padding(0, 1).
			MarginBottom(1)

	// Status badges
	StatusBadgeBase = lipgloss.NewStyle().
			Bold(true).
			Padding(0, 1)

	StatusBooked = StatusBadgeBase.
			Foreground(TeslaWhite).
			Background(StatusBlue)

	StatusInProgress = StatusBadgeBase.
				Foreground(TeslaWhite).
				Background(StatusYellow)

	StatusDelivered = StatusBadgeBase.
			Foreground(TeslaWhite).
			Background(StatusGreen)

	StatusCancelled = StatusBadgeBase.
			Foreground(TeslaWhite).
			Background(StatusRed)

	// Table
	TableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(TeslaWhite).
				Background(TeslaGray).
				Padding(0, 1)

	TableRowStyle = lipgloss.NewStyle().
			Padding(0, 1)

	TableSelectedStyle = lipgloss.NewStyle().
				Foreground(TeslaWhite).
				Background(TeslaRed).
				Bold(true).
				Padding(0, 1)

	// Tabs
	TabStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(Muted)

	ActiveTabStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(TeslaWhite).
			Background(TeslaRed).
			Bold(true)

	TabBarStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(Muted)

	// Detail items
	LabelStyle = lipgloss.NewStyle().
			Foreground(Muted).
			Width(24).
			Align(lipgloss.Right)

	SubheadingStyle = lipgloss.NewStyle().
			Foreground(TeslaWhite).
			Bold(true)

	ValueStyle = lipgloss.NewStyle().
			Foreground(TeslaWhite)

	ChangedValueStyle = lipgloss.NewStyle().
				Foreground(Highlight).
				Bold(true)

	OldValueStyle = lipgloss.NewStyle().
			Foreground(Muted).
			Strikethrough(true)

	// Help
	HelpStyle = lipgloss.NewStyle().
			Foreground(Muted).
			MarginTop(1)

	// Error
	ErrorStyle = lipgloss.NewStyle().
			Foreground(StatusRed).
			Bold(true)

	// Success
	SuccessStyle = lipgloss.NewStyle().
			Foreground(StatusGreen).
			Bold(true)

	// Box/Card
	CardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(TeslaGray).
			Padding(1, 2).
			MarginBottom(1)

	// Spinner
	SpinnerStyle = lipgloss.NewStyle().
			Foreground(TeslaRed)

	// Task status
	TaskCompleteStyle = lipgloss.NewStyle().
				Foreground(StatusGreen)

	TaskIncompleteStyle = lipgloss.NewStyle().
				Foreground(Muted)

	// JSON
	JSONKeyStyle = lipgloss.NewStyle().
			Foreground(StatusBlue)

	JSONStringStyle = lipgloss.NewStyle().
			Foreground(StatusGreen)

	JSONNumberStyle = lipgloss.NewStyle().
			Foreground(StatusYellow)

	JSONBoolStyle = lipgloss.NewStyle().
			Foreground(TeslaRed)

	// Diff
	DiffAddedStyle = lipgloss.NewStyle().
			Foreground(StatusGreen).
			Bold(true)

	DiffRemovedStyle = lipgloss.NewStyle().
			Foreground(StatusRed).
			Strikethrough(true)

	// Toast notifications
	ToastStyle = lipgloss.NewStyle().
			Foreground(TeslaWhite).
			Background(StatusGreen).
			Padding(0, 1).
			Bold(true)

	ToastErrorStyle = lipgloss.NewStyle().
			Foreground(TeslaWhite).
			Background(StatusRed).
			Padding(0, 1).
			Bold(true)

	// Section box style
	SectionBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(TeslaGray).
			Padding(0, 1)

	// Login card
	LoginCardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(TeslaRed).
			Padding(1, 2).
			Width(70)

	// JSON null
	JSONNullStyle = lipgloss.NewStyle().
			Foreground(Muted).
			Italic(true)

	// Help key/desc styles for bubbles/help
	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(TeslaRed).
			Bold(true)

	HelpDescStyle = lipgloss.NewStyle().
			Foreground(Muted)
)

// GetStatusBadgeStyle returns the appropriate style for an order status
func GetStatusBadgeStyle(status string) lipgloss.Style {
	switch {
	case containsAny(status, "booked", "book"):
		return StatusBooked
	case containsAny(status, "progress", "pending", "processing"):
		return StatusInProgress
	case containsAny(status, "delivered", "complete"):
		return StatusDelivered
	case containsAny(status, "cancel"):
		return StatusCancelled
	default:
		return StatusBadgeBase.Background(TeslaGray)
	}
}

// containsAny checks if s contains any of the substrings
func containsAny(s string, substrs ...string) bool {
	lower := toLower(s)
	for _, sub := range substrs {
		if contains(lower, sub) {
			return true
		}
	}
	return false
}

// toLower converts string to lowercase
func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		result[i] = c
	}
	return string(result)
}

// contains checks if s contains substr
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
