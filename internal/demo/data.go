package demo

import (
	"encoding/json"
	"time"

	"github.com/marcelblijleven/tesla-delivery-tui/internal/model"
)

// GetDemoOrders returns mock order data for demo/recording purposes
func GetDemoOrders() []model.CombinedOrder {
	// Model Y VIN: XP7 (Berlin) + Y (Model Y) + A (SUV LHD) + C + E (Electric) + F (LR AWD) + 9 + T (2026) + B (Berlin) + 123456
	vin := "XP7YACEF9TB123456"
	mktOptions := "APBS,IPB11,PPSW,SC04,MDLY,WY19P,MTY52,STY5S,CPF0,TW01"

	return []model.CombinedOrder{
		{
			Order: model.TeslaOrder{
				ReferenceNumber: "RN123456789",
				OrderStatus:     "BOOKED",
				ModelCode:       "my",
				VIN:             &vin,
				IsB2B:           false,
				IsUsed:          false,
				MktOptions:      &mktOptions,
			},
			Details: model.OrderDetails{
				Tasks: model.OrderTasks{
					Scheduling: &model.SchedulingTask{
						TeslaTask: model.TeslaTask{
							ID:       "scheduling",
							Complete: true,
							Enabled:  true,
							Required: true,
							Order:    1,
							Card: &model.TeslaTaskCard{
								Title:    "Schedule Delivery",
								Subtitle: "Your delivery is scheduled",
							},
						},
						DeliveryWindowDisplay:  "May - Jun 2026",
						ApptDateTimeAddressStr: "June 15, 2026 at 10:00 AM",
						DeliveryType:           "PICKUP_SERVICE_CENTER",
						DeliveryAddressTitle:   "Utrecht - Eendrachtlaan",
					},
					Registration: &model.RegistrationTask{
						TeslaTask: model.TeslaTask{
							ID:       "registration",
							Complete: true,
							Enabled:  true,
							Required: true,
							Order:    2,
							Card: &model.TeslaTaskCard{
								Title:    "Registration",
								Subtitle: "Registration complete",
							},
						},
						OrderDetails: &model.RegistrationOrderDetails{
							VehicleRoutingLocation: "Tilburg Factory",
							VehicleOdometer:        "50",
							VehicleOdometerType:    "km",
							ReservationDate:        "2024-01-15",
							OrderBookedDate:        "2024-03-20",
						},
					},
					FinalPayment: &model.FinalPaymentTask{
						TeslaTask: model.TeslaTask{
							ID:       "finalPayment",
							Complete: false,
							Enabled:  true,
							Required: true,
							Order:    3,
							Card: &model.TeslaTaskCard{
								Title:    "Final Payment",
								Subtitle: "Complete your payment",
							},
						},
						Data: &model.FinalPaymentData{
							ETAToDeliveryCenter: "June 10, 2026",
						},
					},
					DeliveryDetails: &model.DeliveryDetailsTask{
						TeslaTask: model.TeslaTask{
							ID:       "deliveryDetails",
							Complete: true,
							Enabled:  true,
							Required: false,
							Order:    4,
							Card: &model.TeslaTaskCard{
								Title:    "Delivery Details",
								Subtitle: "Review delivery information",
							},
						},
						RegData: &model.DeliveryDetailsRegData{
							ReggieLicensePlate: "AB-123-CD",
						},
					},
					Raw: createDemoTasksRaw(),
				},
				RawJSON: createDemoRawJSON(),
			},
		},
	}
}

