package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const checklistDirName = "checklists"

// ChecklistItem represents a single checklist item
type ChecklistItem struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

// ChecklistSection represents a group of checklist items
type ChecklistSection struct {
	Title string          `json:"title"`
	Items []ChecklistItem `json:"items"`
}

// ChecklistState stores which items have been checked per order
type ChecklistState struct {
	ReferenceNumber string          `json:"referenceNumber"`
	Checked         map[string]bool `json:"checked"`
}

// DeliveryChecklist defines the standard delivery checklist sections.
//
// Based on Tesla's official delivery guidance (tesla.com/support/taking-delivery,
// tesla.com/support/delivery-day) and community best practices
// (github.com/mykeln/teslaprep).
var DeliveryChecklist = []ChecklistSection{
	{
		Title: "Before Delivery",
		Items: []ChecklistItem{
			{ID: "finance_sorted", Text: "Financing or payment method confirmed in the Tesla app"},
			{ID: "insured", Text: "Vehicle added to insurance policy"},
			{ID: "home_charger", Text: "Home charging setup ready (wall connector / outlet)"},
			{ID: "docs_reviewed", Text: "Motor Vehicle Purchase Agreement reviewed and signed"},
			{ID: "tradein_ready", Text: "Trade-in vehicle cleaned and paperwork prepared"},
			{ID: "pickup_route", Text: "Route to delivery center or pickup location planned"},
		},
	},
}

// Checklist manages checklist persistence
type Checklist struct {
	baseDir string
}

// NewChecklist creates a new Checklist instance
func NewChecklist(configDir string) (*Checklist, error) {
	checklistDir := filepath.Join(configDir, checklistDirName)
	if err := os.MkdirAll(checklistDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create checklist directory: %w", err)
	}

	return &Checklist{baseDir: checklistDir}, nil
}

func (c *Checklist) filePath(referenceNumber string) string {
	return filepath.Join(c.baseDir, referenceNumber+".json")
}

// LoadState loads the checklist state for a specific order
func (c *Checklist) LoadState(referenceNumber string) (*ChecklistState, error) {
	data, err := os.ReadFile(c.filePath(referenceNumber))
	if err != nil {
		if os.IsNotExist(err) {
			return &ChecklistState{
				ReferenceNumber: referenceNumber,
				Checked:         make(map[string]bool),
			}, nil
		}
		return nil, fmt.Errorf("failed to read checklist file: %w", err)
	}

	var state ChecklistState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse checklist file: %w", err)
	}
	if state.Checked == nil {
		state.Checked = make(map[string]bool)
	}

	return &state, nil
}

// SaveState saves the checklist state for a specific order
func (c *Checklist) SaveState(state *ChecklistState) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal checklist: %w", err)
	}

	if err := os.WriteFile(c.filePath(state.ReferenceNumber), data, 0600); err != nil {
		return fmt.Errorf("failed to write checklist file: %w", err)
	}

	return nil
}

// ToggleItem toggles a checklist item and persists the change
func (c *Checklist) ToggleItem(referenceNumber, itemID string) (bool, error) {
	state, err := c.LoadState(referenceNumber)
	if err != nil {
		return false, err
	}

	state.Checked[itemID] = !state.Checked[itemID]
	newValue := state.Checked[itemID]

	if err := c.SaveState(state); err != nil {
		return false, err
	}

	return newValue, nil
}

// CountCompleted returns (completed, total) counts for all checklist items
func CountCompleted(checked map[string]bool) (int, int) {
	total := 0
	completed := 0
	for _, section := range DeliveryChecklist {
		for _, item := range section.Items {
			total++
			if checked[item.ID] {
				completed++
			}
		}
	}
	return completed, total
}
