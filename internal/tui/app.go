package tui

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os/exec"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/pkg/browser"

	"github.com/marcelblijleven/tesla-delivery-tui/internal/api"
	"github.com/marcelblijleven/tesla-delivery-tui/internal/config"
	"github.com/marcelblijleven/tesla-delivery-tui/internal/data"
	"github.com/marcelblijleven/tesla-delivery-tui/internal/demo"
	"github.com/marcelblijleven/tesla-delivery-tui/internal/model"
	"github.com/marcelblijleven/tesla-delivery-tui/internal/storage"
)

// View represents the current view
type View int

const (
	ViewLogin View = iota
	ViewOrders
	ViewDetail
	ViewHelp
)

// Tab represents tabs in the detail view
type Tab int

const (
	TabDetails Tab = iota
	TabTasks
	TabChecklist
	TabHistory
	TabJSON
)

// Messages
type (
	// AuthResultMsg contains the result of authentication
	AuthResultMsg struct {
		Tokens *model.TeslaTokens
		Error  error
	}

	// OrdersLoadedMsg contains loaded orders
	OrdersLoadedMsg struct {
		Orders []model.CombinedOrder
		Diffs  map[string][]model.OrderDiff
		Error  error
	}

	// TickMsg for auto-refresh
	TickMsg time.Time

	// ErrMsg for errors
	ErrMsg struct{ error }

	// BrowserOpenedMsg indicates browser was opened for auth
	BrowserOpenedMsg struct {
		Session *api.AuthSession
		Error   error
	}

	// DemoLoadedMsg indicates demo data was loaded
	DemoLoadedMsg struct {
		Orders  []model.CombinedOrder
		Diffs   map[string][]model.OrderDiff
		History map[string]*model.OrderHistory
	}

	// ToastMsg displays a temporary notification
	ToastMsg struct {
		Message string
		IsError bool
	}

	// ClearToastMsg clears the toast notification
	ClearToastMsg struct{}

	// AutoRefreshTickMsg triggers auto-refresh
	AutoRefreshTickMsg time.Time

	// ClipboardMsg indicates text was copied to clipboard
	ClipboardMsg struct {
		Text    string
		Success bool
		Error   error
	}

	// LogoutMsg indicates the user has been logged out
	LogoutMsg struct{}

	// ChecklistToggleMsg indicates a checklist item was toggled
	ChecklistToggleMsg struct {
		ItemID   string
		Checked  bool
		Error    error
	}
)

// Compiled regexes for JSON syntax highlighting
var (
	jsonKeyRe    = regexp.MustCompile(`^(\s*)"([^"]+)"\s*:`)
	jsonStringRe = regexp.MustCompile(`:\s*"([^"]*)"`)
	jsonNumberRe = regexp.MustCompile(`:\s*(-?\d+\.?\d*(?:[eE][+-]?\d+)?)`)
	jsonBoolRe   = regexp.MustCompile(`:\s*(true|false)`)
	jsonNullRe   = regexp.MustCompile(`:\s*(null)`)
)

// Model is the main application model
type Model struct {
	// Dependencies
	config    *config.Config
	client    *api.Client
	history   *storage.History
	checklist *storage.Checklist

	// State
	view             View
	previousView     View // for returning from help
	tokens           *model.TeslaTokens
	orders           []model.CombinedOrder
	diffs            map[string][]model.OrderDiff
	selectedOrder    int
	selectedTab      Tab
	err              error
	loading          bool
	authenticating   bool
	authSession      *api.AuthSession
	demoMode         bool
	demoHistory      map[string]*model.OrderHistory
	confirmingLogout bool

	// Checklist
	checklistState  *storage.ChecklistState
	checklistCursor int

	// Toast notification
	toastMessage string
	toastIsError bool

	// Auto-refresh
	autoRefresh         bool
	autoRefreshInterval time.Duration
	lastRefresh         time.Time

	// UI Components
	spinner   spinner.Model
	textInput textinput.Model
	viewport  viewport.Model
	help      help.Model
	keys      KeyMap

	// Window size
	width  int
	height int
}

// New creates a new Model
func New(cfg *config.Config, client *api.Client, hist *storage.History, cl *storage.Checklist) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = SpinnerStyle

	ti := textinput.New()
	ti.Placeholder = "Paste callback URL here..."
	ti.CharLimit = 2000
	ti.Width = 60

	vp := viewport.New(80, 20)
	vp.MouseWheelEnabled = true
	vp.MouseWheelDelta = 3

	h := help.New()
	h.Styles.FullKey = HelpKeyStyle
	h.Styles.FullDesc = HelpDescStyle
	h.Styles.FullSeparator = HelpDescStyle
	h.Styles.ShortKey = HelpKeyStyle
	h.Styles.ShortDesc = HelpDescStyle
	h.Styles.ShortSeparator = HelpDescStyle
	h.ShowAll = true

	return Model{
		config:    cfg,
		client:    client,
		history:   hist,
		checklist: cl,
		view:      ViewLogin,
		keys:      DefaultKeyMap,
		spinner:   s,
		textInput: ti,
		viewport:  vp,
		help:      h,
		diffs:     make(map[string][]model.OrderDiff),
	}
}

// WithDemoMode enables demo mode with mock data
func (m Model) WithDemoMode() Model {
	m.demoMode = true
	return m
}

// WithAutoRefresh enables automatic refresh at the specified interval
func (m Model) WithAutoRefresh(interval time.Duration) Model {
	m.autoRefresh = true
	m.autoRefreshInterval = interval
	return m
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	if m.demoMode {
		return tea.Batch(
			m.spinner.Tick,
			m.loadDemoData,
		)
	}
	return tea.Batch(
		m.spinner.Tick,
		m.checkSavedTokens,
	)
}

// loadDemoData loads mock data for demo mode
func (m Model) loadDemoData() tea.Msg {
	return DemoLoadedMsg{
		Orders:  demo.GetDemoOrders(),
		Diffs:   demo.GetDemoDiffs(),
		History: demo.GetDemoHistory(),
	}
}

// checkSavedTokens checks for saved tokens on startup
func (m Model) checkSavedTokens() tea.Msg {
	tokens, err := m.config.LoadTokens()
	if err != nil {
		return ErrMsg{err}
	}
	if tokens == nil {
		return nil
	}

	// If tokens are still valid, use them
	if !tokens.IsExpired() {
		return AuthResultMsg{Tokens: tokens}
	}

	// Access token expired, try to refresh using the refresh token
	if tokens.RefreshToken != "" {
		newTokens, err := m.client.Auth().RefreshTokens(tokens.RefreshToken)
		if err == nil {
			// Save the refreshed tokens
			if saveErr := m.config.SaveTokens(newTokens); saveErr != nil {
				return AuthResultMsg{Error: fmt.Errorf("failed to save refreshed tokens: %w", saveErr)}
			}
			return AuthResultMsg{Tokens: newTokens}
		}
		// Refresh failed, show the error so the user knows why re-authentication is needed
		return AuthResultMsg{Error: fmt.Errorf("session expired, please sign in again (%w)", err)}
	}

	return AuthResultMsg{Error: fmt.Errorf("session expired, please sign in again")}
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width - 4
		// Update viewport size (leave room for header, tabs, footer, and padding)
		// Header: combined title+order line(1) + tabs(1) + tabBorder(1) + tabMarginBottom(1) = 4
		// Footer: toast slot(2) + helpMarginTop(1) + help text(1) = 4 lines
		// AppStyle padding: top(1) + bottom(1) = 2 lines
		// Safety: 2 lines
		reservedHeight := 4 + 4 + 2 + 2
		m.viewport.Width = msg.Width - 4 // account for horizontal padding
		m.viewport.Height = msg.Height - reservedHeight
		if m.viewport.Height < 5 {
			m.viewport.Height = 5
		}
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case BrowserOpenedMsg:
		if msg.Error != nil {
			m.err = msg.Error
			m.authenticating = false
			return m, nil
		}
		m.authSession = msg.Session
		m.textInput.Focus()
		return m, textinput.Blink

	case AuthResultMsg:
		m.authenticating = false
		m.authSession = nil
		m.textInput.SetValue("")
		m.textInput.Blur()
		if msg.Error != nil {
			m.err = msg.Error
			return m, nil
		}
		m.tokens = msg.Tokens
		m.client.SetTokens(msg.Tokens)
		if err := m.config.SaveTokens(msg.Tokens); err != nil {
			m.err = err
			return m, nil
		}
		m.view = ViewOrders
		m.loading = true
		return m, tea.Batch(m.spinner.Tick, m.loadOrders)

	case OrdersLoadedMsg:
		m.loading = false
		m.lastRefresh = time.Now()
		if msg.Error != nil {
			m.err = msg.Error
			// Still schedule next auto-refresh even on error
			if m.autoRefresh {
				return m, m.scheduleAutoRefresh()
			}
			return m, nil
		}
		m.orders = msg.Orders
		m.diffs = msg.Diffs
		m.err = nil

		// Show toast notification with refresh result
		changeCount := len(msg.Diffs)
		if changeCount > 0 {
			m.toastMessage = fmt.Sprintf("✓ Refreshed - %d order(s) with changes", changeCount)
		} else {
			m.toastMessage = "✓ Refreshed - no changes detected"
		}
		m.toastIsError = false

		// Schedule next auto-refresh if enabled
		var cmds []tea.Cmd
		cmds = append(cmds, m.clearToastAfterDelay())
		if m.autoRefresh {
			cmds = append(cmds, m.scheduleAutoRefresh())
		}
		return m, tea.Batch(cmds...)

	case ToastMsg:
		m.toastMessage = msg.Message
		m.toastIsError = msg.IsError
		return m, m.clearToastAfterDelay()

	case ClearToastMsg:
		m.toastMessage = ""
		m.toastIsError = false
		return m, nil

	case DemoLoadedMsg:
		m.loading = false
		m.orders = msg.Orders
		m.diffs = msg.Diffs
		m.demoHistory = msg.History
		m.view = ViewOrders
		m.err = nil
		return m, nil

	case ErrMsg:
		m.err = msg.error
		m.loading = false
		return m, nil

	case AutoRefreshTickMsg:
		// Only refresh if we're on the orders view and not already loading
		if m.view == ViewOrders && !m.loading && m.tokens != nil {
			m.loading = true
			return m, tea.Batch(m.spinner.Tick, m.loadOrders)
		}
		// Reschedule if we couldn't refresh now
		if m.autoRefresh {
			return m, m.scheduleAutoRefresh()
		}
		return m, nil

	case LogoutMsg:
		m.tokens = nil
		m.orders = nil
		m.diffs = make(map[string][]model.OrderDiff)
		m.checklistState = nil
		m.view = ViewLogin
		m.err = nil
		return m, nil

	case ChecklistToggleMsg:
		if msg.Error != nil {
			m.toastMessage = "✗ Failed to save checklist"
			m.toastIsError = true
			return m, m.clearToastAfterDelay()
		}
		// Reload checklist state
		if m.selectedOrder < len(m.orders) {
			ref := m.orders[m.selectedOrder].Order.ReferenceNumber
			state, err := m.checklist.LoadState(ref)
			if err == nil {
				m.checklistState = state
			}
		}
		m.viewport.SetContent(m.getTabContent())
		return m, nil

	case ClipboardMsg:
		if msg.Success {
			label := msg.Text
			if len(label) > 40 {
				label = "JSON"
			}
			m.toastMessage = fmt.Sprintf("✓ Copied: %s", label)
			m.toastIsError = false
		} else {
			m.toastMessage = "✗ Failed to copy to clipboard"
			m.toastIsError = true
		}
		return m, m.clearToastAfterDelay()

	case tea.MouseMsg:
		return m.handleMouseEvent(msg)
	}

	return m, nil
}

