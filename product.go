package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

// ProductSize represents a product size with availability
type ProductSize struct {
	Size      string `json:"size"`
	SKU       string `json:"sku"`
	Available bool   `json:"available"`
	Stock     string `json:"stock"`
}

// Product represents a Zalando product
type Product struct {
	Name        string        `json:"name"`
	Brand       string        `json:"brand"`
	Price       string        `json:"price"`
	Currency    string        `json:"currency"`
	ProductURL  string        `json:"url"`
	ImageURL    string        `json:"image"`
	ProductID   string        `json:"productId"`
	Sizes       []ProductSize `json:"sizes"`
	Description string        `json:"description"`
}

// JSONLDProduct represents JSON-LD product data
type JSONLDProduct struct {
	Context     string      `json:"@context"`
	Type        string      `json:"@type"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Brand       BrandInfo   `json:"brand"`
	Image       interface{} `json:"image"`
	Offers      interface{} `json:"offers"`
	SKU         string      `json:"sku"`
}

// BrandInfo represents brand information
type BrandInfo struct {
	Type string `json:"@type"`
	Name string `json:"name"`
}

// Offer represents a product offer
type Offer struct {
	Type          string `json:"@type"`
	Price         string `json:"price"`
	PriceCurrency string `json:"priceCurrency"`
	Availability  string `json:"availability"`
	SKU           string `json:"sku"`
	URL           string `json:"url"`
}

// ProductFetcher handles product-related operations
type ProductFetcher struct {
	client *HTTPClient
}

// NewProductFetcher creates a new product fetcher
func NewProductFetcher(client *HTTPClient) *ProductFetcher {
	return &ProductFetcher{client: client}
}

// FetchProduct fetches product details from a Zalando URL
func (pf *ProductFetcher) FetchProduct(productURL string) (*Product, error) {
	req, err := http.NewRequest("GET", productURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := pf.client.Do(req, false)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch product: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	product, err := pf.parseProductPage(string(body), productURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse product: %w", err)
	}

	return product, nil
}

// parseProductPage parses the product page HTML to extract product info
func (pf *ProductFetcher) parseProductPage(html, productURL string) (*Product, error) {
	product := &Product{
		ProductURL: productURL,
		Sizes:      make([]ProductSize, 0),
	}

	// Extract JSON-LD data
	jsonLDData := pf.extractJSONLD(html)
	if jsonLDData != nil {
		product.Name = jsonLDData.Name
		product.Description = jsonLDData.Description
		if jsonLDData.Brand.Name != "" {
			product.Brand = jsonLDData.Brand.Name
		}
		product.ProductID = jsonLDData.SKU

		// Parse offers for size information
		sizes := pf.parseOffers(jsonLDData.Offers)
		if len(sizes) > 0 {
			product.Sizes = sizes
		}
	}

	// Try to extract from __NEXT_DATA__ or similar embedded JSON
	if len(product.Sizes) == 0 {
		sizes := pf.extractSizesFromNextData(html)
		if len(sizes) > 0 {
			product.Sizes = sizes
		}
	}

	// Extract from inline JavaScript
	if len(product.Sizes) == 0 {
		sizes := pf.extractSizesFromScript(html)
		if len(sizes) > 0 {
			product.Sizes = sizes
		}
	}

	// Extract price
	if product.Price == "" {
		product.Price = pf.extractPrice(html)
	}

	// Extract product name if not found
	if product.Name == "" {
		product.Name = pf.extractProductName(html)
	}

	// Extract brand if not found
	if product.Brand == "" {
		product.Brand = pf.extractBrand(html)
	}

	return product, nil
}

// extractJSONLD extracts JSON-LD data from the page
func (pf *ProductFetcher) extractJSONLD(html string) *JSONLDProduct {
	// Find JSON-LD script tags
	pattern := regexp.MustCompile(`<script[^>]*type="application/ld\+json"[^>]*>([\s\S]*?)</script>`)
	matches := pattern.FindAllStringSubmatch(html, -1)

	for _, match := range matches {
		if len(match) > 1 {
			jsonStr := strings.TrimSpace(match[1])
			var data JSONLDProduct
			if err := json.Unmarshal([]byte(jsonStr), &data); err == nil {
				if data.Type == "Product" {
					return &data
				}
			}

			// Try parsing as array
			var dataArray []JSONLDProduct
			if err := json.Unmarshal([]byte(jsonStr), &dataArray); err == nil {
				for _, d := range dataArray {
					if d.Type == "Product" {
						return &d
					}
				}
			}
		}
	}

	return nil
}

// parseOffers parses offers from JSON-LD data
func (pf *ProductFetcher) parseOffers(offersData interface{}) []ProductSize {
	sizes := make([]ProductSize, 0)

	if offersData == nil {
		return sizes
	}

	// Handle single offer
	if offerMap, ok := offersData.(map[string]interface{}); ok {
		size := pf.parseOfferMap(offerMap)
		if size.SKU != "" {
			sizes = append(sizes, size)
		}
	}

	// Handle array of offers
	if offerArray, ok := offersData.([]interface{}); ok {
		for _, offer := range offerArray {
			if offerMap, ok := offer.(map[string]interface{}); ok {
				size := pf.parseOfferMap(offerMap)
				if size.SKU != "" {
					sizes = append(sizes, size)
				}
			}
		}
	}

	return sizes
}

// parseOfferMap parses a single offer map
func (pf *ProductFetcher) parseOfferMap(offerMap map[string]interface{}) ProductSize {
	size := ProductSize{}

	if sku, ok := offerMap["sku"].(string); ok {
		size.SKU = sku
		size.Size = pf.extractSizeFromSKU(sku)
	}

	if availability, ok := offerMap["availability"].(string); ok {
		size.Available = strings.Contains(strings.ToLower(availability), "instock")
		if size.Available {
			size.Stock = "In Stock"
		} else {
			size.Stock = "Out of Stock"
		}
	}

	return size
}

// extractSizeFromSKU extracts size from SKU string
func (pf *ProductFetcher) extractSizeFromSKU(sku string) string {
	// Shoe size pattern: ends with numbers like 40, 41, 42
	shoeSizePattern := regexp.MustCompile(`(\d{2,3}(?:[.,]\d)?)$`)
	if matches := shoeSizePattern.FindStringSubmatch(sku); len(matches) > 1 {
		return matches[1]
	}

	// Clothing size pattern: ends with letters like M, L, XL, XXL
	clothingSizePattern := regexp.MustCompile(`(XXS|XS|S|M|L|XL|XXL|XXXL|3XL|4XL|5XL)$`)
	if matches := clothingSizePattern.FindStringSubmatch(sku); len(matches) > 1 {
		return matches[1]
	}

	// Try to find size in the middle of SKU separated by underscore or dash
	underscorePattern := regexp.MustCompile(`[_-]([A-Z]{1,4}|[0-9]{2,3}(?:[.,]\d)?)[_-]`)
	if matches := underscorePattern.FindStringSubmatch(sku); len(matches) > 1 {
		return matches[1]
	}

	return sku
}

// extractSizesFromNextData extracts sizes from __NEXT_DATA__ script
func (pf *ProductFetcher) extractSizesFromNextData(html string) []ProductSize {
	sizes := make([]ProductSize, 0)

	// Find __NEXT_DATA__ script
	pattern := regexp.MustCompile(`<script[^>]*id="__NEXT_DATA__"[^>]*>([\s\S]*?)</script>`)
	matches := pattern.FindStringSubmatch(html)

	if len(matches) < 2 {
		return sizes
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(matches[1]), &data); err != nil {
		return sizes
	}

	// Navigate to product data
	sizes = pf.findSizesInJSON(data)
	return sizes
}

// findSizesInJSON recursively searches for size information in JSON
func (pf *ProductFetcher) findSizesInJSON(data interface{}) []ProductSize {
	sizes := make([]ProductSize, 0)

	switch v := data.(type) {
	case map[string]interface{}:
		// Check for size array keys
		for key, value := range v {
			lowerKey := strings.ToLower(key)
			if lowerKey == "sizes" || lowerKey == "simples" || lowerKey == "units" {
				if arr, ok := value.([]interface{}); ok {
					for _, item := range arr {
						if itemMap, ok := item.(map[string]interface{}); ok {
							size := pf.parseSizeItem(itemMap)
							if size.Size != "" || size.SKU != "" {
								sizes = append(sizes, size)
							}
						}
					}
				}
			}
		}

		// Recurse into nested objects if no sizes found
		if len(sizes) == 0 {
			for _, value := range v {
				found := pf.findSizesInJSON(value)
				if len(found) > 0 {
					sizes = append(sizes, found...)
				}
			}
		}
	case []interface{}:
		for _, item := range v {
			found := pf.findSizesInJSON(item)
			if len(found) > 0 {
				sizes = append(sizes, found...)
			}
		}
	}

	return sizes
}

// parseSizeItem parses a size item from JSON
func (pf *ProductFetcher) parseSizeItem(item map[string]interface{}) ProductSize {
	size := ProductSize{}

	// Extract size name
	if sizeName, ok := item["size"].(string); ok {
		size.Size = sizeName
	} else if sizeName, ok := item["sizeLabel"].(string); ok {
		size.Size = sizeName
	} else if sizeName, ok := item["displayName"].(string); ok {
		size.Size = sizeName
	} else if sizeName, ok := item["name"].(string); ok {
		size.Size = sizeName
	}

	// Extract SKU
	if sku, ok := item["sku"].(string); ok {
		size.SKU = sku
	} else if sku, ok := item["simpleSku"].(string); ok {
		size.SKU = sku
	} else if sku, ok := item["id"].(string); ok {
		size.SKU = sku
	}

	// Extract availability
	if available, ok := item["available"].(bool); ok {
		size.Available = available
	} else if available, ok := item["inStock"].(bool); ok {
		size.Available = available
	} else if stock, ok := item["stock"].(float64); ok {
		size.Available = stock > 0
	} else if offer, ok := item["offer"].(map[string]interface{}); ok {
		if stock, ok := offer["stock"].(map[string]interface{}); ok {
			if quantity, ok := stock["quantity"].(float64); ok {
				size.Available = quantity > 0
			}
		}
	}

	if size.Available {
		size.Stock = "In Stock"
	} else {
		size.Stock = "Out of Stock"
	}

	// If size is empty but SKU exists, extract size from SKU
	if size.Size == "" && size.SKU != "" {
		size.Size = pf.extractSizeFromSKU(size.SKU)
	}

	return size
}

// extractSizesFromScript extracts sizes from inline scripts
func (pf *ProductFetcher) extractSizesFromScript(html string) []ProductSize {
	sizes := make([]ProductSize, 0)

	// Look for size data in various script patterns
	patterns := []string{
		`"simples":\s*(\[[^\]]+\])`,
		`"sizes":\s*(\[[^\]]+\])`,
		`"availableSizes":\s*(\[[^\]]+\])`,
	}

	for _, patternStr := range patterns {
		pattern := regexp.MustCompile(patternStr)
		matches := pattern.FindStringSubmatch(html)
		if len(matches) > 1 {
			var sizeData []map[string]interface{}
			if err := json.Unmarshal([]byte(matches[1]), &sizeData); err == nil {
				for _, item := range sizeData {
					size := pf.parseSizeItem(item)
					if size.Size != "" || size.SKU != "" {
						sizes = append(sizes, size)
					}
				}
				if len(sizes) > 0 {
					return sizes
				}
			}
		}
	}

	return sizes
}

// extractPrice extracts price from HTML
func (pf *ProductFetcher) extractPrice(html string) string {
	patterns := []string{
		`"price":\s*"?([0-9,.]+)"?`,
		`class="[^"]*price[^"]*"[^>]*>([0-9,.\s]+\s*(?:kr|SEK|€))`,
		`itemprop="price"[^>]*content="([^"]+)"`,
	}

	for _, patternStr := range patterns {
		pattern := regexp.MustCompile(patternStr)
		matches := pattern.FindStringSubmatch(html)
		if len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
	}

	return ""
}

