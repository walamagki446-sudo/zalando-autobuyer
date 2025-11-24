# Zalando Autobuyer - Python + Selenium

A Python-based automated purchase bot for Zalando using Selenium WebDriver in **visible mode** (headful browser). Watch the automation in real-time as it navigates through the entire purchase flow.

## Features

✅ **Interactive CLI** - User-friendly command-line interface  
✅ **Visible Browser Automation** - Watch Chrome browser perform actions  
✅ **Smart Size Detection** - Multiple methods to detect available sizes  
✅ **Intelligent Pickup Point Selection** - Priority: Instabox > Budbee > Closest > Farthest  
✅ **Automatic Session Management** - Detects payment session IDs automatically  
✅ **BNPL Payment Support** - Buy Now Pay Later payment method  
✅ **API Integration** - Combines browser automation with API calls for reliability  
✅ **Error Handling** - Robust error handling with multiple fallback strategies  

## Technology Stack

- **Python 3.9+**
- **Selenium WebDriver** - Browser automation
- **Chrome/ChromeDriver** - Visible browser (headful mode)
- **Requests** - HTTP client for API calls
- **webdriver-manager** - Automatic ChromeDriver management

## Project Structure

```
zalando-autobuyer/
├── main.py                     # Entry point
├── src/
│   ├── __init__.py
│   ├── browser.py              # Selenium setup
│   ├── auth.py                 # Login automation
│   ├── product.py              # Size detection & selection
│   ├── cart.py                 # Add to cart
│   ├── checkout.py             # Checkout flow
│   ├── pickup.py               # Pickup point logic
│   ├── payment.py              # BNPL payment
│   └── api_client.py           # HTTP client
├── requirements.txt
├── README.md
└── .gitignore
```

## Installation

### Prerequisites

- Python 3.9 or higher
- Google Chrome browser installed

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

The `webdriver-manager` will automatically download the appropriate ChromeDriver version.

## Usage

Run the autobuyer:

```bash
python main.py
```

### Complete User Flow

#### 1. **Login**
```
Enter credentials (email:password): your.email@example.com:yourpassword
```
- Format: `email:password`
- Browser opens and navigates to Zalando login
- Credentials are entered automatically
- Session cookies are saved

#### 2. **Product Selection**
```
Enter product URL: https://www.zalando.se/product-name-sku.html
```
- Navigates to product page
- Waits for page to load

#### 3. **Size Detection & Selection**
```
Available sizes: XS, S, M, L, XL, XXL
Enter your size: L
```
- Automatically detects available sizes using multiple methods
- User selects desired size
- Size is selected in the browser

#### 4. **Add to Cart**
- Automatically clicks "Handla" button
- Verifies item added to cart
- Uses multiple strategies to find the button

#### 5. **Checkout**
- Navigates to checkout page
- Waits for page to load
- Extracts checkout ID

#### 6. **Pickup Point Selection**
Smart priority logic:
1. **Instabox** locations (Priority 1)
2. **Budbee** locations (Priority 2)
3. **Closest** to address (Priority 3)
4. **Farthest** available (Priority 4)

#### 7. **Payment Method**
- Automatically detects payment session ID
- Selects BNPL (Buy Now Pay Later) payment method

#### 8. **Final Confirmation**
```
⚠️  Type 'YES' to complete purchase: YES
```
- User must type 'YES' to confirm
- Executes purchase via API
- Displays confirmation

## Console Output Example

```
==============================================================
║          ZALANDO AUTOBUYER - PYTHON + SELENIUM           ║
==============================================================

Enter credentials (email:password): user@example.com:password123

🔄 Opening Chrome browser...
🔐 Logging in to Zalando...
✅ Login successful!

Enter product URL: https://www.zalando.se/adidas-spezial-ad116b0d9-a11.html

🔄 Navigating to product...
📦 Detecting available sizes...
✅ Found sizes: XS, S, M, L, XL, XXL

Enter your size: L

🔄 Selecting size L...
✅ Size L selected!
🛒 Clicking 'Handla' button...
✅ Added to cart!

🔄 Navigating to checkout...
📍 Searching for pickup points...

Found 8 pickup points:
  ✓ Priority 1: Instabox Stockholm City (Instabox)

✅ Selected: Instabox Stockholm City

💳 Detecting payment session...
✅ Session ID: ce32de21-55ed-4a5d-8eb4-b0aaa00591ff
💳 Selecting BNPL payment...
✅ Payment method selected!

==============================================================
⚠️  FINAL CONFIRMATION
==============================================================

Type 'YES' to complete purchase: YES

🔄 Executing purchase...

==============================================================
✅ 🎉 PURCHASE COMPLETED!
==============================================================
📧 Check your email for order confirmation

Press Enter to close browser...
```

## Key Features Explained

### Visible Browser Mode
Unlike headless browsers, this autobuyer runs Chrome in **visible mode**. You can watch every action being performed:
- Login process
- Product navigation
- Size selection
- Adding to cart
- Checkout process

### Multiple Detection Strategies

#### Size Detection
The bot uses 3 different methods to detect available sizes:
1. **Role-based detection** - Uses ARIA roles and attributes
2. **ARIA controls** - Follows dropdown relationships
3. **Pattern matching** - Recognizes common size patterns (XS-XXL, 34-52)

#### "Handla" Button Detection
Multiple strategies to find the add-to-cart button:
1. Text content matching
2. Span text matching
3. Class pattern matching
4. ARIA label matching
5. Data attribute matching

### Smart Pickup Point Selection
Automatically selects the best pickup point based on:
1. **Provider priority** - Instabox and Budbee preferred
2. **Distance** - Closest location if no preferred providers
3. **Fallback** - Farthest location as last resort

### Session Management
- Automatically detects payment session IDs from:
  - Network traffic logs
  - JavaScript context
  - API responses
- Transfers cookies between browser and API client
- Maintains session throughout the flow

## Error Handling

The autobuyer includes robust error handling:
- Multiple fallback strategies for each action
- Graceful degradation if automation fails
- Informative error messages
- Safe cleanup on exit

## Security Notes

⚠️ **Important Security Considerations:**

1. **Credentials** - Never commit credentials to version control
2. **API Keys** - Store sensitive data securely
3. **Session Tokens** - Handled in memory only
4. **HTTPS** - All API calls use secure connections

## Limitations

- Requires active Chrome browser window
- Dependent on Zalando's website structure
- May need updates if website changes
- Not suitable for high-frequency automation

## Troubleshooting

### ChromeDriver Issues
If you encounter ChromeDriver errors:
```bash
pip install --upgrade webdriver-manager
```

### Timeout Errors
If pages load slowly, the bot includes wait times. You may need to adjust timeouts in the code for slower connections.

### Element Not Found
If elements aren't found, Zalando may have updated their website. Check the selectors in the relevant modules.

## Legal Disclaimer

This tool is for **educational purposes only**. Users are responsible for:
- Complying with Zalando's Terms of Service
- Following all applicable laws and regulations
- Using the tool responsibly and ethically

Automated purchasing may violate terms of service. Use at your own risk.

## Contributing

Contributions are welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## License

This project is provided as-is for educational purposes.

## Support

For issues or questions:
1. Check existing issues on GitHub
2. Review the troubleshooting section
3. Open a new issue with detailed information

## Acknowledgments

- Selenium WebDriver team
- webdriver-manager contributors
- Python community

---

**Remember:** Use this tool responsibly and in accordance with Zalando's terms of service.
