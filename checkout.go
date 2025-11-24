package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
)

// CheckoutSession represents a checkout session
type CheckoutSession struct {
	ID               string
	PaymentSessionID string
}

// PickupPoint represents a pickup point location
type PickupPoint struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Provider string  `json:"provider"`
	Address  string  `json:"address"`
	Distance float64 `json:"distance"`
	Lat      float64 `json:"latitude"`
	Lon      float64 `json:"longitude"`
}

// Address represents a delivery address
type Address struct {
	Street     string `json:"street"`
	City       string `json:"city"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

// CreateCheckoutSession creates a new checkout session
func (hc *HTTPClient) CreateCheckoutSession() (*CheckoutSession, error) {
	LogInfo("Skapar checkout-session...")

	req, err := http.NewRequest("GET", "https://www.zalando.se/checkout/v3/fetch-or-create-checkout-trampoline?checkout_variant=PHYSICAL", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("User-Agent", hc.UserAgent)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", "https://www.zalando.se/cart")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")

	resp, err := hc.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create checkout: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("kunde inte skapa checkout: %d - %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	session := &CheckoutSession{}

	// Extract checkout ID from response
	if checkoutID, ok := result["checkout_id"].(string); ok {
		session.ID = checkoutID
	} else if id, ok := result["id"].(string); ok {
		session.ID = id
	}

	// Extract payment session ID if available
	if payment, ok := result["payment"].(map[string]interface{}); ok {
		if sessionID, ok := payment["session_id"].(string); ok {
			session.PaymentSessionID = sessionID
		}
	}

	LogSuccess("Checkout-session skapad: %s", session.ID)
	return session, nil
}

// SearchPickupPoints searches for pickup points near the given address
func (hc *HTTPClient) SearchPickupPoints(address Address) ([]PickupPoint, error) {
	LogInfo("Söker efter upphämtningsställen...")

	searchData := map[string]interface{}{
		"address": map[string]string{
			"street":      address.Street,
			"city":        address.City,
			"postal_code": address.PostalCode,
			"country":     address.Country,
		},
		"limit": 20,
	}

	jsonData, err := json.Marshal(searchData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", "https://www.zalando.se/api/checkout/search-pickup-points-by-address", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("User-Agent", hc.UserAgent)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Origin", "https://www.zalando.se")
	req.Header.Set("Referer", "https://www.zalando.se/checkout")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")

	resp, err := hc.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to search pickup points: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("kunde inte söka upphämtningsställen: %d - %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	var pickupPoints []PickupPoint
	allowedProviders := []string{"instabox", "budbee"}

	if points, ok := result["pickup_points"].([]interface{}); ok {
		for _, p := range points {
			if point, ok := p.(map[string]interface{}); ok {
				pp := PickupPoint{}

				if id, ok := point["id"].(string); ok {
					pp.ID = id
				}
				if name, ok := point["name"].(string); ok {
					pp.Name = name
				}
				if provider, ok := point["provider"].(string); ok {
					pp.Provider = strings.ToLower(provider)
				}
				if address, ok := point["address"].(string); ok {
					pp.Address = address
				}
				if distance, ok := point["distance"].(float64); ok {
					pp.Distance = distance
				}
				if lat, ok := point["latitude"].(float64); ok {
					pp.Lat = lat
				}
				if lon, ok := point["longitude"].(float64); ok {
					pp.Lon = lon
				}

				// Only include Instabox or Budbee pickup points
				isAllowed := false
				for _, allowed := range allowedProviders {
					if strings.Contains(strings.ToLower(pp.Provider), allowed) {
						isAllowed = true
						break
					}
				}

				if isAllowed {
					pickupPoints = append(pickupPoints, pp)
				}
			}
		}
	}

	LogSuccess("Hittade %d upphämtningsställen (endast Instabox/Budbee)", len(pickupPoints))
	return pickupPoints, nil
}

// SelectBestPickupPoint selects the best pickup point from Instabox/Budbee only
func SelectBestPickupPoint(points []PickupPoint, preferFarthest bool) *PickupPoint {
	if len(points) == 0 {
		return nil
	}

	// All points should already be Instabox or Budbee from filtering
	LogInfo("Väljer från %d upphämtningsställen (Instabox/Budbee)", len(points))

	// Find closest or farthest based on preference
	var selected *PickupPoint
	extremeDistance := math.Inf(-1)
	if !preferFarthest {
		extremeDistance = math.Inf(1)
	}

	for i := range points {
		if preferFarthest {
			if points[i].Distance > extremeDistance {
				extremeDistance = points[i].Distance
				selected = &points[i]
			}
		} else {
			if points[i].Distance < extremeDistance {
				extremeDistance = points[i].Distance
				selected = &points[i]
			}
		}
	}

	return selected
}

// SelectPickupPoint selects a pickup point for delivery
func (hc *HTTPClient) SelectPickupPoint(pickupPointID string) error {
	LogInfo("Väljer upphämtningsställe...")

	selectData := map[string]interface{}{
		"pickup_point_id": pickupPointID,
	}

	jsonData, err := json.Marshal(selectData)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", "https://www.zalando.se/api/checkout/select-pickup-point", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("User-Agent", hc.UserAgent)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Origin", "https://www.zalando.se")
	req.Header.Set("Referer", "https://www.zalando.se/checkout")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")

	resp, err := hc.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to select pickup point: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		return fmt.Errorf("kunde inte välja upphämtningsställe: %d - %s", resp.StatusCode, string(body))
	}

	LogSuccess("Upphämtningsställe valt!")
	return nil
}

// NextStep proceeds to the next checkout step
func (hc *HTTPClient) NextStep() error {
	LogInfo("Går vidare till nästa steg...")

	req, err := http.NewRequest("POST", "https://www.zalando.se/api/checkout/next-step", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("User-Agent", hc.UserAgent)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Origin", "https://www.zalando.se")
	req.Header.Set("Referer", "https://www.zalando.se/checkout")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")

	resp, err := hc.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to proceed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		return fmt.Errorf("kunde inte gå vidare: %d - %s", resp.StatusCode, string(body))
	}

	LogSuccess("Fortsätter till betalning...")
	return nil
}
