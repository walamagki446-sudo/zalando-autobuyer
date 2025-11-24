"""
HTTP client for Zalando API calls.
"""
import requests


class ZalandoAPIClient:
    """HTTP client for API interactions."""
    
    def __init__(self, base_url=None):
        self.base_url = base_url or "https://www.zalando.se/api/"
        self.session = requests.Session()
        self.cookies = {}
    
    def set_cookies(self, cookies_dict):
        """
        Set cookies from browser session.
        
        Args:
            cookies_dict: Dictionary of cookies
        """
        self.cookies = cookies_dict
        for name, value in cookies_dict.items():
            self.session.cookies.set(name, value)
    
    def get(self, endpoint, **kwargs):
        """
        Perform GET request.
        
        Args:
            endpoint: API endpoint
            **kwargs: Additional arguments for requests
            
        Returns:
            Response object
        """
        url = self._build_url(endpoint)
        return self.session.get(url, **kwargs)
    
    def post(self, endpoint, **kwargs):
        """
        Perform POST request.
        
        Args:
            endpoint: API endpoint
            **kwargs: Additional arguments for requests
            
        Returns:
            Response object
        """
        url = self._build_url(endpoint)
        return self.session.post(url, **kwargs)
    
    def options(self, endpoint, **kwargs):
        """
        Perform OPTIONS request.
        
        Args:
            endpoint: API endpoint
            **kwargs: Additional arguments for requests
            
        Returns:
            Response object
        """
        url = self._build_url(endpoint)
        return self.session.options(url, **kwargs)
    
    def _build_url(self, endpoint):
        """Build full URL from endpoint."""
        if endpoint.startswith('http'):
            return endpoint
        
        # Remove leading slash if present
        endpoint = endpoint.lstrip('/')
        
        # Handle special endpoints
        if endpoint.startswith('sessions/'):
            return f"https://purchase-session.client-api.payment.zalando.com/{endpoint}"
        
        return f"{self.base_url}{endpoint}"
