"""
Checkout flow automation.
"""
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
import time
import requests


class CheckoutManager:
    """Handles checkout process."""
    
    def __init__(self, driver, api_client):
        self.driver = driver
        self.api_client = api_client
        self.checkout_id = None
    
    def navigate_to_checkout(self):
        """
        Navigate to checkout page.
        
        Returns:
            bool: True if navigation successful
        """
        print("🔄 Navigating to checkout...")
        
        try:
            # Method 1: Click checkout button
            checkout_selectors = [
                (By.XPATH, "//button[contains(., 'Checkout')]"),
                (By.XPATH, "//button[contains(., 'Till kassan')]"),
                (By.XPATH, "//a[contains(., 'Checkout')]"),
                (By.XPATH, "//a[contains(., 'Till kassan')]"),
                (By.CSS_SELECTOR, "button[data-testid='checkout-button']"),
                (By.CSS_SELECTOR, "a[href*='checkout']"),
            ]
            
            for by, selector in checkout_selectors:
                try:
                    button = WebDriverWait(self.driver, 5).until(
                        EC.element_to_be_clickable((by, selector))
                    )
                    button.click()
                    time.sleep(2)
                    return True
                except:
                    continue
            
            # Method 2: Direct URL navigation with API call
            try:
                response = self.api_client.get(
                    "checkout/v3/fetch-or-create-checkout-trampoline?checkout_variant=PHYSICAL"
                )
                if response.status_code == 200:
                    data = response.json()
                    if 'checkoutId' in data:
                        self.checkout_id = data['checkoutId']
                    
                    # Navigate to checkout URL
                    checkout_url = self.driver.current_url.split('/')[0:3]
                    checkout_url = '/'.join(checkout_url) + '/checkout'
                    self.driver.get(checkout_url)
                    time.sleep(2)
                    
                    print("✅ Navigated to checkout!")
                    return True
            except:
                pass
            
            # Method 3: Fallback direct URL
            checkout_url = self.driver.current_url.split('/')[0:3]
            checkout_url = '/'.join(checkout_url) + '/checkout'
            self.driver.get(checkout_url)
            time.sleep(2)
            
            print("✅ Navigated to checkout!")
            return True
            
        except Exception as e:
            print(f"❌ Failed to navigate to checkout: {str(e)}")
            return False
    
    def wait_for_checkout_load(self):
        """Wait for checkout page to fully load."""
        try:
            WebDriverWait(self.driver, 10).until(
                EC.presence_of_element_located((By.TAG_NAME, "body"))
            )
            time.sleep(2)
            return True
        except:
            return False
