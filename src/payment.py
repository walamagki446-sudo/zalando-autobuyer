"""Payment processing for Zalando checkout."""

import logging
from typing import Optional

from .api_client import ZalandoAPIClient

logger = logging.getLogger(__name__)


class ZalandoPayment:
    """Handle payment processing."""
    
    def __init__(self, client: ZalandoAPIClient):
        """
        Initialize payment handler.
        
        Args:
            client: ZalandoAPIClient instance
        """
        self.client = client
        self.order_id: Optional[str] = None
    
    def execute_payment(self, checkout_id: str) -> bool:
        """
        Execute payment for checkout.
        
        Args:
            checkout_id: Checkout session ID
            
        Returns:
            True if successful, False otherwise
        """
        logger.info(f"Executing payment for checkout: {checkout_id}")
        
        try:
            payment_url = f"{self.client.base_url}/api/checkout/{checkout_id}/payment"
            
            payload = {
                'paymentMethod': 'default'
            }
            
            response = self.client.post(
                payment_url,
                json=payload,
                headers={'Content-Type': 'application/json'}
            )
            
            data = response.json()
            
            if data.get('success'):
                self.order_id = data.get('orderId')
                logger.info(f"Payment successful. Order ID: {self.order_id}")
                return True
            
            logger.error("Payment failed")
            return False
            
        except Exception as e:
            logger.error(f"Payment execution failed: {e}")
            return False
    
    def complete_order(self, checkout_id: str) -> bool:
        """
        Complete the order after payment.
        
        Args:
            checkout_id: Checkout session ID
            
        Returns:
            True if successful, False otherwise
        """
        logger.info(f"Completing order for checkout: {checkout_id}")
        
        try:
            complete_url = f"{self.client.base_url}/api/checkout/{checkout_id}/complete"
            
            response = self.client.post(
                complete_url,
                headers={'Content-Type': 'application/json'}
            )
            
            data = response.json()
            
            if data.get('success'):
                logger.info("Order completed successfully")
                return True
            
            logger.error("Failed to complete order")
            return False
            
        except Exception as e:
            logger.error(f"Order completion failed: {e}")
            return False
    
    def get_order_id(self) -> Optional[str]:
        """Get order ID if payment was successful."""
        return self.order_id
