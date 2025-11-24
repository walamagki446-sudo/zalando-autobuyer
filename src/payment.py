"""
Payment method selection: BNPL (Buy Now Pay Later).
Includes automatic session ID detection.
"""
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
import time
import json
import re


class PaymentManager:
    """Handles payment method selection."""
    
    def __init__(self, driver, api_client):
        self.driver = driver
        self.api_client = api_client
        self.session_id = None
    
    def detect_session_id(self):
        """
        Detect payment session ID using multiple methods.
        
        Returns:
            str: Session ID or None
        """
        print("💳 Detecting payment session...")
        
        # Method 1: From performance logs
        session_id = self._detect_from_network()
        if session_id:
            self.session_id = session_id
            print(f"✅ Session ID: {session_id}")
            return session_id
        
        # Method 2: From page context / JavaScript
        session_id = self._detect_from_javascript()
        if session_id:
            self.session_id = session_id
            print(f"✅ Session ID: {session_id}")
            return session_id
        
        # Method 3: From API response
        session_id = self._detect_from_api()
        if session_id:
            self.session_id = session_id
            print(f"✅ Session ID: {session_id}")
            return session_id
        
        print("⚠️  Could not detect session ID")
        return None
    
    def _detect_from_network(self):
        """Detect session ID from network traffic."""
        try:
            logs = self.driver.get_log('performance')
            
            for log in logs:
                message = json.loads(log['message'])
                
                # Look for payment session API calls
                if 'message' in message:
                    msg_data = message['message']
                    
                    if 'params' in msg_data:
                        params = msg_data['params']
                        
                        # Check request
                        if 'request' in params:
                            url = params['request'].get('url', '')
                            if 'purchase-session.client-api.payment.zalando.com' in url:
                                # Extract session ID from URL
                                match = re.search(r'/sessions/([a-f0-9\-]+)', url)
                                if match:
                                    return match.group(1)
                        
                        # Check response
                        if 'response' in params:
                            url = params['response'].get('url', '')
                            if 'purchase-session.client-api.payment.zalando.com' in url:
                                match = re.search(r'/sessions/([a-f0-9\-]+)', url)
                                if match:
                                    return match.group(1)
        except Exception as e:
            print(f"Note: Network detection failed: {str(e)}")
        
        return None
    
    def _detect_from_javascript(self):
        """Detect session ID from page JavaScript context."""
        try:
            # Try multiple JavaScript paths
            scripts = [
                "return window.__INITIAL_STATE__?.payment?.sessionId;",
                "return window.zalandoPaymentSession?.id;",
                "return window.paymentSessionId;",
                "return document.querySelector('[data-payment-session-id]')?.getAttribute('data-payment-session-id');",
            ]
            
            for script in scripts:
                try:
                    result = self.driver.execute_script(script)
                    if result:
                        return result
                except:
                    continue
        except Exception as e:
            print(f"Note: JavaScript detection failed: {str(e)}")
        
        return None
    
    def _detect_from_api(self):
        """Detect session ID from API response."""
        try:
            response = self.api_client.post("checkout/next-step")
            if response.status_code == 200:
                data = response.json()
                return data.get('paymentSessionId') or data.get('sessionId')
        except Exception as e:
            print(f"Note: API detection failed: {str(e)}")
        
        return None
    
    def select_bnpl(self):
        """
        Select BNPL payment method.
        
        Returns:
            bool: True if selection successful
        """
        print("💳 Selecting BNPL payment...")
        
        # Ensure we have session ID
        if not self.session_id:
            self.detect_session_id()
        
        # Method 1: Browser interaction
        success = self._select_bnpl_browser()
        if success:
            print("✅ Payment method selected!")
            return True
        
        # Method 2: API call
        if self.session_id:
            success = self._select_bnpl_api()
            if success:
                print("✅ Payment method selected!")
                return True
        
        print("⚠️  Could not select BNPL payment")
        return False
    
    def _select_bnpl_browser(self):
        """Select BNPL via browser interaction."""
        try:
            # Look for BNPL payment option
            selectors = [
                (By.XPATH, "//button[contains(., 'BNPL')]"),
                (By.XPATH, "//button[contains(., 'Buy Now Pay Later')]"),
                (By.XPATH, "//label[contains(., 'BNPL')]"),
                (By.CSS_SELECTOR, "button[data-payment-method='bnpl']"),
                (By.CSS_SELECTOR, "input[value='bnpl']"),
                (By.XPATH, "//div[contains(@class, 'payment')]//button[contains(., 'Betala senare')]"),
            ]
            
            for by, selector in selectors:
                try:
                    element = WebDriverWait(self.driver, 5).until(
                        EC.element_to_be_clickable((by, selector))
                    )
                    element.click()
                    time.sleep(1)
                    return True
                except:
                    continue
            
        except Exception as e:
            print(f"Note: Browser selection failed: {str(e)}")
        
        return False
    
    def _select_bnpl_api(self):
        """Select BNPL via API."""
        try:
            # OPTIONS request first
            options_url = f"sessions/{self.session_id}/checkout/payment-methods/bnpl"
            self.api_client.options(options_url)
            
            # POST request to update payment method
            response = self.api_client.post(
                "checkout/update-payment",
                json={
                    "paymentMethod": "bnpl",
                    "sessionId": self.session_id
                }
            )
            
            return response.status_code == 200
            
        except Exception as e:
            print(f"Note: API selection failed: {str(e)}")
        
        return False