// handleKeyPress handles key presses based on current view
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle logout confirmation first
	if m.confirmingLogout {
		switch msg.String() {
		case "y", "Y":
			m.confirmingLogout = false
			return m, m.logout
		case "n", "N", "esc":
			m.confirmingLogout = false
			return m, nil
		}
		return m, nil
	}

	// Global keys
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "?":
		// Toggle help view; skip when already showing help (handled by handleHelpKeys)
		if m.view == ViewHelp {
			m.view = m.previousView
			return m, nil
		}
		if m.view != ViewLogin || (!m.authenticating && m.authSession == nil) {
			m.previousView = m.view
			m.view = ViewHelp
			return m, nil
		}
	}

	// View-specific keys
	switch m.view {
	case ViewLogin:
		return m.handleLoginKeys(msg)
	case ViewOrders:
		return m.handleOrdersKeys(msg)
	case ViewDetail:
		return m.handleDetailKeys(msg)
	case ViewHelp:
		return m.handleHelpKeys(msg)
	}

	return m, nil
}

// handleHelpKeys handles keys in help view
func (m Model) handleHelpKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "?", "enter", "backspace":
		m.view = m.previousView
		return m, nil
	}
	return m, nil
}

// handleLoginKeys handles keys in login view
func (m Model) handleLoginKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// If we're waiting for URL input
	if m.authSession != nil {
		switch msg.String() {
		case "esc":
			m.authSession = nil
			m.authenticating = false
			m.textInput.SetValue("")
			m.textInput.Blur()
			return m, nil
		case "enter":
			return m.submitCallbackURL()
		default:
			var cmd tea.Cmd
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
		}
	}

	if m.authenticating {
		return m, nil
	}

	switch msg.String() {
	case "enter":
		m.authenticating = true
		m.err = nil
		return m, tea.Batch(m.spinner.Tick, m.openBrowserForAuth)
	}

	return m, nil
}

// handleOrdersKeys handles keys in orders view
func (m Model) handleOrdersKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.loading {
		return m, nil
	}

	switch msg.String() {
	case "up", "k":
		if m.selectedOrder > 0 {
			m.selectedOrder--
		}
	case "down", "j":
		if m.selectedOrder < len(m.orders)-1 {
			m.selectedOrder++
		}
	case "enter":
		if len(m.orders) > 0 {
			m.view = ViewDetail
			m.selectedTab = TabDetails
			m.viewport.SetContent(m.getTabContent())
			m.viewport.GotoTop()
		}
	case "r":
		m.loading = true
		return m, tea.Batch(m.spinner.Tick, m.loadOrders)
	case "L":
		m.confirmingLogout = true
		return m, nil
	case "y", "c":
		// Copy VIN of selected order to clipboard
		if len(m.orders) > 0 && m.selectedOrder < len(m.orders) {
			vin := m.orders[m.selectedOrder].Order.GetVIN()
			if vin != "" && vin != "N/A" {
				return m, copyToClipboard(vin)
			}
			m.toastMessage = "No VIN available to copy"
			m.toastIsError = true
			return m, m.clearToastAfterDelay()
		}
	}

	return m, nil
}

// handleDetailKeys handles keys in detail view
func (m Model) handleDetailKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	const numTabs = 5 // Details, Tasks, Checklist, History, JSON

	// Checklist-specific keys
	if m.selectedTab == TabChecklist {
		switch msg.String() {
		case "up", "k":
			if m.checklistCursor > 0 {
				m.checklistCursor--
				m.viewport.SetContent(m.getTabContent())
			}
			return m, nil
		case "down", "j":
			totalItems := 0
			for _, section := range storage.DeliveryChecklist {
				totalItems += len(section.Items)
			}
			if m.checklistCursor < totalItems-1 {
				m.checklistCursor++
				m.viewport.SetContent(m.getTabContent())
			}
			return m, nil
		case "enter", " ":
			if m.selectedOrder < len(m.orders) {
				ref := m.orders[m.selectedOrder].Order.ReferenceNumber
				itemID := m.getChecklistItemAtCursor()
				if itemID != "" {
					return m, func() tea.Msg {
						checked, err := m.checklist.ToggleItem(ref, itemID)
						return ChecklistToggleMsg{ItemID: itemID, Checked: checked, Error: err}
					}
				}
			}
			return m, nil
		}
	}

	switch msg.String() {
	case "esc", "backspace":
		m.view = ViewOrders
		m.viewport.GotoTop()
		return m, nil
	case "tab":
		m.selectedTab = Tab((int(m.selectedTab) + 1) % numTabs)
		m.onTabSwitch()
		m.viewport.SetContent(m.getTabContent())
		m.viewport.GotoTop()
		return m, nil
	case "shift+tab":
		if m.selectedTab == 0 {
			m.selectedTab = TabJSON
		} else {
			m.selectedTab--
		}
		m.onTabSwitch()
		m.viewport.SetContent(m.getTabContent())
		m.viewport.GotoTop()
		return m, nil
	case "r":
		m.loading = true
		return m, tea.Batch(m.spinner.Tick, m.loadOrders)
	case "y", "c":
		if m.selectedOrder < len(m.orders) {
			if m.selectedTab == TabJSON {
				// Copy full JSON on the JSON tab
				return m, m.copyJSON()
			}
			// Copy VIN on other tabs
			vin := m.orders[m.selectedOrder].Order.GetVIN()
			if vin != "" && vin != "N/A" {
				return m, copyToClipboard(vin)
			}
			m.toastMessage = "No VIN available to copy"
			m.toastIsError = true
			return m, m.clearToastAfterDelay()
		}
	}

	// Pass other keys to viewport for scrolling
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// onTabSwitch performs setup when switching tabs
func (m *Model) onTabSwitch() {
	if m.selectedTab == TabChecklist && m.selectedOrder < len(m.orders) {
		ref := m.orders[m.selectedOrder].Order.ReferenceNumber
		state, err := m.checklist.LoadState(ref)
		if err == nil {
			m.checklistState = state
		}
		m.checklistCursor = 0
	}
}

// getChecklistItemAtCursor returns the checklist item ID at the current cursor position
func (m Model) getChecklistItemAtCursor() string {
	idx := 0
	for _, section := range storage.DeliveryChecklist {
		for _, item := range section.Items {
			if idx == m.checklistCursor {
				return item.ID
			}
			idx++
		}
	}
	return ""
}

// openBrowserForAuth opens the browser for Tesla login
func (m Model) openBrowserForAuth() tea.Msg {
	session, err := m.client.Auth().CreateAuthSession()
	if err != nil {
		return BrowserOpenedMsg{Error: err}
	}

	// Open browser with auth URL
	if err := browser.OpenURL(session.AuthURL); err != nil {
		return BrowserOpenedMsg{Error: fmt.Errorf("failed to open browser: %w", err)}
	}

	return BrowserOpenedMsg{Session: session}
}

// submitCallbackURL processes the pasted callback URL
func (m Model) submitCallbackURL() (tea.Model, tea.Cmd) {
	callbackURL := m.textInput.Value()
	if callbackURL == "" {
		m.err = fmt.Errorf("please paste the callback URL")
		return m, nil
	}

	// Parse the URL to extract the code
	code, err := extractCodeFromURL(callbackURL)
	if err != nil {
		m.err = err
		return m, nil
	}

	// Exchange code for tokens
	m.err = nil
	return m, func() tea.Msg {
		tokens, err := m.client.Auth().ExchangeCode(code, m.authSession.CodeVerifier)
		return AuthResultMsg{Tokens: tokens, Error: err}
	}
}

