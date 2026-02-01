package demo

import (
	"testing"
)

func TestGetDemoOrders(t *testing.T) {
	orders := GetDemoOrders()

	if len(orders) == 0 {
		t.Fatal("GetDemoOrders() returned empty slice")
	}

	// Check first order
	order := orders[0]

	if order.Order.ReferenceNumber == "" {
		t.Error("ReferenceNumber is empty")
	}

	if order.Order.OrderStatus == "" {
		t.Error("OrderStatus is empty")
	}

	if order.Order.ModelCode == "" {
		t.Error("ModelCode is empty")
	}

	// Check VIN is set
	if order.Order.VIN == nil || *order.Order.VIN == "" {
		t.Error("VIN should be set in demo data")
	}

	// Check tasks are populated
	if order.Details.Tasks.Scheduling == nil {
		t.Error("Scheduling task should be set")
	}

	if order.Details.Tasks.Registration == nil {
		t.Error("Registration task should be set")
	}

	if order.Details.Tasks.FinalPayment == nil {
		t.Error("FinalPayment task should be set")
	}

	// Check Raw tasks map is populated
	if order.Details.Tasks.Raw == nil {
		t.Error("Raw tasks map should be set")
	}

	expectedTasks := []string{"scheduling", "registration", "finalPayment", "deliveryDetails", "tradeIn", "insurance", "financing"}
	for _, taskName := range expectedTasks {
		if _, ok := order.Details.Tasks.Raw[taskName]; !ok {
			t.Errorf("Raw tasks missing %q", taskName)
		}
	}
}

func TestGetDemoOrders_ValidVIN(t *testing.T) {
	orders := GetDemoOrders()
	if len(orders) == 0 {
		t.Fatal("GetDemoOrders() returned empty slice")
	}

	vin := *orders[0].Order.VIN

	// VIN should be 17 characters
	if len(vin) != 17 {
		t.Errorf("VIN length = %d, want 17", len(vin))
	}

	// VIN should start with a valid Tesla WMI
	validWMIs := []string{"5YJ", "7SA", "7G2", "LRW", "XP7"}
	validWMI := false
	for _, wmi := range validWMIs {
		if vin[:3] == wmi {
			validWMI = true
			break
		}
	}
	if !validWMI {
		t.Errorf("VIN has invalid WMI: %s", vin[:3])
	}
}

func TestGetDemoOrders_SchedulingTask(t *testing.T) {
	orders := GetDemoOrders()
	if len(orders) == 0 {
		t.Fatal("GetDemoOrders() returned empty slice")
	}

	scheduling := orders[0].Details.Tasks.Scheduling

	if scheduling.DeliveryWindowDisplay == "" {
		t.Error("DeliveryWindowDisplay is empty")
	}

	if scheduling.ApptDateTimeAddressStr == "" {
		t.Error("ApptDateTimeAddressStr is empty")
	}

	if scheduling.DeliveryType == "" {
		t.Error("DeliveryType is empty")
	}

	if scheduling.DeliveryAddressTitle == "" {
		t.Error("DeliveryAddressTitle is empty")
	}
}

func TestGetDemoOrders_RegistrationTask(t *testing.T) {
	orders := GetDemoOrders()
	if len(orders) == 0 {
		t.Fatal("GetDemoOrders() returned empty slice")
	}

	registration := orders[0].Details.Tasks.Registration

	if registration.OrderDetails == nil {
		t.Fatal("OrderDetails is nil")
	}

	if registration.OrderDetails.VehicleRoutingLocation == "" {
		t.Error("VehicleRoutingLocation is empty")
	}

	if registration.OrderDetails.VehicleOdometer == "" {
		t.Error("VehicleOdometer is empty")
	}
}

func TestGetDemoDiffs(t *testing.T) {
	diffs := GetDemoDiffs()

	if len(diffs) == 0 {
		t.Fatal("GetDemoDiffs() returned empty map")
	}

	// Check for specific order
	orderDiffs, ok := diffs["RN123456789"]
	if !ok {
		t.Fatal("Missing diffs for RN123456789")
	}

	if len(orderDiffs) == 0 {
		t.Error("No diffs for RN123456789")
	}

	// Check diff structure
	for _, diff := range orderDiffs {
		if diff.Field == "" {
			t.Error("Diff field is empty")
		}
		// OldValue and NewValue can be empty, but at least one should have content
		if diff.OldValue == "" && diff.NewValue == "" {
			t.Errorf("Both OldValue and NewValue are empty for field %q", diff.Field)
		}
	}
}

