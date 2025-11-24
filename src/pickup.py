"""
Smart pickup point selection logic.
Priority: Instabox > Budbee > Closest > Farthest
"""
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
import time


class PickupPointManager:
    """Handles pickup point selection with smart prioritization."""
    
    def __init__(self, driver, api_client):
        self.driver = driver
        self.api_client = api_client
    
    def select_best_pickup_point(self, address=None):
        """
        Select best pickup point using smart logic:
        1. Priority 1: Instabox locations
        2. Priority 2: Budbee locations
        3. Priority 3: Closest to address
        4. Priority 4: Farthest available
        
        Args:
            address: Optional delivery address
            
        Returns:
            dict: Selected pickup point information
        """
        print("📍 Searching for pickup points...")
        
        # Try to get pickup points via API
        pickup_points = self._fetch_pickup_points_api(address)
        
        if not pickup_points:
            # Fallback: Get from browser
            pickup_points = self._fetch_pickup_points_browser()
        
        if not pickup_points:
            print("⚠️  No pickup points found")
            return None
        
        print(f"\nFound {len(pickup_points)} pickup points:")
        
        # Apply smart selection logic
        selected = self._apply_selection_logic(pickup_points)
        
        if selected:
            # Select the pickup point
            self._select_pickup_point(selected)
            print(f"\n✅ Selected: {selected.get('name', 'Unknown location')}")
            return selected
        
        return None
    
    def _fetch_pickup_points_api(self, address):
        """Fetch pickup points via API."""
        try:
            if address:
                response = self.api_client.post(
                    "checkout/search-pickup-points-by-address",
                    json={"address": address}
                )
            else:
                response = self.api_client.get("checkout/pickup-points")
            
            if response.status_code == 200:
                data = response.json()
                return data.get("pickupPoints", [])
        except Exception as e:
            print(f"Note: API fetch failed: {str(e)}")
        
        return []
    
    def _fetch_pickup_points_browser(self):
        """Fetch pickup points from browser DOM."""
        pickup_points = []
        
        try:
            # Look for pickup point elements
            selectors = [
                "[data-testid='pickup-point']",
                ".pickup-point",
                "[role='option']",
                "button[aria-label*='pickup']",
            ]
            
            for selector in selectors:
                try:
                    elements = self.driver.find_elements(By.CSS_SELECTOR, selector)
                    for elem in elements:
                        name = elem.text.strip() or elem.get_attribute("aria-label")
                        if name:
                            pickup_points.append({
                                'id': elem.get_attribute('data-id') or str(len(pickup_points)),
                                'name': name,
                                'element': elem
                            })
                    
                    if pickup_points:
                        break
                except:
                    continue
        except:
            pass
        
        return pickup_points
    
    def _apply_selection_logic(self, pickup_points):
        """
        Apply smart selection logic to pickup points.
        
        Returns:
            dict: Selected pickup point
        """
        # Priority 1: Instabox
        for point in pickup_points:
            name = point.get('name', '').lower()
            if 'instabox' in name:
                print(f"  ✓ Priority 1: {point['name']} (Instabox)")
                return point
        
        # Priority 2: Budbee
        for point in pickup_points:
            name = point.get('name', '').lower()
            if 'budbee' in name:
                print(f"  ✓ Priority 2: {point['name']} (Budbee)")
                return point
        
        # Priority 3: Closest (sort by distance)
        sorted_by_distance = sorted(
            pickup_points, 
            key=lambda p: p.get('distance', 999)
        )
        
        if sorted_by_distance:
            closest = sorted_by_distance[0]
            distance = closest.get('distance', 'Unknown')
            print(f"  ✓ Priority 3: {closest['name']} (Closest - {distance})")
            return closest
        
        # Priority 4: Farthest (if no close ones work)
        if pickup_points:
            sorted_by_distance_desc = sorted(
                pickup_points, 
                key=lambda p: p.get('distance', 0), 
                reverse=True
            )
            farthest = sorted_by_distance_desc[0]
            distance = farthest.get('distance', 'Unknown')
            print(f"  ✓ Priority 4: {farthest['name']} (Farthest - {distance})")
            return farthest
        
        return None
    
    def _select_pickup_point(self, point):
        """
        Select a pickup point in the browser or via API.
        
        Args:
            point: Pickup point dictionary
        """
        try:
            # Method 1: Click element if available
            if 'element' in point:
                point['element'].click()
                time.sleep(1)
                return True
            
            # Method 2: Use API
            point_id = point.get('id')
            if point_id:
                response = self.api_client.post(
                    "checkout/select-pickup-point",
                    json={"pickupPointId": point_id}
                )
                if response.status_code == 200:
                    time.sleep(1)
                    return True
            
        except Exception as e:
            print(f"Note: Selection error: {str(e)}")
        
        return False
    
    def _print_pickup_points(self, pickup_points):
        """Print pickup points for user visibility."""
        for i, point in enumerate(pickup_points, 1):
            name = point.get('name', 'Unknown')
            distance = point.get('distance', 'N/A')
            print(f"  {i}. {name} (Distance: {distance})")