// extractCodeFromURL extracts the authorization code from a callback URL
func extractCodeFromURL(callbackURL string) (string, error) {
	parsed, err := url.Parse(callbackURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL format")
	}

	// Try query parameters first
	code := parsed.Query().Get("code")
	if code != "" {
		return code, nil
	}

	// Try hash fragment
	if parsed.Fragment != "" {
		fragParams, err := url.ParseQuery(parsed.Fragment)
		if err == nil {
			code = fragParams.Get("code")
			if code != "" {
				return code, nil
			}
		}
	}

	return "", fmt.Errorf("could not find authorization code in URL")
}


// loadOrders loads orders from the API
func (m Model) loadOrders() tea.Msg {
	orders, err := m.client.GetAllOrderData()
	if err != nil {
		return OrdersLoadedMsg{Error: err}
	}

	// Check for changes and store history
	diffs := make(map[string][]model.OrderDiff)
	for _, order := range orders {
		orderDiffs, err := m.history.AddSnapshot(order)
		if err != nil {
			// Log but don't fail
			continue
		}
		if len(orderDiffs) > 0 {
			diffs[order.Order.ReferenceNumber] = orderDiffs
		}
	}

	return OrdersLoadedMsg{Orders: orders, Diffs: diffs}
}

// logout logs out the user
func (m Model) logout() tea.Msg {
	if err := m.config.DeleteTokens(); err != nil {
		return ErrMsg{err}
	}
	return LogoutMsg{}
}

// clearToastAfterDelay returns a command that clears the toast after 3 seconds
func (m Model) clearToastAfterDelay() tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return ClearToastMsg{}
	})
}

// scheduleAutoRefresh schedules the next auto-refresh tick
func (m Model) scheduleAutoRefresh() tea.Cmd {
	return tea.Tick(m.autoRefreshInterval, func(t time.Time) tea.Msg {
		return AutoRefreshTickMsg(t)
	})
}

// copyJSON copies the full JSON of the selected order to the clipboard
func (m Model) copyJSON() tea.Cmd {
	if m.selectedOrder >= len(m.orders) {
		return nil
	}
	order := m.orders[m.selectedOrder]
	combined := map[string]interface{}{
		"order": order.Order,
	}
	if order.Details.RawJSON != nil {
		combined["details"] = order.Details.RawJSON
	} else {
		combined["details"] = order.Details
	}
	jsonBytes, err := json.MarshalIndent(combined, "", "  ")
	if err != nil {
		return func() tea.Msg {
			return ClipboardMsg{Text: "JSON", Success: false, Error: err}
		}
	}
	return copyToClipboard(string(jsonBytes))
}

// copyToClipboard copies text to the system clipboard using platform-native tools
func copyToClipboard(text string) tea.Cmd {
	return func() tea.Msg {
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "darwin":
			cmd = exec.Command("pbcopy")
		case "linux":
			// Try xclip first, fall back to xsel
			if _, err := exec.LookPath("xclip"); err == nil {
				cmd = exec.Command("xclip", "-selection", "clipboard")
			} else {
				cmd = exec.Command("xsel", "--clipboard", "--input")
			}
		default:
			return ClipboardMsg{Text: text, Success: false, Error: fmt.Errorf("unsupported platform")}
		}

		cmd.Stdin = strings.NewReader(text)
		if err := cmd.Run(); err != nil {
			return ClipboardMsg{Text: text, Success: false, Error: err}
		}
		return ClipboardMsg{Text: text, Success: true}
	}
}

// handleMouseEvent handles mouse clicks and scroll
func (m Model) handleMouseEvent(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// Handle scroll wheel events
	switch msg.Button {
	case tea.MouseButtonWheelUp, tea.MouseButtonWheelDown:
		if m.view == ViewDetail {
			// Pass scroll events to viewport
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		} else if m.view == ViewOrders && len(m.orders) > 0 {
			// Scroll through order list
			if msg.Button == tea.MouseButtonWheelUp {
				if m.selectedOrder > 0 {
					m.selectedOrder--
				}
			} else {
				if m.selectedOrder < len(m.orders)-1 {
					m.selectedOrder++
				}
			}
		}
		return m, nil
	}

	// Only handle clicks in certain views
	if m.view != ViewOrders && m.view != ViewDetail {
		return m, nil
	}

	// Handle left clicks
	if msg.Button != tea.MouseButtonLeft || msg.Action != tea.MouseActionRelease {
		return m, nil
	}

	switch m.view {
	case ViewOrders:
		// Calculate which row was clicked (accounting for header, padding, and table borders)
		// AppStyle padding: 1 line top
		// Title: 1 line
		// Blank line before table: 1 line
		// Table top border: 1 line
		// Table header row: 1 line
		// Table header border: 1 line
		headerLines := 6 // padding + title + blank + top border + header + header border
		clickedRow := msg.Y - headerLines

		if clickedRow >= 0 && clickedRow < len(m.orders) {
			if clickedRow == m.selectedOrder {
				// Double-click behavior: if same row, open details
				m.view = ViewDetail
				m.selectedTab = TabDetails
				m.viewport.SetContent(m.getTabContent())
				m.viewport.GotoTop()
			} else {
				m.selectedOrder = clickedRow
			}
		}

	case ViewDetail:
		// Check if clicking on tabs area
		// Tabs are at approximately line 4 (after header and subtitle)
		tabLine := 4
		if msg.Y == tabLine {
			// Calculate which tab was clicked based on X position
			// Each tab is roughly 10 characters wide with padding
			tabWidth := 10
			clickedTab := msg.X / tabWidth
			if clickedTab >= 0 && clickedTab < 4 {
				m.selectedTab = Tab(clickedTab)
				m.viewport.SetContent(m.getTabContent())
				m.viewport.GotoTop()
			}
		}
	}

	return m, nil
}

// Minimum terminal size
const (
	minTerminalWidth  = 80
	minTerminalHeight = 24
)

// View renders the UI
func (m Model) View() string {
	// Check minimum terminal size
	if m.width > 0 && m.height > 0 && (m.width < minTerminalWidth || m.height < minTerminalHeight) {
		return m.viewTerminalTooSmall()
	}

	switch m.view {
	case ViewLogin:
		return m.viewLogin()
	case ViewOrders:
		return m.viewOrders()
	case ViewDetail:
		return m.viewDetail()
	case ViewHelp:
		return m.viewHelp()
	default:
		return "Unknown view"
	}
}

// viewTerminalTooSmall renders a warning when terminal is too small
func (m Model) viewTerminalTooSmall() string {
	warning := lipgloss.JoinVertical(lipgloss.Center,
		"",
		ErrorStyle.Render("Terminal too small"),
		"",
		HelpStyle.Render(fmt.Sprintf("Minimum: %d×%d", minTerminalWidth, minTerminalHeight)),
		HelpStyle.Render(fmt.Sprintf("Current: %d×%d", m.width, m.height)),
		"",
		HelpStyle.Render("Please resize your terminal window."),
	)

	// Center the warning
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, warning)
}

// layoutWithFooter creates a layout with content at top and footer pinned to bottom
func (m Model) layoutWithFooter(content, footer string) string {
	contentHeight := lipgloss.Height(content)
	footerHeight := lipgloss.Height(footer)

	// Always reserve space for the toast line to prevent layout shift
	const toastReserved = 2 // toast line + spacing

	toastContent := ""
	if m.toastMessage != "" {
		style := ToastStyle
		if m.toastIsError {
			style = ToastErrorStyle
		}
		toastContent = style.Render(m.toastMessage)
	}

	// Account for AppStyle padding (1 line top + 1 line bottom = 2 lines)
	paddingHeight := 2

	// Calculate gap — always subtract toastReserved so footer stays stable
	availableHeight := m.height - contentHeight - footerHeight - toastReserved - paddingHeight
	if availableHeight < 1 {
		availableHeight = 1
	}

	gap := strings.Repeat("\n", availableHeight)

	// Build footer section: toast slot is always present (empty or filled)
	var footerSection string
	if toastContent != "" {
		footerSection = lipgloss.JoinVertical(lipgloss.Left, toastContent, footer)
	} else {
		// Empty line reserves the toast slot
		footerSection = lipgloss.JoinVertical(lipgloss.Left, "", footer)
	}

	return AppStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			content,
			gap,
			footerSection,
		),
	)
}

// relativeTime returns a human-readable relative time string
func relativeTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case diff < 48*time.Hour:
		return "yesterday"
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		return fmt.Sprintf("%d days ago", days)
	case diff < 30*24*time.Hour:
		weeks := int(diff.Hours() / 24 / 7)
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	default:
		return t.Format("Jan 02, 2006")
	}
}

// viewLogin renders the login view
func (m Model) viewLogin() string {
	title := TitleStyle.Render("⚡ Tesla Delivery Status")
	subtitle := SubtitleStyle.Render("Track your Tesla order delivery status")

	var cardContent string
	var helpText string

	if m.authSession != nil {
		// Waiting for user to paste callback URL
		cardContent = fmt.Sprintf(`1. Complete the Tesla login in the browser window
2. After login, you'll see a "Page Not Found" page
3. Copy the entire URL from your browser's address bar
4. Paste it below and press Enter

%s`,
			m.textInput.View(),
		)

		if m.err != nil {
			cardContent += "\n\n" + ErrorStyle.Render("Error: "+m.err.Error())
		}

		helpText = HelpStyle.Render("enter: submit • esc: cancel")
	} else if m.authenticating {
		cardContent = fmt.Sprintf("%s Opening browser for authentication...", m.spinner.View())
		helpText = HelpStyle.Render(LoginKeys())
	} else if m.err != nil {
		cardContent = fmt.Sprintf("%s\n\nPress Enter to try again.", ErrorStyle.Render("Error: "+m.err.Error()))
		helpText = HelpStyle.Render(LoginKeys())
	} else {
		cardContent = "Press Enter to login with your Tesla account."
		helpText = HelpStyle.Render(LoginKeys())
	}

	// Wrap in login card and center horizontally
	card := LoginCardStyle.Render(cardContent)
	cardWidth := lipgloss.Width(card)
	leftMargin := 0
	if m.width > cardWidth+4 {
		leftMargin = (m.width - cardWidth - 4) / 2
	}
	centeredCard := lipgloss.NewStyle().MarginLeft(leftMargin).Render(card)

	topContent := lipgloss.JoinVertical(lipgloss.Left, title, subtitle, "", centeredCard)
	return m.layoutWithFooter(topContent, helpText)
}

