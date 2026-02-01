package model

import (
	"testing"
	"time"
)

func TestTeslaTokens_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		want      bool
	}{
		{
			name:      "expired token",
			expiresAt: time.Now().Add(-1 * time.Hour),
			want:      true,
		},
		{
			name:      "valid token",
			expiresAt: time.Now().Add(1 * time.Hour),
			want:      false,
		},
		{
			name:      "just expired",
			expiresAt: time.Now().Add(-1 * time.Second),
			want:      true,
		},
		{
			name:      "zero time (expired)",
			expiresAt: time.Time{},
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := &TeslaTokens{
				ExpiresAt: tt.expiresAt,
			}
			if got := tokens.IsExpired(); got != tt.want {
				t.Errorf("IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTeslaOrder_GetVIN(t *testing.T) {
	vin := "5YJ3E1EA1LF123456"
	emptyVin := ""

	tests := []struct {
		name string
		vin  *string
		want string
	}{
		{
			name: "with VIN",
			vin:  &vin,
			want: vin,
		},
		{
			name: "nil VIN",
			vin:  nil,
			want: "N/A",
		},
		{
			name: "empty VIN",
			vin:  &emptyVin,
			want: "N/A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &TeslaOrder{VIN: tt.vin}
			if got := o.GetVIN(); got != tt.want {
				t.Errorf("GetVIN() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTeslaOrder_GetModelName(t *testing.T) {
	tests := []struct {
		name      string
		modelCode string
		want      string
	}{
		{"Model S lowercase", "ms", "Model S"},
		{"Model S uppercase", "MS", "Model S"},
		{"Model S single char", "s", "Model S"},
		{"Model 3 lowercase", "m3", "Model 3"},
		{"Model 3 uppercase", "M3", "Model 3"},
		{"Model 3 single char", "3", "Model 3"},
		{"Model X lowercase", "mx", "Model X"},
		{"Model X uppercase", "MX", "Model X"},
		{"Model X single char", "x", "Model X"},
		{"Model Y lowercase", "my", "Model Y"},
		{"Model Y uppercase", "MY", "Model Y"},
		{"Model Y single char", "y", "Model Y"},
		{"Cybertruck lowercase", "ct", "Cybertruck"},
		{"Cybertruck uppercase", "CT", "Cybertruck"},
		{"Cybertruck full", "cybertruck", "Cybertruck"},
		{"Unknown model", "unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &TeslaOrder{ModelCode: tt.modelCode}
			if got := o.GetModelName(); got != tt.want {
				t.Errorf("GetModelName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCombinedOrder_GetDeliveryWindow(t *testing.T) {
	tests := []struct {
		name  string
		order CombinedOrder
		want  string
	}{
		{
			name: "with delivery window",
			order: CombinedOrder{
				Details: OrderDetails{
					Tasks: OrderTasks{
						Scheduling: &SchedulingTask{
							DeliveryWindowDisplay: "Jan - Feb 2026",
						},
					},
				},
			},
			want: "Jan - Feb 2026",
		},
		{
			name: "nil scheduling",
			order: CombinedOrder{
				Details: OrderDetails{
					Tasks: OrderTasks{},
				},
			},
			want: "N/A",
		},
		{
			name: "empty delivery window",
			order: CombinedOrder{
				Details: OrderDetails{
					Tasks: OrderTasks{
						Scheduling: &SchedulingTask{
							DeliveryWindowDisplay: "",
						},
					},
				},
			},
			want: "N/A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.order.GetDeliveryWindow(); got != tt.want {
				t.Errorf("GetDeliveryWindow() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCombinedOrder_GetDeliveryAppointment(t *testing.T) {
	tests := []struct {
		name  string
		order CombinedOrder
		want  string
	}{
		{
			name: "with appointment",
			order: CombinedOrder{
				Details: OrderDetails{
					Tasks: OrderTasks{
						Scheduling: &SchedulingTask{
							ApptDateTimeAddressStr: "June 15, 2026 at 10:00 AM",
						},
					},
				},
			},
			want: "June 15, 2026 at 10:00 AM",
		},
		{
			name:  "nil scheduling",
			order: CombinedOrder{},
			want:  "N/A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.order.GetDeliveryAppointment(); got != tt.want {
				t.Errorf("GetDeliveryAppointment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCombinedOrder_GetETAToDeliveryCenter(t *testing.T) {
	tests := []struct {
		name  string
		order CombinedOrder
		want  string
	}{
		{
			name: "with ETA",
			order: CombinedOrder{
				Details: OrderDetails{
					Tasks: OrderTasks{
						FinalPayment: &FinalPaymentTask{
							Data: &FinalPaymentData{
								ETAToDeliveryCenter: "June 10, 2026",
							},
						},
					},
				},
			},
			want: "June 10, 2026",
		},
		{
			name: "nil final payment",
			order: CombinedOrder{},
			want: "N/A",
		},
		{
			name: "nil data",
			order: CombinedOrder{
				Details: OrderDetails{
					Tasks: OrderTasks{
						FinalPayment: &FinalPaymentTask{},
					},
				},
			},
			want: "N/A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.order.GetETAToDeliveryCenter(); got != tt.want {
				t.Errorf("GetETAToDeliveryCenter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCombinedOrder_GetVehicleLocation(t *testing.T) {
	tests := []struct {
		name  string
		order CombinedOrder
		want  string
	}{
		{
			name: "with location",
			order: CombinedOrder{
				Details: OrderDetails{
					Tasks: OrderTasks{
						Registration: &RegistrationTask{
							OrderDetails: &RegistrationOrderDetails{
								VehicleRoutingLocation: "Tilburg Factory",
							},
						},
					},
				},
			},
			want: "Tilburg Factory",
		},
		{
			name:  "nil registration",
			order: CombinedOrder{},
			want:  "N/A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.order.GetVehicleLocation(); got != tt.want {
				t.Errorf("GetVehicleLocation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCombinedOrder_GetOdometer(t *testing.T) {
	tests := []struct {
		name  string
		order CombinedOrder
		want  string
	}{
		{
			name: "with odometer and type",
			order: CombinedOrder{
				Details: OrderDetails{
					Tasks: OrderTasks{
						Registration: &RegistrationTask{
							OrderDetails: &RegistrationOrderDetails{
								VehicleOdometer:     "50",
								VehicleOdometerType: "km",
							},
						},
					},
				},
			},
			want: "50 km",
		},
		{
			name: "with odometer only",
			order: CombinedOrder{
				Details: OrderDetails{
					Tasks: OrderTasks{
						Registration: &RegistrationTask{
							OrderDetails: &RegistrationOrderDetails{
								VehicleOdometer: "100",
							},
						},
					},
				},
			},
			want: "100",
		},
		{
			name:  "nil registration",
			order: CombinedOrder{},
			want:  "N/A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.order.GetOdometer(); got != tt.want {
				t.Errorf("GetOdometer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCombinedOrder_GetLicensePlate(t *testing.T) {
	tests := []struct {
		name  string
		order CombinedOrder
		want  string
	}{
		{
			name: "with license plate",
			order: CombinedOrder{
				Details: OrderDetails{
					Tasks: OrderTasks{
						DeliveryDetails: &DeliveryDetailsTask{
							RegData: &DeliveryDetailsRegData{
								ReggieLicensePlate: "AB-123-CD",
							},
						},
					},
				},
			},
			want: "AB-123-CD",
		},
		{
			name:  "nil delivery details",
			order: CombinedOrder{},
			want:  "N/A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.order.GetLicensePlate(); got != tt.want {
				t.Errorf("GetLicensePlate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCombinedOrder_GetDeliveryType(t *testing.T) {
	tests := []struct {
		name  string
		order CombinedOrder
		want  string
	}{
		{
			name: "with delivery type",
			order: CombinedOrder{
				Details: OrderDetails{
					Tasks: OrderTasks{
						Scheduling: &SchedulingTask{
							DeliveryType: "PICKUP_SERVICE_CENTER",
						},
					},
				},
			},
			want: "PICKUP_SERVICE_CENTER",
		},
		{
			name:  "nil scheduling",
			order: CombinedOrder{},
			want:  "N/A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.order.GetDeliveryType(); got != tt.want {
				t.Errorf("GetDeliveryType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCombinedOrder_GetDeliveryCenter(t *testing.T) {
	tests := []struct {
		name  string
		order CombinedOrder
		want  string
	}{
		{
			name: "with delivery center",
			order: CombinedOrder{
				Details: OrderDetails{
					Tasks: OrderTasks{
						Scheduling: &SchedulingTask{
							DeliveryAddressTitle: "Utrecht - Eendrachtlaan",
						},
					},
				},
			},
			want: "Utrecht - Eendrachtlaan",
		},
		{
			name:  "nil scheduling",
			order: CombinedOrder{},
			want:  "N/A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.order.GetDeliveryCenter(); got != tt.want {
				t.Errorf("GetDeliveryCenter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseAppointment(t *testing.T) {
	tests := []struct {
		name        string
		raw         string
		wantNil     bool
		wantDate    string
		wantTime    string
		wantAddress string
	}{
		{
			name:    "empty string",
			raw:     "",
			wantNil: true,
		},
		{
			name:    "N/A",
			raw:     "N/A",
			wantNil: true,
		},
		{
			name:        "full appointment",
			raw:         "August 15, 2024 at 10:00 AM - Tesla Delivery Center, 123 Electric Ave",
			wantDate:    "August 15, 2024",
			wantTime:    "10:00 AM",
			wantAddress: "Tesla Delivery Center, 123 Electric Ave",
		},
		{
			name:        "date and time only",
			raw:         "June 20, 2026 at 2:30 PM",
			wantDate:    "June 20, 2026",
			wantTime:    "2:30 PM",
			wantAddress: "",
		},
		{
			name:        "date only no time separator",
			raw:         "March 5, 2026",
			wantDate:    "March 5, 2026",
			wantTime:    "",
			wantAddress: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseAppointment(tt.raw)
			if tt.wantNil {
				if got != nil {
					t.Errorf("ParseAppointment(%q) = %+v, want nil", tt.raw, got)
				}
				return
			}
			if got == nil {
				t.Fatalf("ParseAppointment(%q) = nil, want non-nil", tt.raw)
			}
			if got.Date != tt.wantDate {
				t.Errorf("Date = %q, want %q", got.Date, tt.wantDate)
			}
			if got.Time != tt.wantTime {
				t.Errorf("Time = %q, want %q", got.Time, tt.wantTime)
			}
			if got.Address != tt.wantAddress {
				t.Errorf("Address = %q, want %q", got.Address, tt.wantAddress)
			}
		})
	}
}

func TestCombinedOrder_GetReservationDate(t *testing.T) {
	tests := []struct {
		name  string
		order CombinedOrder
		want  string
	}{
		{
			name: "with reservation date",
			order: CombinedOrder{
				Details: OrderDetails{
					Tasks: OrderTasks{
						Registration: &RegistrationTask{
							OrderDetails: &RegistrationOrderDetails{
								ReservationDate: "2024-01-15",
							},
						},
					},
				},
			},
			want: "2024-01-15",
		},
		{
			name:  "nil registration",
			order: CombinedOrder{},
			want:  "N/A",
		},
		{
			name: "nil order details",
			order: CombinedOrder{
				Details: OrderDetails{
					Tasks: OrderTasks{
						Registration: &RegistrationTask{},
					},
				},
			},
			want: "N/A",
		},
		{
			name: "empty reservation date",
			order: CombinedOrder{
				Details: OrderDetails{
					Tasks: OrderTasks{
						Registration: &RegistrationTask{
							OrderDetails: &RegistrationOrderDetails{
								ReservationDate: "",
							},
						},
					},
				},
			},
			want: "N/A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.order.GetReservationDate(); got != tt.want {
				t.Errorf("GetReservationDate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCombinedOrder_GetOrderBookedDate(t *testing.T) {
	tests := []struct {
		name  string
		order CombinedOrder
		want  string
	}{
		{
			name: "with order booked date",
			order: CombinedOrder{
				Details: OrderDetails{
					Tasks: OrderTasks{
						Registration: &RegistrationTask{
							OrderDetails: &RegistrationOrderDetails{
								OrderBookedDate: "2024-02-20",
							},
						},
					},
				},
			},
			want: "2024-02-20",
		},
		{
			name:  "nil registration",
			order: CombinedOrder{},
			want:  "N/A",
		},
		{
			name: "nil order details",
			order: CombinedOrder{
				Details: OrderDetails{
					Tasks: OrderTasks{
						Registration: &RegistrationTask{},
					},
				},
			},
			want: "N/A",
		},
		{
			name: "empty order booked date",
			order: CombinedOrder{
				Details: OrderDetails{
					Tasks: OrderTasks{
						Registration: &RegistrationTask{
							OrderDetails: &RegistrationOrderDetails{
								OrderBookedDate: "",
							},
						},
					},
				},
			},
			want: "N/A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.order.GetOrderBookedDate(); got != tt.want {
				t.Errorf("GetOrderBookedDate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCombinedOrder_GetParsedAppointment(t *testing.T) {
	tests := []struct {
		name        string
		order       CombinedOrder
		wantNil     bool
		wantDate    string
		wantTime    string
		wantAddress string
	}{
		{
			name:    "no scheduling",
			order:   CombinedOrder{},
			wantNil: true,
		},
		{
			name: "with full appointment",
			order: CombinedOrder{
				Details: OrderDetails{
					Tasks: OrderTasks{
						Scheduling: &SchedulingTask{
							ApptDateTimeAddressStr: "July 4, 2026 at 9:00 AM - Tesla Utrecht, Eendrachtlaan 100",
						},
					},
				},
			},
			wantDate:    "July 4, 2026",
			wantTime:    "9:00 AM",
			wantAddress: "Tesla Utrecht, Eendrachtlaan 100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.order.GetParsedAppointment()
			if tt.wantNil {
				if got != nil {
					t.Errorf("GetParsedAppointment() = %+v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Fatalf("GetParsedAppointment() = nil, want non-nil")
			}
			if got.Date != tt.wantDate {
				t.Errorf("Date = %q, want %q", got.Date, tt.wantDate)
			}
			if got.Time != tt.wantTime {
				t.Errorf("Time = %q, want %q", got.Time, tt.wantTime)
			}
			if got.Address != tt.wantAddress {
				t.Errorf("Address = %q, want %q", got.Address, tt.wantAddress)
			}
		})
	}
}

func TestCompareOrders_NoDifferences(t *testing.T) {
	order := CombinedOrder{
		Order: TeslaOrder{
			OrderStatus: "PENDING",
			ModelCode:   "my",
		},
	}

	diffs := CompareOrders(order, order)
	if len(diffs) != 0 {
		t.Errorf("CompareOrders() returned %d diffs for identical orders, want 0", len(diffs))
	}
}

func TestCompareOrders_AllFields(t *testing.T) {
	vin1 := "5YJ3E1EA1LF000001"
	vin2 := "5YJ3E1EA1LF000002"
	opts1 := "OPTION_A,OPTION_B"
	opts2 := "OPTION_A,OPTION_C"

	oldOrder := CombinedOrder{
		Order: TeslaOrder{
			OrderStatus: "PENDING",
			VIN:         &vin1,
			MktOptions:  &opts1,
		},
		Details: OrderDetails{
			Tasks: OrderTasks{
				Scheduling: &SchedulingTask{
					DeliveryWindowDisplay: "Jan - Feb 2026",
					ApptDateTimeAddressStr: "Jan 15, 2026 at 10:00 AM - Tesla Center",
					DeliveryType:          "PICKUP",
					DeliveryAddressTitle:  "Utrecht",
				},
				FinalPayment: &FinalPaymentTask{
					Data: &FinalPaymentData{
						ETAToDeliveryCenter: "Jan 10",
					},
				},
				Registration: &RegistrationTask{
					OrderDetails: &RegistrationOrderDetails{
						VehicleRoutingLocation: "Factory",
						VehicleOdometer:        "10",
						VehicleOdometerType:    "km",
						ReservationDate:        "2024-01-01",
						OrderBookedDate:        "2024-01-05",
					},
				},
				DeliveryDetails: &DeliveryDetailsTask{
					RegData: &DeliveryDetailsRegData{
						ReggieLicensePlate: "AA-111-BB",
					},
				},
			},
		},
	}

	newOrder := CombinedOrder{
		Order: TeslaOrder{
			OrderStatus: "DELIVERED",
			VIN:         &vin2,
			MktOptions:  &opts2,
		},
		Details: OrderDetails{
			Tasks: OrderTasks{
				Scheduling: &SchedulingTask{
					DeliveryWindowDisplay: "Feb - Mar 2026",
					ApptDateTimeAddressStr: "Feb 20, 2026 at 2:00 PM - Tesla Center 2",
					DeliveryType:          "DELIVERY",
					DeliveryAddressTitle:  "Amsterdam",
				},
				FinalPayment: &FinalPaymentTask{
					Data: &FinalPaymentData{
						ETAToDeliveryCenter: "Feb 15",
					},
				},
				Registration: &RegistrationTask{
					OrderDetails: &RegistrationOrderDetails{
						VehicleRoutingLocation: "Port",
						VehicleOdometer:        "50",
						VehicleOdometerType:    "km",
						ReservationDate:        "2024-02-01",
						OrderBookedDate:        "2024-02-10",
					},
				},
				DeliveryDetails: &DeliveryDetailsTask{
					RegData: &DeliveryDetailsRegData{
						ReggieLicensePlate: "CC-222-DD",
					},
				},
			},
		},
	}

	diffs := CompareOrders(oldOrder, newOrder)

	expectedFields := map[string]bool{
		"Order Status":           true,
		"VIN":                    true,
		"Delivery Window":        true,
		"Delivery Appointment":   true,
		"ETA to Delivery Center": true,
		"Vehicle Location":       true,
		"Delivery Method":        true,
		"Delivery Center":        true,
		"Odometer":               true,
		"License Plate":          true,
		"Reservation Date":       true,
		"Order Booked Date":      true,
		"Vehicle Options":        true,
	}

	if len(diffs) != len(expectedFields) {
		t.Errorf("CompareOrders() returned %d diffs, want %d", len(diffs), len(expectedFields))
	}

	for _, diff := range diffs {
		if !expectedFields[diff.Field] {
			t.Errorf("Unexpected diff field: %q", diff.Field)
		}
		delete(expectedFields, diff.Field)
	}

	for field := range expectedFields {
		t.Errorf("Missing diff for field: %q", field)
	}
}

func TestCompareOrders_MktOptions_NilHandling(t *testing.T) {
	opts := "OPTION_A"

	tests := []struct {
		name      string
		oldOpts   *string
		newOpts   *string
		wantDiffs int
	}{
		{"both nil", nil, nil, 0},
		{"old nil new set", nil, &opts, 1},
		{"old set new nil", &opts, nil, 1},
		{"both same", &opts, &opts, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			old := CombinedOrder{Order: TeslaOrder{MktOptions: tt.oldOpts}}
			newOrd := CombinedOrder{Order: TeslaOrder{MktOptions: tt.newOpts}}
			diffs := CompareOrders(old, newOrd)

			optsDiffs := 0
			for _, d := range diffs {
				if d.Field == "Vehicle Options" {
					optsDiffs++
				}
			}
			if optsDiffs != tt.wantDiffs {
				t.Errorf("Vehicle Options diffs = %d, want %d", optsDiffs, tt.wantDiffs)
			}
		})
	}
}