// extractProductName extracts product name from HTML
func (pf *ProductFetcher) extractProductName(html string) string {
	patterns := []string{
		`<h1[^>]*>([^<]+)</h1>`,
		`"name":\s*"([^"]+)"`,
		`itemprop="name"[^>]*>([^<]+)<`,
	}

	for _, patternStr := range patterns {
		pattern := regexp.MustCompile(patternStr)
		matches := pattern.FindStringSubmatch(html)
		if len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
	}

	return ""
}

// extractBrand extracts brand from HTML
func (pf *ProductFetcher) extractBrand(html string) string {
	patterns := []string{
		`"brand":\s*(?:{"name":\s*)?"([^"]+)"`,
		`itemprop="brand"[^>]*>([^<]+)<`,
		`class="[^"]*brand[^"]*"[^>]*>([^<]+)<`,
	}

	for _, patternStr := range patterns {
		pattern := regexp.MustCompile(patternStr)
		matches := pattern.FindStringSubmatch(html)
		if len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
	}

	return ""
}

// ValidateProductURL validates if the URL is a valid Zalando product URL
func ValidateProductURL(urlStr string) bool {
	// Must be a valid URL with zalando domain and have a product path
	if !strings.Contains(urlStr, "zalando.") {
		return false
	}
	// Should have a path after the domain (i.e., not just the homepage)
	parts := strings.SplitN(urlStr, "zalando.", 2)
	if len(parts) < 2 {
		return false
	}
	afterDomain := parts[1]
	// Check if there's content after the TLD (e.g., ".se/product")
	slashIdx := strings.Index(afterDomain, "/")
	if slashIdx == -1 {
		return false
	}
	// Should have something after the first slash
	return len(afterDomain) > slashIdx+1
}

// GetAvailableSizes returns only available sizes
func (p *Product) GetAvailableSizes() []ProductSize {
	available := make([]ProductSize, 0)
	for _, size := range p.Sizes {
		if size.Available {
			available = append(available, size)
		}
	}
	return available
}

// GetSizeBySKU returns a size by its SKU
func (p *Product) GetSizeBySKU(sku string) *ProductSize {
	for _, size := range p.Sizes {
		if size.SKU == sku {
			return &size
		}
	}
	return nil
}

// GetSizeByName returns a size by its display name
func (p *Product) GetSizeByName(name string) *ProductSize {
	name = strings.ToUpper(strings.TrimSpace(name))
	for _, size := range p.Sizes {
		if strings.ToUpper(size.Size) == name {
			return &size
		}
	}
	return nil
}
