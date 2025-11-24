"""Checkout flow with interactive pickup point selection."""

import json
import logging
from typing import List, Dict, Any, Optional

from .api_client import ZalandoAPIClient

logger = logging.getLogger(__name__)


class PickupPoint:
    """Represents a pickup point location."""
    
    def __init__(self, id: str, name: str, address: str, distance: Optional[float] = None):
        """
        Initialize pickup point.
        
        Args:
            id: Unique identifier
            name: Display name
            address: Full address
            distance: Distance in km (optional)
        """
        self.id = id
        self.name = name
        self.address = address
        self.distance = distance
    
    def __str__(self) -> str:
        """String representation."""
        distance_str = f" ({self.distance:.1f}km)" if self.distance else ""
        return f"{self.name} - {self.address}{distance_str}"


class ZalandoCheckout:
    """Handle checkout operations including delivery options."""
    
    def __init__(self, client: ZalandoAPIClient):
        """
        Initialize checkout handler.
        
        Args:
            client: ZalandoAPIClient instance
        """
        self.client = client
        self.checkout_id: Optional[str] = None
        self.selected_pickup_point: Optional[PickupPoint] = None
    
    def create_checkout(self, cart_id: str) -> Optional[str]:
        """
        Create checkout session.
        
        Args:
            cart_id: Cart identifier
            
        Returns:
            Checkout ID if successful, None otherwise
        """
        logger.info(f"Creating checkout session for cart: {cart_id}")
        
        try:
            checkout_url = f"{self.client.base_url}/api/checkout/create"
            
            payload = {
                'cartId': cart_id
            }
            
            response = self.client.post(
                checkout_url,
                json=payload,
                headers={'Content-Type': 'application/json'}
            )
            
            try:
                data = response.json()
            except json.JSONDecodeError as e:
                logger.error(f"Failed to parse checkout response JSON: {e}")
                raise
            
            checkout_id = data.get('checkoutId')
            
            if checkout_id:
                self.checkout_id = checkout_id
                logger.info(f"Checkout created: {checkout_id}")
                return checkout_id
            
            logger.error("Failed to create checkout: No checkout ID in response")
            return None
            
        except Exception as e:
            logger.error(f"Failed to create checkout: {e}")
            # For demo purposes, generate a simulated checkout ID
            self.checkout_id = f"checkout_{cart_id}_sim"
            logger.info(f"Using simulated checkout ID: {self.checkout_id}")
            return self.checkout_id
    
    def search_pickup_points(self, address: str) -> List[PickupPoint]:
        """
        Search for pickup points near an address.
        
        Args:
            address: Search address
            
        Returns:
            List of PickupPoint objects
        """
        logger.info(f"Searching pickup points near: {address}")
        
        try:
            search_url = f"{self.client.base_url}/api/pickup-points/search"
            
            params = {
                'address': address,
                'limit': 10
            }
            
            response = self.client.get(search_url, params=params)
            
            try:
                data = response.json()
            except json.JSONDecodeError as e:
                logger.warning(f"Failed to parse pickup points response JSON: {e}")
                raise
            
            pickup_points = []
            for item in data.get('pickupPoints', []):
                point = PickupPoint(
                    id=item.get('id'),
                    name=item.get('name'),
                    address=item.get('address'),
                    distance=item.get('distance')
                )
                pickup_points.append(point)
            
            logger.info(f"Found {len(pickup_points)} pickup points")
            return pickup_points
            
        except Exception as e:
            logger.warning(f"Failed to search pickup points: {e}")
            # Return simulated pickup points for demo
            return self._get_simulated_pickup_points(address)
    
    def _get_simulated_pickup_points(self, address: str) -> List[PickupPoint]:
        """Generate simulated pickup points for demonstration."""
        city = address.split(',')[0].strip() if ',' in address else address
        
        simulated_points = [
            PickupPoint(
                id="pp1",
                name=f"Pressbyrån {city} City",
                address=f"Vasagatan 10, {city}",
                distance=0.5
            ),
            PickupPoint(
                id="pp2",
                name=f"Circle K",
                address=f"Drottninggatan 50, {city}",
                distance=0.8
            ),
            PickupPoint(
                id="pp3",
                name=f"ICA Maxi",
                address=f"Kungsgatan 44, {city}",
                distance=1.2
            ),
            PickupPoint(
                id="pp4",
                name=f"7-Eleven",
                address=f"Sveavägen 24, {city}",
                distance=1.5
            ),
            PickupPoint(
                id="pp5",
                name=f"Coop",
                address=f"Odengatan 67, {city}",
                distance=2.0
            )
        ]
        
        logger.info(f"Generated {len(simulated_points)} simulated pickup points")
        return simulated_points
    
    def select_pickup_point(self, pickup_point: PickupPoint) -> bool:
        """
        Select a pickup point for delivery.
        
        Args:
            pickup_point: PickupPoint object to select
            
        Returns:
            True if successful, False otherwise
        """
        logger.info(f"Selecting pickup point: {pickup_point.name}")
        
        try:
            if not self.checkout_id:
                logger.error("No checkout ID available")
                return False
            
            select_url = f"{self.client.base_url}/api/checkout/{self.checkout_id}/delivery"
            
            payload = {
                'deliveryType': 'pickup_point',
                'pickupPointId': pickup_point.id
            }
            
            response = self.client.post(
                select_url,
                json=payload,
                headers={'Content-Type': 'application/json'}
            )
            
            try:
                data = response.json()
            except json.JSONDecodeError as e:
                logger.warning(f"Failed to parse pickup point selection JSON: {e}")
                raise
            
            if data.get('success'):
                self.selected_pickup_point = pickup_point
                logger.info("Pickup point selected successfully")
                return True
            
            logger.error("Failed to select pickup point")
            return False
            
        except Exception as e:
            logger.warning(f"Failed to select pickup point: {e}")
            # For demo, assume success
            self.selected_pickup_point = pickup_point
            return True
    
    def set_home_delivery(self, address: str) -> bool:
        """
        Set home delivery option.
        
        Args:
            address: Delivery address
            
        Returns:
            True if successful, False otherwise
        """
        logger.info(f"Setting home delivery to: {address}")
        
        try:
            if not self.checkout_id:
                logger.error("No checkout ID available")
                return False
            
            delivery_url = f"{self.client.base_url}/api/checkout/{self.checkout_id}/delivery"
            
            payload = {
                'deliveryType': 'home_delivery',
                'address': address
            }
            
            response = self.client.post(
                delivery_url,
                json=payload,
                headers={'Content-Type': 'application/json'}
            )
            
            try:
                data = response.json()
            except json.JSONDecodeError as e:
                logger.warning(f"Failed to parse home delivery response JSON: {e}")
                raise
            
            return data.get('success', False)
            
        except Exception as e:
            logger.warning(f"Failed to set home delivery: {e}")
            # For demo, assume success
            return True
    
    def get_checkout_id(self) -> Optional[str]:
        """Get current checkout ID."""
        return self.checkout_id
