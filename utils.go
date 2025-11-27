package main

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync"
	"time"

	"golang.org/x/net/publicsuffix"
)

// UserAgents is a list of realistic user agents for rotation
var UserAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:120.0) Gecko/20100101 Firefox/120.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36 Edg/129.0.0.0",
}

// HTTPClient wraps http.Client with additional functionality
type HTTPClient struct {
	client    *http.Client
	jar       *cookiejar.Jar
	userAgent string
	csrfToken string
	mutex     sync.RWMutex
	proxyURL  *url.URL
}

// NewHTTPClient creates a new HTTP client with cookie jar and optional proxy
func NewHTTPClient(proxyStr string) (*HTTPClient, error) {
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar: %w", err)
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}

	// Parse and set proxy if provided
	var proxyURL *url.URL
	if proxyStr != "" {
		proxyURL, err = parseProxy(proxyStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse proxy: %w", err)
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	client := &http.Client{
		Transport: transport,
		Jar:       jar,
		Timeout:   30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	return &HTTPClient{
		client:    client,
		jar:       jar,
		userAgent: GetRandomUserAgent(),
		proxyURL:  proxyURL,
	}, nil
}

// parseProxy parses proxy string in format host:port:username:password
func parseProxy(proxyStr string) (*url.URL, error) {
	// Expected format: host:port:username:password
	// Example: budget.waveproxies.com:1337:user:pass
	var host, port, username, password string

	// Try to parse as host:port:user:pass format
	parts := splitProxyString(proxyStr)
	if len(parts) == 4 {
		host = parts[0]
		port = parts[1]
		username = parts[2]
		password = parts[3]
	} else if len(parts) == 2 {
		host = parts[0]
		port = parts[1]
	} else {
		return nil, fmt.Errorf("invalid proxy format, expected host:port or host:port:user:pass")
	}

	proxyURLStr := fmt.Sprintf("http://%s:%s", host, port)
	proxyURL, err := url.Parse(proxyURLStr)
	if err != nil {
		return nil, err
	}

	if username != "" && password != "" {
		proxyURL.User = url.UserPassword(username, password)
	}

	return proxyURL, nil
}

// splitProxyString splits proxy string handling the complex password
func splitProxyString(s string) []string {
	// Split on first 3 colons for host:port:user, rest is password
	result := make([]string, 0, 4)
	start := 0
	colonCount := 0

	for i, char := range s {
		if char == ':' {
			colonCount++
			if colonCount <= 3 {
				result = append(result, s[start:i])
				start = i + 1
			}
		}
	}
	// Add the remaining part (could be password with colons)
	if start < len(s) {
		result = append(result, s[start:])
	}

	return result
}

// GetRandomUserAgent returns a random user agent from the list
func GetRandomUserAgent() string {
	return UserAgents[rand.Intn(len(UserAgents))]
}

// RotateUserAgent changes the user agent
func (c *HTTPClient) RotateUserAgent() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.userAgent = GetRandomUserAgent()
}

// SetCSRFToken sets the CSRF token
func (c *HTTPClient) SetCSRFToken(token string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.csrfToken = token
}

// GetCSRFToken returns the current CSRF token
func (c *HTTPClient) GetCSRFToken() string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.csrfToken
}

// SetDefaultHeaders sets default headers for Zalando requests
func (c *HTTPClient) SetDefaultHeaders(req *http.Request, isJSON bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept-Language", "sv-SE,sv;q=0.9,en;q=0.8")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Sec-Ch-Ua", `"Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"`)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)
	req.Header.Set("x-zalando-header-mode", "desktop")
	req.Header.Set("x-zalando-footer-mode", "desktop")
	req.Header.Set("x-zalando-checkout-app", "web")

	if isJSON {
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Sec-Fetch-Dest", "empty")
		req.Header.Set("Sec-Fetch-Mode", "cors")
		req.Header.Set("Sec-Fetch-Site", "same-origin")
	} else {
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
		req.Header.Set("Sec-Fetch-Dest", "document")
		req.Header.Set("Sec-Fetch-Mode", "navigate")
		req.Header.Set("Sec-Fetch-Site", "none")
	}

	if c.csrfToken != "" {
		req.Header.Set("x-xsrf-token", c.csrfToken)
	}
}

// Do executes the HTTP request with automatic header setting
func (c *HTTPClient) Do(req *http.Request, isJSON bool) (*http.Response, error) {
	c.SetDefaultHeaders(req, isJSON)
	return c.client.Do(req)
}

// GetCookies returns cookies for a URL
func (c *HTTPClient) GetCookies(u *url.URL) []*http.Cookie {
	return c.jar.Cookies(u)
}

// GetCookieValue gets a specific cookie value
func (c *HTTPClient) GetCookieValue(urlStr, name string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}
	for _, cookie := range c.jar.Cookies(u) {
		if cookie.Name == name {
			return cookie.Value
		}
	}
	return ""
}

// RandomDelay adds a random delay between min and max milliseconds
func RandomDelay(minMs, maxMs int) {
	delay := minMs + rand.Intn(maxMs-minMs+1)
	time.Sleep(time.Duration(delay) * time.Millisecond)
}

// HumanDelay adds a human-like delay between requests
func HumanDelay() {
	RandomDelay(100, 800)
}

// ErrorDelay adds a longer delay after errors
func ErrorDelay() {
	RandomDelay(1500, 2500)
}

// Base64Encode encodes string to base64
func Base64Encode(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

// Base64Decode decodes base64 string
func Base64Decode(s string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
