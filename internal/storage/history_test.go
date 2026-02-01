package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/marcelblijleven/tesla-delivery-tui/internal/model"
)

func TestNewHistory(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tesla-tui-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	history, err := NewHistory(tempDir)
	if err != nil {
		t.Fatalf("NewHistory() error = %v", err)
	}

	if history == nil {
		t.Fatal("NewHistory() returned nil")
	}

	// Check history directory was created
	historyDir := filepath.Join(tempDir, historyDirName)
	if _, err := os.Stat(historyDir); os.IsNotExist(err) {
		t.Error("History directory was not created")
	}
}

func TestHistory_LoadHistory_NoFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tesla-tui-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	history, _ := NewHistory(tempDir)

	// Load non-existent history
	loaded, err := history.LoadHistory("RN123456789")
	if err != nil {
		t.Fatalf("LoadHistory() error = %v", err)
	}

	if loaded == nil {
		t.Fatal("LoadHistory() returned nil")
	}

	if loaded.ReferenceNumber != "RN123456789" {
		t.Errorf("ReferenceNumber = %q, want %q", loaded.ReferenceNumber, "RN123456789")
	}

	if len(loaded.Snapshots) != 0 {
		t.Errorf("Snapshots length = %d, want 0", len(loaded.Snapshots))
	}
}

func TestHistory_SaveAndLoadHistory(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tesla-tui-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	history, _ := NewHistory(tempDir)

	vin := "5YJ3E1EA1LF123456"
	orderHistory := &model.OrderHistory{
		ReferenceNumber: "RN123456789",
		Snapshots: []model.HistoricalSnapshot{
			{
				Timestamp: time.Now().Add(-24 * time.Hour),
				Data: model.CombinedOrder{
					Order: model.TeslaOrder{
						ReferenceNumber: "RN123456789",
						OrderStatus:     "BOOKED",
						VIN:             &vin,
					},
				},
			},
		},
	}

	// Save
	if err := history.SaveHistory(orderHistory); err != nil {
		t.Fatalf("SaveHistory() error = %v", err)
	}

	// Load
	loaded, err := history.LoadHistory("RN123456789")
	if err != nil {
		t.Fatalf("LoadHistory() error = %v", err)
	}

	if len(loaded.Snapshots) != 1 {
		t.Fatalf("Snapshots length = %d, want 1", len(loaded.Snapshots))
	}

	if loaded.Snapshots[0].Data.Order.OrderStatus != "BOOKED" {
		t.Errorf("OrderStatus = %q, want %q", loaded.Snapshots[0].Data.Order.OrderStatus, "BOOKED")
	}
}

func TestHistory_SaveHistory_PrunesOldEntries(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tesla-tui-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	history, _ := NewHistory(tempDir)

	// Create history with more than maxHistoryEntries
	orderHistory := &model.OrderHistory{
		ReferenceNumber: "RN123456789",
		Snapshots:       make([]model.HistoricalSnapshot, 25), // More than maxHistoryEntries (20)
	}

	for i := 0; i < 25; i++ {
		orderHistory.Snapshots[i] = model.HistoricalSnapshot{
			Timestamp: time.Now().Add(time.Duration(-25+i) * time.Hour),
			Data: model.CombinedOrder{
				Order: model.TeslaOrder{
					ReferenceNumber: "RN123456789",
					OrderStatus:     "BOOKED",
				},
			},
		}
	}

	// Save (should prune)
	if err := history.SaveHistory(orderHistory); err != nil {
		t.Fatalf("SaveHistory() error = %v", err)
	}

	// Load and check
	loaded, _ := history.LoadHistory("RN123456789")
	if len(loaded.Snapshots) != maxHistoryEntries {
		t.Errorf("Snapshots length = %d, want %d (should be pruned)", len(loaded.Snapshots), maxHistoryEntries)
	}
}

