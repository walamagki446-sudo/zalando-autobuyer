package main

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// AutobuyerApp holds the application state
type AutobuyerApp struct {
	app            fyne.App
	window         fyne.Window
	httpClient     *HTTPClient
	proxyManager   *ProxyManager
	authenticator  *Authenticator
	productFetcher *ProductFetcher
	cartManager    *CartManager
	checkoutMgr    *CheckoutManager

	// UI elements
	productURLEntry  *widget.Entry
	emailEntry       *widget.Entry
	passwordEntry    *widget.Entry
	streetEntry      *widget.Entry
	zipCodeEntry     *widget.Entry
	cityEntry        *widget.Entry
	firstNameEntry   *widget.Entry
	lastNameEntry    *widget.Entry
	phoneEntry       *widget.Entry
	proxyEntry       *widget.Entry
	sizeSelect       *widget.Select
	statusLog        *widget.Entry
	pickupPointLabel *widget.Label
	progressBar      *widget.ProgressBar

	// State
	currentProduct  *Product
	selectedSize    *ProductSize
	running         bool
	stopChan        chan struct{}
	mutex           sync.Mutex
}

func main() {
	autobuyerApp := NewAutobuyerApp()
	autobuyerApp.Run()
}

// NewAutobuyerApp creates a new autobuyer application
func NewAutobuyerApp() *AutobuyerApp {
	return &AutobuyerApp{
		proxyManager: NewProxyManager(),
		stopChan:     make(chan struct{}),
	}
}

// Run starts the application
func (a *AutobuyerApp) Run() {
	a.app = app.New()
	a.window = a.app.NewWindow("Zalando Autobuyer")
	a.window.Resize(fyne.NewSize(800, 700))

	// Create UI
	content := a.createUI()
	a.window.SetContent(content)
	a.window.ShowAndRun()
}

