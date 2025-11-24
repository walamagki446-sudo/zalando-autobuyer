package checkout

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/walamagki446-sudo/zalando-autobuyer/internal/api"
	"github.com/walamagki446-sudo/zalando-autobuyer/internal/browser"
	"github.com/walamagki446-sudo/zalando-autobuyer/internal/models"
)

// NavigateToCheckout navigates to the checkout page
func NavigateToCheckout(ctx context.Context, baseURL string) error {
	log.Println("🔄 Navigating to checkout...")

	checkoutURL := baseURL + "/cart"
	if err := browser.Navigate(ctx, checkoutURL); err != nil {
		return fmt.Errorf("failed to navigate to checkout: %w", err)
	}

	// Wait for page to load
	time.Sleep(3 * time.Second)

	// Click checkout button
	checkoutSelectors := []string{
		"button:has-text('Till kassan')",
		"button:has-text('Checkout')",
		"button:has-text('Gå till kassan')",
		"a[href*='checkout']",
	}

	for _, selector := range checkoutSelectors {
		err := chromedp.Run(ctx,
			chromedp.Click(selector, chromedp.ByQuery),
		)
		if err == nil {
			log.Println("✅ Navigated to checkout!")
			time.Sleep(3 * time.Second)
			return nil
		}
	}

	// Try JavaScript approach
	script := `
		const buttons = document.querySelectorAll('button, a');
		for (let btn of buttons) {
			const text = btn.textContent.toLowerCase();
			if (text.includes('kassan') || text.includes('checkout')) {
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
		return fmt.Errorf("failed to navigate to checkout")
	}

	log.Println("✅ Navigated to checkout!")
	time.Sleep(3 * time.Second)
	return nil
}

// SelectBestPickupPoint selects the best available pickup point
func SelectBestPickupPoint(points []models.PickupPoint, userAddress string) *models.PickupPoint {
	if len(points) == 0 {
		return nil
	}

	log.Printf("📍 Found %d pickup points", len(points))

	// Priority 1: Instabox
	for _, p := range points {
		if strings.Contains(strings.ToLower(p.Name), "instabox") {
			log.Printf("✓ Selected Instabox: %s", p.Name)
			return &p
		}
	}

	// Priority 2: Budbee
	for _, p := range points {
		if strings.Contains(strings.ToLower(p.Name), "budbee") {
			log.Printf("✓ Selected Budbee: %s", p.Name)
			return &p
		}
	}

	// Priority 3: Closest by distance
	sort.Slice(points, func(i, j int) bool {
		return points[i].Distance < points[j].Distance
	})

	log.Printf("✓ Selected closest: %s (%.2f km)", points[0].Name, points[0].Distance)
	return &points[0]
}

// SelectPickupPointOnPage selects a pickup point on the checkout page
func SelectPickupPointOnPage(ctx context.Context, client *api.Client, address string) error {
	log.Println("📍 Searching for pickup points...")

	// Try to find pickup points via API
	points, err := client.SearchPickupPoints(address)
	if err != nil {
		log.Printf("⚠️  API search failed, will use page selection: %v", err)
		return selectPickupPointViaPage(ctx)
	}

	// Select best pickup point
	selectedPoint := SelectBestPickupPoint(points, address)
	if selectedPoint == nil {
		return fmt.Errorf("no pickup points found")
	}

	log.Printf("✅ Selecting pickup point: %s", selectedPoint.Name)

	// Try to select via API first
	if err := client.SelectPickupPoint(selectedPoint.ID); err != nil {
		log.Printf("⚠️  API selection failed, trying page selection: %v", err)
		return selectPickupPointViaPage(ctx)
	}

	log.Println("✅ Pickup point selected!")
	time.Sleep(2 * time.Second)
	return nil
}

// selectPickupPointViaPage selects a pickup point using browser automation
func selectPickupPointViaPage(ctx context.Context) error {
	log.Println("📍 Selecting pickup point via page...")

	// Look for Instabox or Budbee in the page
	script := `
		const options = document.querySelectorAll('[data-pickup-point], .pickup-point, button[role="radio"]');
		for (let opt of options) {
			const text = opt.textContent.toLowerCase();
			if (text.includes('instabox') || text.includes('budbee')) {
				opt.click();
				return opt.textContent.trim();
			}
		}
		// If not found, click first available
		if (options.length > 0) {
			options[0].click();
			return options[0].textContent.trim();
		}
		return null;
	`

	var selected string
	err := chromedp.Run(ctx,
		chromedp.Evaluate(script, &selected),
	)

	if err != nil || selected == "" {
		return fmt.Errorf("failed to select pickup point on page")
	}

	log.Printf("✅ Selected pickup point: %s", selected)
	time.Sleep(2 * time.Second)
	return nil
}

// ExtractSessionID attempts to extract the payment session ID
func ExtractSessionID(ctx context.Context) (string, error) {
	log.Println("🔍 Extracting session ID...")

	// Try multiple methods to get session ID
	scripts := []string{
		`window.zalandoSession?.id || null`,
		`document.querySelector('[data-session-id]')?.dataset.sessionId || null`,
		`window.__NEXT_DATA__?.props?.pageProps?.sessionId || null`,
	}

	for _, script := range scripts {
		var sessionID string
		err := chromedp.Run(ctx,
			chromedp.Evaluate(script, &sessionID),
		)

		if err == nil && sessionID != "" {
			log.Printf("✅ Session ID found: %s", sessionID[:8]+"...")
			return sessionID, nil
		}
	}

	// Generate a placeholder if not found
	log.Println("⚠️  Session ID not found, using placeholder")
	return "session-" + fmt.Sprintf("%d", time.Now().Unix()), nil
}
