package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// UpdatePayment updates the payment method to BNPL (Buy Now Pay Later / Faktura)
func (hc *HTTPClient) UpdatePayment() error {
	LogInfo("Uppdaterar betalningsmetod till BNPL (Faktura)...")

	paymentData := map[string]interface{}{
		"payment_method": "bnpl",
		"payment_type":   "INVOICE",
	}

	jsonData, err := json.Marshal(paymentData)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", "https://www.zalando.se/api/checkout/update-payment", bytes.NewBuffer(jsonData))
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
		return fmt.Errorf("failed to update payment: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		return fmt.Errorf("kunde inte uppdatera betalning: %d - %s", resp.StatusCode, string(body))
	}

	LogSuccess("Betalningsmetod uppdaterad till Faktura (BNPL)!")
	return nil
}

// CompletePurchase completes the purchase
func (hc *HTTPClient) CompletePurchase() error {
	LogInfo("Slutför köpet...")

	req, err := http.NewRequest("POST", "https://www.zalando.se/checkout/payment-complete", nil)
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
		return fmt.Errorf("failed to complete purchase: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != 200 && resp.StatusCode != 302 && resp.StatusCode != 303 {
		return fmt.Errorf("kunde inte slutföra köp: %d - %s", resp.StatusCode, string(body))
	}

	LogSuccess("Köp slutfört!")
	return nil
}

// GetPaymentMethods retrieves available payment methods
func (hc *HTTPClient) GetPaymentMethods() ([]string, error) {
	LogInfo("Hämtar tillgängliga betalningsmetoder...")

	req, err := http.NewRequest("GET", "https://www.zalando.se/api/checkout/payment-methods", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("User-Agent", hc.UserAgent)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", "https://www.zalando.se/checkout")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")

	resp, err := hc.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch payment methods: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("kunde inte hämta betalningsmetoder: %d - %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	var methods []string
	if paymentMethods, ok := result["payment_methods"].([]interface{}); ok {
		for _, method := range paymentMethods {
			if methodMap, ok := method.(map[string]interface{}); ok {
				if methodType, ok := methodMap["type"].(string); ok {
					methods = append(methods, methodType)
				}
			}
		}
	}

	return methods, nil
}