// createDemoTasksRaw creates the Raw map for tasks (needed for tasks tab rendering)
func createDemoTasksRaw() map[string]json.RawMessage {
	tasks := map[string]interface{}{
		"scheduling": map[string]interface{}{
			"id":       "scheduling",
			"complete": true,
			"enabled":  true,
			"required": true,
			"order":    1,
			"card": map[string]interface{}{
				"title":    "Schedule Delivery",
				"subtitle": "Your delivery is scheduled",
			},
			"deliveryWindowDisplay":  "May - Jun 2026",
			"apptDateTimeAddressStr": "June 15, 2026 at 10:00 AM",
			"deliveryType":           "PICKUP_SERVICE_CENTER",
			"deliveryAddressTitle":   "Utrecht - Eendrachtlaan",
		},
		"registration": map[string]interface{}{
			"id":       "registration",
			"complete": true,
			"enabled":  true,
			"required": true,
			"order":    2,
			"card": map[string]interface{}{
				"title":    "Registration",
				"subtitle": "Registration complete",
			},
			"orderDetails": map[string]interface{}{
				"vehicleRoutingLocation": "Tilburg Factory",
				"vehicleOdometer":        "50",
				"vehicleOdometerType":    "km",
				"reservationDate":        "2024-01-15",
				"orderBookedDate":        "2024-03-20",
				"reservationAmountReceived": 250,
				"orderAdjustments": []map[string]interface{}{
					{
						"label":  "Referral Credit",
						"amount": -2500,
					},
				},
				"currencyFormat": map[string]interface{}{
					"currencyCode": "EUR",
				},
			},
		},
		"financing": map[string]interface{}{
			"id":       "financing",
			"complete": true,
			"enabled":  true,
			"required": true,
			"order":    7,
			"card": map[string]interface{}{
				"title":        "Financing",
				"subtitle":     "Payment method selected",
				"messageTitle": "Pay With",
				"messageBody":  "Cash",
			},
		},
		"finalPayment": map[string]interface{}{
			"id":       "finalPayment",
			"complete": false,
			"enabled":  true,
			"required": true,
			"order":    3,
			"card": map[string]interface{}{
				"title":    "Final Payment",
				"subtitle": "Complete your payment before delivery",
			},
			"amountDue": 39120,
			"currencyFormat": map[string]interface{}{
				"currencyCode": "EUR",
			},
		},
		"deliveryDetails": map[string]interface{}{
			"id":       "deliveryDetails",
			"complete": true,
			"enabled":  true,
			"required": false,
			"order":    4,
			"card": map[string]interface{}{
				"title":    "Delivery Details",
				"subtitle": "Review your delivery information",
			},
		},
		"tradeIn": map[string]interface{}{
			"id":       "tradeIn",
			"complete": true,
			"enabled":  true,
			"required": false,
			"order":    5,
			"card": map[string]interface{}{
				"title":    "Trade-In",
				"subtitle": "Trade-in vehicle submitted",
			},
			"tradeInVehicle": map[string]interface{}{
				"make":                 "Volkswagen",
				"model":                "Golf",
				"year":                 "2019",
				"vin":                  "WVWZZZ1KZAW123456",
				"trim":                 "Comfortline 1.5 TSI 130pk",
				"mileage":              69500,
				"mileageUnitOfMeasure": "km",
				"condition":            "Fair",
				"licensePlate":         "XY-123-ZZ",
				"tradeInCredit":        10470,
			},
			"currentVehicle": map[string]interface{}{
				"finalOffer": 10470,
			},
			"selectedValuation": map[string]interface{}{
				"valuationExpireDate": "2026-03-15",
			},
		},
		"insurance": map[string]interface{}{
			"id":       "insurance",
			"complete": false,
			"enabled":  true,
			"required": false,
			"order":    6,
			"card": map[string]interface{}{
				"title":    "Insurance",
				"subtitle": "Add insurance before delivery",
			},
		},
	}

	raw := make(map[string]json.RawMessage)
	for name, data := range tasks {
		jsonBytes, _ := json.Marshal(data)
		raw[name] = json.RawMessage(jsonBytes)
	}
	return raw
}

// GetDemoDiffs returns mock diffs showing recent changes
func GetDemoDiffs() map[string][]model.OrderDiff {
	return map[string][]model.OrderDiff{
		"RN123456789": {
			{
				Field:    "Delivery Window",
				OldValue: "Apr - May 2026",
				NewValue: "May - Jun 2026",
			},
			{
				Field:    "Vehicle Location",
				OldValue: "N/A",
				NewValue: "Tilburg Factory",
			},
			{
				Field:    "VIN",
				OldValue: "N/A",
				NewValue: "XP7YACEF9TB123456",
			},
		},
	}
}

