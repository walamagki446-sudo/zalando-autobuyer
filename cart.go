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
	ZalandoCartURL    = "https://www.zalando.se/api/cart"
	ZalandoCartAddURL = "https://www.zalando.se/api/cart/items"
)

// CartItem represents an item in the cart
type CartItem struct {
	SKU      string `json:"sku"`
	Quantity int    `json:"quantity"`
	Name     string `json:"name,omitempty"`
	Size     string `json:"size,omitempty"`
	Price    string `json:"price,omitempty"`
}

// Cart represents the shopping cart
type Cart struct {
	Items      []CartItem `json:"items"`
	TotalPrice string     `json:"totalPrice"`
	ItemCount  int        `json:"itemCount"`
}

// AddToCartRequest represents the request to add item to cart
type AddToCartRequest struct {
	SKU      string `json:"sku"`
	Quantity int    `json:"quantity"`
}

// CartResponse represents the API response for cart operations
type CartResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
	Cart    *Cart  `json:"cart,omitempty"`
}

// CartManager handles cart operations
type CartManager struct {
	client *HTTPClient
	cart   *Cart
}

// NewCartManager creates a new cart manager
func NewCartManager(client *HTTPClient) *CartManager {
	return &CartManager{
		client: client,
		cart:   &Cart{Items: make([]CartItem, 0)},
	}
}

// GetCart returns the current cart
func (cm *CartManager) GetCart() *Cart {
	return cm.cart
}

// AddToCart adds a product to the cart
func (cm *CartManager) AddToCart(sku string, quantity int) error {
	if quantity <= 0 {
		quantity = 1
	}

	// Prepare the request payload
	payload := map[string]interface{}{
		"sku":      sku,
		"quantity": quantity,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", ZalandoCartAddURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Origin", ZalandoBaseURL)
	req.Header.Set("Referer", ZalandoBaseURL)

	resp, err := cm.client.Do(req, true)
	if err != nil {
		return fmt.Errorf("failed to add to cart: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Check response status
	if resp.StatusCode == http.StatusTooManyRequests {
		ErrorDelay()
		return fmt.Errorf("rate limited, please try again")
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("failed to add to cart: status %d, body: %s", resp.StatusCode, string(body))
	}

	// Try to parse response
	var cartResp map[string]interface{}
	if err := json.Unmarshal(body, &cartResp); err == nil {
		// Check for errors in response
		if errMsg, ok := cartResp["error"].(string); ok && errMsg != "" {
			return fmt.Errorf("cart error: %s", errMsg)
		}
		if errMsg, ok := cartResp["message"].(string); ok && strings.Contains(strings.ToLower(errMsg), "error") {
			return fmt.Errorf("cart error: %s", errMsg)
		}
	}

	// Update local cart state
	cm.cart.Items = append(cm.cart.Items, CartItem{
		SKU:      sku,
		Quantity: quantity,
	})

	return nil
}

// RemoveFromCart removes an item from the cart
func (cm *CartManager) RemoveFromCart(sku string) error {
	deleteURL := fmt.Sprintf("%s/%s", ZalandoCartAddURL, sku)

	req, err := http.NewRequest("DELETE", deleteURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Origin", ZalandoBaseURL)
	req.Header.Set("Referer", ZalandoBaseURL)

	resp, err := cm.client.Do(req, true)
	if err != nil {
		return fmt.Errorf("failed to remove from cart: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to remove from cart: status %d, body: %s", resp.StatusCode, string(body))
	}

	// Update local cart state
	newItems := make([]CartItem, 0)
	for _, item := range cm.cart.Items {
		if item.SKU != sku {
			newItems = append(newItems, item)
		}
	}
	cm.cart.Items = newItems

	return nil
}

// ClearCart clears all items from the cart
func (cm *CartManager) ClearCart() error {
	for _, item := range cm.cart.Items {
		if err := cm.RemoveFromCart(item.SKU); err != nil {
			return fmt.Errorf("failed to clear cart: %w", err)
		}
		HumanDelay()
	}
	cm.cart.Items = make([]CartItem, 0)
	return nil
}

// FetchCart fetches the current cart from Zalando
func (cm *CartManager) FetchCart() (*Cart, error) {
	req, err := http.NewRequest("GET", ZalandoCartURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := cm.client.Do(req, true)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch cart: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var cartData map[string]interface{}
	if err := json.Unmarshal(body, &cartData); err != nil {
		return nil, fmt.Errorf("failed to parse cart: %w", err)
	}

	// Parse cart items from response
	cart := &Cart{Items: make([]CartItem, 0)}

	if items, ok := cartData["items"].([]interface{}); ok {
		for _, item := range items {
			if itemMap, ok := item.(map[string]interface{}); ok {
				cartItem := CartItem{}
				if sku, ok := itemMap["sku"].(string); ok {
					cartItem.SKU = sku
				}
				if name, ok := itemMap["name"].(string); ok {
					cartItem.Name = name
				}
				if quantity, ok := itemMap["quantity"].(float64); ok {
					cartItem.Quantity = int(quantity)
				}
				cart.Items = append(cart.Items, cartItem)
			}
		}
	}

	if total, ok := cartData["totalPrice"].(string); ok {
		cart.TotalPrice = total
	}

	cart.ItemCount = len(cart.Items)
	cm.cart = cart

	return cart, nil
}

// IsEmpty returns true if the cart is empty
func (cm *CartManager) IsEmpty() bool {
	return len(cm.cart.Items) == 0
}

// GetItemCount returns the number of items in the cart
func (cm *CartManager) GetItemCount() int {
	return len(cm.cart.Items)
}

// HasItem checks if a specific SKU is in the cart
func (cm *CartManager) HasItem(sku string) bool {
	for _, item := range cm.cart.Items {
		if item.SKU == sku {
			return true
		}
	}
	return false
}

// UpdateQuantity updates the quantity of an item in the cart
func (cm *CartManager) UpdateQuantity(sku string, quantity int) error {
	if quantity <= 0 {
		return cm.RemoveFromCart(sku)
	}

	updateURL := fmt.Sprintf("%s/%s", ZalandoCartAddURL, sku)

	payload := map[string]interface{}{
		"quantity": quantity,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("PUT", updateURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Origin", ZalandoBaseURL)
	req.Header.Set("Referer", ZalandoBaseURL)

	resp, err := cm.client.Do(req, true)
	if err != nil {
		return fmt.Errorf("failed to update quantity: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update quantity: status %d, body: %s", resp.StatusCode, string(body))
	}

	// Update local cart state
	for i, item := range cm.cart.Items {
		if item.SKU == sku {
			cm.cart.Items[i].Quantity = quantity
			break
		}
	}

	return nil
}
