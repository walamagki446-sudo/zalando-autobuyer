# Zalando Autobuyer - Interactive CLI Application

A fully interactive command-line tool for automating Zalando purchases with a polished, professional user experience.

## Features

✅ **Interactive CLI** - Guided step-by-step purchase flow  
✅ **SKU Auto-extraction** - Automatically extracts SKU from Zalando URLs  
✅ **Multiple Delivery Options** - Pickup points or home delivery  
✅ **Safe Testing** - Dry-run mode to test without making purchases  
✅ **Multi-region Support** - Works with Zalando in SE, DE, UK, FR, NL, ES, IT  
✅ **Robust Error Handling** - Graceful failures with helpful messages  
✅ **Progress Indicators** - Clear visual feedback at each step  

## Installation

### Prerequisites

- Python 3.7 or higher
- pip (Python package manager)

### Setup

1. Clone the repository:
```bash
git clone https://github.com/walamagki446-sudo/zalando-autobuyer.git
cd zalando-autobuyer
```

2. Install dependencies:
```bash
pip install -r requirements.txt
```

## Usage

### Basic Usage

Run the interactive application:

```bash
python main.py
```

### Command Line Options

- `--dry-run` - Stop before actual purchase (safe testing)
- `--verbose` - Enable debug logging
- `--region REGION` - Select Zalando region (se, de, uk, fr, nl, es, it)

### Examples

#### Standard Purchase

```bash
python main.py
```

#### Dry Run (Safe Testing)

```bash
python main.py --dry-run
```

#### Different Region with Verbose Logging

```bash
python main.py --region de --verbose
```

## Interactive Flow Example

```
╔══════════════════════════════════════════════════════════╗
║               ZALANDO AUTOBUYER                          ║
║            Interactive Mode Activated                    ║
╚══════════════════════════════════════════════════════════╝

🔐 ZALANDO LOGIN
═══════════════════════════════════════════════════════════
Enter credentials (email:password): user@example.com:password123
🔄 Logging in as: user@example.com
✅ Login successful!

🛍️ PRODUCT SELECTION
═══════════════════════════════════════════════════════════
Enter product URL: https://www.zalando.se/adidas-originals-ad116b0d9-a11.html
📦 Detected SKU: ad116b0d9-a11
Enter size (e.g., S, M, L, XL, 38, 40, etc.): 42
Enter quantity (default: 1): 1

🔄 Adding to cart: 1x ad116b0d9-a11 (Size: 42)
✅ Product added to cart!

🔄 Creating checkout session...
✅ Checkout created: checkout_xyz123

📦 DELIVERY OPTIONS
═══════════════════════════════════════════════════════════
Choose delivery type:
1. Pickup Point
2. Home Delivery
Enter choice (1 or 2): 1

Enter your address for pickup point search: Stockholm, 11122
🔄 Searching for pickup points near Stockholm, 11122...

📍 Found 5 pickup points:
1. Pressbyrån Stockholm City - Vasagatan 10, Stockholm (0.5km)
2. Circle K - Drottninggatan 50, Stockholm (0.8km)
3. ICA Maxi - Kungsgatan 44, Stockholm (1.2km)
4. 7-Eleven - Sveavägen 24, Stockholm (1.5km)
5. Coop - Odengatan 67, Stockholm (2.0km)

Select pickup point (1-5), or press Enter for first: 1
🔄 Selecting: Pressbyrån Stockholm City...
✅ Selected: Pressbyrån Stockholm City

💳 PAYMENT
═══════════════════════════════════════════════════════════
⚠️  WARNING: You are about to make a REAL purchase!
Type 'YES' to confirm and proceed with payment: YES

🔄 Executing purchase...
🔄 Completing payment...

═══════════════════════════════════════════════════════════
✅ 🎉 PURCHASE COMPLETED SUCCESSFULLY! 🎉
═══════════════════════════════════════════════════════════
📧 Check your email for order confirmation
📦 Order ID: order_abc123
```

## Module Structure

```
zalando-autobuyer/
├── main.py                 # Interactive CLI orchestrator
├── requirements.txt        # Python dependencies
├── README.md              # This file
└── src/
    ├── __init__.py        # Package initialization
    ├── api_client.py      # Base API client with retry logic
    ├── auth.py            # Authentication handler
    ├── cart.py            # Cart management (GraphQL + REST fallback)
    ├── checkout.py        # Checkout flow with pickup points
    └── payment.py         # Payment processing
```

