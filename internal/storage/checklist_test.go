package storage

import (
	"os"
	"testing"
)

func TestNewChecklist(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tesla-tui-checklist-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cl, err := NewChecklist(tempDir)
	if err != nil {
		t.Fatalf("NewChecklist() error = %v", err)
	}
	if cl == nil {
		t.Fatal("NewChecklist() returned nil")
	}
}

func TestChecklist_LoadState_NoFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tesla-tui-checklist-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cl, _ := NewChecklist(tempDir)

	state, err := cl.LoadState("RN123456789")
	if err != nil {
		t.Fatalf("LoadState() error = %v", err)
	}
	if state == nil {
		t.Fatal("LoadState() returned nil")
	}
	if state.ReferenceNumber != "RN123456789" {
		t.Errorf("ReferenceNumber = %q, want %q", state.ReferenceNumber, "RN123456789")
	}
	if len(state.Checked) != 0 {
		t.Errorf("Checked should be empty, got %d items", len(state.Checked))
	}
}

func TestChecklist_SaveAndLoadState(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tesla-tui-checklist-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cl, _ := NewChecklist(tempDir)

	state := &ChecklistState{
		ReferenceNumber: "RN123456789",
		Checked: map[string]bool{
			"finance_sorted": true,
			"insured":        true,
			"paint_check":    false,
		},
	}

	if err := cl.SaveState(state); err != nil {
		t.Fatalf("SaveState() error = %v", err)
	}

	loaded, err := cl.LoadState("RN123456789")
	if err != nil {
		t.Fatalf("LoadState() error = %v", err)
	}

	if !loaded.Checked["finance_sorted"] {
		t.Error("Expected finance_sorted to be checked")
	}
	if !loaded.Checked["insured"] {
		t.Error("Expected insured to be checked")
	}
	if loaded.Checked["paint_check"] {
		t.Error("Expected paint_check to not be checked")
	}
}

func TestChecklist_ToggleItem(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tesla-tui-checklist-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cl, _ := NewChecklist(tempDir)

	// Toggle on
	checked, err := cl.ToggleItem("RN123456789", "finance_sorted")
	if err != nil {
		t.Fatalf("ToggleItem() error = %v", err)
	}
	if !checked {
		t.Error("Expected finance_sorted to be toggled on")
	}

	// Verify persisted
	state, _ := cl.LoadState("RN123456789")
	if !state.Checked["finance_sorted"] {
		t.Error("Expected finance_sorted to be persisted as checked")
	}

	// Toggle off
	checked, err = cl.ToggleItem("RN123456789", "finance_sorted")
	if err != nil {
		t.Fatalf("ToggleItem() error = %v", err)
	}
	if checked {
		t.Error("Expected finance_sorted to be toggled off")
	}

	// Verify persisted
	state, _ = cl.LoadState("RN123456789")
	if state.Checked["finance_sorted"] {
		t.Error("Expected finance_sorted to be persisted as unchecked")
	}
}

func TestChecklist_ToggleItem_MultipleItems(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tesla-tui-checklist-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cl, _ := NewChecklist(tempDir)

	cl.ToggleItem("RN123", "finance_sorted")
	cl.ToggleItem("RN123", "insured")
	cl.ToggleItem("RN123", "paint_check")

	state, _ := cl.LoadState("RN123")
	if !state.Checked["finance_sorted"] || !state.Checked["insured"] || !state.Checked["paint_check"] {
		t.Error("Expected all three items to be checked")
	}

	// Toggle one off
	cl.ToggleItem("RN123", "insured")
	state, _ = cl.LoadState("RN123")
	if state.Checked["insured"] {
		t.Error("Expected insured to be unchecked after toggle")
	}
	if !state.Checked["finance_sorted"] || !state.Checked["paint_check"] {
		t.Error("Expected finance_sorted and paint_check to remain checked")
	}
}

func TestChecklist_SeparateOrders(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tesla-tui-checklist-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cl, _ := NewChecklist(tempDir)

	// Toggle items for two different orders
	cl.ToggleItem("RN111", "finance_sorted")
	cl.ToggleItem("RN222", "paint_check")

	state1, _ := cl.LoadState("RN111")
	state2, _ := cl.LoadState("RN222")

	if !state1.Checked["finance_sorted"] {
		t.Error("RN111: expected finance_sorted checked")
	}
	if state1.Checked["paint_check"] {
		t.Error("RN111: paint_check should not be checked")
	}

	if state2.Checked["finance_sorted"] {
		t.Error("RN222: finance_sorted should not be checked")
	}
	if !state2.Checked["paint_check"] {
		t.Error("RN222: expected paint_check checked")
	}
}

func TestCountCompleted(t *testing.T) {
	tests := []struct {
		name          string
		checked       map[string]bool
		wantCompleted int
		wantTotal     int
	}{
		{
			name:          "none checked",
			checked:       map[string]bool{},
			wantCompleted: 0,
		},
		{
			name: "some checked",
			checked: map[string]bool{
				"finance_sorted": true,
				"insured":        true,
			},
			wantCompleted: 2,
		},
		{
			name: "false values not counted",
			checked: map[string]bool{
				"finance_sorted": true,
				"insured":        false,
			},
			wantCompleted: 1,
		},
		{
			name: "unknown items not counted",
			checked: map[string]bool{
				"unknown_item":   true,
				"finance_sorted": true,
			},
			wantCompleted: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			completed, total := CountCompleted(tt.checked)
			if completed != tt.wantCompleted {
				t.Errorf("CountCompleted() completed = %d, want %d", completed, tt.wantCompleted)
			}
			// Total should be the sum of all items in DeliveryChecklist
			expectedTotal := 0
			for _, section := range DeliveryChecklist {
				expectedTotal += len(section.Items)
			}
			if total != expectedTotal {
				t.Errorf("CountCompleted() total = %d, want %d", total, expectedTotal)
			}
		})
	}
}

func TestDeliveryChecklist_Structure(t *testing.T) {
	if len(DeliveryChecklist) < 1 {
		t.Errorf("Expected at least 1 checklist section, got %d", len(DeliveryChecklist))
	}

	// Check that all items have IDs and text
	seenIDs := make(map[string]bool)
	for _, section := range DeliveryChecklist {
		if section.Title == "" {
			t.Error("Section has empty title")
		}
		if len(section.Items) == 0 {
			t.Errorf("Section %q has no items", section.Title)
		}
		for _, item := range section.Items {
			if item.ID == "" {
				t.Errorf("Item in section %q has empty ID", section.Title)
			}
			if item.Text == "" {
				t.Errorf("Item %q in section %q has empty text", item.ID, section.Title)
			}
			if seenIDs[item.ID] {
				t.Errorf("Duplicate item ID: %q", item.ID)
			}
			seenIDs[item.ID] = true
		}
	}
}

func TestChecklist_FilePermissions(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tesla-tui-checklist-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cl, _ := NewChecklist(tempDir)

	state := &ChecklistState{
		ReferenceNumber: "RN123",
		Checked:         map[string]bool{"finance_sorted": true},
	}
	cl.SaveState(state)

	info, err := os.Stat(cl.filePath("RN123"))
	if err != nil {
		t.Fatalf("Failed to stat checklist file: %v", err)
	}

	mode := info.Mode().Perm()
	if mode != 0600 {
		t.Errorf("Checklist file permissions = %o, want 0600", mode)
	}
}
