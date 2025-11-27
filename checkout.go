package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	ZalandoCheckoutURL           = "https://www.zalando.se/checkout"
	ZalandoCheckoutAPIURL        = "https://www.zalando.se/api/checkout"
	ZalandoPickupPointsURL       = "https://www.zalando.se/api/checkout/search-pickup-points-by-address"
	ZalandoSelectPickupPointURL  = "https://www.zalando.se/api/checkout/select-pickup-point"
	ZalandoSelectAddressURL      = "https://www.zalando.se/api/checkout/select-address"
	ZalandoSelectPaymentURL      = "https://www.zalando.se/api/checkout/select-payment"
	ZalandoConfirmOrderURL       = "https://www.zalando.se/api/checkout/confirm"
)

// PickupPoint represents a pickup point
type PickupPoint struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Address   string  `json:"address"`
	City      string  `json:"city"`
	ZipCode   string  `json:"zipCode"`
	Carrier   string  `json:"carrier"`
	CarrierID string  `json:"carrierId"`
	Distance  float64 `json:"distance"`
	IsInstabox bool   `json:"isInstabox"`
}

// Address represents a delivery address
type Address struct {
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
	Street     string `json:"street"`
	Additional string `json:"additional,omitempty"`
	ZipCode    string `json:"zipCode"`
	City       string `json:"city"`
	Country    string `json:"country"`
	Phone      string `json:"phone,omitempty"`
	Email      string `json:"email,omitempty"`
}

// CheckoutState represents the current checkout state
type CheckoutState struct {
	AddressSelected     bool
	PickupPointSelected bool
	PaymentSelected     bool
	OrderConfirmed      bool
	SelectedAddress     *Address
	SelectedPickupPoint *PickupPoint
	PaymentMethod       string
	OrderNumber         string
}

// CheckoutManager handles checkout operations
type CheckoutManager struct {
	client *HTTPClient
	state  *CheckoutState
}

// NewCheckoutManager creates a new checkout manager
func NewCheckoutManager(client *HTTPClient) *CheckoutManager {
	return &CheckoutManager{
		client: client,
		state:  &CheckoutState{},
	}
}

// GetState returns the current checkout state
func (cm *CheckoutManager) GetState() *CheckoutState {
	return cm.state
}

