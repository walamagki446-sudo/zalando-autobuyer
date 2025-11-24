package models

import "net/http"

// Session holds session data for API calls and authentication
type Session struct {
	Cookies      []*http.Cookie
	SessionID    string
	CheckoutID   string
	PaymentToken string
	Email        string
}

// PickupPoint represents a delivery pickup location
type PickupPoint struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Address  string  `json:"address"`
	Distance float64 `json:"distance"`
	Type     string  `json:"type"`
}

// ProductInfo holds product details
type ProductInfo struct {
	URL          string
	SKU          string
	Name         string
	Price        string
	Sizes        []string
	SelectedSize string
}

// CartItem represents an item in the shopping cart
type CartItem struct {
	SKU      string `json:"sku"`
	Quantity int    `json:"quantity"`
	Size     string `json:"size"`
}

// PaymentMethod represents a payment option
type PaymentMethod struct {
	Type      string `json:"type"`
	ID        string `json:"id"`
	Name      string `json:"name"`
	Available bool   `json:"available"`
}
