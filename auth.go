package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

const (
	ZalandoBaseURL   = "https://www.zalando.se"
	ZalandoLoginURL  = "https://accounts.zalando.com/api/login/email-password"
	ZalandoAccountsURL = "https://accounts.zalando.com"
)

// LoginRequest represents the login request payload
type LoginRequest struct {
	Email         string      `json:"email"`
	Password      string      `json:"password"`
	Request       AuthRequest `json:"request"`
	RecaptchaData interface{} `json:"recaptcha_data"`
}

// AuthRequest represents the authorization request details
type AuthRequest struct {
	ClientID      string   `json:"client_id"`
	ResponseType  string   `json:"response_type"`
	RedirectURI   string   `json:"redirect_uri"`
	Scope         string   `json:"scope"`
	State         string   `json:"state"`
	Nonce         string   `json:"nonce"`
	UILocales     string   `json:"ui_locales"`
	ResponseMode  string   `json:"response_mode"`
	ACRValues     string   `json:"acr_values"`
	CodeChallenge string   `json:"code_challenge,omitempty"`
	CodeChallengeMethod string `json:"code_challenge_method,omitempty"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	Status     string `json:"status"`
	Action     string `json:"action"`
	TargetURL  string `json:"target_url"`
	ErrorCode  string `json:"error_code,omitempty"`
	ErrorMsg   string `json:"error_message,omitempty"`
}

// AuthSession holds the authentication session data
type AuthSession struct {
	IsLoggedIn  bool
	Email       string
	AccessToken string
	CSRF        string
	SessionID   string
}

// Authenticator handles Zalando authentication
type Authenticator struct {
	client  *HTTPClient
	session *AuthSession
}

// NewAuthenticator creates a new authenticator
func NewAuthenticator(client *HTTPClient) *Authenticator {
	return &Authenticator{
		client:  client,
		session: &AuthSession{},
	}
}

// GetSession returns the current auth session
func (a *Authenticator) GetSession() *AuthSession {
	return a.session
}

// IsLoggedIn returns whether user is logged in
func (a *Authenticator) IsLoggedIn() bool {
	return a.session.IsLoggedIn
}

// Login performs the login flow for Zalando
func (a *Authenticator) Login(email, password string) error {
	// Step 1: Visit Zalando to get initial cookies
	err := a.initSession()
	if err != nil {
		return fmt.Errorf("failed to initialize session: %w", err)
	}

	HumanDelay()

	// Step 2: Get the login page and extract CSRF token
	csrf, state, err := a.getLoginPage()
	if err != nil {
		return fmt.Errorf("failed to get login page: %w", err)
	}

	a.client.SetCSRFToken(csrf)
	HumanDelay()

	// Step 3: Submit login credentials
	err = a.submitLogin(email, password, state)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	a.session.IsLoggedIn = true
	a.session.Email = email

	return nil
}

// initSession visits Zalando homepage to initialize cookies
func (a *Authenticator) initSession() error {
	req, err := http.NewRequest("GET", ZalandoBaseURL, nil)
	if err != nil {
		return err
	}

	resp, err := a.client.Do(req, false)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read body to ensure cookies are set
	io.Copy(io.Discard, resp.Body)

	// Extract CSRF token from cookies
	csrf := a.client.GetCookieValue(ZalandoBaseURL, "frsx")
	if csrf != "" {
		a.client.SetCSRFToken(csrf)
		a.session.CSRF = csrf
	}

	return nil
}

// getLoginPage gets the login page and extracts state/CSRF
func (a *Authenticator) getLoginPage() (csrf string, state string, err error) {
	loginPageURL := fmt.Sprintf("%s/welcome", ZalandoAccountsURL)

	req, err := http.NewRequest("GET", loginPageURL, nil)
	if err != nil {
		return "", "", err
	}

	// Add necessary parameters
	q := req.URL.Query()
	q.Add("client_id", "fashion-store-web")
	q.Add("response_type", "code")
	q.Add("redirect_uri", ZalandoBaseURL+"/resources/oauth/callback")
	q.Add("scope", "openid")
	q.Add("ui_locales", "sv-SE")
	req.URL.RawQuery = q.Encode()

	resp, err := a.client.Do(req, false)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	// Extract state from response
	statePattern := regexp.MustCompile(`"state"\s*:\s*"([^"]+)"`)
	if matches := statePattern.FindSubmatch(body); len(matches) > 1 {
		state = string(matches[1])
	}

	// Extract CSRF token
	csrfPattern := regexp.MustCompile(`"csrf_token"\s*:\s*"([^"]+)"`)
	if matches := csrfPattern.FindSubmatch(body); len(matches) > 1 {
		csrf = string(matches[1])
	}

	// Also check cookies for CSRF
	if csrf == "" {
		csrf = a.client.GetCookieValue(ZalandoAccountsURL, "frsx")
	}

	// Generate random state if not found
	if state == "" {
		state = generateRandomState()
	}

	return csrf, state, nil
}

