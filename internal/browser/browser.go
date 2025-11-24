package browser

import (
	"context"
	"log"
	"time"

	"github.com/chromedp/chromedp"
)

// SetupBrowser creates a new browser context with visible window
func SetupBrowser() (context.Context, context.CancelFunc, error) {
	log.Println("🔄 Setting up browser...")

	// Browser options for visible mode
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false), // VISIBLE browser
		chromedp.Flag("disable-gpu", false),
		chromedp.Flag("disable-dev-shm-usage", false),
		chromedp.Flag("no-sandbox", false),
		chromedp.WindowSize(1920, 1080),
		chromedp.Flag("start-maximized", false),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)

	allocCtx, cancel1 := chromedp.NewExecAllocator(context.Background(), opts...)

	// Create chrome instance with debug logging
	ctx, cancel2 := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))

	// Combine cancel functions
	cancel := func() {
		cancel2()
		cancel1()
	}

	// Start the browser
	if err := chromedp.Run(ctx); err != nil {
		cancel()
		return nil, nil, err
	}

	log.Println("✅ Browser ready!")
	return ctx, cancel, nil
}

// WaitForElement waits for an element to be visible
func WaitForElement(ctx context.Context, selector string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return chromedp.Run(ctx,
		chromedp.WaitVisible(selector, chromedp.ByQuery),
	)
}

// ClickElement clicks on an element
func ClickElement(ctx context.Context, selector string) error {
	return chromedp.Run(ctx,
		chromedp.WaitVisible(selector, chromedp.ByQuery),
		chromedp.Click(selector, chromedp.ByQuery),
	)
}

// GetText retrieves text content from an element
func GetText(ctx context.Context, selector string) (string, error) {
	var text string
	err := chromedp.Run(ctx,
		chromedp.WaitVisible(selector, chromedp.ByQuery),
		chromedp.Text(selector, &text, chromedp.ByQuery),
	)
	return text, err
}

// FillInput fills an input field with text
func FillInput(ctx context.Context, selector, value string) error {
	return chromedp.Run(ctx,
		chromedp.WaitVisible(selector, chromedp.ByQuery),
		chromedp.Clear(selector),
		chromedp.SendKeys(selector, value, chromedp.ByQuery),
	)
}

// Navigate navigates to a URL
func Navigate(ctx context.Context, url string) error {
	return chromedp.Run(ctx,
		chromedp.Navigate(url),
	)
}

// Sleep pauses execution for a duration
func Sleep(ctx context.Context, duration time.Duration) error {
	return chromedp.Run(ctx,
		chromedp.Sleep(duration),
	)
}
