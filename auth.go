package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/publicsuffix"
)

// HTTPClient manages HTTP sessions with cookies
type HTTPClient struct {
	Client    *http.Client
	UserAgent string
	CSRFToken string
}

// NewHTTPClient creates a new HTTP client with cookie jar
func NewHTTPClient() (*HTTPClient, error) {
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar: %v", err)
	}

	client := &http.Client{
		Jar:     jar,
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Allow redirects but preserve cookies
			return nil
		},
	}

	return &HTTPClient{
		Client:    client,
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
	}, nil
}

// extractCSRFToken extracts CSRF token from HTML page
func (hc *HTTPClient) extractCSRFToken(htmlContent string) (string, error) {
	// Try multiple patterns for CSRF token extraction
	patterns := []string{
		`<meta name="csrf-token" content="([^"]+)"`,
		`"csrfToken":"([^"]+)"`,
		`csrf_token":\s*"([^"]+)"`,
		`data-csrf="([^"]+)"`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(htmlContent)
		if len(matches) > 1 {
			return matches[1], nil
		}
	}

	return "", fmt.Errorf("CSRF token not found in page")
}

// GetCSRFToken fetches CSRF token from authenticate page
func (hc *HTTPClient) GetCSRFToken() error {
	LogInfo("Hämtar CSRF-token...")

	req, err := http.NewRequest("GET", "https://accounts.zalando.com/authenticate", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("User-Agent", hc.UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Language", "sv-SE,sv;q=0.9,en;q=0.8")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")

	resp, err := hc.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch authenticate page: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	token, err := hc.extractCSRFToken(string(body))
	if err != nil {
		return err
	}

	hc.CSRFToken = token
	LogSuccess("CSRF-token hämtad: %s", token[:20]+"...")
	return nil
}

// LoginResponse represents the login API response
type LoginResponse struct {
	Status       string `json:"status"`
	RedirectURL  string `json:"redirect_url"`
	ErrorMessage string `json:"error_message"`
}

// Login authenticates user with Zalando
func (hc *HTTPClient) Login(email, password string) error {
	// First get CSRF token
	if err := hc.GetCSRFToken(); err != nil {
		return fmt.Errorf("kunde inte hämta CSRF-token: %v", err)
	}

	LogInfo("Loggar in som %s...", email)

	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		// Prepare login payload
		loginData := map[string]interface{}{
			"email":    email,
			"password": password,
			"remember": true,
		}

		jsonData, err := json.Marshal(loginData)
		if err != nil {
			return fmt.Errorf("failed to marshal login data: %v", err)
		}

		req, err := http.NewRequest("POST", "https://accounts.zalando.com/api/sso/authentications/credentials", bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("failed to create login request: %v", err)
		}

		// Set required headers
		req.Header.Set("User-Agent", hc.UserAgent)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json, text/plain, */*")
		req.Header.Set("Accept-Language", "sv-SE,sv;q=0.9,en;q=0.8")
		req.Header.Set("Origin", "https://accounts.zalando.com")
		req.Header.Set("Referer", "https://accounts.zalando.com/authenticate")
		req.Header.Set("Sec-Fetch-Dest", "empty")
		req.Header.Set("Sec-Fetch-Mode", "cors")
		req.Header.Set("Sec-Fetch-Site", "same-origin")
		req.Header.Set("x-csrf-token", hc.CSRFToken)

		resp, err := hc.Client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to send login request: %v", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		// Handle different status codes
		switch resp.StatusCode {
		case 200, 204, 302, 303:
			LogSuccess("Inloggning lyckades!")
			return nil

		case 401:
			// Parse response for error details
			var loginResp LoginResponse
			if err := json.Unmarshal(body, &loginResp); err == nil && strings.Contains(loginResp.ErrorMessage, "elevated-risk") {
				LogWarning("Förhöjd risk upptäckt, väntar %d sekunder innan nytt försök...", attempt*5)
				time.Sleep(time.Duration(attempt*5) * time.Second)
				continue
			}
			return fmt.Errorf("ogiltiga inloggningsuppgifter (401)")

		case 429:
			waitTime := attempt * 10
			LogWarning("Rate limit nådd (429), väntar %d sekunder...", waitTime)
			time.Sleep(time.Duration(waitTime) * time.Second)
			continue

		case 403:
			return fmt.Errorf("åtkomst nekad (403) - kontot kan vara blockerat")

		default:
			return fmt.Errorf("oväntat svar från server: %d - %s", resp.StatusCode, string(body))
		}
	}

	return fmt.Errorf("inloggning misslyckades efter %d försök", maxRetries)
}

// IsAuthenticated checks if the user is authenticated
func (hc *HTTPClient) IsAuthenticated() bool {
	req, err := http.NewRequest("GET", "https://www.zalando.se/api/cart-gateway/carts", nil)
	if err != nil {
		return false
	}

	req.Header.Set("User-Agent", hc.UserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := hc.Client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200
}

// GetCookies returns current cookies for debugging
func (hc *HTTPClient) GetCookies(urlStr string) []*http.Cookie {
	u, _ := url.Parse(urlStr)
	return hc.Client.Jar.Cookies(u)
}
