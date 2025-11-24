package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/walamagki446-sudo/zalando-autobuyer/internal/models"
)

// Client handles API requests to Zalando
type Client struct {
	httpClient *http.Client
	session    *models.Session
	baseURL    string
}

// NewClient creates a new API client
func NewClient(session *models.Session, baseURL string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		session: session,
		baseURL: baseURL,
	}
}

// AddToCart adds an item to the cart via API
func (c *Client) AddToCart(sku, size string) error {
	url := c.baseURL + "/api/cart-gateway/carts"

	payload := map[string]interface{}{
		"sku":      sku,
		"quantity": 1,
		"size":     size,
	}

	return c.doRequest("POST", url, payload, nil)
}

// GetCart retrieves the current cart
func (c *Client) GetCart() (interface{}, error) {
	url := c.baseURL + "/api/cart-gateway/carts"

	var result interface{}
	err := c.doRequest("GET", url, nil, &result)
	return result, err
}

// CreateCheckout creates a checkout session
func (c *Client) CreateCheckout() (string, error) {
	url := c.baseURL + "/checkout/v3/fetch-or-create-checkout-trampoline?checkout_variant=PHYSICAL"

	var result map[string]interface{}
	err := c.doRequest("GET", url, nil, &result)
	if err != nil {
		return "", err
	}

	// Extract checkout ID from response
	if checkoutID, ok := result["checkout_id"].(string); ok {
		return checkoutID, nil
	}

	return "", fmt.Errorf("checkout ID not found in response")
}

// SearchPickupPoints searches for available pickup points
func (c *Client) SearchPickupPoints(address string) ([]models.PickupPoint, error) {
	url := c.baseURL + "/api/checkout/search-pickup-points-by-address"

	payload := map[string]interface{}{
		"address": address,
	}

	var result []models.PickupPoint
	err := c.doRequest("POST", url, payload, &result)
	return result, err
}

// SelectPickupPoint selects a pickup point
func (c *Client) SelectPickupPoint(pointID string) error {
	url := c.baseURL + "/api/checkout/select-pickup-point"

	payload := map[string]interface{}{
		"pickup_point_id": pointID,
	}

	return c.doRequest("POST", url, payload, nil)
}

// NextStep proceeds to next checkout step
func (c *Client) NextStep() error {
	url := c.baseURL + "/api/checkout/next-step"
	return c.doRequest("POST", url, nil, nil)
}

// UpdatePayment updates the payment method
func (c *Client) UpdatePayment(paymentType string, sessionID string) error {
	url := c.baseURL + "/api/checkout/update-payment"

	payload := map[string]interface{}{
		"payment_type": paymentType,
		"session_id":   sessionID,
	}

	return c.doRequest("POST", url, payload, nil)
}

// BuyNow executes the final purchase
func (c *Client) BuyNow() error {
	url := c.baseURL + "/api/checkout/buy-now"
	return c.doRequest("POST", url, nil, nil)
}

// PaymentComplete completes the payment
func (c *Client) PaymentComplete() error {
	url := c.baseURL + "/checkout/payment-complete"
	return c.doRequest("POST", url, nil, nil)
}

// doRequest performs an HTTP request with session cookies
func (c *Client) doRequest(method, url string, payload interface{}, result interface{}) error {
	var body io.Reader

	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	// Add session cookies
	if c.session != nil && c.session.Cookies != nil {
		for _, cookie := range c.session.Cookies {
			req.AddCookie(cookie)
		}
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response if result is provided
	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}
