"""
Shopping cart operations: Add to cart ("Handla" button).
"""
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
import time


class CartManager:
    """Handles shopping cart operations."""
    
    def __init__(self, driver):
        self.driver = driver
    
    def add_to_cart(self):
        """
        Click "Handla" button to add item to cart.
        Uses multiple strategies to find and click the button.
        
        Returns:
            bool: True if successfully added to cart
        """
        print("🛒 Clicking 'Handla' button...")
        
        # Multiple strategies to find the button
        strategies = [
            # Strategy 1: By text content "Handla"
            (By.XPATH, "//button[contains(., 'Handla')]"),
            (By.XPATH, "//button[contains(text(), 'Handla')]"),
            
            # Strategy 2: By span text
            (By.XPATH, "//span[text()='Handla']/parent::button"),
            (By.XPATH, "//span[contains(text(), 'Handla')]/parent::button"),
            
            # Strategy 3: By class patterns (common Zalando classes)
            (By.CSS_SELECTOR, "button.EKabf7"),
            (By.CSS_SELECTOR, "button.EKabf7.aX2-iv"),
            (By.XPATH, "//button[contains(@class, 'EKabf7')]"),
            
            # Strategy 4: By aria-label or data-testid
            (By.CSS_SELECTOR, "button[aria-label*='Handla']"),
            (By.CSS_SELECTOR, "button[data-testid*='add-to-cart']"),
            (By.CSS_SELECTOR, "button[data-testid*='handla']"),
            
            # Strategy 5: Generic add to cart patterns
            (By.CSS_SELECTOR, "button[type='submit']"),
        ]
        
        for by, selector in strategies:
            try:
                button = WebDriverWait(self.driver, 5).until(
                    EC.element_to_be_clickable((by, selector))
                )
                
                # Verify this is likely the right button
                button_text = button.text.strip().lower()
                if 'handla' in button_text or 'lägg till' in button_text or \
                   'add to' in button_text or button_text == '':
                    button.click()
                    time.sleep(2)
                    
                    # Verify cart confirmation
                    self._verify_cart_addition()
                    
                    print("✅ Added to cart!")
                    return True
            except Exception as e:
                continue
        
        print("❌ Could not find or click 'Handla' button")
        return False
    
    def _verify_cart_addition(self):
        """
        Verify that item was added to cart by checking for confirmation.
        """
        try:
            # Wait for cart confirmation/animation
            time.sleep(1)
            
            # Look for cart icon with item count or confirmation message
            cart_indicators = [
                (By.CSS_SELECTOR, "[data-testid='cart-icon']"),
                (By.CSS_SELECTOR, ".cart-icon"),
                (By.XPATH, "//*[contains(@class, 'cart')]"),
            ]
            
            for by, selector in cart_indicators:
                try:
                    self.driver.find_element(by, selector)
                    return True
                except Exception:
                    continue
                    
        except:
            pass
        
        return False
    
    def navigate_to_cart(self):
        """
        Navigate to shopping cart.
        
        Returns:
            bool: True if navigation successful
        """
        print("🔄 Navigating to cart...")
        
        try:
            # Try to click cart icon
            cart_selectors = [
                (By.CSS_SELECTOR, "[data-testid='cart-icon']"),
                (By.CSS_SELECTOR, "a[href*='cart']"),
                (By.XPATH, "//a[contains(@href, 'cart')]"),
            ]
            
            for by, selector in cart_selectors:
                try:
                    cart_link = self.driver.find_element(by, selector)
                    cart_link.click()
                    time.sleep(2)
                    return True
                except Exception:
                    continue
            
            # Fallback: Navigate directly to cart URL
            cart_url = self.driver.current_url.split('/')[0:3]
            cart_url = '/'.join(cart_url) + '/cart'
            self.driver.get(cart_url)
            time.sleep(2)
            
            return True
            
        except Exception as e:
            print(f"Note: Cart navigation error: {str(e)}")
            return False
