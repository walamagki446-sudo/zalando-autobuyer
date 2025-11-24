package payment

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/walamagki446-sudo/zalando-autobuyer/internal/api"
	"github.com/walamagki446-sudo/zalando-autobuyer/internal/browser"
)

// SelectBNPL selects Buy Now Pay Later as payment method
func SelectBNPL(ctx context.Context, client *api.Client, sessionID string) error {
	log.Println("💳 Setting payment method: BNPL...")

	// Try API first
	if err := client.UpdatePayment("bnpl", sessionID); err != nil {
		log.Printf("⚠️  API payment update failed, trying page selection: %v", err)
		return selectBNPLOnPage(ctx)
	}

	log.Println("✅ BNPL payment method selected via API!")
	time.Sleep(2 * time.Second)
	return nil
}

// selectBNPLOnPage selects BNPL using browser automation
func selectBNPLOnPage(ctx context.Context) error {
	log.Println("💳 Selecting BNPL via page...")

	// Look for BNPL payment option
	script := `
		const paymentOptions = document.querySelectorAll('[data-payment-method], .payment-option, input[type="radio"]');
		for (let opt of paymentOptions) {
			const text = (opt.textContent || opt.value || '').toLowerCase();
			const label = opt.labels?.[0]?.textContent?.toLowerCase() || '';
			if (text.includes('bnpl') || text.includes('pay later') || 
			    label.includes('bnpl') || label.includes('pay later') ||
			    text.includes('betala senare') || label.includes('betala senare')) {
				opt.click();
				return true;
			}
		}
		return false;
	`

	var selected bool
	err := chromedp.Run(ctx,
		chromedp.Evaluate(script, &selected),
	)

	if err != nil || !selected {
		return fmt.Errorf("failed to select BNPL on page")
	}

	log.Println("✅ BNPL payment method selected!")
	time.Sleep(2 * time.Second)
	return nil
}

// CompletePurchase executes the final purchase
func CompletePurchase(ctx context.Context, client *api.Client) error {
	log.Println("⚠️  Final confirmation required!")

	// Prompt user for final confirmation
	fmt.Print("Type 'YES' to complete purchase: ")
	var confirmation string
	fmt.Scanln(&confirmation)

	if confirmation != "YES" {
		return fmt.Errorf("purchase cancelled by user")
	}

	log.Println("🔄 Executing purchase...")

	// Try API purchase first
	if err := client.BuyNow(); err != nil {
		log.Printf("⚠️  API purchase failed, trying page button: %v", err)
		return completePurchaseOnPage(ctx)
	}

	// Complete payment
	if err := client.PaymentComplete(); err != nil {
		log.Printf("⚠️  Payment completion API failed: %v", err)
	}

	log.Println("✅ 🎉 PURCHASE COMPLETED! 🎉")
	log.Println("📧 Check your email for confirmation")

	time.Sleep(3 * time.Second)
	return nil
}

// completePurchaseOnPage clicks the final purchase button
func completePurchaseOnPage(ctx context.Context) error {
	log.Println("🔘 Clicking purchase button...")

	// Try multiple selectors for the buy button
	selectors := []string{
		"button:has-text('Köp nu')",
		"button:has-text('Köp')",
		"button:has-text('Buy now')",
		"button:has-text('Complete purchase')",
		"button[type='submit']:has-text('Köp')",
		"button[data-testid='buy-now-button']",
	}

	for _, selector := range selectors {
		err := browser.ClickElement(ctx, selector)
		if err == nil {
			log.Println("✅ Purchase button clicked!")
			time.Sleep(5 * time.Second)
			return nil
		}
	}

	// Try JavaScript approach
	script := `
		const buttons = document.querySelectorAll('button');
		for (let btn of buttons) {
			const text = btn.textContent.toLowerCase();
			if (text.includes('köp') || text.includes('buy') || text.includes('complete')) {
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
		return fmt.Errorf("failed to click purchase button")
	}

	log.Println("✅ 🎉 PURCHASE COMPLETED! 🎉")
	log.Println("📧 Check your email for confirmation")

	time.Sleep(3 * time.Second)
	return nil
}
