"""
Zalando Autobuyer - Python + Selenium
Entry point for the automated purchase flow.
"""
from src.browser import BrowserManager
from src.auth import ZalandoAuth
from src.product import ProductManager
from src.cart import CartManager
from src.checkout import CheckoutManager
from src.pickup import PickupPointManager
from src.payment import PaymentManager
from src.api_client import ZalandoAPIClient
import time


def print_header():
    """Print application header."""
    print("\n" + "="*62)
    print("║" + " "*10 + "ZALANDO AUTOBUYER - PYTHON + SELENIUM" + " "*11 + "║")
    print("="*62 + "\n")


def parse_credentials(credentials_input):
    """
    Parse email:password format.
    
    Args:
        credentials_input: String in format "email:password"
        
    Returns:
        tuple: (email, password)
    """
    parts = credentials_input.split(':', 1)
    if len(parts) == 2:
        return parts[0].strip(), parts[1].strip()
    return None, None


def confirm_purchase():
    """
    Prompt user for final purchase confirmation.
    
    Returns:
        bool: True if user confirms
    """
    print("\n" + "="*62)
    print("⚠️  FINAL CONFIRMATION")
    print("="*62)
    confirm = input("\nType 'YES' to complete purchase: ").strip()
    return confirm == "YES"


def execute_purchase(api_client):
    """
    Execute final purchase.
    
    Args:
        api_client: API client instance
        
    Returns:
        bool: True if purchase successful
    """
    print("\n🔄 Executing purchase...")
    
    try:
        # Execute purchase API call
        response = api_client.post("checkout/buy-now")
        
        if response.status_code == 200:
            # Complete payment
            payment_response = api_client.post("checkout/payment-complete")
            
            if payment_response.status_code == 200:
                print("\n" + "="*62)
                print("✅ 🎉 PURCHASE COMPLETED!")
                print("="*62)
                print("📧 Check your email for order confirmation\n")
                return True
        
        print("❌ Purchase failed")
        return False
        
    except Exception as e:
        print(f"❌ Purchase error: {str(e)}")
        return False


def main():
    """Main execution flow."""
    browser_manager = None
    
    try:
        # Print header
        print_header()
        
        # Step 1: Get credentials
        credentials = input("Enter credentials (email:password): ").strip()
        email, password = parse_credentials(credentials)
        
        if not email or not password:
            print("❌ Invalid credentials format. Use: email:password")
            return
        
        # Step 2: Setup browser
        browser_manager = BrowserManager()
        driver = browser_manager.setup_driver()
        
        # Initialize API client
        api_client = ZalandoAPIClient()
        
        # Step 3: Login
        auth = ZalandoAuth(driver)
        if not auth.login(email, password):
            print("❌ Login failed. Exiting...")
            return
        
        # Transfer cookies to API client
        api_client.set_cookies(auth.get_cookies_dict())
        
        # Step 4: Get product URL
        print()
        product_url = input("Enter product URL: ").strip()
        
        # Step 5: Navigate to product
        product_manager = ProductManager(driver)
        if not product_manager.navigate_to_product(product_url):
            print("❌ Failed to load product. Exiting...")
            return
        
        # Step 6: Detect sizes
        available_sizes = product_manager.detect_sizes()
        
        if available_sizes and available_sizes[0] != "Unable to detect - manual entry":
            print(f"\nAvailable sizes: {', '.join(available_sizes)}")
        
        # Step 7: Get desired size
        desired_size = input("Enter your size: ").strip()
        
        # Step 8: Select size
        if not product_manager.select_size(desired_size):
            print("❌ Failed to select size. Exiting...")
            return
        
        # Step 9: Add to cart
        cart_manager = CartManager(driver)
        if not cart_manager.add_to_cart():
            print("❌ Failed to add to cart. Exiting...")
            return
        
        # Step 10: Navigate to checkout
        checkout_manager = CheckoutManager(driver, api_client)
        if not checkout_manager.navigate_to_checkout():
            print("❌ Failed to navigate to checkout. Exiting...")
            return
        
        checkout_manager.wait_for_checkout_load()
        
        # Step 11: Select pickup point
        pickup_manager = PickupPointManager(driver, api_client)
        pickup_point = pickup_manager.select_best_pickup_point()
        
        if not pickup_point:
            print("⚠️  No pickup point selected, continuing...")
        
        # Step 12: Detect payment session and select BNPL
        payment_manager = PaymentManager(driver, api_client)
        payment_manager.detect_session_id()
        payment_manager.select_bnpl()
        
        # Step 13: Final confirmation
        if not confirm_purchase():
            print("\n❌ Purchase cancelled by user")
            return
        
        # Step 14: Execute purchase
        execute_purchase(api_client)
        
        # Wait for user to review
        browser_manager.wait_for_user()
        
    except KeyboardInterrupt:
        print("\n\n❌ Process interrupted by user")
    except Exception as e:
        print(f"\n❌ Unexpected error: {str(e)}")
        import traceback
        traceback.print_exc()
    finally:
        # Cleanup
        if browser_manager:
            browser_manager.close()


if __name__ == '__main__':
    main()
