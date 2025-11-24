# Zalando API Endpoints Reference

This document contains the actual Zalando API endpoints discovered through network inspection.
These endpoints are for reference purposes only and should be used responsibly.

⚠️ **WARNING**: Automating interactions with Zalando's APIs may violate their Terms of Service.
Use this information at your own risk and for educational purposes only.

## Authentication

### Login
```
POST https://accounts.zalando.com/authenticate?client_id=fashion-store-web
```
- Handles user authentication
- Returns session cookies and tokens

## Cart Operations

### Get/Create Cart
```
GET/POST https://www.zalando.{region}/api/cart-gateway/carts
```
- Retrieves or creates a shopping cart
- Returns cart ID and contents

### Add to Cart (GraphQL)
```
POST https://www.zalando.{region}/api/graphql/add-to-cart/
```
- Primary method for adding items to cart
- Uses GraphQL mutations

### Buy Now (Quick Checkout)
```
POST https://www.zalando.{region}/api/checkout/buy-now
```
- Initiates quick checkout flow
- Alternative to standard cart flow

## Checkout Process

### Create Checkout Session
```
GET https://www.zalando.{region}/checkout/v3/fetch-or-create-checkout-trampoline?checkout_variant=PHYSICAL
```
- Initializes checkout session
- Returns checkout session ID

### Search Pickup Points
```
POST https://www.zalando.{region}/api/checkout/search-pickup-points-by-address
```
- Searches for available pickup points near an address
- Returns list of pickup locations with details

### Select Pickup Point
```
POST https://www.zalando.{region}/api/checkout/select-pickup-point
```
- Sets the selected pickup point for delivery
- Requires pickup point ID from search results

### Next Step
```
POST https://www.zalando.{region}/api/checkout/next-step
```
- Progresses checkout to next stage
- Returns next required actions

## Payment

### Update Payment Method
```
POST https://www.zalando.{region}/api/checkout/update-payment
```
- Sets the payment method (e.g., BNPL)
- Updates checkout with payment selection

### Payment Methods (BNPL)
```
OPTIONS https://purchase-session.client-api.payment.zalando.com/sessions/{session_id}/checkout/payment-methods/bnpl
```
- Gets available BNPL (Buy Now Pay Later) options
- Session ID is dynamically generated per checkout

### Complete Payment
```
POST https://www.zalando.{region}/checkout/payment-complete
```
- Finalizes the purchase
- Confirms payment and creates order

## Implementation Notes

### Session Management
- Each checkout creates a unique session ID
- Session ID must be detected dynamically from responses
- Cookies must be maintained throughout the flow

### Preferred Delivery Partners
When selecting pickup points, prioritize:
1. **Instabox** - Fast delivery service
2. **Budbee** - Flexible pickup times

### Size Selection
- Product pages have a "Välj Storlek" (Choose Size) button
- JavaScript selector: `document.querySelector("#picker-trigger")`
- Sizes vary by product type (S/M/L for clothes, numeric for shoes)

### Headers Required
```javascript
{
    'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
    'Accept': 'application/json',
    'Accept-Language': 'sv-SE,sv;q=0.9,en-US;q=0.8,en;q=0.7',
    'Content-Type': 'application/json',
    'Origin': 'https://www.zalando.se',
    'Referer': 'https://www.zalando.se/'
}
```

## Security Considerations

1. **Never hardcode credentials** - Always prompt for user input
2. **Respect rate limits** - Add delays between requests
3. **Handle sessions securely** - Don't log sensitive tokens
4. **Use HTTPS only** - All Zalando APIs use secure connections
5. **Terms of Service** - Review Zalando's ToS before automation

## Browser Automation Alternative

For visual monitoring and more reliable automation, consider using:
- **Selenium** or **Playwright** for browser automation
- Allows seeing the bot in action
- Handles JavaScript-heavy pages naturally
- Automatically manages sessions and cookies

Example structure:
```python
from playwright.sync_api import sync_playwright

with sync_playwright() as p:
    browser = p.chromium.launch(headless=False)  # headless=False to see it
    page = browser.new_page()
    page.goto('https://www.zalando.se')
    # Interact with page...
```

## Development Recommendations

1. **Start with browser automation** for rapid prototyping
2. **Record network traffic** to understand API flows
3. **Test in dry-run mode** extensively before real purchases
4. **Implement proper error handling** for API failures
5. **Add logging** to track each step of the process

## Disclaimer

This documentation is provided for educational purposes only. The author is not responsible for:
- Any violations of Zalando's Terms of Service
- Unintended purchases or financial losses
- Legal consequences of automated interactions
- Any other misuse of this information

Always use automation tools responsibly and ethically.
