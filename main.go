package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

// Simple input reader that WORKS in Windows .exe
func readInput(prompt string) string {
	fmt.Print(prompt)
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			log.Fatalf("❌ Fel vid läsning av input: %v", err)
		}
		// EOF or Ctrl+C - return empty string
		return ""
	}
	return strings.TrimSpace(scanner.Text())
}

func main() {
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║          ZALANDO AUTOBUYER - WINDOWS VERSION               ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// STEG 1: Inloggning
	fmt.Println("=== STEG 1: INLOGGNING ===")
	email := readInput("E-postadress: ")
	password := readInput("Lösenord: ")  // Visible for now - works!

	if email == "" || password == "" {
		log.Fatal("❌ E-post och lösenord får inte vara tomma")
	}

	fmt.Printf("\n[INFO] Loggar in som: %s\n", email)
	
	client, err := Login(email, password)
	if err != nil {
		log.Fatalf("❌ Inloggning misslyckades: %v", err)
	}
	fmt.Println("✅ Inloggning lyckades!\n")

	// STEG 2: Produkt
	fmt.Println("=== STEG 2: PRODUKTVAL ===")
	productURL := readInput("Produkt-URL: ")
	
	if productURL == "" {
		log.Fatal("❌ Produkt-URL får inte vara tom")
	}

	fmt.Println("\n[INFO] Hämtar produktinformation...")
	product, sizes, err := FetchProductInfo(client, productURL)
	if err != nil {
		log.Fatalf("❌ Kunde inte hämta produkt: %v", err)
	}

	fmt.Printf("✅ Produkt: %s\n\n", product.Name)
	fmt.Println("📦 Tillgängliga storlekar:")
	for i, size := range sizes {
		status := "✅ I lager"
		if !size.Available {
			status = "❌ Slutsåld"
		}
		fmt.Printf("  %d. %s (%s)\n", i+1, size.Name, status)
	}

	// STEG 3: Storleksval
	selectedSize := readInput("\nVälj storlek (t.ex. L, 42): ")
	
	var chosenSize *Size
	for _, size := range sizes {
		if strings.EqualFold(size.Name, selectedSize) && size.Available {
			chosenSize = &size
			break
		}
	}

	if chosenSize == nil {
		log.Fatalf("❌ Storlek '%s' finns inte eller är slutsåld", selectedSize)
	}

	fmt.Printf("\n[INFO] Lägger till i kundvagn: %s (storlek %s)...\n", product.Name, chosenSize.Name)
	err = AddToCart(client, product.SKU, chosenSize.SKU)
	if err != nil {
		log.Fatalf("❌ Kunde inte lägga till i kundvagn: %v", err)
	}
	fmt.Println("✅ Tillagd i kundvagn!\n")

	// STEG 4: Leveransadress
	fmt.Println("=== STEG 3: LEVERANS ===")
	address := readInput("Leveransadress (för upphämtning): ")
	
	if address == "" {
		log.Fatal("❌ Adress får inte vara tom")
	}

	fmt.Println("\n[INFO] Skapar checkout...")
	checkoutID, err := CreateCheckout(client)
	if err != nil {
		log.Fatalf("❌ Kunde inte skapa checkout: %v", err)
	}
	fmt.Printf("✅ Checkout skapad: %s\n", checkoutID)

	// STEG 5: Upphämtningsställe
	fmt.Println("\n[INFO] Söker upphämtningsställen (prioriterar Instabox/Budbee)...")
	pickupPoints, err := SearchPickupPoints(client, address)
	if err != nil {
		log.Fatalf("❌ Kunde inte söka upphämtningsställen: %v", err)
	}

	var preferredPoints []PickupPoint
	for _, point := range pickupPoints {
		provider := strings.ToLower(point.Provider)
		if strings.Contains(provider, "instabox") || strings.Contains(provider, "budbee") {
			preferredPoints = append(preferredPoints, point)
		}
	}

	if len(preferredPoints) == 0 {
		fmt.Println("⚠️  Inga Instabox/Budbee-ställen hittades, använder närmaste...")
		if len(pickupPoints) > 0 {
			preferredPoints = pickupPoints
		} else {
			log.Fatal("❌ Inga upphämtningsställen hittades")
		}
	}

	selectedPoint := preferredPoints[0]
	fmt.Printf("✅ Valt upphämtningsställe: %s (%s)\n", selectedPoint.Name, selectedPoint.Provider)

	fmt.Println("\n[INFO] Väljer upphämtningsställe...")
	err = SelectPickupPoint(client, selectedPoint.ID)
	if err != nil {
		log.Fatalf("❌ Kunde inte välja upphämtningsställe: %v", err)
	}
	fmt.Println("✅ Upphämtningsställe valt!\n")

	// STEG 6: Betalning
	fmt.Println("=== STEG 4: BETALNING ===")
	fmt.Println("[INFO] Väljer betalmetod BNPL (Faktura)...")
	sessionID, err := GetPaymentSessionID(client)
	if err != nil {
		log.Fatalf("❌ Kunde inte hämta betalningssession: %v", err)
	}

	err = SelectBNPLPayment(client, sessionID)
	if err != nil {
		log.Fatalf("❌ Kunde inte välja BNPL: %v", err)
	}
	fmt.Println("✅ BNPL valt!\n")

	// STEG 7: Bekräftelse
	fmt.Println("=== STEG 5: BEKRÄFTELSE ===")
	fmt.Println("⚠️  VARNING: Du är på väg att slutföra ett RIKTIGT KÖP!")
	confirmation := readInput("Skriv 'JA' för att bekräfta köpet: ")
	
	if strings.ToUpper(confirmation) != "JA" {
		log.Fatal("❌ Köp avbrutet av användaren")
	}

	fmt.Println("\n[INFO] Går vidare till nästa steg...")
	err = NextStep(client)
	if err != nil {
		log.Fatalf("❌ Kunde inte gå vidare: %v", err)
	}

	fmt.Println("[INFO] Slutför köpet...")
	err = CompletePurchase(client)
	if err != nil {
		log.Fatalf("❌ Kunde inte slutföra köpet: %v", err)
	}

	fmt.Println("\n╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║              ✅ KÖP GENOMFÖRT!                              ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println("📧 Kolla din e-post för orderbekräftelse")
	fmt.Println()
	
	readInput("Tryck Enter för att avsluta...")
}