// createUI creates the main UI
func (a *AutobuyerApp) createUI() fyne.CanvasObject {
	// Product Section
	a.productURLEntry = widget.NewEntry()
	a.productURLEntry.SetPlaceHolder("https://www.zalando.se/product-url...")

	fetchSizesBtn := widget.NewButton("Fetch Sizes", a.onFetchSizes)
	fetchSizesBtn.Importance = widget.HighImportance

	a.sizeSelect = widget.NewSelect([]string{}, func(selected string) {
		a.onSizeSelected(selected)
	})
	a.sizeSelect.PlaceHolder = "Select a size..."

	addToCartBtn := widget.NewButton("Add to Cart", a.onAddToCart)
	addToCartBtn.Importance = widget.MediumImportance

	productSection := container.NewVBox(
		widget.NewLabel("Product URL:"),
		a.productURLEntry,
		container.NewHBox(fetchSizesBtn, widget.NewLabel("  "), a.sizeSelect, widget.NewLabel("  "), addToCartBtn),
	)

	// Login Section
	a.emailEntry = widget.NewEntry()
	a.emailEntry.SetPlaceHolder("email@example.com")

	a.passwordEntry = widget.NewPasswordEntry()
	a.passwordEntry.SetPlaceHolder("Password")

	loginSection := container.NewVBox(
		widget.NewLabel("Login Credentials:"),
		container.NewGridWithColumns(2,
			container.NewVBox(widget.NewLabel("Email:"), a.emailEntry),
			container.NewVBox(widget.NewLabel("Password:"), a.passwordEntry),
		),
	)

	// Address Section
	a.firstNameEntry = widget.NewEntry()
	a.firstNameEntry.SetPlaceHolder("First Name")

	a.lastNameEntry = widget.NewEntry()
	a.lastNameEntry.SetPlaceHolder("Last Name")

	a.streetEntry = widget.NewEntry()
	a.streetEntry.SetPlaceHolder("Street Address")

	a.zipCodeEntry = widget.NewEntry()
	a.zipCodeEntry.SetPlaceHolder("12345")

	a.cityEntry = widget.NewEntry()
	a.cityEntry.SetPlaceHolder("Stockholm")

	a.phoneEntry = widget.NewEntry()
	a.phoneEntry.SetPlaceHolder("+46 70 123 4567")

	addressSection := container.NewVBox(
		widget.NewLabel("Delivery Address:"),
		container.NewGridWithColumns(2,
			container.NewVBox(widget.NewLabel("First Name:"), a.firstNameEntry),
			container.NewVBox(widget.NewLabel("Last Name:"), a.lastNameEntry),
		),
		container.NewVBox(widget.NewLabel("Street:"), a.streetEntry),
		container.NewGridWithColumns(2,
			container.NewVBox(widget.NewLabel("Zip Code:"), a.zipCodeEntry),
			container.NewVBox(widget.NewLabel("City:"), a.cityEntry),
		),
		container.NewVBox(widget.NewLabel("Phone (optional):"), a.phoneEntry),
	)

	// Proxy Section
	a.proxyEntry = widget.NewEntry()
	a.proxyEntry.SetPlaceHolder("host:port:username:password (optional)")

	proxySection := container.NewVBox(
		widget.NewLabel("Proxy (host:port:user:pass):"),
		a.proxyEntry,
	)

	// Pickup Point Display
	a.pickupPointLabel = widget.NewLabel("Selected Pickup Point: None")
	a.pickupPointLabel.Wrapping = fyne.TextWrapWord

	// Action Buttons
	checkoutBtn := widget.NewButton("Login & Checkout", a.onCheckout)
	checkoutBtn.Importance = widget.HighImportance
	checkoutBtn.Icon = theme.ConfirmIcon()

	stopBtn := widget.NewButton("Stop", a.onStop)
	stopBtn.Importance = widget.DangerImportance
	stopBtn.Icon = theme.CancelIcon()

	actionSection := container.NewHBox(
		checkoutBtn,
		stopBtn,
	)

	// Progress Bar
	a.progressBar = widget.NewProgressBar()
	a.progressBar.Min = 0
	a.progressBar.Max = 100

	// Status Log
	a.statusLog = widget.NewMultiLineEntry()
	a.statusLog.SetPlaceHolder("Status log will appear here...")
	a.statusLog.Wrapping = fyne.TextWrapWord
	a.statusLog.Disable()

	statusScroll := container.NewScroll(a.statusLog)
	statusScroll.SetMinSize(fyne.NewSize(0, 150))

	statusSection := container.NewVBox(
		widget.NewLabel("Status Log:"),
		statusScroll,
	)

	// Main Layout
	return container.NewVBox(
		widget.NewCard("Product Selection", "", productSection),
		widget.NewCard("Authentication", "", loginSection),
		widget.NewCard("Delivery Address", "", addressSection),
		widget.NewCard("Proxy Configuration", "", proxySection),
		widget.NewCard("Pickup Point", "", a.pickupPointLabel),
		a.progressBar,
		container.NewCenter(actionSection),
		widget.NewCard("Status", "", statusSection),
	)
}

// log adds a message to the status log
func (a *AutobuyerApp) log(message string) {
	timestamp := time.Now().Format("15:04:05")
	logMessage := fmt.Sprintf("[%s] %s\n", timestamp, message)

	a.statusLog.Enable()
	currentText := a.statusLog.Text
	a.statusLog.SetText(currentText + logMessage)
	a.statusLog.Disable()

	// Scroll to bottom
	a.statusLog.CursorRow = len(strings.Split(a.statusLog.Text, "\n")) - 1
}

// initHTTPClient initializes the HTTP client with optional proxy
func (a *AutobuyerApp) initHTTPClient() error {
	proxyStr := strings.TrimSpace(a.proxyEntry.Text)

	client, err := NewHTTPClient(proxyStr)
	if err != nil {
		return fmt.Errorf("failed to create HTTP client: %w", err)
	}

	a.httpClient = client
	a.authenticator = NewAuthenticator(client)
	a.productFetcher = NewProductFetcher(client)
	a.cartManager = NewCartManager(client)
	a.checkoutMgr = NewCheckoutManager(client)

	return nil
}

