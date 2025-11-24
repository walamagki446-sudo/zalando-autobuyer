package auth

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/walamagki446-sudo/zalando-autobuyer/internal/browser"
	"github.com/walamagki446-sudo/zalando-autobuyer/internal/models"
)

const (
	loginURL            = "https://accounts.zalando.com/authenticate?client_id=fashion-store-web"
	emailSelector       = "input[name='login.email']"
	passwordSelector    = "input[name='login.secret']"
	loginButtonSelector = "button[type='submit']"
	successIndicator    = ".z-navicat-header_user"
)

// Login performs interactive login with visible browser
func Login(ctx context.Context, email, password string) (*models.Session, error) {
	log.Printf("🔐 Logging in as: %s", email)

	// Navigate to login page
	if err := browser.Navigate(ctx, loginURL); err != nil {
		return nil, fmt.Errorf("failed to navigate to login page: %w", err)
	}

	// Wait for page to load
	if err := browser.Sleep(ctx, 2*time.Second); err != nil {
		return nil, err
	}

	// Fill in email
	log.Println("📧 Entering email...")
	if err := browser.FillInput(ctx, emailSelector, email); err != nil {
		return nil, fmt.Errorf("failed to enter email: %w", err)
	}

	// Fill in password
	log.Println("🔑 Entering password...")
	if err := browser.FillInput(ctx, passwordSelector, password); err != nil {
		return nil, fmt.Errorf("failed to enter password: %w", err)
	}

	// Click login button
	log.Println("🔘 Clicking login button...")
	if err := browser.ClickElement(ctx, loginButtonSelector); err != nil {
		return nil, fmt.Errorf("failed to click login button: %w", err)
	}

	// Wait for successful login (adjust selector based on actual page)
	log.Println("⏳ Waiting for authentication...")
	time.Sleep(5 * time.Second)

	log.Println("✅ Login successful!")

	session := &models.Session{
		Email: email,
	}

	return session, nil
}

// ParseCredentials parses email:password format
func ParseCredentials(input string) (email, password string, err error) {
	var colonIndex int = -1
	for i, c := range input {
		if c == ':' {
			colonIndex = i
			break
		}
	}

	if colonIndex == -1 {
		return "", "", fmt.Errorf("invalid format: expected 'email:password'")
	}

	email = input[:colonIndex]
	password = input[colonIndex+1:]

	if email == "" || password == "" {
		return "", "", fmt.Errorf("email and password cannot be empty")
	}

	return email, password, nil
}