// GetDemoHistory returns mock history data
func GetDemoHistory() map[string]*model.OrderHistory {
	vin1 := ""
	vin2 := "XP7YACEF9TB123456"
	mktOptions := "APBS,IPB11,PPSW,SC04,MDLY,WY19P,MTY52,STY5S,CPF0,TW01"

	return map[string]*model.OrderHistory{
		"RN123456789": {
			ReferenceNumber: "RN123456789",
			Snapshots: []model.HistoricalSnapshot{
				{
					Timestamp: time.Now().Add(-72 * time.Hour),
					Data: model.CombinedOrder{
						Order: model.TeslaOrder{
							ReferenceNumber: "RN123456789",
							OrderStatus:     "BOOKED",
							ModelCode:       "my",
							VIN:             &vin1,
							MktOptions:      &mktOptions,
						},
						Details: model.OrderDetails{
							Tasks: model.OrderTasks{
								Scheduling: &model.SchedulingTask{
									DeliveryWindowDisplay: "Apr - May 2026",
								},
							},
						},
					},
				},
				{
					Timestamp: time.Now().Add(-24 * time.Hour),
					Data: model.CombinedOrder{
						Order: model.TeslaOrder{
							ReferenceNumber: "RN123456789",
							OrderStatus:     "BOOKED",
							ModelCode:       "my",
							VIN:             &vin2,
							MktOptions:      &mktOptions,
						},
						Details: model.OrderDetails{
							Tasks: model.OrderTasks{
								Scheduling: &model.SchedulingTask{
									DeliveryWindowDisplay: "May - Jun 2026",
								},
							},
						},
					},
				},
			},
		},
	}
}

func createDemoRawJSON() map[string]interface{} {
	rawJSON := `{
		"tasks": {
			"scheduling": {
				"id": "scheduling",
				"complete": true,
				"enabled": true,
				"required": true,
				"order": 1,
				"deliveryWindowDisplay": "May - Jun 2026",
				"apptDateTimeAddressStr": "June 15, 2026 at 10:00 AM",
				"deliveryType": "PICKUP_SERVICE_CENTER",
				"deliveryAddressTitle": "Utrecht - Eendrachtlaan"
			},
			"registration": {
				"id": "registration",
				"complete": true,
				"enabled": true,
				"required": true,
				"order": 2,
				"orderDetails": {
					"vehicleRoutingLocation": "Tilburg Factory",
					"vehicleOdometer": "50",
					"vehicleOdometerType": "km",
					"reservationDate": "2024-01-15",
					"orderBookedDate": "2024-03-20",
					"reservationAmountReceived": 250,
					"orderAdjustments": [
						{"label": "Referral Credit", "amount": -2500}
					],
					"currencyFormat": {"currencyCode": "EUR"}
				}
			},
			"financing": {
				"id": "financing",
				"complete": true,
				"enabled": true,
				"required": true,
				"order": 7,
				"card": {
					"title": "Financing",
					"subtitle": "Payment method selected",
					"messageTitle": "Pay With",
					"messageBody": "Cash"
				}
			},
			"finalPayment": {
				"id": "finalPayment",
				"complete": false,
				"enabled": true,
				"required": true,
				"order": 3,
				"amountDue": 39120,
				"currencyFormat": {"currencyCode": "EUR"},
				"data": {
					"etaToDeliveryCenter": "June 10, 2026"
				}
			},
			"deliveryDetails": {
				"id": "deliveryDetails",
				"complete": true,
				"enabled": true,
				"required": false,
				"order": 4,
				"regData": {
					"reggieLicensePlate": "AB-123-CD"
				}
			},
			"tradeIn": {
				"id": "tradeIn",
				"complete": true,
				"enabled": true,
				"required": false,
				"order": 5,
				"tradeInVehicle": {
					"make": "Volkswagen",
					"model": "Golf",
					"year": "2019",
					"vin": "WVWZZZ1KZAW123456",
					"trim": "Comfortline 1.5 TSI 130pk",
					"mileage": 69500,
					"mileageUnitOfMeasure": "km",
					"condition": "Fair",
					"licensePlate": "XY-123-ZZ",
					"tradeInCredit": 10470
				},
				"currentVehicle": {
					"finalOffer": 10470
				},
				"selectedValuation": {
					"valuationExpireDate": "2026-03-15"
				}
			},
			"insurance": {
				"id": "insurance",
				"complete": false,
				"enabled": true,
				"required": false,
				"order": 6
			}
		}
	}`

	var result map[string]interface{}
	json.Unmarshal([]byte(rawJSON), &result)
	return result
}