## How It Works

### 1. Login
- Enter credentials in `email:password` format
- Authenticates with Zalando API
- Maintains session for subsequent requests

### 2. Product Selection
- Paste Zalando product URL
- SKU is automatically extracted from URL
- Enter desired size and quantity

### 3. Add to Cart
- Attempts GraphQL API first (primary method)
- Falls back to REST API if GraphQL fails
- Verifies cart contents

### 4. Delivery Selection
- Choose between pickup point or home delivery
- For pickup points:
  - Search by address
  - View list of nearby locations
  - Select preferred pickup point

### 5. Payment
- Displays clear warning before purchase
- Requires explicit 'YES' confirmation
- Executes payment and completes order
- In dry-run mode, stops before payment

## SKU Extraction

The tool automatically extracts SKU codes from Zalando product URLs.

**URL Pattern:**
```
https://www.zalando.XX/product-name-SKUCODE.html
```

**Examples:**
- `https://www.zalando.se/adidas-originals-ad116b0d9-a11.html` → SKU: `ad116b0d9-a11`
- `https://www.zalando.de/nike-running-nk123x456-b12.html` → SKU: `nk123x456-b12`

## Safety Features

### Dry-Run Mode
Test the entire flow without making a real purchase:

```bash
python main.py --dry-run
```

The tool will:
- Go through all steps (login, cart, checkout, delivery)
- Stop before executing payment
- Display "DRY RUN MODE" message
- Confirm all steps completed successfully

### Explicit Confirmation
Before making a purchase, you must type exactly `YES` to confirm.
Any other input will cancel the purchase.

### Keyboard Interrupt
Press `Ctrl+C` at any time to safely cancel the operation.

## Error Handling

The tool handles errors gracefully:

- **Invalid credentials** - Clear error message, option to retry
- **Invalid URL/SKU** - Prompts for correct format
- **No pickup points found** - Option to try different address
- **Network errors** - Automatic retry with exponential backoff
- **API failures** - Fallback methods when available

## Logging

### Normal Mode
Shows important progress messages and errors only.

### Verbose Mode
Enable detailed debug logging:

```bash
python main.py --verbose
```

Displays:
- API request details
- Response data
- Internal state changes
- Detailed error traces

## Supported Regions

The tool supports multiple Zalando regions:

- 🇸🇪 Sweden (`se`) - Default
- 🇩🇪 Germany (`de`)
- 🇬🇧 United Kingdom (`uk`)
- 🇫🇷 France (`fr`)
- 🇳🇱 Netherlands (`nl`)
- 🇪🇸 Spain (`es`)
- 🇮🇹 Italy (`it`)

Set region with `--region` flag:
```bash
python main.py --region de
```

## Troubleshooting

### "Could not extract SKU from URL"
- Ensure the URL is from Zalando
- URL should be in format: `https://www.zalando.XX/product-name-SKU.html`
- Try copying the URL directly from the browser address bar

### "Login failed"
- Check credentials are correct
- Ensure format is `email:password` with no extra spaces
- Try accessing Zalando website in a browser first

### "Failed to add product to cart"
- Product might be out of stock
- Size might not be available
- Try with `--verbose` flag to see detailed error

### "No pickup points found"
- Try a more specific address with postal code
- Ensure address is in the correct region

## Development

### Project Structure

The codebase is organized into focused modules:

- **api_client.py** - HTTP client with retry logic and session management
- **auth.py** - Handles login and credential parsing
- **cart.py** - Cart operations with GraphQL primary and REST fallback
- **checkout.py** - Checkout flow including pickup point search and selection
- **payment.py** - Payment execution and order completion
- **main.py** - CLI interface and user interaction flow

### Adding New Features

1. Keep modules focused on single responsibility
2. Add new functionality to appropriate module
3. Update main.py to integrate new features
4. Test with `--dry-run` mode first
5. Update README with new usage examples

## License

This project is for educational purposes. Always comply with Zalando's Terms of Service.

## Disclaimer

This tool is provided as-is for educational purposes. The authors are not responsible for any misuse or violations of Zalando's Terms of Service. Always ensure you have permission to automate actions on any website.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Support

For issues, questions, or suggestions, please open an issue on GitHub.

---

Made with ❤️ for the automation community