// onFetchSizes handles the Fetch Sizes button click
func (a *AutobuyerApp) onFetchSizes() {
	a.mutex.Lock()
	if a.running {
		a.mutex.Unlock()
		return
	}
	a.running = true
	a.mutex.Unlock()

	go func() {
		defer func() {
			a.mutex.Lock()
			a.running = false
			a.mutex.Unlock()
		}()

		productURL := strings.TrimSpace(a.productURLEntry.Text)
		if productURL == "" {
			a.log("Error: Please enter a product URL")
			return
		}

		if !ValidateProductURL(productURL) {
			a.log("Error: Invalid Zalando product URL")
			return
		}

		a.log("Initializing HTTP client...")
		a.progressBar.SetValue(10)

		if err := a.initHTTPClient(); err != nil {
			a.log(fmt.Sprintf("Error: %v", err))
			return
		}

		a.log("Fetching product information...")
		a.progressBar.SetValue(30)

		product, err := a.productFetcher.FetchProduct(productURL)
		if err != nil {
			a.log(fmt.Sprintf("Error fetching product: %v", err))
			ErrorDelay()
			return
		}

		a.currentProduct = product
		a.progressBar.SetValue(80)

		if product.Name != "" {
			a.log(fmt.Sprintf("Product: %s - %s", product.Brand, product.Name))
		}

		if len(product.Sizes) == 0 {
			a.log("Warning: No sizes found for this product")
			dialog.ShowInformation("No Sizes", "Could not detect sizes for this product. The product page structure may have changed.", a.window)
			return
		}

		// Update size selector
		sizeOptions := make([]string, 0, len(product.Sizes))
		for _, size := range product.Sizes {
			status := "✓"
			if !size.Available {
				status = "✗"
			}
			sizeOptions = append(sizeOptions, fmt.Sprintf("%s [%s] %s", size.Size, status, size.SKU))
		}

		a.sizeSelect.Options = sizeOptions
		a.sizeSelect.Refresh()

		a.log(fmt.Sprintf("Found %d sizes (%d available)", len(product.Sizes), len(product.GetAvailableSizes())))
		a.progressBar.SetValue(100)

		// Show available sizes
		for _, size := range product.Sizes {
			status := "In Stock"
			if !size.Available {
				status = "Out of Stock"
			}
			a.log(fmt.Sprintf("  Size: %s - %s (SKU: %s)", size.Size, status, size.SKU))
		}
	}()
}

// onSizeSelected handles size selection
func (a *AutobuyerApp) onSizeSelected(selected string) {
	if a.currentProduct == nil || selected == "" {
		return
	}

	// Extract SKU from selection string
	parts := strings.Split(selected, " ")
	if len(parts) > 0 {
		sku := parts[len(parts)-1]
		a.selectedSize = a.currentProduct.GetSizeBySKU(sku)
		if a.selectedSize != nil {
			a.log(fmt.Sprintf("Selected size: %s (SKU: %s)", a.selectedSize.Size, a.selectedSize.SKU))
		}
	}
}

// onAddToCart handles adding item to cart
func (a *AutobuyerApp) onAddToCart() {
	if a.selectedSize == nil {
		dialog.ShowError(fmt.Errorf("please select a size first"), a.window)
		return
	}

	if !a.selectedSize.Available {
		dialog.ShowConfirm("Out of Stock", "This size is out of stock. Try anyway?", func(ok bool) {
			if ok {
				a.addToCart()
			}
		}, a.window)
		return
	}

	a.addToCart()
}

