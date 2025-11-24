package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/walamagki446-sudo/zalando-autobuyer/internal/api"
	"github.com/walamagki446-sudo/zalando-autobuyer/internal/auth"
	"github.com/walamagki446-sudo/zalando-autobuyer/internal/browser"
	"github.com/walamagki446-sudo/zalando-autobuyer/internal/cart"
	"github.com/walamagki446-sudo/zalando-autobuyer/internal/checkout"
	"github.com/walamagki446-sudo/zalando-autobuyer/internal/payment"
)

const (
	baseURL = "https://www.zalando.se"
)

func main() {
	printBanner()

	// Read credentials
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter credentials (email:password): ")
	credentials, _ := reader.ReadString('\n')
	credentials = strings.TrimSpace(credentials)

	email, password, err := auth.ParseCredentials(credentials)
	if err != nil {
		log.Fatalf("❌ Invalid credentials format: %v", err)
	}

	// Setup browser
	ctx, cancel, err := browser.SetupBrowser()
	if err != nil {
		log.Fatalf("❌ Failed to setup browser: %v", err)
	}
	defer cancel()

	// Login
	session, err := auth.Login(ctx, email, password)
	if err != nil {
		log.Fatalf("❌ Login failed: %v", err)
	}

	// Create API client
	apiClient := api.NewClient(session, baseURL)

	// Get product URL
	fmt.Print("\nEnter product URL: ")
	productURL, _ := reader.ReadString('\n')
	productURL = strings.TrimSpace(productURL)

	// Navigate to product page
	log.Println("🔄 Navigating to product page...")
	if err := browser.Navigate(ctx, productURL); err != nil {
		log.Fatalf("❌ Failed to navigate to product: %v", err)
	}

	// Wait for page to load
	time.Sleep(3 * time.Second)

	// Detect available sizes
	sizes, err := cart.DetectAvailableSizes(ctx)
	if err != nil {
		log.Fatalf("❌ Failed to detect sizes: %v", err)
	}

	// Display available sizes
	fmt.Printf("\nAvailable sizes: %s\n", strings.Join(sizes, ", "))
	fmt.Print("Enter your size: ")
	selectedSize, _ := reader.ReadString('\n')
	selectedSize = strings.TrimSpace(selectedSize)

	// Validate size
	if !cart.FindSizeInList(sizes, selectedSize) {
		log.Fatalf("❌ Size '%s' not available. Choose from: %s", selectedSize, strings.Join(sizes, ", "))
	}

	// Select size
	if err := cart.SelectSize(ctx, selectedSize); err != nil {
		log.Fatalf("❌ Failed to select size: %v", err)
	}

	// Add to cart
	if err := cart.AddToCart(ctx); err != nil {
		log.Fatalf("❌ Failed to add to cart: %v", err)
	}

	// Navigate to checkout
	if err := checkout.NavigateToCheckout(ctx, baseURL); err != nil {
		log.Fatalf("❌ Failed to navigate to checkout: %v", err)
	}

	// Create checkout session via API
	log.Println("🔄 Creating checkout session...")
	checkoutID, err := apiClient.CreateCheckout()
	if err != nil {
		log.Printf("⚠️  Failed to create checkout via API: %v", err)
	} else {
		log.Printf("✅ Checkout created: %s", checkoutID)
		session.CheckoutID = checkoutID
	}

	// Select pickup point
	userAddress := "Stockholm" // Default address
	if err := checkout.SelectPickupPointOnPage(ctx, apiClient, userAddress); err != nil {
		log.Printf("⚠️  Failed to select pickup point: %v", err)
		log.Println("Continuing anyway...")
	}

	// Proceed to next step
	log.Println("🔄 Proceeding to payment...")
	if err := apiClient.NextStep(); err != nil {
		log.Printf("⚠️  Next step API failed: %v", err)
	}

	time.Sleep(2 * time.Second)

	// Extract session ID for payment
	sessionID, err := checkout.ExtractSessionID(ctx)
	if err != nil {
		log.Printf("⚠️  Failed to extract session ID: %v", err)
		sessionID = "unknown"
	}

	// Select BNPL payment
	if err := payment.SelectBNPL(ctx, apiClient, sessionID); err != nil {
		log.Printf("⚠️  Failed to select BNPL: %v", err)
		log.Println("Continuing anyway...")
	}

	// Complete purchase
	if err := payment.CompletePurchase(ctx, apiClient); err != nil {
		log.Fatalf("❌ Purchase failed: %v", err)
	}

	// Keep browser open for a few seconds
	log.Println("✅ Process completed! Browser will close in 5 seconds...")
	time.Sleep(5 * time.Second)
}

func printBanner() {
	banner := `
╔══════════════════════════════════════════════════════════╗
║            ZALANDO AUTOBUYER - GO EDITION                ║
╚══════════════════════════════════════════════════════════╝
`
	fmt.Println(banner)
}