// InitCheckout initializes the checkout process
func (cm *CheckoutManager) InitCheckout() error {
	req, err := http.NewRequest("GET", ZalandoCheckoutURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := cm.client.Do(req, false)
	if err != nil {
		return fmt.Errorf("failed to initialize checkout: %w", err)
	}
	defer resp.Body.Close()

	// Read body to get any CSRF tokens
	io.Copy(io.Discard, resp.Body)

	// Extract CSRF token from cookies
	csrf := cm.client.GetCookieValue(ZalandoBaseURL, "frsx")
	if csrf != "" {
		cm.client.SetCSRFToken(csrf)
	}

	return nil
}

// SetAddress sets the delivery address
func (cm *CheckoutManager) SetAddress(address *Address) error {
	payload := map[string]interface{}{
		"address": map[string]interface{}{
			"firstName":  address.FirstName,
			"lastName":   address.LastName,
			"street":     address.Street,
			"additional": address.Additional,
			"zipCode":    address.ZipCode,
			"city":       address.City,
			"country":    address.Country,
			"phone":      address.Phone,
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal address: %w", err)
	}

	req, err := http.NewRequest("POST", ZalandoSelectAddressURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Origin", ZalandoBaseURL)
	req.Header.Set("Referer", ZalandoCheckoutURL)

	resp, err := cm.client.Do(req, true)
	if err != nil {
		return fmt.Errorf("failed to set address: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to set address: status %d, body: %s", resp.StatusCode, string(body))
	}

	cm.state.AddressSelected = true
	cm.state.SelectedAddress = address

	return nil
}

// SearchPickupPoints searches for pickup points near an address
func (cm *CheckoutManager) SearchPickupPoints(address *Address) ([]PickupPoint, error) {
	payload := map[string]interface{}{
		"address": map[string]interface{}{
			"street":  address.Street,
			"zipCode": address.ZipCode,
			"city":    address.City,
			"country": address.Country,
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", ZalandoPickupPointsURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Origin", ZalandoBaseURL)
	req.Header.Set("Referer", ZalandoCheckoutURL)

	resp, err := cm.client.Do(req, true)
	if err != nil {
		return nil, fmt.Errorf("failed to search pickup points: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("failed to search pickup points: status %d, body: %s", resp.StatusCode, string(body))
	}

	var responseData map[string]interface{}
	if err := json.Unmarshal(body, &responseData); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return cm.parsePickupPoints(responseData)
}

// parsePickupPoints parses pickup points from API response
func (cm *CheckoutManager) parsePickupPoints(data map[string]interface{}) ([]PickupPoint, error) {
	pickupPoints := make([]PickupPoint, 0)

	// Try different response structures
	var pointsData []interface{}

	if points, ok := data["pickupPoints"].([]interface{}); ok {
		pointsData = points
	} else if points, ok := data["servicePoints"].([]interface{}); ok {
		pointsData = points
	} else if points, ok := data["points"].([]interface{}); ok {
		pointsData = points
	} else if points, ok := data["data"].([]interface{}); ok {
		pointsData = points
	}

	for _, point := range pointsData {
		if pointMap, ok := point.(map[string]interface{}); ok {
			pp := cm.parsePickupPoint(pointMap)
			pickupPoints = append(pickupPoints, pp)
		}
	}

	return pickupPoints, nil
}

// parsePickupPoint parses a single pickup point
func (cm *CheckoutManager) parsePickupPoint(data map[string]interface{}) PickupPoint {
	pp := PickupPoint{}

	if id, ok := data["id"].(string); ok {
		pp.ID = id
	}
	if name, ok := data["name"].(string); ok {
		pp.Name = name
	}
	if carrier, ok := data["carrier"].(string); ok {
		pp.Carrier = carrier
		pp.IsInstabox = strings.ToUpper(carrier) == "INS" || strings.ToLower(carrier) == "instabox"
	}
	if carrierID, ok := data["carrierId"].(string); ok {
		pp.CarrierID = carrierID
		if pp.CarrierID == "INS" {
			pp.IsInstabox = true
		}
	}
	if distance, ok := data["distance"].(float64); ok {
		pp.Distance = distance
	}

	// Parse address
	if addr, ok := data["address"].(map[string]interface{}); ok {
		if street, ok := addr["street"].(string); ok {
			pp.Address = street
		}
		if city, ok := addr["city"].(string); ok {
			pp.City = city
		}
		if zipCode, ok := addr["zipCode"].(string); ok {
			pp.ZipCode = zipCode
		}
	} else {
		if street, ok := data["street"].(string); ok {
			pp.Address = street
		}
		if city, ok := data["city"].(string); ok {
			pp.City = city
		}
		if zipCode, ok := data["zipCode"].(string); ok {
			pp.ZipCode = zipCode
		}
	}

	return pp
}

// GetInstaboxPickupPoints filters for Instabox pickup points only
func (cm *CheckoutManager) GetInstaboxPickupPoints(pickupPoints []PickupPoint) []PickupPoint {
	instaboxPoints := make([]PickupPoint, 0)
	for _, pp := range pickupPoints {
		if pp.IsInstabox {
			instaboxPoints = append(instaboxPoints, pp)
		}
	}
	return instaboxPoints
}

// SelectPickupPoint selects a pickup point for delivery
func (cm *CheckoutManager) SelectPickupPoint(pickupPoint *PickupPoint) error {
	payload := map[string]interface{}{
		"pickupPointId": pickupPoint.ID,
		"carrier":       pickupPoint.Carrier,
		"carrierId":     pickupPoint.CarrierID,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", ZalandoSelectPickupPointURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Origin", ZalandoBaseURL)
	req.Header.Set("Referer", ZalandoCheckoutURL)

	resp, err := cm.client.Do(req, true)
	if err != nil {
		return fmt.Errorf("failed to select pickup point: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to select pickup point: status %d, body: %s", resp.StatusCode, string(body))
	}

	cm.state.PickupPointSelected = true
	cm.state.SelectedPickupPoint = pickupPoint

	return nil
}

// SelectFakturaPayment selects Faktura as payment method
func (cm *CheckoutManager) SelectFakturaPayment() error {
	payload := map[string]interface{}{
		"paymentMethod": "INVOICE",
		"paymentType":   "invoice",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", ZalandoSelectPaymentURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Origin", ZalandoBaseURL)
	req.Header.Set("Referer", ZalandoCheckoutURL)

	resp, err := cm.client.Do(req, true)
	if err != nil {
		return fmt.Errorf("failed to select payment: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to select payment: status %d, body: %s", resp.StatusCode, string(body))
	}

	cm.state.PaymentSelected = true
	cm.state.PaymentMethod = "Faktura"

	return nil
}

// ConfirmOrder confirms and places the order
func (cm *CheckoutManager) ConfirmOrder() (string, error) {
	payload := map[string]interface{}{
		"confirm": true,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", ZalandoConfirmOrderURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Origin", ZalandoBaseURL)
	req.Header.Set("Referer", ZalandoCheckoutURL)

	resp, err := cm.client.Do(req, true)
	if err != nil {
		return "", fmt.Errorf("failed to confirm order: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("failed to confirm order: status %d, body: %s", resp.StatusCode, string(body))
	}

	var responseData map[string]interface{}
	if err := json.Unmarshal(body, &responseData); err == nil {
		if orderNumber, ok := responseData["orderNumber"].(string); ok {
			cm.state.OrderConfirmed = true
			cm.state.OrderNumber = orderNumber
			return orderNumber, nil
		}
		if orderId, ok := responseData["orderId"].(string); ok {
			cm.state.OrderConfirmed = true
			cm.state.OrderNumber = orderId
			return orderId, nil
		}
	}

	cm.state.OrderConfirmed = true
	return "Order confirmed", nil
}

// FullCheckout performs the complete checkout process
func (cm *CheckoutManager) FullCheckout(address *Address) (string, error) {
	// Step 1: Initialize checkout
	if err := cm.InitCheckout(); err != nil {
		return "", fmt.Errorf("checkout init failed: %w", err)
	}
	HumanDelay()

	// Step 2: Set address
	if err := cm.SetAddress(address); err != nil {
		return "", fmt.Errorf("address selection failed: %w", err)
	}
	HumanDelay()

	// Step 3: Search for Instabox pickup points
	pickupPoints, err := cm.SearchPickupPoints(address)
	if err != nil {
		return "", fmt.Errorf("pickup point search failed: %w", err)
	}
	HumanDelay()

	// Step 4: Filter and select Instabox
	instaboxPoints := cm.GetInstaboxPickupPoints(pickupPoints)
	if len(instaboxPoints) == 0 {
		return "", fmt.Errorf("no Instabox pickup points found")
	}

	// Select the first/closest Instabox
	if err := cm.SelectPickupPoint(&instaboxPoints[0]); err != nil {
		return "", fmt.Errorf("pickup point selection failed: %w", err)
	}
	HumanDelay()

	// Step 5: Select Faktura payment
	if err := cm.SelectFakturaPayment(); err != nil {
		return "", fmt.Errorf("payment selection failed: %w", err)
	}
	HumanDelay()

	// Step 6: Confirm order
	orderNumber, err := cm.ConfirmOrder()
	if err != nil {
		return "", fmt.Errorf("order confirmation failed: %w", err)
	}

	return orderNumber, nil
}

// ValidateCheckoutReady validates that checkout can proceed
func (cm *CheckoutManager) ValidateCheckoutReady() error {
	if cm.state.OrderConfirmed {
		return fmt.Errorf("order already confirmed")
	}
	return nil
}

// GetSelectedPickupPoint returns the selected pickup point
func (cm *CheckoutManager) GetSelectedPickupPoint() *PickupPoint {
	return cm.state.SelectedPickupPoint
}

// IsOrderComplete returns true if order has been confirmed
func (cm *CheckoutManager) IsOrderComplete() bool {
	return cm.state.OrderConfirmed
}

// GetOrderNumber returns the order number
func (cm *CheckoutManager) GetOrderNumber() string {
	return cm.state.OrderNumber
}