func TestHistory_AddSnapshot_FirstSnapshot(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tesla-tui-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	history, _ := NewHistory(tempDir)

	order := model.CombinedOrder{
		Order: model.TeslaOrder{
			ReferenceNumber: "RN123456789",
			OrderStatus:     "BOOKED",
		},
	}

	// Add first snapshot
	diffs, err := history.AddSnapshot(order)
	if err != nil {
		t.Fatalf("AddSnapshot() error = %v", err)
	}

	// First snapshot should have no diffs
	if len(diffs) != 0 {
		t.Errorf("First snapshot diffs = %d, want 0", len(diffs))
	}

	// Verify it was saved
	loaded, _ := history.LoadHistory("RN123456789")
	if len(loaded.Snapshots) != 1 {
		t.Errorf("Snapshots length = %d, want 1", len(loaded.Snapshots))
	}
}

func TestHistory_AddSnapshot_DetectsChanges(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tesla-tui-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	history, _ := NewHistory(tempDir)

	// First snapshot
	order1 := model.CombinedOrder{
		Order: model.TeslaOrder{
			ReferenceNumber: "RN123456789",
			OrderStatus:     "BOOKED",
		},
	}
	history.AddSnapshot(order1)

	// Second snapshot with changes
	vin := "5YJ3E1EA1LF123456"
	order2 := model.CombinedOrder{
		Order: model.TeslaOrder{
			ReferenceNumber: "RN123456789",
			OrderStatus:     "READY",
			VIN:             &vin,
		},
	}

	diffs, err := history.AddSnapshot(order2)
	if err != nil {
		t.Fatalf("AddSnapshot() error = %v", err)
	}

	// Should detect status and VIN changes
	if len(diffs) < 2 {
		t.Errorf("Expected at least 2 diffs, got %d", len(diffs))
	}

	// Check for specific diffs
	foundStatus := false
	foundVIN := false
	for _, diff := range diffs {
		if diff.Field == "Order Status" {
			foundStatus = true
			if diff.OldValue != "BOOKED" || diff.NewValue != "READY" {
				t.Errorf("Order Status diff: old=%v new=%v", diff.OldValue, diff.NewValue)
			}
		}
		if diff.Field == "VIN" {
			foundVIN = true
		}
	}

	if !foundStatus {
		t.Error("Order Status change not detected")
	}
	if !foundVIN {
		t.Error("VIN change not detected")
	}
}

func TestHistory_AddSnapshot_NoChanges(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tesla-tui-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	history, _ := NewHistory(tempDir)

	order := model.CombinedOrder{
		Order: model.TeslaOrder{
			ReferenceNumber: "RN123456789",
			OrderStatus:     "BOOKED",
		},
	}

	// Add same order twice
	history.AddSnapshot(order)
	diffs, _ := history.AddSnapshot(order)

	// No changes = no diffs
	if len(diffs) != 0 {
		t.Errorf("No changes but got %d diffs", len(diffs))
	}

	// Should not add duplicate snapshot
	loaded, _ := history.LoadHistory("RN123456789")
	if len(loaded.Snapshots) != 1 {
		t.Errorf("Snapshots = %d, want 1 (no duplicate)", len(loaded.Snapshots))
	}
}

func TestHistory_GetLatestSnapshot(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tesla-tui-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	history, _ := NewHistory(tempDir)

	// No snapshots
	snapshot, err := history.GetLatestSnapshot("RN123456789")
	if err != nil {
		t.Fatalf("GetLatestSnapshot() error = %v", err)
	}
	if snapshot != nil {
		t.Error("GetLatestSnapshot() should return nil for no snapshots")
	}

	// Add some snapshots
	for i := 0; i < 3; i++ {
		order := model.CombinedOrder{
			Order: model.TeslaOrder{
				ReferenceNumber: "RN123456789",
				OrderStatus:     "STATUS" + string(rune('A'+i)),
			},
		}
		history.AddSnapshot(order)
	}

	// Get latest
	snapshot, err = history.GetLatestSnapshot("RN123456789")
	if err != nil {
		t.Fatalf("GetLatestSnapshot() error = %v", err)
	}

	if snapshot == nil {
		t.Fatal("GetLatestSnapshot() returned nil")
	}

	if snapshot.Data.Order.OrderStatus != "STATUSC" {
		t.Errorf("Latest status = %q, want STATUSC", snapshot.Data.Order.OrderStatus)
	}
}

