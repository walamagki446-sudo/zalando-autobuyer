"""Base API client with retry logic for Zalando API interactions."""

import requests
import time
import logging
from typing import Optional, Dict, Any

logger = logging.getLogger(__name__)


class ZalandoAPIClient:
    """Base API client for Zalando with retry logic and session management."""
    
    def __init__(self, region: str = "se", session: Optional[requests.Session] = None):
        """
        Initialize Zalando API client.
        
        Args:
            region: Zalando region code (se, de, uk, fr, nl, es, it)
            session: Optional requests.Session object
        """
        self.region = region.lower()
        self.base_url = f"https://www.zalando.{self.region}"
        self.session = session or requests.Session()
        
        # Set default headers
        self.session.headers.update({
            'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
            'Accept': 'application/json',
            'Accept-Language': 'en-US,en;q=0.9',
        })
    
    def request(
        self,
        method: str,
        url: str,
        max_retries: int = 3,
        retry_delay: float = 1.0,
        **kwargs
    ) -> requests.Response:
        """
        Make HTTP request with retry logic.
        
        Args:
            method: HTTP method (GET, POST, etc.)
            url: URL to request
            max_retries: Maximum number of retries
            retry_delay: Delay between retries in seconds
            **kwargs: Additional arguments to pass to requests
            
        Returns:
            Response object
            
        Raises:
            requests.RequestException: If all retries fail
        """
        for attempt in range(max_retries):
            try:
                response = self.session.request(method, url, **kwargs)
                response.raise_for_status()
                return response
            except requests.RequestException as e:
                if attempt == max_retries - 1:
                    logger.error(f"Request failed after {max_retries} attempts: {e}")
                    raise
                logger.warning(f"Request attempt {attempt + 1} failed: {e}. Retrying...")
                time.sleep(retry_delay * (attempt + 1))
        
        raise requests.RequestException("Request failed after all retries")
    
    def get(self, url: str, **kwargs) -> requests.Response:
        """Make GET request."""
        return self.request('GET', url, **kwargs)
    
    def post(self, url: str, **kwargs) -> requests.Response:
        """Make POST request."""
        return self.request('POST', url, **kwargs)
    
    def put(self, url: str, **kwargs) -> requests.Response:
        """Make PUT request."""
        return self.request('PUT', url, **kwargs)
    
    def delete(self, url: str, **kwargs) -> requests.Response:
        """Make DELETE request."""
        return self.request('DELETE', url, **kwargs)
