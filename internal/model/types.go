package model

import (
	"encoding/json"
	"strings"
	"time"
)

// TeslaTokens represents OAuth2 tokens from Tesla's API
type TeslaTokens struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresIn    int       `json:"expires_in"`
	Scope        string    `json:"scope"`
	TokenType    string    `json:"token_type"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// IsExpired checks if the access token has expired
func (t *TeslaTokens) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// TeslaOrder represents a Tesla vehicle order
type TeslaOrder struct {
	ReferenceNumber  string  `json:"referenceNumber"`
	OrderStatus      string  `json:"orderStatus"`
	ModelCode        string  `json:"modelCode"`
	VIN              *string `json:"vin,omitempty"`
	IsB2B            bool    `json:"isB2b"`
	OwnerCompanyName *string `json:"ownerCompanyName,omitempty"`
	IsUsed           bool    `json:"isUsed"`
	MktOptions       *string `json:"mktOptions,omitempty"`
}

// GetVIN returns the VIN or "N/A" if not available
func (o *TeslaOrder) GetVIN() string {
	if o.VIN != nil && *o.VIN != "" {
		return *o.VIN
	}
	return "N/A"
}

// GetModelName returns a human-readable model name
func (o *TeslaOrder) GetModelName() string {
	switch o.ModelCode {
	case "ms", "MS", "s", "S":
		return "Model S"
	case "m3", "M3", "3":
		return "Model 3"
	case "mx", "MX", "x", "X":
		return "Model X"
	case "my", "MY", "y", "Y":
		return "Model Y"
	case "ct", "CT", "cybertruck", "CYBERTRUCK":
		return "Cybertruck"
	default:
		return o.ModelCode
	}
}

// TeslaTaskCard represents a task card's display information
type TeslaTaskCard struct {
	Title        string `json:"title"`
	Subtitle     string `json:"subtitle"`
	MessageBody  string `json:"messageBody,omitempty"`
	MessageTitle string `json:"messageTitle,omitempty"`
	ButtonText   *struct {
		CTA string `json:"cta"`
	} `json:"buttonText,omitempty"`
	Target string `json:"target,omitempty"`
}

// TeslaTask represents a task in the order process
type TeslaTask struct {
	ID       string         `json:"id"`
	Complete bool           `json:"complete"`
	Enabled  bool           `json:"enabled"`
	Required bool           `json:"required"`
	Order    int            `json:"order"`
	Card     *TeslaTaskCard `json:"card,omitempty"`
}

// SchedulingTask represents scheduling-specific task data
type SchedulingTask struct {
	TeslaTask
	DeliveryWindowDisplay      string `json:"deliveryWindowDisplay,omitempty"`
	ApptDateTimeAddressStr     string `json:"apptDateTimeAddressStr,omitempty"`
	DeliveryType               string `json:"deliveryType,omitempty"`
	DeliveryAddressTitle       string `json:"deliveryAddressTitle,omitempty"`
	IsSelfSchedulingAvailable  bool   `json:"isSelfSchedulingAvailable,omitempty"`
	SelfSchedulingURL          string `json:"selfSchedulingUrl,omitempty"`
}

// AppointmentDetails holds parsed appointment information
type AppointmentDetails struct {
	Date    string
	Time    string
	Address string
}

// ParseAppointment parses the apptDateTimeAddressStr into structured parts.
// Expected format: "August 15, 2024 at 10:00 AM - Tesla Delivery Center, 123 Electric Ave"
func ParseAppointment(raw string) *AppointmentDetails {
	if raw == "" || raw == "N/A" {
		return nil
	}

	parts := strings.SplitN(raw, " at ", 2)
	date := strings.TrimSpace(parts[0])

	if len(parts) < 2 {
		return &AppointmentDetails{Date: date}
	}

	timeAndAddress := strings.SplitN(parts[1], " - ", 2)
	apptTime := strings.TrimSpace(timeAndAddress[0])
	address := ""
	if len(timeAndAddress) > 1 {
		address = strings.TrimSpace(timeAndAddress[1])
	}

	return &AppointmentDetails{
		Date:    date,
		Time:    apptTime,
		Address: address,
	}
}

// RegistrationOrderDetails contains order details from registration task
type RegistrationOrderDetails struct {
	VehicleRoutingLocation string `json:"vehicleRoutingLocation,omitempty"`
	VehicleOdometer        string `json:"vehicleOdometer,omitempty"`
	VehicleOdometerType    string `json:"vehicleOdometerType,omitempty"`
	ReservationDate        string `json:"reservationDate,omitempty"`
	OrderBookedDate        string `json:"orderBookedDate,omitempty"`
}

// RegistrationTask represents registration-specific task data
type RegistrationTask struct {
	TeslaTask
	OrderDetails *RegistrationOrderDetails `json:"orderDetails,omitempty"`
}

// FinalPaymentData contains payment-related data
type FinalPaymentData struct {
	ETAToDeliveryCenter string `json:"etaToDeliveryCenter,omitempty"`
}

// FinalPaymentTask represents final payment task data
type FinalPaymentTask struct {
	TeslaTask
	Data *FinalPaymentData `json:"data,omitempty"`
}

// DeliveryDetailsRegData contains registration data
type DeliveryDetailsRegData struct {
	ReggieLicensePlate string `json:"reggieLicensePlate,omitempty"`
}

// DeliveryDetailsTask represents delivery details task
type DeliveryDetailsTask struct {
	TeslaTask
	RegData *DeliveryDetailsRegData `json:"regData,omitempty"`
}

// OrderTasks contains all the tasks associated with an order
type OrderTasks struct {
	Scheduling       *SchedulingTask      `json:"scheduling,omitempty"`
	Registration     *RegistrationTask    `json:"registration,omitempty"`
	FinalPayment     *FinalPaymentTask    `json:"finalPayment,omitempty"`
	DeliveryDetails  *DeliveryDetailsTask `json:"deliveryDetails,omitempty"`
	// Generic map for other tasks we might not have typed
	Raw map[string]json.RawMessage `json:"-"`
}

// OrderDetails represents the detailed information about an order
type OrderDetails struct {
	Tasks   OrderTasks             `json:"tasks"`
	RawJSON map[string]interface{} `json:"-"` // Full raw JSON for display
}

// CombinedOrder combines basic order info with detailed task data
type CombinedOrder struct {
	Order   TeslaOrder   `json:"order"`
	Details OrderDetails `json:"details"`
}

// GetDeliveryWindow returns the delivery window display string
func (c *CombinedOrder) GetDeliveryWindow() string {
	if c.Details.Tasks.Scheduling != nil && c.Details.Tasks.Scheduling.DeliveryWindowDisplay != "" {
		return c.Details.Tasks.Scheduling.DeliveryWindowDisplay
	}
	return "N/A"
}

// GetDeliveryAppointment returns the appointment date/time/address
func (c *CombinedOrder) GetDeliveryAppointment() string {
	if c.Details.Tasks.Scheduling != nil && c.Details.Tasks.Scheduling.ApptDateTimeAddressStr != "" {
		return c.Details.Tasks.Scheduling.ApptDateTimeAddressStr
	}
	return "N/A"
}

// GetETAToDeliveryCenter returns the ETA to delivery center
func (c *CombinedOrder) GetETAToDeliveryCenter() string {
	if c.Details.Tasks.FinalPayment != nil && c.Details.Tasks.FinalPayment.Data != nil {
		return c.Details.Tasks.FinalPayment.Data.ETAToDeliveryCenter
	}
	return "N/A"
}

// GetVehicleLocation returns the vehicle routing location
func (c *CombinedOrder) GetVehicleLocation() string {
	if c.Details.Tasks.Registration != nil && c.Details.Tasks.Registration.OrderDetails != nil {
		return c.Details.Tasks.Registration.OrderDetails.VehicleRoutingLocation
	}
	return "N/A"
}

// GetDeliveryType returns the delivery type
func (c *CombinedOrder) GetDeliveryType() string {
	if c.Details.Tasks.Scheduling != nil && c.Details.Tasks.Scheduling.DeliveryType != "" {
		return c.Details.Tasks.Scheduling.DeliveryType
	}
	return "N/A"
}

// GetDeliveryCenter returns the delivery center name
func (c *CombinedOrder) GetDeliveryCenter() string {
	if c.Details.Tasks.Scheduling != nil && c.Details.Tasks.Scheduling.DeliveryAddressTitle != "" {
		return c.Details.Tasks.Scheduling.DeliveryAddressTitle
	}
	return "N/A"
}

// GetOdometer returns the vehicle odometer reading
func (c *CombinedOrder) GetOdometer() string {
	if c.Details.Tasks.Registration != nil && c.Details.Tasks.Registration.OrderDetails != nil {
		od := c.Details.Tasks.Registration.OrderDetails
		if od.VehicleOdometer != "" {
			if od.VehicleOdometerType != "" {
				return od.VehicleOdometer + " " + od.VehicleOdometerType
			}
			return od.VehicleOdometer
		}
	}
	return "N/A"
}

// GetLicensePlate returns the assigned license plate
func (c *CombinedOrder) GetLicensePlate() string {
	if c.Details.Tasks.DeliveryDetails != nil && c.Details.Tasks.DeliveryDetails.RegData != nil {
		return c.Details.Tasks.DeliveryDetails.RegData.ReggieLicensePlate
	}
	return "N/A"
}

// HistoricalSnapshot represents a point-in-time snapshot of order data
type HistoricalSnapshot struct {
	Timestamp time.Time     `json:"timestamp"`
	Data      CombinedOrder `json:"data"`
}

// OrderHistory contains the history of snapshots for an order
type OrderHistory struct {
	ReferenceNumber string               `json:"referenceNumber"`
	Snapshots       []HistoricalSnapshot `json:"snapshots"`
}

// OrderDiff represents a change between two snapshots
type OrderDiff struct {
	Field    string      `json:"field"`
	OldValue interface{} `json:"oldValue"`
	NewValue interface{} `json:"newValue"`
}

// GetReservationDate returns the reservation date
func (c *CombinedOrder) GetReservationDate() string {
	if c.Details.Tasks.Registration != nil && c.Details.Tasks.Registration.OrderDetails != nil {
		if d := c.Details.Tasks.Registration.OrderDetails.ReservationDate; d != "" {
			return d
		}
	}
	return "N/A"
}

// GetOrderBookedDate returns the order booked date
func (c *CombinedOrder) GetOrderBookedDate() string {
	if c.Details.Tasks.Registration != nil && c.Details.Tasks.Registration.OrderDetails != nil {
		if d := c.Details.Tasks.Registration.OrderDetails.OrderBookedDate; d != "" {
			return d
		}
	}
	return "N/A"
}

// GetParsedAppointment returns structured appointment details
func (c *CombinedOrder) GetParsedAppointment() *AppointmentDetails {
	return ParseAppointment(c.GetDeliveryAppointment())
}

// CompareOrders compares two CombinedOrders and returns the differences
func CompareOrders(old, new CombinedOrder) []OrderDiff {
	var diffs []OrderDiff

	addDiff := func(field string, oldVal, newVal interface{}) {
		oldStr, _ := oldVal.(string)
		newStr, _ := newVal.(string)
		if oldStr != newStr {
			diffs = append(diffs, OrderDiff{Field: field, OldValue: oldVal, NewValue: newVal})
		}
	}

	addDiff("Order Status", old.Order.OrderStatus, new.Order.OrderStatus)
	addDiff("VIN", old.Order.GetVIN(), new.Order.GetVIN())
	addDiff("Delivery Window", old.GetDeliveryWindow(), new.GetDeliveryWindow())
	addDiff("Delivery Appointment", old.GetDeliveryAppointment(), new.GetDeliveryAppointment())
	addDiff("ETA to Delivery Center", old.GetETAToDeliveryCenter(), new.GetETAToDeliveryCenter())
	addDiff("Vehicle Location", old.GetVehicleLocation(), new.GetVehicleLocation())
	addDiff("Delivery Method", old.GetDeliveryType(), new.GetDeliveryType())
	addDiff("Delivery Center", old.GetDeliveryCenter(), new.GetDeliveryCenter())
	addDiff("Odometer", old.GetOdometer(), new.GetOdometer())
	addDiff("License Plate", old.GetLicensePlate(), new.GetLicensePlate())
	addDiff("Reservation Date", old.GetReservationDate(), new.GetReservationDate())
	addDiff("Order Booked Date", old.GetOrderBookedDate(), new.GetOrderBookedDate())

	// Compare MktOptions via pointer
	oldOpts := "N/A"
	newOpts := "N/A"
	if old.Order.MktOptions != nil {
		oldOpts = *old.Order.MktOptions
	}
	if new.Order.MktOptions != nil {
		newOpts = *new.Order.MktOptions
	}
	if oldOpts != newOpts {
		diffs = append(diffs, OrderDiff{Field: "Vehicle Options", OldValue: oldOpts, NewValue: newOpts})
	}

	return diffs
}

// OrdersResponse represents the API response for orders endpoint
type OrdersResponse struct {
	Response []TeslaOrder `json:"response"`
}
