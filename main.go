package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/term"
)

func main() {
	// Disable default log flags for cleaner output
	log.SetFlags(0)

	// Print welcome banner
	printBanner()

	// Create HTTP client
	client, err := NewHTTPClient()
	if err != nil {
		LogError("Kunde inte skapa HTTP-klient: %v", err)
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)

	// Step 1: Login
	fmt.Println()
	LogInfo("=== STEG 1: INLOGGNING ===")
	fmt.Print("E-postadress: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)

	fmt.Print("Lösenord: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		LogError("Kunde inte läsa lösenord: %v", err)
		os.Exit(1)
	}
	password := strings.TrimSpace(string(passwordBytes))
	fmt.Println() // New line after hidden password input

	if err := client.Login(email, password); err != nil {
		LogError("Inloggning misslyckades: %v", err)
		os.Exit(1)
	}

	// Step 2: Product Selection
	fmt.Println()
	LogInfo("=== STEG 2: PRODUKTVAL ===")
	fmt.Print("Produkt-URL: ")
	productURL, _ := reader.ReadString('\n')
	productURL = strings.TrimSpace(productURL)

	product, err := client.FetchProduct(productURL)
	if err != nil {
		LogError("Kunde inte hämta produkt: %v", err)
		os.Exit(1)
	}

	// Display product info
	fmt.Println()
	LogSuccess("Produkt: %s", product.Name)
	LogInfo("SKU: %s", product.SKU)

	// Display available sizes
	if len(product.Sizes) > 0 {
		fmt.Println()
		LogInfo("Tillgängliga storlekar:")
		for i, size := range product.Sizes {
			status := "I lager"
			if !size.Available {
				status = "Slut i lager"
			}
			fmt.Printf("  %d. %s (%s) - %s\n", i+1, size.Size, size.SKU, status)
		}

		fmt.Println()
		fmt.Print("Välj storlek (nummer): ")
		sizeInput, _ := reader.ReadString('\n')
		sizeInput = strings.TrimSpace(sizeInput)

		sizeIndex, err := strconv.Atoi(sizeInput)
		if err != nil || sizeIndex < 1 || sizeIndex > len(product.Sizes) {
			LogError("Ogiltigt val")
			os.Exit(1)
		}

		selectedSize := product.Sizes[sizeIndex-1]
		LogSuccess("Vald storlek: %s", selectedSize.Size)

		// Add to cart
		if err := client.AddToCart(selectedSize.SKU, product.ConfigID); err != nil {
			LogError("Kunde inte lägga till i varukorg: %v", err)
			os.Exit(1)
		}
	} else {
		LogWarning("Inga storlekar hittades, lägger till med huvud-SKU...")
		if err := client.AddToCart(product.SKU, product.ConfigID); err != nil {
			LogError("Kunde inte lägga till i varukorg: %v", err)
			os.Exit(1)
		}
	}

	// Step 3: Delivery Address
	fmt.Println()
	LogInfo("=== STEG 3: LEVERANSADRESS ===")
	fmt.Print("Gatuadress: ")
	street, _ := reader.ReadString('\n')
	street = strings.TrimSpace(street)

	fmt.Print("Stad: ")
	city, _ := reader.ReadString('\n')
	city = strings.TrimSpace(city)

	fmt.Print("Postnummer: ")
	postalCode, _ := reader.ReadString('\n')
	postalCode = strings.TrimSpace(postalCode)

	address := Address{
		Street:     street,
		City:       city,
		PostalCode: postalCode,
		Country:    "SE",
	}

	// Step 4: Create Checkout
	fmt.Println()
	LogInfo("=== STEG 4: CHECKOUT ===")
	checkoutSession, err := client.CreateCheckoutSession()
	if err != nil {
		LogError("Kunde inte skapa checkout: %v", err)
		os.Exit(1)
	}

	// Step 5: Pickup Point Selection
	fmt.Println()
	LogInfo("=== STEG 5: UPPHÄMTNINGSSTÄLLE ===")
	pickupPoints, err := client.SearchPickupPoints(address)
	if err != nil {
		LogError("Kunde inte söka upphämtningsställen: %v", err)
		os.Exit(1)
	}

	if len(pickupPoints) == 0 {
		LogError("Inga upphämtningsställen hittades för denna adress")
		os.Exit(1)
	}

	// Display pickup points
	fmt.Println()
	LogInfo("Hittade upphämtningsställen:")
	for i, point := range pickupPoints {
		if i >= 10 {
			break // Show only first 10
		}
		fmt.Printf("  %d. %s (%s) - %.2f km\n", i+1, point.Name, point.Provider, point.Distance/1000)
		fmt.Printf("     %s\n", point.Address)
	}

	fmt.Println()
	fmt.Print("Välj upphämtningsställe (nummer) eller tryck Enter för automatiskt val: ")
	pickupInput, _ := reader.ReadString('\n')
	pickupInput = strings.TrimSpace(pickupInput)

	var selectedPickupPoint *PickupPoint

	if pickupInput == "" {
		// Auto-select best pickup point
		selectedPickupPoint = SelectBestPickupPoint(pickupPoints, false)
		LogInfo("Automatiskt valt: %s", selectedPickupPoint.Name)
	} else {
		pickupIndex, err := strconv.Atoi(pickupInput)
		if err != nil || pickupIndex < 1 || pickupIndex > len(pickupPoints) {
			LogError("Ogiltigt val")
			os.Exit(1)
		}
		selectedPickupPoint = &pickupPoints[pickupIndex-1]
	}

	LogSuccess("Valt upphämtningsställe: %s (%s)", selectedPickupPoint.Name, selectedPickupPoint.Provider)

	if err := client.SelectPickupPoint(selectedPickupPoint.ID); err != nil {
		LogError("Kunde inte välja upphämtningsställe: %v", err)
		os.Exit(1)
	}

	if err := client.NextStep(); err != nil {
		LogError("Kunde inte gå vidare: %v", err)
		os.Exit(1)
	}

	// Step 6: Payment
	fmt.Println()
	LogInfo("=== STEG 6: BETALNING ===")

	if err := client.UpdatePayment(); err != nil {
		LogError("Kunde inte uppdatera betalning: %v", err)
		os.Exit(1)
	}

	// Step 7: Complete Purchase
	fmt.Println()
	LogInfo("=== STEG 7: SLUTFÖR KÖP ===")
	fmt.Print("Bekräfta köp? (ja/nej): ")
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))

	if confirm == "ja" || confirm == "j" || confirm == "yes" || confirm == "y" {
		if err := client.CompletePurchase(); err != nil {
			LogError("Kunde inte slutföra köp: %v", err)
			os.Exit(1)
		}

		// Success message
		fmt.Println()
		printSuccessMessage(checkoutSession.ID)
	} else {
		LogWarning("Köp avbrutet av användaren")
	}
}

func printBanner() {
	banner := `
╔═══════════════════════════════════════════════════════════╗
║                                                           ║
║           ZALANDO AUTOBUYER v1.0                          ║
║           Automatiserad köpprocess för Zalando.se         ║
║                                                           ║
║           ENDAST FÖR UTBILDNINGSÄNDAMÅL                   ║
║                                                           ║
╚═══════════════════════════════════════════════════════════╝
`
	fmt.Println(banner)
}

func printSuccessMessage(checkoutID string) {
	success := `
╔═══════════════════════════════════════════════════════════╗
║                                                           ║
║                    ✓ KÖP SLUTFÖRT!                        ║
║                                                           ║
║   Din beställning har genomförts framgångsrikt.          ║
║   Du kommer att få en bekräftelse via e-post.            ║
║                                                           ║
`
	if checkoutID != "" {
		success += fmt.Sprintf("║   Checkout-ID: %-42s ║\n║                                                           ║\n", checkoutID)
	}
	success += `╚═══════════════════════════════════════════════════════════╝
`
	fmt.Println(success)
}
