# Zalando Autobuyer

A complete Zalando autobuyer application written in Go with a graphical user interface.

## Features

- **Product Detection**: Parse Zalando product URLs to detect all available sizes
- **Size Selection**: View and select from available sizes with stock status
- **Cart Management**: Add selected items to cart automatically
- **Authentication**: Login using email and password
- **Pickup Point Selection**: Automatically detects and selects Instabox pickup points
- **Checkout**: Complete checkout with Faktura (invoice) payment method
- **Anti-Detection**: User-Agent rotation, realistic headers, and human-like delays
- **Proxy Support**: Built-in proxy support for anonymity

## Requirements

- Go 1.19 or higher
- Internet connection
- Valid Zalando account (for checkout)

## Installation

### From Source

1. Clone the repository:
```bash
git clone https://github.com/walamagki446-sudo/zalando-autobuyer.git
cd zalando-autobuyer
```

2. Install dependencies:
```bash
go mod tidy
```

3. Build the application:
```bash
# For Linux
go build -o zalando-autobuyer .

# For Windows
GOOS=windows GOARCH=amd64 go build -o zalando-autobuyer.exe .

# For macOS
GOOS=darwin GOARCH=amd64 go build -o zalando-autobuyer-mac .
```

4. Run the application:
```bash
./zalando-autobuyer
```

## Usage

### Step 1: Enter Product URL
1. Open the application
2. Paste a Zalando product URL in the "Product URL" field
3. Click "Fetch Sizes" to load available sizes

### Step 2: Select Size
1. Choose your desired size from the dropdown menu
2. Sizes marked with ✓ are in stock, ✗ are out of stock
3. Click "Add to Cart" to add the item to your cart

### Step 3: Configure Delivery
1. Enter your login credentials (email and password)
2. Fill in your delivery address details:
   - First Name
   - Last Name
   - Street Address
   - Zip Code
   - City
   - Phone (optional)

### Step 4: Complete Checkout
1. Optionally configure a proxy in the "Proxy" field
2. Click "Login & Checkout" to start the checkout process
3. The application will:
   - Log you in
   - Add the item to cart (if not already added)
   - Search for Instabox pickup points near your address
   - Select the first available Instabox
   - Choose Faktura as payment method
   - Confirm the order

### Step 5: Monitor Progress
- Watch the status log for real-time updates
- The progress bar shows the current step
- If successful, you'll see the order confirmation number

## Proxy Configuration

The proxy field accepts proxies in the format:
```
host:port:username:password
```

Example:
```
budget.waveproxies.com:1337:username:password
```

## Anti-Detection Features

The application includes several anti-detection measures:

- **User-Agent Rotation**: Cycles through realistic browser user agents
- **Realistic Headers**: Includes all necessary headers that a real browser would send
- **Random Delays**: Human-like delays between requests (100-800ms)
- **Error Delays**: Longer delays after errors to avoid rate limiting (1500-2500ms)
- **Cookie Management**: Proper cookie jar handling and session persistence
- **CSRF Token Handling**: Automatic extraction and usage of CSRF tokens

## File Structure

```
zalando-autobuyer/
├── main.go         # GUI and main application logic
├── auth.go         # Login and authentication
├── product.go      # Product fetching and size detection
├── cart.go         # Cart operations
├── checkout.go     # Checkout process including pickup points
├── proxy.go        # Proxy management
├── utils.go        # HTTP client, headers, delays
├── go.mod          # Go module file
├── go.sum          # Go dependencies checksum
└── README.md       # This file
```

## Troubleshooting

### "No sizes found"
- The product page structure may have changed
- Try a different product URL
- Check if the product is available in your region

### "Login failed"
- Verify your credentials are correct
- Check if your account is not locked
- Wait and try again (rate limiting)

### "No pickup points found"
- Verify your address is correct
- Instabox may not be available in your area
- The application will fall back to other pickup points

### "Rate limited"
- Wait a few minutes before trying again
- Consider using a proxy
- Reduce the frequency of requests

## Disclaimer

This application is for educational purposes only. Use at your own risk. Automated purchasing may violate Zalando's terms of service.

## License

MIT License - see LICENSE file for details.
