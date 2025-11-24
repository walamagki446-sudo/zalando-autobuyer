"""
Authentication and login automation for Zalando.
"""
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
import time


class ZalandoAuth:
    """Handles Zalando authentication."""
    
    LOGIN_URL = "https://accounts.zalando.com/authenticate?client_id=fashion-store-web"
    
    def __init__(self, driver):
        self.driver = driver
        self.cookies = None
    
    def login(self, email, password):
        """
        Automate login process with visible browser.
        
        Args:
            email: User email
            password: User password
            
        Returns:
            bool: True if login successful
        """
        print("🔐 Logging in to Zalando...")
        
        try:
            # Navigate to login page
            self.driver.get(self.LOGIN_URL)
            time.sleep(2)
            
            # Wait for email field and fill it
            email_field = WebDriverWait(self.driver, 10).until(
                EC.presence_of_element_located((By.ID, "login.email"))
            )
            email_field.clear()
            email_field.send_keys(email)
            
            # Wait for password field and fill it
            password_field = WebDriverWait(self.driver, 10).until(
                EC.presence_of_element_located((By.ID, "login.password"))
            )
            password_field.clear()
            password_field.send_keys(password)
            
            # Find and click login button
            login_button = WebDriverWait(self.driver, 10).until(
                EC.element_to_be_clickable((By.CSS_SELECTOR, "button[type='submit']"))
            )
            login_button.click()
            
            # Wait for successful authentication
            # Check for redirect or successful login indicators
            time.sleep(3)
            
            # Store session cookies
            self.save_cookies()
            
            print("✅ Login successful!")
            return True
            
        except Exception as e:
            print(f"❌ Login failed: {str(e)}")
            return False
    
    def save_cookies(self):
        """Save session cookies for API requests."""
        self.cookies = self.driver.get_cookies()
    
    def get_cookies_dict(self):
        """Return cookies as dictionary for requests library."""
        if not self.cookies:
            return {}
        return {cookie['name']: cookie['value'] for cookie in self.cookies}