// submitLogin submits the login form
func (a *Authenticator) submitLogin(email, password, state string) error {
	loginReq := LoginRequest{
		Email:    email,
		Password: password,
		Request: AuthRequest{
			ClientID:     "fashion-store-web",
			ResponseType: "code",
			RedirectURI:  ZalandoBaseURL + "/resources/oauth/callback",
			Scope:        "openid",
			State:        state,
			UILocales:    "sv-SE",
			ResponseMode: "query",
			ACRValues:    "zalando",
		},
	}

	jsonData, err := json.Marshal(loginReq)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", ZalandoLoginURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Origin", ZalandoAccountsURL)
	req.Header.Set("Referer", ZalandoAccountsURL+"/welcome")

	resp, err := a.client.Do(req, true)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var loginResp LoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		// Check if we got a redirect (successful login)
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}
		return fmt.Errorf("failed to parse login response: %w", err)
	}

	// Check for errors
	if loginResp.ErrorCode != "" {
		return fmt.Errorf("login error: %s - %s", loginResp.ErrorCode, loginResp.ErrorMsg)
	}

	// Check for rate limiting
	if strings.Contains(string(body), "rate_limit") || strings.Contains(string(body), "too_many_requests") {
		ErrorDelay()
		return fmt.Errorf("rate limited, please wait and try again")
	}

	// Follow the redirect if there's a target URL
	if loginResp.TargetURL != "" {
		return a.followLoginRedirect(loginResp.TargetURL)
	}

	return nil
}

// followLoginRedirect follows the login redirect to complete authentication
func (a *Authenticator) followLoginRedirect(targetURL string) error {
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return err
	}

	resp, err := a.client.Do(req, false)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	io.Copy(io.Discard, resp.Body)

	// Update session with new cookies
	if csrf := a.client.GetCookieValue(ZalandoBaseURL, "frsx"); csrf != "" {
		a.client.SetCSRFToken(csrf)
		a.session.CSRF = csrf
	}

	return nil
}

// generateRandomState generates a random state parameter
func generateRandomState() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 32)
	for i := range b {
		b[i] = charset[int(RandomIntBetween(0, len(charset)-1))]
	}
	return string(b)
}

// RandomIntBetween returns a random integer between min and max (inclusive)
func RandomIntBetween(min, max int) int {
	if min >= max {
		return min
	}
	return min + rand.Intn(max-min+1)
}

// CheckZalandoLogin checks if the provided credentials are valid
// This is the function mentioned in the requirements
func CheckZalandoLogin(client *HTTPClient, email, password string) (bool, error) {
	auth := NewAuthenticator(client)
	err := auth.Login(email, password)
	if err != nil {
		return false, err
	}
	return auth.IsLoggedIn(), nil
}

// VerifySession checks if the current session is still valid
func (a *Authenticator) VerifySession() bool {
	req, err := http.NewRequest("GET", ZalandoBaseURL+"/myaccount", nil)
	if err != nil {
		return false
	}

	resp, err := a.client.Do(req, false)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// If we get redirected to login, session is invalid
	finalURL := resp.Request.URL.String()
	if strings.Contains(finalURL, "accounts.zalando.com") || strings.Contains(finalURL, "login") {
		a.session.IsLoggedIn = false
		return false
	}

	return true
}

// Logout logs out of the Zalando session
func (a *Authenticator) Logout() error {
	logoutURL := ZalandoBaseURL + "/logout"

	req, err := http.NewRequest("GET", logoutURL, nil)
	if err != nil {
		return err
	}

	resp, err := a.client.Do(req, false)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	io.Copy(io.Discard, resp.Body)

	a.session.IsLoggedIn = false
	a.session.Email = ""
	a.session.AccessToken = ""

	return nil
}

// ExtractCSRFFromURL extracts CSRF token from URL cookies
func (a *Authenticator) ExtractCSRFFromURL(urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}

	for _, cookie := range a.client.GetCookies(u) {
		if cookie.Name == "frsx" || cookie.Name == "csrf" || cookie.Name == "xsrf" {
			return cookie.Value
		}
	}

	return ""
}