func TestCompareOrders_AllFields(t *testing.T) {
	vin1 := "VIN1"
	vin2 := "VIN2"
	opts1 := "OPT1"
	opts2 := "OPT2"

	old := model.CombinedOrder{
		Order: model.TeslaOrder{
			OrderStatus: "BOOKED",
			VIN:         &vin1,
			MktOptions:  &opts1,
		},
		Details: model.OrderDetails{
			Tasks: model.OrderTasks{
				Scheduling: &model.SchedulingTask{
					DeliveryWindowDisplay:  "Jan-Feb",
					ApptDateTimeAddressStr: "Jan 15",
					DeliveryType:           "PICKUP",
					DeliveryAddressTitle:   "Center A",
				},
				FinalPayment: &model.FinalPaymentTask{
					Data: &model.FinalPaymentData{
						ETAToDeliveryCenter: "Jan 10",
					},
				},
				Registration: &model.RegistrationTask{
					OrderDetails: &model.RegistrationOrderDetails{
						VehicleRoutingLocation: "Location A",
						VehicleOdometer:        "10",
						ReservationDate:        "2024-01-01",
						OrderBookedDate:        "2024-01-02",
					},
				},
				DeliveryDetails: &model.DeliveryDetailsTask{
					RegData: &model.DeliveryDetailsRegData{
						ReggieLicensePlate: "ABC-123",
					},
				},
			},
		},
	}

	newOrder := model.CombinedOrder{
		Order: model.TeslaOrder{
			OrderStatus: "READY",
			VIN:         &vin2,
			MktOptions:  &opts2,
		},
		Details: model.OrderDetails{
			Tasks: model.OrderTasks{
				Scheduling: &model.SchedulingTask{
					DeliveryWindowDisplay:  "Feb-Mar",
					ApptDateTimeAddressStr: "Feb 15",
					DeliveryType:           "DELIVERY",
					DeliveryAddressTitle:   "Center B",
				},
				FinalPayment: &model.FinalPaymentTask{
					Data: &model.FinalPaymentData{
						ETAToDeliveryCenter: "Feb 10",
					},
				},
				Registration: &model.RegistrationTask{
					OrderDetails: &model.RegistrationOrderDetails{
						VehicleRoutingLocation: "Location B",
						VehicleOdometer:        "20",
						ReservationDate:        "2024-02-01",
						OrderBookedDate:        "2024-02-02",
					},
				},
				DeliveryDetails: &model.DeliveryDetailsTask{
					RegData: &model.DeliveryDetailsRegData{
						ReggieLicensePlate: "XYZ-789",
					},
				},
			},
		},
	}

	diffs := compareOrders(old, newOrder)

	expectedFields := []string{
		"Order Status",
		"VIN",
		"Vehicle Options",
		"Delivery Window",
		"Delivery Appointment",
		"ETA to Delivery Center",
		"Vehicle Location",
		"Delivery Method",
		"Delivery Center",
		"Odometer",
		"License Plate",
		"Reservation Date",
		"Order Booked Date",
	}

	foundFields := make(map[string]bool)
	for _, diff := range diffs {
		foundFields[diff.Field] = true
	}

	for _, field := range expectedFields {
		if !foundFields[field] {
			t.Errorf("Missing diff for field: %s", field)
		}
	}

	if len(diffs) != len(expectedFields) {
		t.Errorf("Got %d diffs, want %d. Fields: %v", len(diffs), len(expectedFields), foundFields)
	}
}

func TestHistory_FilePermissions(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tesla-tui-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	history, _ := NewHistory(tempDir)

	orderHistory := &model.OrderHistory{
		ReferenceNumber: "RN123456789",
		Snapshots: []model.HistoricalSnapshot{
			{
				Timestamp: time.Now(),
				Data: model.CombinedOrder{
					Order: model.TeslaOrder{ReferenceNumber: "RN123456789"},
				},
			},
		},
	}

	history.SaveHistory(orderHistory)

	// Check file permissions
	filePath := filepath.Join(tempDir, historyDirName, "RN123456789.json")
	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("Failed to stat history file: %v", err)
	}

	mode := info.Mode().Perm()
	if mode != 0600 {
		t.Errorf("History file permissions = %o, want 0600", mode)
	}
}
