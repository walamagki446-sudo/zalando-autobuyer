"""
Selenium browser setup and management for visible Chrome automation.
"""
from selenium import webdriver
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.chrome.service import Service
from webdriver_manager.chrome import ChromeDriverManager
import time


class BrowserManager:
    """Manages Chrome browser instance for automation."""
    
    def __init__(self):
        self.driver = None
    
    def setup_driver(self):
        """
        Initialize Chrome WebDriver in visible mode (headful).
        User can watch the automation process.
        """
        print("🔄 Opening Chrome browser...")
        
        options = Options()
        # NO headless mode - user can see browser
        options.add_argument("--start-maximized")
        options.add_argument("--disable-blink-features=AutomationControlled")
        options.add_experimental_option("excludeSwitches", ["enable-automation"])
        options.add_experimental_option('useAutomationExtension', False)
        
        # Enable performance logging for network monitoring
        options.set_capability('goog:loggingPrefs', {'performance': 'ALL'})
        
        # Initialize driver
        service = Service(ChromeDriverManager().install())
        self.driver = webdriver.Chrome(service=service, options=options)
        
        # Execute CDP commands to evade detection
        self.driver.execute_cdp_cmd('Page.addScriptToEvaluateOnNewDocument', {
            'source': '''
                Object.defineProperty(navigator, 'webdriver', {
                    get: () => undefined
                });
            '''
        })
        
        return self.driver
    
    def close(self):
        """Close the browser."""
        if self.driver:
            self.driver.quit()
    
    def wait_for_user(self):
        """Wait for user input before closing browser."""
        input("\nPress Enter to close browser...")
        self.close()
