# Zalando Autobuyer - Go Edition

A Go-based automated buyer for Zalando.se with **visible browser automation** that allows users to monitor the entire purchase process in real-time.

## Features

✅ **Visible Browser Mode** - Watch every action in real-time  
✅ **Interactive CLI** - User-friendly terminal prompts  
✅ **Automatic Size Detection** - Extracts available sizes from product page  
✅ **Smart Pickup Point Selection** - Prioritizes Instabox/Budbee  
✅ **BNPL Payment Support** - Buy Now Pay Later integration  
✅ **Comprehensive API Integration** - Uses Zalando's APIs for reliability  
✅ **Cross-Platform** - Works on Windows, Linux, and macOS  

## Technology Stack

- **Language**: Go 1.21+
- **Browser Automation**: Chromedp (headful mode)
- **HTTP Client**: Native Go net/http
- **UI**: Terminal-based interactive prompts

## Prerequisites

- Go 1.21 or higher
- Chrome/Chromium browser installed
- Active Zalando account

## Installation

### 1. Clone the repository
```bash
git clone https://github.com/walamagki446-sudo/zalando-autobuyer.git
cd zalando-autobuyer
```

### 2. Install dependencies
```bash
go mod download
```

### 3. Build the application
```bash
# For your current platform
go build -o zalando-autobuyer main.go

# For Windows (from Linux/Mac)
GOOS=windows GOARCH=amd64 go build -o zalando-autobuyer.exe main.go

# For Linux (from Windows/Mac)
GOOS=linux GOARCH=amd64 go build -o zalando-autobuyer main.go

# For macOS (from Windows/Linux)
GOOS=darwin GOARCH=amd64 go build -o zalando-autobuyer main.go
```

## Usage

### Running the Application

```bash
# On Linux/Mac
./zalando-autobuyer

# On Windows
zalando-autobuyer.exe
```

### Example Session

```
╔══════════════════════════════════════════════════════════╗
║            ZALANDO AUTOBUYER - GO EDITION                ║
╚══════════════════════════════════════════════════════════╝

Enter credentials (email:password): your-email@example.com:your-password
🔄 Setting up browser...
✅ Browser ready!
🔐 Logging in as: your-email@example.com
📧 Entering email...
🔑 Entering password...
🔘 Clicking login button...
⏳ Waiting for authentication...
✅ Login successful!

Enter product URL: https://www.zalando.se/adidas-originals-spezial-ad116b0d9-a11.html
🔄 Navigating to product page...
📦 Fetching available sizes...
📏 Found 6 sizes: [XS S M L XL XXL]

Available sizes: XS, S, M, L, XL, XXL
Enter your size: L
✅ Selecting size: L
✅ Size L selected!
🛒 Clicking 'Handla' button...
✅ Added to cart!

🔄 Navigating to checkout...
✅ Navigated to checkout!
🔄 Creating checkout session...
✅ Checkout created: checkout-12345
📍 Searching for pickup points...
📍 Found 8 pickup points
✓ Selected Instabox: Instabox Stockholm City
✅ Pickup point selected!

🔄 Proceeding to payment...
🔍 Extracting session ID...
✅ Session ID found: session-67890
💳 Setting payment method: BNPL...
✅ BNPL payment method selected via API!

⚠️  Final confirmation required!
Type 'YES' to complete purchase: YES
🔄 Executing purchase...
✅ 🎉 PURCHASE COMPLETED! 🎉
📧 Check your email for confirmation

✅ Process completed! Browser will close in 5 seconds...
```

## Project Structure

```
zalando-autobuyer/
├── main.go                    # Entry point with CLI
├── go.mod                     # Go modules
├── go.sum                     # Dependencies
├── internal/
│   ├── auth/
│   │   └── auth.go           # Login automation
│   ├── browser/
│   │   └── browser.go        # Browser setup (chromedp)
│   ├── cart/
│   │   └── cart.go           # Cart operations
│   ├── checkout/
│   │   └── checkout.go       # Checkout flow
│   ├── payment/
│   │   └── payment.go        # Payment processing
│   ├── api/
│   │   └── client.go         # HTTP API client
│   └── models/
│       └── models.go         # Data structures
├── README.md
└── .gitignore
```

## How It Works

### 1. Authentication
- Opens a visible Chrome browser
- Navigates to Zalando login page
- Automatically fills in credentials
- Waits for successful authentication

### 2. Product Selection
- Navigates to the specified product URL
- Waits for page to fully load
- Extracts product information

### 3. Size Detection & Selection
- Clicks the size picker to open dropdown
- Extracts all available sizes dynamically
- Displays options to user
- Selects user's chosen size

### 4. Add to Cart
- Clicks the "Handla" (Add to Cart) button
- Waits for cart confirmation
- Fallback API support

### 5. Checkout
- Navigates to checkout page
- Creates checkout session via API
- Displays progress in real-time

### 6. Pickup Point Selection
- Searches for available pickup points
- **Priority 1**: Instabox locations
- **Priority 2**: Budbee locations
- **Priority 3**: Closest available location
- Confirms selection

### 7. Payment
- Extracts payment session ID
- Selects BNPL (Buy Now Pay Later) method
- Updates payment via API

### 8. Purchase Completion
- Requires user confirmation (type 'YES')
- Executes final purchase
- Completes payment flow
- Displays success message

## API Endpoints Used

The application integrates with these Zalando APIs:

- `POST /api/cart-gateway/carts` - Cart management
- `GET /checkout/v3/fetch-or-create-checkout-trampoline` - Checkout creation
- `POST /api/checkout/search-pickup-points-by-address` - Pickup point search
- `POST /api/checkout/select-pickup-point` - Pickup point selection
- `POST /api/checkout/next-step` - Checkout navigation
- `POST /api/checkout/update-payment` - Payment method update
- `POST /api/checkout/buy-now` - Purchase execution
- `POST /checkout/payment-complete` - Payment completion

## Key Selectors

```go
// Size picker
#picker-trigger

// Size options
[role='option']

// Add to cart button
button[data-testid='pdp_add-to-cart-button']

// Checkout buttons
button:has-text('Till kassan')
button:has-text('Checkout')
```

## Error Handling

- **Retry Logic**: Failed clicks are retried with alternative selectors
- **Timeout Handling**: Configurable timeouts for slow pages
- **Fallback Mechanisms**: API calls fallback to page interactions
- **Detailed Logging**: Real-time console output for debugging

## Troubleshooting

### Browser Not Opening
Ensure Chrome/Chromium is installed:
```bash
# Ubuntu/Debian
sudo apt-get install chromium-browser

# macOS
brew install --cask google-chrome

# Windows
# Download from https://www.google.com/chrome/
```

### Size Not Found
- Verify the product page loaded completely
- Check that the size picker button is visible
- Ensure JavaScript is enabled

### Payment Failure
- Verify BNPL is available for your account
- Check that you have sufficient credit limit
- Ensure your payment information is up to date

## Security Considerations

⚠️ **Important Security Notes**:
- Never commit credentials to version control
- Use environment variables for sensitive data
- Keep your Zalando session secure
- This tool is for educational purposes

## License

This project is for educational purposes only. Use at your own risk.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Disclaimer

This tool is for educational purposes only. The authors are not responsible for any misuse or damages caused by this software. Always comply with Zalando's Terms of Service and applicable laws.

## Author

Created by walamagki446-sudo

## Support

For issues and questions, please open an issue on GitHub.
