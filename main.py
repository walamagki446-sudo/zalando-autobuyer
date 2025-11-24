#!/usr/bin/env python3
"""
Zalando Autobuyer - Interactive CLI Application

A fully interactive command-line tool for automating Zalando purchases.
"""

import argparse
import logging
import re
import sys
from typing import Optional, Tuple

from src.api_client import ZalandoAPIClient
from src.auth import ZalandoAuth
from src.cart import ZalandoCart
from src.checkout import ZalandoCheckout, PickupPoint
from src.payment import ZalandoPayment


# ASCII Art and UI Elements
HEADER = """
╔══════════════════════════════════════════════════════════╗
║               ZALANDO AUTOBUYER                          ║
║            Interactive Mode Activated                    ║
╚══════════════════════════════════════════════════════════╝
"""

SEPARATOR = "═" * 63


def setup_logging(verbose: bool = False):
    """Configure logging based on verbosity level."""
    level = logging.DEBUG if verbose else logging.INFO
    logging.basicConfig(
        level=level,
        format='%(levelname)s: %(message)s' if not verbose else '%(asctime)s - %(name)s - %(levelname)s - %(message)s',
        datefmt='%Y-%m-%d %H:%M:%S'
    )


def print_header():
    """Print the welcome header."""
    print(HEADER)


def print_section(title: str):
    """Print a section header."""
    print(f"\n{title}")
    print(SEPARATOR)


def extract_sku_from_url(product_url: str) -> Optional[str]:
    """
    Extract SKU from Zalando product URL.
    
    Pattern: https://www.zalando.XX/product-name-SKUCODE.html
    Example: https://www.zalando.se/adidas-originals-ad116b0d9-a11.html
    
    Args:
        product_url: Zalando product URL
        
    Returns:
        SKU code if found, None otherwise
    """
    # Pattern to match Zalando URLs
    pattern = r'zalando\.[a-z]{2}/[^/]+-([a-z0-9]+-[a-z0-9]+)\.html'
    match = re.search(pattern, product_url.lower())
    
    if match:
        return match.group(1)
    
    # Try alternative pattern without .html
    pattern = r'zalando\.[a-z]{2}/[^/]+-([a-z0-9]+-[a-z0-9]+)'
    match = re.search(pattern, product_url.lower())
    
    if match:
        return match.group(1)
    
    return None


def get_login_credentials() -> Tuple[str, str]:
    """
    Get login credentials from user.
    
    Returns:
        Tuple of (email, password)
    """
    print_section("🔐 ZALANDO LOGIN")
    
    credentials = input("Enter credentials (email:password): ").strip()
    
    try:
        email, password = ZalandoAuth.parse_credentials(credentials)
        return email, password
    except ValueError as e:
        print(f"❌ Error: {e}")
        print("\nPlease enter credentials in format: email:password")
        return get_login_credentials()


def get_product_details() -> Tuple[str, str, int]:
    """
    Get product details from user.
    
    Returns:
        Tuple of (sku, size, quantity)
    """
    print_section("🛍️  PRODUCT SELECTION")
    
    # Get product URL
    product_url = input("Enter product URL: ").strip()
    
    # Extract SKU
    sku = extract_sku_from_url(product_url)
    
    if not sku:
        print("❌ Could not extract SKU from URL")
        print("Please ensure URL is in format: https://www.zalando.XX/product-name-SKU.html")
        return get_product_details()
    
    print(f"📦 Detected SKU: {sku}")
    
    # Get size
    size = input("Enter size (e.g., S, M, L, XL, 38, 40, etc.): ").strip()
    
    if not size:
        print("❌ Size cannot be empty")
        return get_product_details()
    
    # Get quantity
    quantity_str = input("Enter quantity (default: 1): ").strip()
    quantity = 1
    
    if quantity_str:
        try:
            quantity = int(quantity_str)
            if quantity < 1:
                print("❌ Quantity must be at least 1")
                return get_product_details()
        except ValueError:
            print("❌ Invalid quantity, using default: 1")
            quantity = 1
    
    return sku, size, quantity


def select_delivery_method(checkout: ZalandoCheckout) -> bool:
    """
    Interactive delivery method selection.
    
    Args:
        checkout: ZalandoCheckout instance
        
    Returns:
        True if successful, False otherwise
    """
    print_section("📦 DELIVERY OPTIONS")
    
    print("Choose delivery type:")
    print("1. Pickup Point")
    print("2. Home Delivery")
    
    choice = input("Enter choice (1 or 2): ").strip()
    
    if choice == "1":
        return handle_pickup_point_selection(checkout)
    elif choice == "2":
        return handle_home_delivery(checkout)
    else:
        print("❌ Invalid choice. Please enter 1 or 2")
        return select_delivery_method(checkout)


