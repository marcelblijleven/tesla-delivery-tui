package api

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/marcelblijleven/tesla-delivery-tui/internal/model"
)

const (
	ordersAPIURL            = "https://owner-api.teslamotors.com/api/1/users/orders"
	orderDetailsAPITemplate = "https://akamai-apigateway-vfx.tesla.com/tasks?deviceLanguage=en&deviceCountry=US&referenceNumber={ORDER_ID}&appVersion=9.99.9-9999"
)

// GetOrders fetches all orders for the authenticated user
func (c *Client) GetOrders() ([]model.TeslaOrder, error) {
	resp, err := c.Get(ordersAPIURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch orders: %w", err)
	}
	defer resp.Body.Close()

	// Read the body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for API errors
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Tesla API wraps orders in a "response" field
	var ordersResp model.OrdersResponse
	if err := json.Unmarshal(body, &ordersResp); err != nil {
		return nil, fmt.Errorf("failed to decode orders response: %w", err)
	}

	return ordersResp.Response, nil
}

// GetOrderDetails fetches detailed information for a specific order
func (c *Client) GetOrderDetails(referenceNumber string) (*model.OrderDetails, error) {
	url := strings.Replace(orderDetailsAPITemplate, "{ORDER_ID}", referenceNumber, 1)

	resp, err := c.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch order details: %w", err)
	}
	defer resp.Body.Close()

	// Read the body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for API errors
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Store raw JSON for display
	var rawJSON map[string]interface{}
	if err := json.Unmarshal(body, &rawJSON); err != nil {
		return nil, fmt.Errorf("failed to decode raw JSON: %w", err)
	}

	// The order details API returns the tasks object directly
	var rawTasks map[string]json.RawMessage
	if err := json.Unmarshal(body, &rawTasks); err != nil {
		return nil, fmt.Errorf("failed to decode order details response: %w", err)
	}

	// Parse into our structure
	details := &model.OrderDetails{
		RawJSON: rawJSON,
	}

	// Try to parse known task types
	if tasksRaw, ok := rawTasks["tasks"]; ok {
		var tasksMap map[string]json.RawMessage
		if err := json.Unmarshal(tasksRaw, &tasksMap); err == nil {
			// Parse scheduling task
			if schedulingRaw, ok := tasksMap["scheduling"]; ok {
				var scheduling model.SchedulingTask
				if err := json.Unmarshal(schedulingRaw, &scheduling); err == nil {
					details.Tasks.Scheduling = &scheduling
				}
			}

			// Parse registration task
			if registrationRaw, ok := tasksMap["registration"]; ok {
				var registration model.RegistrationTask
				if err := json.Unmarshal(registrationRaw, &registration); err == nil {
					details.Tasks.Registration = &registration
				}
			}

			// Parse final payment task
			if finalPaymentRaw, ok := tasksMap["finalPayment"]; ok {
				var finalPayment model.FinalPaymentTask
				if err := json.Unmarshal(finalPaymentRaw, &finalPayment); err == nil {
					details.Tasks.FinalPayment = &finalPayment
				}
			}

			// Parse delivery details task
			if deliveryDetailsRaw, ok := tasksMap["deliveryDetails"]; ok {
				var deliveryDetails model.DeliveryDetailsTask
				if err := json.Unmarshal(deliveryDetailsRaw, &deliveryDetails); err == nil {
					details.Tasks.DeliveryDetails = &deliveryDetails
				}
			}

			// Store raw for tasks view
			details.Tasks.Raw = tasksMap
		}
	} else {
		// The response might be the tasks directly (without a "tasks" wrapper)
		var scheduling model.SchedulingTask
		if schedulingRaw, ok := rawTasks["scheduling"]; ok {
			if err := json.Unmarshal(schedulingRaw, &scheduling); err == nil {
				details.Tasks.Scheduling = &scheduling
			}
		}

		var registration model.RegistrationTask
		if registrationRaw, ok := rawTasks["registration"]; ok {
			if err := json.Unmarshal(registrationRaw, &registration); err == nil {
				details.Tasks.Registration = &registration
			}
		}

		var finalPayment model.FinalPaymentTask
		if finalPaymentRaw, ok := rawTasks["finalPayment"]; ok {
			if err := json.Unmarshal(finalPaymentRaw, &finalPayment); err == nil {
				details.Tasks.FinalPayment = &finalPayment
			}
		}

		var deliveryDetails model.DeliveryDetailsTask
		if deliveryDetailsRaw, ok := rawTasks["deliveryDetails"]; ok {
			if err := json.Unmarshal(deliveryDetailsRaw, &deliveryDetails); err == nil {
				details.Tasks.DeliveryDetails = &deliveryDetails
			}
		}

		details.Tasks.Raw = rawTasks
	}

	return details, nil
}

// GetAllOrderData fetches all orders with their details
func (c *Client) GetAllOrderData() ([]model.CombinedOrder, error) {
	orders, err := c.GetOrders()
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}

	if len(orders) == 0 {
		return []model.CombinedOrder{}, nil
	}

	combinedOrders := make([]model.CombinedOrder, 0, len(orders))

	for _, order := range orders {
		details, err := c.GetOrderDetails(order.ReferenceNumber)
		if err != nil {
			// Log but continue with other orders
			fmt.Printf("Warning: failed to get details for order %s: %v\n", order.ReferenceNumber, err)
			details = &model.OrderDetails{}
		}

		combinedOrders = append(combinedOrders, model.CombinedOrder{
			Order:   order,
			Details: *details,
		})
	}

	return combinedOrders, nil
}