// viewOrders renders the orders list view
func (m Model) viewOrders() string {
	title := TitleStyle.Render("⚡ Tesla Delivery Status")

	var help string
	if m.confirmingLogout {
		help = HelpStyle.Render("Logout? Press 'y' to confirm, 'n' or 'esc' to cancel")
	} else {
		help = HelpStyle.Render(OrdersKeys())
	}

	var content string

	if m.confirmingLogout {
		content = m.renderLogoutConfirmation()
	} else if m.loading {
		content = fmt.Sprintf("\n%s Loading orders...", m.spinner.View())
	} else if m.err != nil {
		content = ErrorStyle.Render(fmt.Sprintf("\nError: %s\n\nPress 'r' to retry.", m.err.Error()))
	} else if len(m.orders) == 0 {
		content = m.renderEmptyState()
	} else {
		// Build orders table with lipgloss/table
		tableWidth := m.width - 4
		if tableWidth < 80 {
			tableWidth = 80
		}

		selectedOrder := m.selectedOrder
		orderDiffs := m.diffs

		var tableRows [][]string
		for i, order := range m.orders {
			vin := order.Order.GetVIN()
			if len(vin) > 17 {
				vin = vin[:17]
			}

			deliveryWindow := order.GetDeliveryWindow()
			if len(deliveryWindow) > 25 {
				deliveryWindow = deliveryWindow[:22] + "..."
			}

			changeIndicator := " "
			if _, hasChanges := orderDiffs[order.Order.ReferenceNumber]; hasChanges {
				changeIndicator = "✓"
			}

			modelName := order.Order.GetModelName()
			if i == selectedOrder {
				modelName = "▸ " + modelName
			}

			tableRows = append(tableRows, []string{
				modelName,
				order.Order.OrderStatus,
				vin,
				deliveryWindow,
				changeIndicator,
			})
		}

		t := table.New().
			Headers("Model", "Status", "VIN", "Delivery Window", "Changed").
			Rows(tableRows...).
			Border(lipgloss.RoundedBorder()).
			BorderStyle(lipgloss.NewStyle().Foreground(TeslaGray)).
			Width(tableWidth).
			StyleFunc(func(row, col int) lipgloss.Style {
				s := lipgloss.NewStyle().Padding(0, 1)

				if row == table.HeaderRow {
					return s.Bold(true).Foreground(TeslaWhite).Background(TeslaRed)
				}

				// Selection highlight — bold + accent color, no background
				if row == selectedOrder {
					return s.Foreground(Highlight).Bold(true)
				}

				// Change indicator column
				if col == 4 {
					return s.Foreground(StatusGreen)
				}

				// Zebra striping (odd rows only, so first data row has no background)
				if row%2 == 1 {
					return s.Background(SubtleBg)
				}

				return s
			})

		content = "\n" + t.Render()
	}

	// Calculate content and create layout with footer at bottom
	topContent := lipgloss.JoinVertical(lipgloss.Left, title, content)
	return m.layoutWithFooter(topContent, help)
}

// viewDetail renders the order detail view
func (m Model) viewDetail() string {
	if m.selectedOrder >= len(m.orders) {
		return "No order selected"
	}

	order := m.orders[m.selectedOrder]

	// Title on the left, order info on the right — same line
	titleLeft := TitleStyle.MarginBottom(0).Render("⚡ Tesla Delivery Status")
	statusStyle := GetStatusBadgeStyle(order.Order.OrderStatus)
	refStyle := lipgloss.NewStyle().Foreground(Muted)
	orderInfo := lipgloss.JoinHorizontal(lipgloss.Center,
		SubheadingStyle.Render(order.Order.GetModelName()),
		"  ",
		statusStyle.Render(order.Order.OrderStatus),
		"  ",
		refStyle.Render(order.Order.ReferenceNumber),
	)
	headerWidth := m.width - 4 // account for AppStyle horizontal padding
	headerLine := lipgloss.JoinHorizontal(lipgloss.Center,
		titleLeft,
		lipgloss.PlaceHorizontal(headerWidth-lipgloss.Width(titleLeft), lipgloss.Right, orderInfo),
	)

	// Tabs
	tabs := m.renderTabs()

	// Build scrollbar indicator
	scrollPercent := ""
	if m.viewport.TotalLineCount() > m.viewport.Height {
		scrollPercent = fmt.Sprintf(" (%d%%)", int(m.viewport.ScrollPercent()*100))
	}

	help := HelpStyle.Render(DetailKeys(m.selectedTab) + scrollPercent)

	topContent := lipgloss.JoinVertical(lipgloss.Left,
		headerLine,
		"",
		tabs,
		m.viewport.View(),
	)

	return m.layoutWithFooter(topContent, help)
}