// addToCart performs the add to cart operation
func (a *AutobuyerApp) addToCart() {
	a.mutex.Lock()
	if a.running {
		a.mutex.Unlock()
		return
	}
	a.running = true
	a.mutex.Unlock()

	go func() {
		defer func() {
			a.mutex.Lock()
			a.running = false
			a.mutex.Unlock()
		}()

		a.log(fmt.Sprintf("Adding to cart: %s (SKU: %s)...", a.selectedSize.Size, a.selectedSize.SKU))
		a.progressBar.SetValue(50)

		if a.cartManager == nil {
			if err := a.initHTTPClient(); err != nil {
				a.log(fmt.Sprintf("Error: %v", err))
				return
			}
		}

		HumanDelay()

		err := a.cartManager.AddToCart(a.selectedSize.SKU, 1)
		if err != nil {
			a.log(fmt.Sprintf("Error adding to cart: %v", err))
			ErrorDelay()
			return
		}

		a.log("Successfully added to cart!")
		a.progressBar.SetValue(100)
		dialog.ShowInformation("Success", "Item added to cart successfully!", a.window)
	}()
}

// onCheckout handles the login and checkout process
func (a *AutobuyerApp) onCheckout() {
	// Validate inputs
	if err := a.validateCheckoutInputs(); err != nil {
		dialog.ShowError(err, a.window)
		return
	}

	a.mutex.Lock()
	if a.running {
		a.mutex.Unlock()
		return
	}
	a.running = true
	a.stopChan = make(chan struct{})
	a.mutex.Unlock()

	go func() {
		defer func() {
			a.mutex.Lock()
			a.running = false
			a.mutex.Unlock()
		}()

		// Initialize client
		a.log("Initializing checkout process...")
		a.progressBar.SetValue(5)

		if err := a.initHTTPClient(); err != nil {
			a.log(fmt.Sprintf("Error: %v", err))
			return
		}

		// Check for stop signal
		select {
		case <-a.stopChan:
			a.log("Process stopped by user")
			return
		default:
		}

		// Login
		a.log("Logging in...")
		a.progressBar.SetValue(15)

		email := strings.TrimSpace(a.emailEntry.Text)
		password := a.passwordEntry.Text

		success, err := CheckZalandoLogin(a.httpClient, email, password)
		if err != nil || !success {
			a.log(fmt.Sprintf("Login failed: %v", err))
			ErrorDelay()
			dialog.ShowError(fmt.Errorf("login failed: %v", err), a.window)
			return
		}

		a.log("Login successful!")
		a.progressBar.SetValue(25)
		HumanDelay()

		// Check for stop signal
		select {
		case <-a.stopChan:
			a.log("Process stopped by user")
			return
		default:
		}

		// Add to cart if not already added
		if a.selectedSize != nil {
			shouldAddToCart := true
			if a.cartManager != nil && a.cartManager.HasItem(a.selectedSize.SKU) {
				shouldAddToCart = false
			}
			if shouldAddToCart {
				a.log(fmt.Sprintf("Adding to cart: %s...", a.selectedSize.SKU))
				if err := a.cartManager.AddToCart(a.selectedSize.SKU, 1); err != nil {
					a.log(fmt.Sprintf("Error adding to cart: %v", err))
					ErrorDelay()
					return
				}
				a.log("Added to cart!")
				HumanDelay()
			}
		}
		a.progressBar.SetValue(40)

		// Check for stop signal
		select {
		case <-a.stopChan:
			a.log("Process stopped by user")
			return
		default:
		}

		// Initialize checkout
		a.log("Initializing checkout...")
		if err := a.checkoutMgr.InitCheckout(); err != nil {
			a.log(fmt.Sprintf("Error initializing checkout: %v", err))
			return
		}
		a.progressBar.SetValue(50)
		HumanDelay()

		// Set address
		address := &Address{
			FirstName: strings.TrimSpace(a.firstNameEntry.Text),
			LastName:  strings.TrimSpace(a.lastNameEntry.Text),
			Street:    strings.TrimSpace(a.streetEntry.Text),
			ZipCode:   strings.TrimSpace(a.zipCodeEntry.Text),
			City:      strings.TrimSpace(a.cityEntry.Text),
			Country:   "SE",
			Phone:     strings.TrimSpace(a.phoneEntry.Text),
			Email:     email,
		}

		a.log("Setting delivery address...")
		if err := a.checkoutMgr.SetAddress(address); err != nil {
			a.log(fmt.Sprintf("Error setting address: %v", err))
			return
		}
		a.progressBar.SetValue(60)
		HumanDelay()

		// Check for stop signal
		select {
		case <-a.stopChan:
			a.log("Process stopped by user")
			return
		default:
		}

		// Search for pickup points
		a.log("Searching for Instabox pickup points...")
		pickupPoints, err := a.checkoutMgr.SearchPickupPoints(address)
		if err != nil {
			a.log(fmt.Sprintf("Error searching pickup points: %v", err))
			return
		}

		instaboxPoints := a.checkoutMgr.GetInstaboxPickupPoints(pickupPoints)
		a.log(fmt.Sprintf("Found %d pickup points (%d Instabox)", len(pickupPoints), len(instaboxPoints)))
		a.progressBar.SetValue(70)

		if len(instaboxPoints) == 0 {
			a.log("No Instabox pickup points found. Using first available pickup point.")
			if len(pickupPoints) == 0 {
				a.log("Error: No pickup points available")
				dialog.ShowError(fmt.Errorf("no pickup points available in your area"), a.window)
				return
			}
			instaboxPoints = pickupPoints
		}

		// Select first Instabox
		selectedPoint := &instaboxPoints[0]
		a.log(fmt.Sprintf("Selecting pickup point: %s (%s)", selectedPoint.Name, selectedPoint.Address))
		a.pickupPointLabel.SetText(fmt.Sprintf("Selected Pickup Point: %s\n%s, %s %s", selectedPoint.Name, selectedPoint.Address, selectedPoint.ZipCode, selectedPoint.City))

		if err := a.checkoutMgr.SelectPickupPoint(selectedPoint); err != nil {
			a.log(fmt.Sprintf("Error selecting pickup point: %v", err))
			return
		}
		a.progressBar.SetValue(80)
		HumanDelay()

		// Check for stop signal
		select {
		case <-a.stopChan:
			a.log("Process stopped by user")
			return
		default:
		}

		// Select Faktura payment
		a.log("Selecting Faktura payment...")
		if err := a.checkoutMgr.SelectFakturaPayment(); err != nil {
			a.log(fmt.Sprintf("Error selecting payment: %v", err))
			return
		}
		a.progressBar.SetValue(90)
		HumanDelay()

		// Confirm order
		a.log("Confirming order...")
		orderNumber, err := a.checkoutMgr.ConfirmOrder()
		if err != nil {
			a.log(fmt.Sprintf("Error confirming order: %v", err))
			dialog.ShowError(fmt.Errorf("order confirmation failed: %v", err), a.window)
			return
		}

		a.progressBar.SetValue(100)
		a.log(fmt.Sprintf("Order confirmed! Order number: %s", orderNumber))
		dialog.ShowInformation("Success", fmt.Sprintf("Order confirmed successfully!\nOrder Number: %s", orderNumber), a.window)
	}()
}

// validateCheckoutInputs validates all required inputs
func (a *AutobuyerApp) validateCheckoutInputs() error {
	if strings.TrimSpace(a.emailEntry.Text) == "" {
		return fmt.Errorf("please enter your email")
	}
	if a.passwordEntry.Text == "" {
		return fmt.Errorf("please enter your password")
	}
	if strings.TrimSpace(a.firstNameEntry.Text) == "" {
		return fmt.Errorf("please enter first name")
	}
	if strings.TrimSpace(a.lastNameEntry.Text) == "" {
		return fmt.Errorf("please enter last name")
	}
	if strings.TrimSpace(a.streetEntry.Text) == "" {
		return fmt.Errorf("please enter street address")
	}
	if strings.TrimSpace(a.zipCodeEntry.Text) == "" {
		return fmt.Errorf("please enter zip code")
	}
	if strings.TrimSpace(a.cityEntry.Text) == "" {
		return fmt.Errorf("please enter city")
	}
	return nil
}

// onStop handles the stop button click
func (a *AutobuyerApp) onStop() {
	a.mutex.Lock()
	if a.running {
		close(a.stopChan)
		a.log("Stopping...")
	}
	a.mutex.Unlock()
}
