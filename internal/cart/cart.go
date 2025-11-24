package cart

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/walamagki446-sudo/zalando-autobuyer/internal/browser"
)

const (
	sizePickerButton  = "#picker-trigger"
	sizeLabel         = "#label-element"
	addToCartButton   = "button[data-testid='pdp_add-to-cart-button']"
	sizeOptionsParent = "[role='listbox']"
	sizeOption        = "[role='option']"
)

// DetectAvailableSizes extracts all available sizes from the product page
func DetectAvailableSizes(ctx context.Context) ([]string, error) {
	log.Println("📦 Fetching available sizes...")

	// Click size picker to open dropdown
	if err := browser.ClickElement(ctx, sizePickerButton); err != nil {
		return nil, fmt.Errorf("failed to click size picker: %w", err)
	}

	// Wait for dropdown to appear
	if err := browser.WaitForElement(ctx, sizeOptionsParent, 5*time.Second); err != nil {
		return nil, fmt.Errorf("size dropdown did not appear: %w", err)
	}

	// Wait a bit for all options to render
	time.Sleep(500 * time.Millisecond)

	// Extract all size options
	var sizes []string
	err := chromedp.Run(ctx,
		chromedp.Evaluate(`
			Array.from(document.querySelectorAll('[role="option"]'))
				.map(el => el.textContent.trim())
				.filter(text => text.length > 0)
		`, &sizes),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to extract sizes: %w", err)
	}

	if len(sizes) == 0 {
		return nil, fmt.Errorf("no sizes found")
	}

	log.Printf("📏 Found %d sizes: %v", len(sizes), sizes)
	return sizes, nil
}

// SelectSize selects a specific size from the dropdown
func SelectSize(ctx context.Context, size string) error {
	log.Printf("✅ Selecting size: %s", size)

	// Find and click the size option
	selector := fmt.Sprintf("[role='option']:has-text('%s')", size)

	// Try multiple strategies to select the size
	var err error

	// Strategy 1: Direct click on option
	err = chromedp.Run(ctx,
		chromedp.Click(selector, chromedp.ByQuery),
	)

	if err != nil {
		// Strategy 2: Use JavaScript to click
		script := fmt.Sprintf(`
			const options = document.querySelectorAll('[role="option"]');
			for (let opt of options) {
				if (opt.textContent.trim() === '%s') {
					opt.click();
					return true;
				}
			}
			return false;
		`, size)

		var clicked bool
		err = chromedp.Run(ctx,
			chromedp.Evaluate(script, &clicked),
		)

		if err != nil || !clicked {
			return fmt.Errorf("failed to select size %s: %w", size, err)
		}
	}

	// Wait for selection to register
	time.Sleep(1 * time.Second)

	log.Printf("✅ Size %s selected!", size)
	return nil
}

// AddToCart clicks the "Handla" (Add to Cart) button
func AddToCart(ctx context.Context) error {
	log.Println("🛒 Clicking 'Handla' button...")

	// Try multiple selectors for the add to cart button
	selectors := []string{
		"button[data-testid='pdp_add-to-cart-button']",
		"button:has-text('Handla')",
		"button:has-text('LÄGG TILL I KUNDVAGNEN')",
		"button:has-text('Add to cart')",
		"button[type='button']:has-text('Handla')",
	}

	var lastErr error
	for _, selector := range selectors {
		err := chromedp.Run(ctx,
			chromedp.Click(selector, chromedp.ByQuery),
		)

		if err == nil {
			log.Println("✅ Added to cart!")
			time.Sleep(2 * time.Second)
			return nil
		}
		lastErr = err
	}

	// If all selectors failed, try JavaScript approach
	script := `
		const buttons = document.querySelectorAll('button');
		for (let btn of buttons) {
			const text = btn.textContent.toLowerCase();
			if (text.includes('handla') || text.includes('add to cart') || text.includes('lägg till')) {
				btn.click();
				return true;
			}
		}
		return false;
	`

	var clicked bool
	err := chromedp.Run(ctx,
		chromedp.Evaluate(script, &clicked),
	)

	if err != nil || !clicked {
		return fmt.Errorf("failed to click add to cart button: %w", lastErr)
	}

	log.Println("✅ Added to cart!")
	time.Sleep(2 * time.Second)
	return nil
}

// FindSizeInList checks if a size exists in the available sizes
func FindSizeInList(sizes []string, targetSize string) bool {
	targetLower := strings.ToLower(strings.TrimSpace(targetSize))

	for _, size := range sizes {
		if strings.ToLower(strings.TrimSpace(size)) == targetLower {
			return true
		}
	}

	return false
}