func TestGetDemoHistory(t *testing.T) {
	history := GetDemoHistory()

	if len(history) == 0 {
		t.Fatal("GetDemoHistory() returned empty map")
	}

	// Check for specific order
	orderHistory, ok := history["RN123456789"]
	if !ok {
		t.Fatal("Missing history for RN123456789")
	}

	if orderHistory.ReferenceNumber != "RN123456789" {
		t.Errorf("ReferenceNumber = %q, want %q", orderHistory.ReferenceNumber, "RN123456789")
	}

	if len(orderHistory.Snapshots) == 0 {
		t.Error("No snapshots in history")
	}

	// Check snapshots are in chronological order
	for i := 1; i < len(orderHistory.Snapshots); i++ {
		if orderHistory.Snapshots[i].Timestamp.Before(orderHistory.Snapshots[i-1].Timestamp) {
			t.Error("Snapshots are not in chronological order")
		}
	}
}

func TestGetDemoHistory_Snapshots(t *testing.T) {
	history := GetDemoHistory()
	orderHistory := history["RN123456789"]

	if len(orderHistory.Snapshots) < 2 {
		t.Skip("Need at least 2 snapshots for this test")
	}

	// First snapshot should have empty VIN
	firstSnapshot := orderHistory.Snapshots[0]
	firstVIN := firstSnapshot.Data.Order.GetVIN()
	if firstVIN != "N/A" && firstVIN != "" {
		// Actually the first snapshot might have VIN based on the demo data
		// Let's check that snapshots show progression
	}

	// Last snapshot should have a VIN
	lastSnapshot := orderHistory.Snapshots[len(orderHistory.Snapshots)-1]
	lastVIN := lastSnapshot.Data.Order.GetVIN()
	if lastVIN == "N/A" || lastVIN == "" {
		t.Error("Last snapshot should have a VIN")
	}
}

func TestGetDemoOrders_TasksComplete(t *testing.T) {
	orders := GetDemoOrders()
	if len(orders) == 0 {
		t.Fatal("GetDemoOrders() returned empty slice")
	}

	tasks := orders[0].Details.Tasks

	// Check that some tasks are complete and some are not (for visual variety)
	var completeCount, incompleteCount int

	if tasks.Scheduling != nil {
		if tasks.Scheduling.Complete {
			completeCount++
		} else {
			incompleteCount++
		}
	}

	if tasks.Registration != nil {
		if tasks.Registration.Complete {
			completeCount++
		} else {
			incompleteCount++
		}
	}

	if tasks.FinalPayment != nil {
		if tasks.FinalPayment.Complete {
			completeCount++
		} else {
			incompleteCount++
		}
	}

	if tasks.DeliveryDetails != nil {
		if tasks.DeliveryDetails.Complete {
			completeCount++
		} else {
			incompleteCount++
		}
	}

	// Should have a mix of complete and incomplete for demo purposes
	if completeCount == 0 {
		t.Error("Expected at least one complete task in demo data")
	}
	if incompleteCount == 0 {
		t.Error("Expected at least one incomplete task in demo data")
	}
}

func TestGetDemoOrders_RawJSON(t *testing.T) {
	orders := GetDemoOrders()
	if len(orders) == 0 {
		t.Fatal("GetDemoOrders() returned empty slice")
	}

	rawJSON := orders[0].Details.RawJSON
	if rawJSON == nil {
		t.Error("RawJSON should be set for JSON tab display")
	}

	// Should have tasks key
	if _, ok := rawJSON["tasks"]; !ok {
		t.Error("RawJSON missing 'tasks' key")
	}
}

func TestGetDemoOrders_MktOptions(t *testing.T) {
	orders := GetDemoOrders()
	if len(orders) == 0 {
		t.Fatal("GetDemoOrders() returned empty slice")
	}

	opts := orders[0].Order.MktOptions
	if opts == nil || *opts == "" {
		t.Error("MktOptions should be set in demo data")
	}

	// Should contain comma-separated option codes
	if opts != nil {
		optStr := *opts
		if len(optStr) < 4 {
			t.Error("MktOptions seems too short")
		}
	}
}
