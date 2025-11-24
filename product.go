package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

// Product represents a Zalando product
type Product struct {
	Name     string
	SKU      string
	ConfigID string
	Sizes    []Size
}

// Size represents an available size option
type Size struct {
	Size      string
	SKU       string
	Available bool
}

// extractProductData extracts product information from HTML
func extractProductData(htmlContent string) (map[string]interface{}, error) {
	// Try to find JSON data in script tags
	patterns := []string{
		`window\.__INITIAL_STATE__\s*=\s*({.+?});`,
		`<script[^>]*>window\.__INITIAL_STATE__\s*=\s*({.+?})</script>`,
		`data-product='({.+?})'`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(htmlContent)
		if len(matches) > 1 {
			var data map[string]interface{}
			if err := json.Unmarshal([]byte(matches[1]), &data); err == nil {
				return data, nil
			}
		}
	}

	return nil, fmt.Errorf("kunde inte hitta produktdata i HTML")
}

// extractProductName extracts product name from HTML
func extractProductName(htmlContent string) string {
	patterns := []string{
		`<h1[^>]*class="[^"]*product[^"]*name[^"]*"[^>]*>([^<]+)</h1>`,
		`"productName":"([^"]+)"`,
		`<title>([^|]+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(htmlContent)
		if len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
	}

	return "Okänd produkt"
}

// extractConfigID extracts config_id from HTML
func extractConfigID(htmlContent string) string {
	patterns := []string{
		`"configId":"([^"]+)"`,
		`config_id["\s:]+([A-Z0-9-]+)`,
		`data-config-id="([^"]+)"`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(htmlContent)
		if len(matches) > 1 {
			return matches[1]
		}
	}

	return ""
}

// extractSKU extracts SKU from URL or HTML
func extractSKU(productURL, htmlContent string) string {
	// Try to extract from URL first (format: /product-name-ABC123.html)
	urlPattern := regexp.MustCompile(`([A-Z0-9]{6,})\.(html|htm)`)
	if matches := urlPattern.FindStringSubmatch(productURL); len(matches) > 1 {
		return matches[1]
	}

	// Try to extract from HTML
	patterns := []string{
		`"sku":"([^"]+)"`,
		`"productSku":"([^"]+)"`,
		`data-sku="([^"]+)"`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(htmlContent)
		if len(matches) > 1 {
			return matches[1]
		}
	}

	return ""
}

// extractSizes extracts available sizes from product data
func extractSizes(data map[string]interface{}) []Size {
	var sizes []Size

	// Navigate through the JSON structure to find simples array
	if product, ok := data["product"].(map[string]interface{}); ok {
		if simples, ok := product["simples"].([]interface{}); ok {
			for _, simple := range simples {
				if simpleMap, ok := simple.(map[string]interface{}); ok {
					size := Size{}

					if sizeValue, ok := simpleMap["size"].(string); ok {
						size.Size = sizeValue
					}
					if skuValue, ok := simpleMap["sku"].(string); ok {
						size.SKU = skuValue
					}

					// Check availability
					if availability, ok := simpleMap["availability"].(map[string]interface{}); ok {
						if available, ok := availability["available"].(bool); ok {
							size.Available = available
						}
					} else if stock, ok := simpleMap["stock"].(float64); ok {
						size.Available = stock > 0
					} else {
						// Assume available if no explicit info
						size.Available = true
					}

					sizes = append(sizes, size)
				}
			}
		}
	}

	// Fallback: try alternative structure
	if len(sizes) == 0 {
		if simples, ok := data["simples"].([]interface{}); ok {
			for _, simple := range simples {
				if simpleMap, ok := simple.(map[string]interface{}); ok {
					size := Size{
						Available: true,
					}
					if sizeValue, ok := simpleMap["size"].(string); ok {
						size.Size = sizeValue
					}
					if skuValue, ok := simpleMap["sku"].(string); ok {
						size.SKU = skuValue
					}
					sizes = append(sizes, size)
				}
			}
		}
	}

	return sizes
}

// FetchProduct fetches product information from Zalando
func (hc *HTTPClient) FetchProduct(productURL string) (*Product, error) {
	LogInfo("Hämtar produktinformation...")

	req, err := http.NewRequest("GET", productURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("User-Agent", hc.UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "sv-SE,sv;q=0.9")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")

	resp, err := hc.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch product: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	htmlContent := string(body)

	// Extract product information
	product := &Product{
		Name:     extractProductName(htmlContent),
		SKU:      extractSKU(productURL, htmlContent),
		ConfigID: extractConfigID(htmlContent),
	}

	// Extract sizes
	data, err := extractProductData(htmlContent)
	if err != nil {
		LogWarning("Kunde inte extrahera strukturerad produktdata: %v", err)
		// Continue anyway, maybe we got basic info
	} else {
		product.Sizes = extractSizes(data)
	}

	if product.Name == "Okänd produkt" && product.SKU == "" {
		return nil, fmt.Errorf("kunde inte extrahera produktinformation")
	}

	LogSuccess("Produkt hittad: %s (SKU: %s)", product.Name, product.SKU)
	return product, nil
}

// AddToCartRequest represents the GraphQL mutation for adding to cart
type AddToCartRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

// AddToCart adds a product to the cart using GraphQL
func (hc *HTTPClient) AddToCart(sku, configID string) error {
	LogInfo("Lägger till produkt i varukorgen...")

	// GraphQL mutation for adding to cart
	mutation := `mutation AddToCart($input: AddToCartInput!) {
		addToCart(input: $input) {
			cart {
				id
				items {
					id
					quantity
				}
			}
		}
	}`

	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"sku":      sku,
			"quantity": 1,
		},
	}

	// If we have configID, include it
	if configID != "" {
		variables["input"].(map[string]interface{})["configId"] = configID
	}

	requestBody := AddToCartRequest{
		Query:     mutation,
		Variables: variables,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", "https://www.zalando.se/api/graphql/add-to-cart/", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("User-Agent", hc.UserAgent)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Origin", "https://www.zalando.se")
	req.Header.Set("Referer", "https://www.zalando.se/")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")

	resp, err := hc.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to add to cart: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return fmt.Errorf("kunde inte lägga till i varukorg: %d - %s", resp.StatusCode, string(body))
	}

	LogSuccess("Produkt tillagd i varukorgen!")
	return nil
}