def handle_pickup_point_selection(checkout: ZalandoCheckout) -> bool:
    """
    Handle pickup point selection flow.
    
    Args:
        checkout: ZalandoCheckout instance
        
    Returns:
        True if successful, False otherwise
    """
    address = input("\nEnter your address for pickup point search: ").strip()
    
    if not address:
        print("❌ Address cannot be empty")
        return handle_pickup_point_selection(checkout)
    
    print(f"🔄 Searching for pickup points near {address}...")
    pickup_points = checkout.search_pickup_points(address)
    
    if not pickup_points:
        print("❌ No pickup points found")
        retry = input("Try another address? (y/n): ").strip().lower()
        if retry == 'y':
            return handle_pickup_point_selection(checkout)
        return False
    
    print(f"\n📍 Found {len(pickup_points)} pickup points:")
    for i, point in enumerate(pickup_points, 1):
        print(f"{i}. {point}")
    
    selection = input(f"\nSelect pickup point (1-{len(pickup_points)}), or press Enter for first: ").strip()
    
    if not selection:
        selected_index = 0
    else:
        try:
            selected_index = int(selection) - 1
            if selected_index < 0 or selected_index >= len(pickup_points):
                print(f"❌ Invalid selection. Please enter a number between 1 and {len(pickup_points)}")
                return handle_pickup_point_selection(checkout)
        except ValueError:
            print("❌ Invalid input")
            return handle_pickup_point_selection(checkout)
    
    selected_point = pickup_points[selected_index]
    print(f"🔄 Selecting: {selected_point.name}...")
    
    if checkout.select_pickup_point(selected_point):
        print(f"✅ Selected: {selected_point.name}")
        return True
    else:
        print("❌ Failed to select pickup point")
        return False


def handle_home_delivery(checkout: ZalandoCheckout) -> bool:
    """
    Handle home delivery flow.
    
    Args:
        checkout: ZalandoCheckout instance
        
    Returns:
        True if successful, False otherwise
    """
    address = input("\nEnter delivery address: ").strip()
    
    if not address:
        print("❌ Address cannot be empty")
        return handle_home_delivery(checkout)
    
    print(f"🔄 Setting home delivery to: {address}...")
    
    if checkout.set_home_delivery(address):
        print("✅ Home delivery configured")
        return True
    else:
        print("❌ Failed to configure home delivery")
        return False


def confirm_payment(dry_run: bool = False) -> bool:
    """
    Get payment confirmation from user.
    
    Args:
        dry_run: If True, skip actual payment
        
    Returns:
        True if user confirmed, False otherwise
    """
    print_section("💳 PAYMENT")
    
    if dry_run:
        print("⚠️  DRY RUN MODE - Stopping before payment")
        print("✅ All steps completed successfully!")
        print("\nTo make a real purchase, run without --dry-run flag")
        return False
    
    print("⚠️  WARNING: You are about to make a REAL purchase!")
    confirmation = input("Type 'YES' to confirm and proceed with payment: ").strip()
    
    return confirmation == 'YES'


def main():
    """Main interactive CLI flow."""
    # Parse command line arguments
    parser = argparse.ArgumentParser(
        description="Zalando Autobuyer - Interactive CLI Application",
        formatter_class=argparse.RawDescriptionHelpFormatter
    )
    
    parser.add_argument(
        '--dry-run',
        action='store_true',
        help='Stop before actual purchase (safe testing)'
    )
    
    parser.add_argument(
        '--verbose',
        action='store_true',
        help='Enable debug logging'
    )
    
    parser.add_argument(
        '--region',
        default='se',
        choices=['se', 'de', 'uk', 'fr', 'nl', 'es', 'it'],
        help='Select Zalando region (default: se)'
    )
    
    args = parser.parse_args()
    
    # Setup logging
    setup_logging(args.verbose)
    
    # Print welcome header
    print_header()
    
    try:
        # Initialize API client
        client = ZalandoAPIClient(region=args.region)
        
        # Step 1: Login
        email, password = get_login_credentials()
        print(f"🔄 Logging in as: {email}")
        
        auth = ZalandoAuth(client)
        if not auth.login(email, password):
            print("❌ Login failed. Please check your credentials.")
            sys.exit(1)
        
        print("✅ Login successful!")
        
        # Step 2: Product Selection
        sku, size, quantity = get_product_details()
        
        # Step 3: Add to Cart
        print(f"\n🔄 Adding to cart: {quantity}x {sku} (Size: {size})")
        cart = ZalandoCart(client)
        
        if not cart.add_to_cart(sku, size, quantity):
            print("❌ Failed to add product to cart")
            sys.exit(1)
        
        print("✅ Product added to cart!")
        
        # Step 4: Create Checkout
        cart_id = cart.get_cart_id()
        if not cart_id:
            print("❌ No cart ID available")
            sys.exit(1)
        
        print(f"\n🔄 Creating checkout session...")
        checkout = ZalandoCheckout(client)
        checkout_id = checkout.create_checkout(cart_id)
        
        if not checkout_id:
            print("❌ Failed to create checkout")
            sys.exit(1)
        
        print(f"✅ Checkout created: {checkout_id}")
        
        # Step 5: Delivery Selection
        if not select_delivery_method(checkout):
            print("❌ Failed to configure delivery")
            sys.exit(1)
        
        # Step 6: Payment
        if not confirm_payment(args.dry_run):
            sys.exit(0)
        
        print("\n🔄 Executing purchase...")
        payment = ZalandoPayment(client)
        
        if not payment.execute_payment(checkout_id):
            print("❌ Payment failed")
            sys.exit(1)
        
        print("🔄 Completing payment...")
        
        if not payment.complete_order(checkout_id):
            print("❌ Failed to complete order")
            sys.exit(1)
        
        # Success!
        print(f"\n{SEPARATOR}")
        print("✅ 🎉 PURCHASE COMPLETED SUCCESSFULLY! 🎉")
        print(SEPARATOR)
        print("📧 Check your email for order confirmation")
        
        order_id = payment.get_order_id()
        if order_id:
            print(f"📦 Order ID: {order_id}")
        
    except KeyboardInterrupt:
        print("\n\n❌ Operation cancelled by user")
        sys.exit(1)
    except Exception as e:
        logging.error(f"Unexpected error: {e}", exc_info=args.verbose)
        print(f"\n❌ An error occurred: {e}")
        sys.exit(1)


if __name__ == '__main__':
    main()
