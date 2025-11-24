"""
Product page automation: size detection and selection.
"""
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
import time


class ProductManager:
    """Handles product page interactions."""
    
    def __init__(self, driver):
        self.driver = driver
    
    def navigate_to_product(self, product_url):
        """
        Navigate to product page.
        
        Args:
            product_url: Full URL of the product
            
        Returns:
            bool: True if navigation successful
        """
        print("🔄 Navigating to product...")
        
        try:
            self.driver.get(product_url)
            time.sleep(2)
            
            # Verify product page loaded
            WebDriverWait(self.driver, 10).until(
                EC.presence_of_element_located((By.TAG_NAME, "body"))
            )
            
            print("✅ Product page loaded!")
            return True
            
        except Exception as e:
            print(f"❌ Failed to load product page: {str(e)}")
            return False
    
    def detect_sizes(self):
        """
        Detect available sizes using multiple methods.
        
        Returns:
            list: List of available sizes
        """
        print("📦 Detecting available sizes...")
        
        sizes = []
        
        # Method 1: Try to click size picker and get options
        try:
            # Find picker trigger
            picker_selectors = [
                (By.ID, "picker-trigger"),
                (By.CSS_SELECTOR, "button[data-testid='pdp-size-picker-trigger']"),
                (By.CSS_SELECTOR, "button[aria-label*='size']"),
                (By.CSS_SELECTOR, "button[aria-label*='Size']")
            ]
            
            picker = None
            for by, selector in picker_selectors:
                try:
                    picker = self.driver.find_element(by, selector)
                    break
                except:
                    continue
            
            if picker:
                picker.click()
                time.sleep(1)
                
                # Try multiple methods to find size options
                
                # Method 1a: Via role='option'
                try:
                    options = self.driver.find_elements(By.CSS_SELECTOR, "[role='option']")
                    sizes = [opt.text.strip() for opt in options if opt.text.strip()]
                    if sizes:
                        # Filter out non-size text
                        sizes = [s for s in sizes if s not in ["Välj storlek", "Select size", ""]]
                except:
                    pass
                
                # Method 1b: Via aria-controls
                if not sizes:
                    try:
                        dropdown_id = picker.get_attribute("aria-controls")
                        if dropdown_id:
                            options = self.driver.find_elements(
                                By.CSS_SELECTOR, 
                                f"#{dropdown_id} button, #{dropdown_id} li"
                            )
                            sizes = [opt.text.strip() for opt in options if opt.text.strip()]
                            sizes = [s for s in sizes if s not in ["Välj storlek", "Select size", ""]]
                    except:
                        pass
                
                # Method 1c: Look for listbox
                if not sizes:
                    try:
                        options = self.driver.find_elements(
                            By.CSS_SELECTOR, 
                            "[role='listbox'] [role='option']"
                        )
                        sizes = [opt.text.strip() for opt in options if opt.text.strip()]
                        sizes = [s for s in sizes if s not in ["Välj storlek", "Select size", ""]]
                    except:
                        pass
                
        except Exception as e:
            print(f"Note: Picker click method failed: {str(e)}")
        
        # Method 2: Look for visible size buttons
        if not sizes:
            try:
                size_patterns = ["XS", "S", "M", "L", "XL", "XXL", "XXXL"] + \
                               [str(i) for i in range(34, 52)]
                
                visible_buttons = self.driver.find_elements(By.CSS_SELECTOR, "button[type='button']")
                for btn in visible_buttons:
                    text = btn.text.strip()
                    if text in size_patterns or (text.replace('.', '').replace(',', '').isdigit()):
                        if text not in sizes:
                            sizes.append(text)
            except:
                pass
        
        # Method 3: Look for size-specific attributes
        if not sizes:
            try:
                size_elements = self.driver.find_elements(
                    By.CSS_SELECTOR, 
                    "button[data-size], button[aria-label*='size']"
                )
                for elem in size_elements:
                    text = elem.text.strip() or elem.get_attribute("data-size")
                    if text and text not in sizes:
                        sizes.append(text)
            except:
                pass
        
        if sizes:
            print(f"✅ Found sizes: {', '.join(sizes)}")
        else:
            print("⚠️  Could not detect sizes automatically")
            sizes = ["Unable to detect - manual entry"]
        
        return sizes
    
    def select_size(self, size):
        """
        Select a specific size.
        
        Args:
            size: Size to select (e.g., "L", "42")
            
        Returns:
            bool: True if size selected successfully
        """
        print(f"🔄 Selecting size {size}...")
        
        try:
            # Multiple strategies to find and click the size
            
            # Strategy 1: Find by text in role='option'
            try:
                options = self.driver.find_elements(By.CSS_SELECTOR, "[role='option']")
                for opt in options:
                    if opt.text.strip() == size:
                        opt.click()
                        time.sleep(1)
                        print(f"✅ Size {size} selected!")
                        return True
            except:
                pass
            
            # Strategy 2: Find by button text
            try:
                buttons = self.driver.find_elements(By.CSS_SELECTOR, "button[type='button']")
                for btn in buttons:
                    if btn.text.strip() == size:
                        btn.click()
                        time.sleep(1)
                        print(f"✅ Size {size} selected!")
                        return True
            except:
                pass
            
            # Strategy 3: Find by aria-label
            try:
                button = self.driver.find_element(
                    By.CSS_SELECTOR, 
                    f"button[aria-label*='{size}']"
                )
                button.click()
                time.sleep(1)
                print(f"✅ Size {size} selected!")
                return True
            except:
                pass
            
            # Strategy 4: Find by data-size attribute
            try:
                button = self.driver.find_element(
                    By.CSS_SELECTOR, 
                    f"button[data-size='{size}']"
                )
                button.click()
                time.sleep(1)
                print(f"✅ Size {size} selected!")
                return True
            except:
                pass
            
            print(f"❌ Could not select size {size}")
            return False
            
        except Exception as e:
            print(f"❌ Error selecting size: {str(e)}")
            return False