// viewHelp renders the help screen
func (m Model) viewHelp() string {
	title := TitleStyle.Render("⚡ Tesla Delivery Status")
	sectionTitle := SubheadingStyle.Render("Keyboard Shortcuts")

	// Use bubbles/help for formatted key/description columns
	helpContent := m.help.View(m.keys)

	var lines []string
	lines = append(lines, "")
	lines = append(lines, helpContent)

	// Show auto-refresh status if enabled
	if m.autoRefresh {
		lines = append(lines, "")
		lines = append(lines, SubheadingStyle.Render("Auto-Refresh"))
		lines = append(lines, SuccessStyle.Render(fmt.Sprintf("  ● Enabled (every %s)", m.autoRefreshInterval)))
		if !m.lastRefresh.IsZero() {
			lines = append(lines, HelpStyle.Render(fmt.Sprintf("  Last refresh: %s", relativeTime(m.lastRefresh))))
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)

	// Wrap in a card/box
	boxContent := CardStyle.Render(content)

	helpFooter := HelpStyle.Render("Press Esc, ?, or Enter to close")

	topContent := lipgloss.JoinVertical(lipgloss.Left, title, sectionTitle, "", boxContent)
	return m.layoutWithFooter(topContent, helpFooter)
}

// renderEmptyState renders a friendly empty state message
func (m Model) renderEmptyState() string {
	var lines []string

	lines = append(lines, "")
	lines = append(lines, SubheadingStyle.Render("No Tesla orders found"))
	lines = append(lines, "")
	lines = append(lines, HelpStyle.Render("If you have an active Tesla order, it should appear here."))
	lines = append(lines, "")
	lines = append(lines, HelpStyle.Render("Try the following:"))
	lines = append(lines, HelpStyle.Render("  • Press 'r' to refresh"))
	lines = append(lines, HelpStyle.Render("  • Press 'L' to logout and login again"))
	lines = append(lines, HelpStyle.Render("  • Verify your order at tesla.com"))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// renderLogoutConfirmation renders the logout confirmation dialog
func (m Model) renderLogoutConfirmation() string {
	var lines []string

	lines = append(lines, "")

	// Create a confirmation box
	confirmContent := lipgloss.JoinVertical(lipgloss.Center,
		"",
		SubheadingStyle.Render("Are you sure you want to logout?"),
		"",
		HelpStyle.Render("This will clear your saved credentials."),
		"",
		ValueStyle.Render("[Y]es    [N]o"),
		"",
	)

	box := CardStyle.Width(50).Render(confirmContent)
	lines = append(lines, box)

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// getTabContent returns the content for the current tab
func (m Model) getTabContent() string {
	if m.selectedOrder >= len(m.orders) {
		return ""
	}

	order := m.orders[m.selectedOrder]
	diffs := m.diffs[order.Order.ReferenceNumber]

	switch m.selectedTab {
	case TabDetails:
		return m.renderDetailsTab(order, diffs)
	case TabTasks:
		return m.renderTasksTab(order)
	case TabChecklist:
		return m.renderChecklistTab(order)
	case TabHistory:
		return m.renderHistoryTab(order)
	case TabJSON:
		return m.renderJSONTab(order)
	}
	return ""
}

// renderTabs renders the tab bar
func (m Model) renderTabs() string {
	tabNames := []string{"Details", "Tasks", "Checklist", "History", "JSON"}

	// Add checklist progress badge
	if m.selectedOrder < len(m.orders) {
		ref := m.orders[m.selectedOrder].Order.ReferenceNumber
		state, err := m.checklist.LoadState(ref)
		if err == nil {
			completed, total := storage.CountCompleted(state.Checked)
			tabNames[2] = fmt.Sprintf("Checklist %d/%d", completed, total)
		}
	}

	// Add history count badge
	if m.selectedOrder < len(m.orders) {
		ref := m.orders[m.selectedOrder].Order.ReferenceNumber
		var historyCount int
		if m.demoMode && m.demoHistory != nil {
			if h, ok := m.demoHistory[ref]; ok {
				historyCount = len(h.Snapshots)
			}
		} else {
			h, err := m.history.LoadHistory(ref)
			if err == nil {
				historyCount = len(h.Snapshots)
			}
		}
		if historyCount > 0 {
			tabNames[3] = fmt.Sprintf("History (%d)", historyCount)
		}
	}

	var tabs []string
	for i, name := range tabNames {
		style := TabStyle
		if Tab(i) == m.selectedTab {
			style = ActiveTabStyle
		}
		tabs = append(tabs, style.Render(name))
	}

	tabBar := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
	return TabBarStyle.Width(m.width - 4).Render(tabBar)
}

// renderLabelValue renders a label-value pair with consistent alignment using LabelStyle and ValueStyle
func renderLabelValue(label, value string) string {
	return fmt.Sprintf("  %s %s",
		LabelStyle.Render(label+":"),
		ValueStyle.Render(value))
}

// currencySymbol returns the symbol for a currency code
func currencySymbol(code string) string {
	switch strings.ToUpper(code) {
	case "EUR":
		return "\u20ac"
	case "USD":
		return "$"
	case "GBP":
		return "\u00a3"
	case "CHF":
		return "CHF"
	case "NOK", "SEK", "DKK":
		return "kr"
	case "CNY":
		return "\u00a5"
	case "JPY":
		return "\u00a5"
	case "CAD":
		return "CA$"
	case "AUD":
		return "A$"
	default:
		return code + " "
	}
}

// formatThousands formats an integer with comma thousand separators (e.g. 39120 → "39,120")
func formatThousands(n int64) string {
	if n < 0 {
		return "-" + formatThousands(-n)
	}
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	var result strings.Builder
	remainder := len(s) % 3
	if remainder > 0 {
		result.WriteString(s[:remainder])
	}
	for i := remainder; i < len(s); i += 3 {
		if result.Len() > 0 {
			result.WriteByte(',')
		}
		result.WriteString(s[i : i+3])
	}
	return result.String()
}

// renderDetailsTab renders the details tab content
func (m Model) renderDetailsTab(order model.CombinedOrder, diffs []model.OrderDiff) string {
	diffMap := make(map[string]model.OrderDiff)
	for _, d := range diffs {
		diffMap[d.Field] = d
	}

	var lines []string

	// Show change summary banner if there are recent changes
	if len(diffs) > 0 {
		changesBanner := DiffAddedStyle.Render(fmt.Sprintf("● %d change(s) detected since last check:", len(diffs)))
		lines = append(lines, changesBanner)
		for _, diff := range diffs {
			lines = append(lines, HelpStyle.Render(fmt.Sprintf("  • %s: %v → %v", diff.Field, diff.OldValue, diff.NewValue)))
		}
		lines = append(lines, "")
	}

	renderField := func(label, value string) string {
		valueStyle := ValueStyle
		prefix := "  "
		suffix := ""
		if diff, ok := diffMap[label]; ok {
			valueStyle = ChangedValueStyle
			prefix = DiffAddedStyle.Render("● ")
			suffix = OldValueStyle.Render(fmt.Sprintf(" (was: %v)", diff.OldValue))
		}
		return fmt.Sprintf("%s%s %s%s",
			prefix,
			LabelStyle.Render(label+":"),
			valueStyle.Render(value),
			suffix,
		)
	}

	// Order Timeline
	lines = append(lines, m.renderOrderTimeline(order))
	lines = append(lines, "")

	// Delivery Countdown
	if countdown := m.renderCountdown(order); countdown != "" {
		lines = append(lines, countdown)
		lines = append(lines, "")
	}

	// Order Details Section
	var detailFields []string
	detailFields = append(detailFields, renderField("VIN", order.Order.GetVIN()))
	detailFields = append(detailFields, renderField("License Plate", order.GetLicensePlate()))
	detailFields = append(detailFields, renderField("Delivery Window", order.GetDeliveryWindow()))

	// Parsed appointment details
	if appt := order.GetParsedAppointment(); appt != nil {
		detailFields = append(detailFields, renderField("Appointment Date", appt.Date))
		if appt.Time != "" {
			detailFields = append(detailFields, renderField("Appointment Time", appt.Time))
		}
		if appt.Address != "" {
			detailFields = append(detailFields, renderField("Appointment Location", appt.Address))
		}
	} else {
		detailFields = append(detailFields, renderField("Delivery Appointment", order.GetDeliveryAppointment()))
	}

	detailFields = append(detailFields, renderField("ETA to Delivery Center", order.GetETAToDeliveryCenter()))
	detailFields = append(detailFields, renderField("Vehicle Location", order.GetVehicleLocation()))
	detailFields = append(detailFields, renderField("Delivery Method", order.GetDeliveryType()))
	detailFields = append(detailFields, renderField("Delivery Center", data.GetStoreName(order.GetDeliveryCenter())))
	detailFields = append(detailFields, renderField("Odometer", order.GetOdometer()))

	// Reservation and order dates
	if order.GetReservationDate() != "N/A" {
		detailFields = append(detailFields, renderField("Reservation Date", order.GetReservationDate()))
	}
	if order.GetOrderBookedDate() != "N/A" {
		detailFields = append(detailFields, renderField("Order Booked Date", order.GetOrderBookedDate()))
	}

	if order.Order.IsB2B && order.Order.OwnerCompanyName != nil {
		detailFields = append(detailFields, renderField("Company", *order.Order.OwnerCompanyName))
	}

	lines = append(lines, SubheadingStyle.Render("Order Details"))
	lines = append(lines, SectionBoxStyle.Width(m.sectionWidth()).Render(lipgloss.JoinVertical(lipgloss.Left, detailFields...)))

	// Payment Summary Section
	if paymentSection := m.renderPaymentSummary(order); paymentSection != "" {
		lines = append(lines, "")
		lines = append(lines, paymentSection)
	}

	// VIN Decoder Section
	if order.Order.VIN != nil && *order.Order.VIN != "" {
		lines = append(lines, "")
		lines = append(lines, m.renderVINDecoder(*order.Order.VIN))
	}

	// Vehicle Options Section
	if order.Order.MktOptions != nil {
		var optLines []string

		// Decode options
		decodedOptions := model.DecodeOptions(*order.Order.MktOptions)
		categories := model.CategorizeOptions(decodedOptions)

		// Display by category
		categoryOrder := []string{"Model", "Paint", "Interior", "Wheels", "Autopilot", "Charging", "Other"}
		for _, category := range categoryOrder {
			opts := categories[category]
			if len(opts) == 0 {
				continue
			}
			optLines = append(optLines, HelpStyle.Render(fmt.Sprintf("  %s:", category)))
			for _, opt := range opts {
				if opt.Description != "" {
					optLines = append(optLines, ValueStyle.Render(fmt.Sprintf("    • %s (%s)", opt.Description, opt.Code)))
				} else {
					optLines = append(optLines, ValueStyle.Render(fmt.Sprintf("    • %s", opt.Code)))
				}
			}
		}

		lines = append(lines, "")
		lines = append(lines, SubheadingStyle.Render("Vehicle Options"))
		lines = append(lines, SectionBoxStyle.Width(m.sectionWidth()).Render(lipgloss.JoinVertical(lipgloss.Left, optLines...)))
	}

	// Trade-In Section
	tradeInSection := m.renderTradeInDetails(order)
	if tradeInSection != "" {
		lines = append(lines, "")
		lines = append(lines, tradeInSection)
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// renderOrderTimeline renders the order progress timeline
func (m Model) renderOrderTimeline(order model.CombinedOrder) string {
	var timelineLines []string

	// Determine which stages are complete (must be sequential)
	hasVIN := order.Order.VIN != nil && *order.Order.VIN != ""
	hasTransitInfo := order.GetETAToDeliveryCenter() != "N/A" || order.GetVehicleLocation() != "N/A"
	hasAppointment := order.GetDeliveryAppointment() != "N/A"
	isDelivered := strings.Contains(strings.ToLower(order.Order.OrderStatus), "deliver")

	// Sequential completion - each stage requires previous stages
	vinComplete := hasVIN
	transitComplete := hasVIN && hasTransitInfo
	readyComplete := hasVIN && hasAppointment
	deliveredComplete := isDelivered

	// Timeline stages with sequential logic
	stageNames := []string{"Order Placed", "VIN Assigned", "In Transit", "Ready for Delivery", "Delivered"}
	stageComplete := []bool{true, vinComplete, transitComplete, readyComplete, deliveredComplete}

	// Find current stage (first incomplete, or last if all complete)
	currentStage := len(stageNames) - 1
	for i, complete := range stageComplete {
		if !complete {
			currentStage = i
			break
		}
	}

	// Style for incomplete items (HelpStyle has MarginTop which breaks inline rendering)
	mutedStyle := lipgloss.NewStyle().Foreground(Muted)

	// Render timeline
	for i, name := range stageNames {
		var icon string
		var nameStyle lipgloss.Style

		if stageComplete[i] {
			icon = TaskCompleteStyle.Render("●")
			nameStyle = ValueStyle
		} else if i == currentStage {
			icon = ChangedValueStyle.Render("◐")
			nameStyle = ChangedValueStyle
		} else {
			icon = mutedStyle.Render("○")
			nameStyle = mutedStyle
		}

		timelineLines = append(timelineLines, icon+" "+nameStyle.Render(name))

		// Connector line (except for last stage)
		if i < len(stageNames)-1 {
			if stageComplete[i+1] || i+1 == currentStage {
				timelineLines = append(timelineLines, TaskCompleteStyle.Render("  │"))
			} else {
				timelineLines = append(timelineLines, mutedStyle.Render("  │"))
			}
		}
	}

	// Wrap timeline content in a box
	timelineContent := strings.Join(timelineLines, "\n")
	boxedTimeline := SectionBoxStyle.Width(m.sectionWidth()).Render(timelineContent)

	return lipgloss.JoinVertical(lipgloss.Left,
		SubheadingStyle.Render("Order Timeline"),
		boxedTimeline,
	)
}

// renderDeliveryGates renders the delivery readiness checklist
func (m Model) renderDeliveryGates(order model.CombinedOrder) string {
	var lines []string
	lines = append(lines, SubheadingStyle.Render("Delivery Readiness:"))
	lines = append(lines, "")

	type gate struct {
		name       string
		complete   bool
		owner      string // "Customer" or "Tesla"
		isBlocker  bool
	}

	var gates []gate

	// Parse tasks to determine gate status
	if order.Details.Tasks.Raw != nil {
		// Customer tasks
		if raw, ok := order.Details.Tasks.Raw["registration"]; ok {
			var task struct {
				Complete bool `json:"complete"`
				Required bool `json:"required"`
			}
			json.Unmarshal(raw, &task)
			gates = append(gates, gate{"Complete Registration", task.Complete, "Customer", task.Required && !task.Complete})
		}

		if raw, ok := order.Details.Tasks.Raw["scheduling"]; ok {
			var task struct {
				Complete bool `json:"complete"`
				Required bool `json:"required"`
			}
			json.Unmarshal(raw, &task)
			gates = append(gates, gate{"Schedule Delivery", task.Complete, "Customer", task.Required && !task.Complete})
		}

		if raw, ok := order.Details.Tasks.Raw["finalPayment"]; ok {
			var task struct {
				Complete bool `json:"complete"`
				Required bool `json:"required"`
			}
			json.Unmarshal(raw, &task)
			gates = append(gates, gate{"Final Payment", task.Complete, "Customer", task.Required && !task.Complete})
		}

		if raw, ok := order.Details.Tasks.Raw["insurance"]; ok {
			var task struct {
				Complete bool `json:"complete"`
				Required bool `json:"required"`
			}
			json.Unmarshal(raw, &task)
			gates = append(gates, gate{"Insurance", task.Complete, "Customer", task.Required && !task.Complete})
		}

		if raw, ok := order.Details.Tasks.Raw["tradeIn"]; ok {
			var task struct {
				Complete bool `json:"complete"`
				Required bool `json:"required"`
			}
			json.Unmarshal(raw, &task)
			gates = append(gates, gate{"Trade-In", task.Complete, "Customer", task.Required && !task.Complete})
		}
	}

	// Tesla internal gates (inferred from data)
	hasVIN := order.Order.VIN != nil && *order.Order.VIN != ""
	gates = append(gates, gate{"VIN Assignment", hasVIN, "Tesla", false})

	inTransit := order.GetVehicleLocation() != "N/A"
	gates = append(gates, gate{"Vehicle in Transit", inTransit, "Tesla", false})

	// Count blockers
	blockerCount := 0
	for _, g := range gates {
		if g.isBlocker {
			blockerCount++
		}
	}

	if blockerCount > 0 {
		lines = append(lines, ErrorStyle.Render(fmt.Sprintf("  ⚠ %d blocker(s) remaining", blockerCount)))
		lines = append(lines, "")
	} else {
		lines = append(lines, SuccessStyle.Render("  ✓ No blockers - Ready for delivery!"))
		lines = append(lines, "")
	}

	// Style for incomplete items (HelpStyle has MarginTop which breaks inline rendering)
	mutedStyle := lipgloss.NewStyle().Foreground(Muted)

	// Group by owner
	lines = append(lines, ValueStyle.Bold(true).Render("  Customer Tasks:"))
	for _, g := range gates {
		if g.owner != "Customer" {
			continue
		}
		icon := mutedStyle.Render("○")
		nameStyle := mutedStyle
		if g.complete {
			icon = TaskCompleteStyle.Render("●")
			nameStyle = ValueStyle
		} else if g.isBlocker {
			icon = ErrorStyle.Render("○")
			nameStyle = ErrorStyle
		}
		lines = append(lines, "    "+icon+" "+nameStyle.Render(g.name))
	}

	lines = append(lines, "")
	lines = append(lines, ValueStyle.Bold(true).Render("  Tesla Tasks:"))
	for _, g := range gates {
		if g.owner != "Tesla" {
			continue
		}
		icon := mutedStyle.Render("○")
		nameStyle := mutedStyle
		if g.complete {
			icon = TaskCompleteStyle.Render("●")
			nameStyle = ValueStyle
		}
		lines = append(lines, "    "+icon+" "+nameStyle.Render(g.name))
	}

	return strings.Join(lines, "\n")
}

// renderVINDecoder renders decoded VIN information
func (m Model) renderVINDecoder(vin string) string {
	vinInfo := model.DecodeVIN(vin)
	if vinInfo == nil {
		return lipgloss.JoinVertical(lipgloss.Left,
			SubheadingStyle.Render("VIN Decoder"),
			HelpStyle.Render("  Invalid VIN"),
		)
	}

	renderVINField := func(label, value string) string {
		return fmt.Sprintf("  %s %s",
			HelpStyle.Render(fmt.Sprintf("%-20s", label+":")),
			ValueStyle.Render(value))
	}

	var fields []string
	fields = append(fields, renderVINField("Manufacturer", vinInfo.Manufacturer))
	fields = append(fields, renderVINField("Model", vinInfo.Model))
	fields = append(fields, renderVINField("Body Type", vinInfo.BodyType))
	fields = append(fields, renderVINField("Powertrain", vinInfo.Powertrain))
	fields = append(fields, renderVINField("Model Year", vinInfo.ModelYear))
	fields = append(fields, renderVINField("Plant", vinInfo.ManufacturingPlant))
	fields = append(fields, renderVINField("Serial Number", vinInfo.SerialNumber))

	return lipgloss.JoinVertical(lipgloss.Left,
		SubheadingStyle.Render("VIN Decoder"),
		SectionBoxStyle.Width(m.sectionWidth()).Render(lipgloss.JoinVertical(lipgloss.Left, fields...)),
	)
}

// renderTradeInDetails renders trade-in information if available
func (m Model) renderTradeInDetails(order model.CombinedOrder) string {
	// Check if trade-in task exists and has data
	if order.Details.Tasks.Raw == nil {
		return ""
	}

	raw, ok := order.Details.Tasks.Raw["tradeIn"]
	if !ok {
		return ""
	}

	var tradeIn struct {
		Complete     bool `json:"complete"`
		TradeInVehicle *struct {
			Make         string      `json:"make"`
			Model        string      `json:"model"`
			Year         string      `json:"year"`
			VIN          string      `json:"vin"`
			Trim         string      `json:"trim"`
			Mileage      json.Number `json:"mileage"`
			MileageUnit  string      `json:"mileageUnitOfMeasure"`
			Condition    string      `json:"condition"`
			TradeInCredit json.Number `json:"tradeInCredit"`
			LicensePlate string      `json:"licensePlate"`
		} `json:"tradeInVehicle"`
		CurrentVehicle *struct {
			FinalOffer json.Number `json:"finalOffer"`
		} `json:"currentVehicle"`
		SelectedValuation *struct {
			ValuationExpireDate string `json:"valuationExpireDate"`
		} `json:"selectedValuation"`
	}

	if err := json.Unmarshal(raw, &tradeIn); err != nil {
		return ""
	}

	// Only show if there's actual trade-in vehicle data
	if tradeIn.TradeInVehicle == nil {
		return ""
	}

	tv := tradeIn.TradeInVehicle
	var fields []string

	// Combined vehicle line: "2020 Volkswagen ID.3"
	vehicleParts := []string{}
	if tv.Year != "" {
		vehicleParts = append(vehicleParts, tv.Year)
	}
	if tv.Make != "" {
		vehicleParts = append(vehicleParts, tv.Make)
	}
	if tv.Model != "" {
		vehicleParts = append(vehicleParts, tv.Model)
	}
	if len(vehicleParts) > 0 {
		fields = append(fields, renderLabelValue("Vehicle", strings.Join(vehicleParts, " ")))
	}

	// Trim (truncated at 60 chars)
	if tv.Trim != "" {
		trim := tv.Trim
		if len(trim) > 60 {
			trim = trim[:57] + "..."
		}
		fields = append(fields, renderLabelValue("Trim", trim))
	}

	// VIN
	if tv.VIN != "" {
		fields = append(fields, renderLabelValue("VIN", tv.VIN))
	}

	// Registration plate
	if tv.LicensePlate != "" {
		fields = append(fields, renderLabelValue("Registration", tv.LicensePlate))
	}

	// Mileage + unit
	if mileageStr := tv.Mileage.String(); mileageStr != "" && mileageStr != "0" {
		mileage, err := tv.Mileage.Int64()
		if err == nil && mileage > 0 {
			unit := tv.MileageUnit
			if unit == "" {
				unit = "km"
			}
			fields = append(fields, renderLabelValue("Mileage", formatThousands(mileage)+" "+unit))
		}
	}

	// Condition
	if tv.Condition != "" {
		fields = append(fields, renderLabelValue("Condition", tv.Condition))
	}

	// Trade-in value: prefer currentVehicle.finalOffer, fall back to tradeInVehicle.tradeInCredit
	tradeValue := int64(0)
	if tradeIn.CurrentVehicle != nil {
		if v, err := tradeIn.CurrentVehicle.FinalOffer.Int64(); err == nil && v > 0 {
			tradeValue = v
		}
	}
	if tradeValue == 0 {
		if v, err := tv.TradeInCredit.Int64(); err == nil && v > 0 {
			tradeValue = v
		}
	}
	if tradeValue > 0 {
		// Try to get currency symbol from finalPayment task
		symbol := ""
		if fpRaw, fpOk := order.Details.Tasks.Raw["finalPayment"]; fpOk {
			var fp struct {
				CurrencyFormat *struct {
					CurrencyCode string `json:"currencyCode"`
				} `json:"currencyFormat"`
			}
			if json.Unmarshal(fpRaw, &fp) == nil && fp.CurrencyFormat != nil {
				symbol = currencySymbol(fp.CurrencyFormat.CurrencyCode)
			}
		}
		fields = append(fields, renderLabelValue("Trade-In Value", symbol+formatThousands(tradeValue)))
	}

	// Offer expiry date
	if tradeIn.SelectedValuation != nil && tradeIn.SelectedValuation.ValuationExpireDate != "" {
		fields = append(fields, renderLabelValue("Offer Expires", tradeIn.SelectedValuation.ValuationExpireDate))
	}

	if len(fields) == 0 {
		return ""
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		SubheadingStyle.Render("Trade-In Details"),
		SectionBoxStyle.Width(m.sectionWidth()).Render(lipgloss.JoinVertical(lipgloss.Left, fields...)),
	)
}

// renderPaymentSummary renders payment information parsed from raw task JSON
func (m Model) renderPaymentSummary(order model.CombinedOrder) string {
	if order.Details.Tasks.Raw == nil {
		return ""
	}

	var fields []string

	// Payment type from financing task: financing.card.messageTitle / messageBody
	if raw, ok := order.Details.Tasks.Raw["financing"]; ok {
		var financing struct {
			Card *struct {
				MessageTitle string `json:"messageTitle"`
				MessageBody  string `json:"messageBody"`
			} `json:"card"`
		}
		if err := json.Unmarshal(raw, &financing); err == nil && financing.Card != nil {
			if financing.Card.MessageBody != "" {
				fields = append(fields, renderLabelValue("Pay With", financing.Card.MessageBody))
			} else if financing.Card.MessageTitle != "" {
				fields = append(fields, renderLabelValue("Payment", financing.Card.MessageTitle))
			}
		}
	}

	// Amount due from finalPayment task
	if raw, ok := order.Details.Tasks.Raw["finalPayment"]; ok {
		var payment struct {
			AmountDue      json.Number `json:"amountDue"`
			CurrencyFormat *struct {
				CurrencyCode string `json:"currencyCode"`
			} `json:"currencyFormat"`
		}
		if err := json.Unmarshal(raw, &payment); err == nil {
			if amountStr := payment.AmountDue.String(); amountStr != "" && amountStr != "0" {
				amount, aErr := payment.AmountDue.Int64()
				if aErr == nil && amount > 0 {
					symbol := ""
					if payment.CurrencyFormat != nil && payment.CurrencyFormat.CurrencyCode != "" {
						symbol = currencySymbol(payment.CurrencyFormat.CurrencyCode)
					}
					fields = append(fields, renderLabelValue("Amount Due", symbol+formatThousands(amount)))
				}
			}
		}
	}

	// Order adjustments (e.g. referral credit) from registration task
	if raw, ok := order.Details.Tasks.Raw["registration"]; ok {
		var reg struct {
			OrderDetails *struct {
				OrderAdjustments []struct {
					Label  string      `json:"label"`
					Amount json.Number `json:"amount"`
				} `json:"orderAdjustments"`
				ReservationAmountReceived json.Number `json:"reservationAmountReceived"`
				CurrencyFormat            *struct {
					CurrencyCode string `json:"currencyCode"`
				} `json:"currencyFormat"`
			} `json:"orderDetails"`
		}
		if err := json.Unmarshal(raw, &reg); err == nil && reg.OrderDetails != nil {
			symbol := ""
			if reg.OrderDetails.CurrencyFormat != nil && reg.OrderDetails.CurrencyFormat.CurrencyCode != "" {
				symbol = currencySymbol(reg.OrderDetails.CurrencyFormat.CurrencyCode)
			}

			// If no symbol from registration, try to get from finalPayment
			if symbol == "" {
				if fpRaw, fpOk := order.Details.Tasks.Raw["finalPayment"]; fpOk {
					var fp struct {
						CurrencyFormat *struct {
							CurrencyCode string `json:"currencyCode"`
						} `json:"currencyFormat"`
					}
					if json.Unmarshal(fpRaw, &fp) == nil && fp.CurrencyFormat != nil {
						symbol = currencySymbol(fp.CurrencyFormat.CurrencyCode)
					}
				}
			}

			for _, adj := range reg.OrderDetails.OrderAdjustments {
				if adj.Label != "" {
					amount, aErr := adj.Amount.Int64()
					if aErr == nil && amount != 0 {
						prefix := "-"
						absAmount := amount
						if amount < 0 {
							absAmount = -amount
						} else {
							prefix = ""
						}
						fields = append(fields, renderLabelValue(adj.Label, prefix+symbol+formatThousands(absAmount)))
					}
				}
			}

			// Order deposit
			if depStr := reg.OrderDetails.ReservationAmountReceived.String(); depStr != "" && depStr != "0" {
				deposit, dErr := reg.OrderDetails.ReservationAmountReceived.Int64()
				if dErr == nil && deposit > 0 {
					fields = append(fields, renderLabelValue("Order Deposit", symbol+formatThousands(deposit)))
				}
			}
		}
	}

	if len(fields) == 0 {
		return ""
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		SubheadingStyle.Render("Payment Summary"),
		SectionBoxStyle.Width(m.sectionWidth()).Render(lipgloss.JoinVertical(lipgloss.Left, fields...)),
	)
}

// taskSortInfo holds task name and order for sorting
type taskSortInfo struct {
	name  string
	order int
}

// renderTasksTab renders the tasks tab content
func (m Model) renderTasksTab(order model.CombinedOrder) string {
	var lines []string

	// Delivery Readiness section
	lines = append(lines, m.renderDeliveryGates(order))
	lines = append(lines, "")

	// Order Tasks section
	lines = append(lines, SubheadingStyle.Render("Order Tasks:"))
	lines = append(lines, "")

	tasks := order.Details.Tasks

	// Skip metadata keys
	skipKeys := map[string]bool{
		"state":   true,
		"strings": true,
	}

	// Get all tasks with their order field for sorting
	var taskList []taskSortInfo
	for name, rawData := range tasks.Raw {
		if skipKeys[name] {
			continue
		}
		// Parse just the order field
		var orderInfo struct {
			Order int `json:"order"`
		}
		json.Unmarshal(rawData, &orderInfo)
		taskList = append(taskList, taskSortInfo{name: name, order: orderInfo.Order})
	}

	// Sort by order field (as in Tesla app)
	sort.Slice(taskList, func(i, j int) bool {
		return taskList[i].order < taskList[j].order
	})

	// Render each task from raw data
	for _, task := range taskList {
		name := task.name
		rawData := tasks.Raw[name]

		// Parse the task to get completion status and title
		var taskData struct {
			Complete bool `json:"complete"`
			Enabled  bool `json:"enabled"`
			Card     *struct {
				Title    string `json:"title"`
				Subtitle string `json:"subtitle"`
			} `json:"card"`
		}

		if err := json.Unmarshal(rawData, &taskData); err != nil {
			// If we can't parse, just show the name
			lines = append(lines, TaskIncompleteStyle.Render(fmt.Sprintf("  ○ %s", formatTaskName(name))))
			continue
		}

		// Render with completion status
		icon := "○"
		style := TaskIncompleteStyle
		statusText := ""
		if taskData.Complete {
			icon = "●"
			style = TaskCompleteStyle
			statusText = " ✓"
		}

		// Always show the formatted task name as the primary identifier
		taskLabel := formatTaskName(name)

		// Build the line with task name and status
		line := style.Render(fmt.Sprintf("  %s %s%s", icon, taskLabel, statusText))

		// Only show card details for incomplete tasks
		if !taskData.Complete && taskData.Card != nil {
			// Add card title as description if it's different and meaningful
			if taskData.Card.Title != "" && taskData.Card.Title != "Complete" &&
				strings.ToLower(taskData.Card.Title) != strings.ToLower(name) {
				line += "\n" + HelpStyle.Render(fmt.Sprintf("      %s", taskData.Card.Title))
			}

			// Add subtitle if available
			if taskData.Card.Subtitle != "" {
				line += "\n" + HelpStyle.Render(fmt.Sprintf("      %s", taskData.Card.Subtitle))
			}
		}

		lines = append(lines, line)
		lines = append(lines, "") // Add spacing between tasks
	}

	if len(lines) <= 2 {
		lines = append(lines, TaskIncompleteStyle.Render("  No tasks available"))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// formatTaskName converts camelCase task names to readable format
func formatTaskName(name string) string {
	// Map of known task names to readable versions
	taskNames := map[string]string{
		"deliveryAcceptance": "Delivery Acceptance",
		"deliveryDetails":    "Delivery Details",
		"finalPayment":       "Final Payment",
		"financing":          "Financing",
		"insurance":          "Insurance",
		"registration":       "Registration",
		"scheduling":         "Scheduling",
		"tradeIn":            "Trade-In",
	}

	if readable, ok := taskNames[name]; ok {
		return readable
	}

	// Convert camelCase to Title Case with spaces
	var result strings.Builder
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune(' ')
		}
		if i == 0 {
			result.WriteRune(rune(strings.ToUpper(string(r))[0]))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// renderCountdown renders a countdown to the delivery appointment
func (m Model) renderCountdown(order model.CombinedOrder) string {
	appt := order.GetParsedAppointment()
	if appt == nil {
		return ""
	}

	// Try to parse the date - format: "August 15, 2024"
	dateStr := appt.Date
	if appt.Time != "" {
		dateStr = appt.Date + " " + appt.Time
	}

	// Try common date formats
	var targetTime time.Time
	var err error
	formats := []string{
		"January 2, 2006 3:04 PM",
		"January 2, 2006 03:04 PM",
		"January 2, 2006",
		"Jan 2, 2006 3:04 PM",
		"Jan 2, 2006",
		"2006-01-02",
	}

	for _, format := range formats {
		targetTime, err = time.Parse(format, dateStr)
		if err == nil {
			break
		}
	}

	if err != nil {
		return ""
	}

	now := time.Now()
	diff := targetTime.Sub(now)

	if diff <= 0 {
		return SectionBoxStyle.Width(m.sectionWidth()).Render(
			SuccessStyle.Render("  It's Delivery Day! Congratulations!  "),
		)
	}

	days := int(diff.Hours() / 24)
	hours := int(diff.Hours()) % 24
	minutes := int(diff.Minutes()) % 60

	var countdown string
	if days > 0 {
		countdown = fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	} else if hours > 0 {
		countdown = fmt.Sprintf("%dh %dm", hours, minutes)
	} else {
		countdown = fmt.Sprintf("%dm", minutes)
	}

	content := fmt.Sprintf("  Delivery in %s  ", ChangedValueStyle.Render(countdown))
	return SectionBoxStyle.Width(m.sectionWidth()).Render(
		lipgloss.JoinVertical(lipgloss.Center,
			SubheadingStyle.Render("Delivery Countdown"),
			content,
		),
	)
}

// renderChecklistTab renders the delivery checklist tab
func (m Model) renderChecklistTab(order model.CombinedOrder) string {
	var lines []string

	// Load checklist state if needed
	checkState := m.checklistState
	if checkState == nil {
		var err error
		checkState, err = m.checklist.LoadState(order.Order.ReferenceNumber)
		if err != nil {
			return ErrorStyle.Render("Failed to load checklist: " + err.Error())
		}
	}

	// Progress summary
	completed, total := storage.CountCompleted(checkState.Checked)
	progressPct := 0
	if total > 0 {
		progressPct = completed * 100 / total
	}

	progressText := fmt.Sprintf("Progress: %d/%d (%d%%)", completed, total, progressPct)
	if completed == total {
		lines = append(lines, SuccessStyle.Render("  "+progressText+" - All done!"))
	} else {
		lines = append(lines, SubheadingStyle.Render("  "+progressText))
	}

	// Progress bar using bubbles/progress
	pct := 0.0
	if total > 0 {
		pct = float64(completed) / float64(total)
	}
	prog := progress.New(
		progress.WithSolidFill("#3B82F6"),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)
	lines = append(lines, "  "+prog.ViewAs(pct))
	lines = append(lines, "")

	// Render sections
	itemIdx := 0
	for _, section := range storage.DeliveryChecklist {
		lines = append(lines, SubheadingStyle.Render("  "+section.Title))
		lines = append(lines, "")

		for _, item := range section.Items {
			checked := checkState.Checked[item.ID]

			var icon string
			var style lipgloss.Style
			if checked {
				icon = TaskCompleteStyle.Render("●")
				style = lipgloss.NewStyle().Foreground(Muted).Strikethrough(true)
			} else {
				icon = TaskIncompleteStyle.Render("○")
				style = ValueStyle
			}

			cursor := "  "
			if itemIdx == m.checklistCursor {
				cursor = ChangedValueStyle.Render("> ")
			}

			lines = append(lines, fmt.Sprintf("  %s%s %s", cursor, icon, style.Render(item.Text)))
			itemIdx++
		}
		lines = append(lines, "")
	}

	lines = append(lines, HelpStyle.Render("  ↑/↓: navigate • enter/space: toggle • tab: next tab"))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// renderJSONTab renders the JSON tab content
func (m Model) renderJSONTab(order model.CombinedOrder) string {
	// Create a combined view with order info and raw API response
	combined := map[string]interface{}{
		"order": order.Order,
	}

	// Use raw JSON from API if available, otherwise use parsed details
	if order.Details.RawJSON != nil {
		combined["details"] = order.Details.RawJSON
	} else {
		combined["details"] = order.Details
	}

	// Marshal with indentation
	jsonBytes, err := json.MarshalIndent(combined, "", "  ")
	if err != nil {
		return ErrorStyle.Render("Failed to render JSON: " + err.Error())
	}

	return highlightJSON(string(jsonBytes))
}

// highlightJSON applies syntax highlighting to JSON output
func highlightJSON(jsonStr string) string {
	lines := strings.Split(jsonStr, "\n")
	var result []string

	for _, line := range lines {
		highlighted := line

		// Check for key: value patterns
		if loc := jsonKeyRe.FindStringSubmatchIndex(line); loc != nil {
			indent := line[loc[2]:loc[3]]
			key := line[loc[4]:loc[5]]
			rest := line[loc[1]:]

			// Highlight the value part
			var styledValue string
			switch {
			case jsonNullRe.MatchString(line):
				valLoc := jsonNullRe.FindStringSubmatchIndex(line)
				styledValue = line[loc[1]:valLoc[2]] + JSONNullStyle.Render(line[valLoc[2]:valLoc[3]]) + line[valLoc[3]:]
			case jsonBoolRe.MatchString(line):
				valLoc := jsonBoolRe.FindStringSubmatchIndex(line)
				styledValue = line[loc[1]:valLoc[2]] + JSONBoolStyle.Render(line[valLoc[2]:valLoc[3]]) + line[valLoc[3]:]
			case jsonNumberRe.MatchString(line):
				valLoc := jsonNumberRe.FindStringSubmatchIndex(line)
				styledValue = line[loc[1]:valLoc[2]] + JSONNumberStyle.Render(line[valLoc[2]:valLoc[3]]) + line[valLoc[3]:]
			case jsonStringRe.MatchString(line):
				valLoc := jsonStringRe.FindStringSubmatchIndex(line)
				// valLoc[2]:valLoc[3] is string content; include surrounding quotes
				quoteStart := valLoc[2] - 1
				quoteEnd := valLoc[3] + 1
				styledValue = line[loc[1]:quoteStart] + JSONStringStyle.Render(line[quoteStart:quoteEnd]) + line[quoteEnd:]
			default:
				styledValue = rest
			}

			highlighted = indent + JSONKeyStyle.Render(fmt.Sprintf(`"%s"`, key)) + styledValue
		}

		result = append(result, highlighted)
	}

	return strings.Join(result, "\n")
}

// renderHistoryTab renders the history tab content
func (m Model) renderHistoryTab(order model.CombinedOrder) string {
	var lines []string
	lines = append(lines, SubheadingStyle.Render("Order History:"))
	lines = append(lines, "")

	// Load history from storage (or demo data)
	var history *model.OrderHistory
	var err error

	if m.demoMode && m.demoHistory != nil {
		history = m.demoHistory[order.Order.ReferenceNumber]
		if history == nil {
			history = &model.OrderHistory{ReferenceNumber: order.Order.ReferenceNumber}
		}
	} else {
		history, err = m.history.LoadHistory(order.Order.ReferenceNumber)
		if err != nil {
			return ErrorStyle.Render("Failed to load history: " + err.Error())
		}
	}

	if len(history.Snapshots) == 0 {
		lines = append(lines, HelpStyle.Render("No history available yet."))
		lines = append(lines, HelpStyle.Render("History is recorded when changes are detected."))
		return lipgloss.JoinVertical(lipgloss.Left, lines...)
	}

	// Show snapshots in reverse chronological order (newest first)
	for i := len(history.Snapshots) - 1; i >= 0; i-- {
		snapshot := history.Snapshots[i]
		relTime := relativeTime(snapshot.Timestamp)
		fullTime := snapshot.Timestamp.Format("Jan 02, 2006 at 03:04 PM")

		// Snapshot header with relative time and full timestamp
		if i == len(history.Snapshots)-1 {
			lines = append(lines, ValueStyle.Render(fmt.Sprintf("● %s (Current)", relTime)))
			lines = append(lines, HelpStyle.Render(fmt.Sprintf("  %s", fullTime)))
		} else {
			lines = append(lines, HelpStyle.Render(fmt.Sprintf("○ %s", relTime)))
			lines = append(lines, HelpStyle.Render(fmt.Sprintf("  %s", fullTime)))
		}

		// Show key details at this snapshot
		data := snapshot.Data
		lines = append(lines, HelpStyle.Render(fmt.Sprintf("  Status: %s", data.Order.OrderStatus)))
		lines = append(lines, HelpStyle.Render(fmt.Sprintf("  VIN: %s", data.Order.GetVIN())))
		lines = append(lines, HelpStyle.Render(fmt.Sprintf("  Delivery Window: %s", data.GetDeliveryWindow())))

		// Show what changed compared to previous snapshot
		if i > 0 {
			prevSnapshot := history.Snapshots[i-1]
			changes := m.compareSnapshots(prevSnapshot.Data, snapshot.Data)
			if len(changes) > 0 {
				lines = append(lines, DiffAddedStyle.Render("    Changes:"))
				for _, change := range changes {
					lines = append(lines, fmt.Sprintf("      %s %s → %s",
						DiffAddedStyle.Render("•"),
						change.Field,
						DiffAddedStyle.Render(fmt.Sprintf("%v", change.NewValue)),
					))
				}
			}
		}

		lines = append(lines, "") // spacing between snapshots
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// sectionWidth returns the width for SectionBoxStyle content areas so borders span full width.
// Accounts for AppStyle horizontal padding (4) and SectionBoxStyle border (2).
func (m Model) sectionWidth() int {
	w := m.width - 6
	if w < 40 {
		w = 40
	}
	return w
}

// compareSnapshots compares two order snapshots using the canonical comparison
func (m Model) compareSnapshots(old, new model.CombinedOrder) []model.OrderDiff {
	return model.CompareOrders(old, new)
}
