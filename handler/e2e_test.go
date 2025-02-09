package handler

import (
	"context"
	_ "log"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

func TestLoginE2E(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := chromedp.DefaultExecAllocatorOptions[:]
	opts = append(opts, chromedp.NoDefaultBrowserCheck, chromedp.Flag("headless", false))
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(ctx, opts...)
	defer cancelAlloc()

	taskCtx, cancelTask := chromedp.NewContext(allocCtx)
	defer cancelTask()

	serverURL := "http://localhost:3000/login"

	err := chromedp.Run(taskCtx,
		chromedp.Navigate(serverURL),
		chromedp.WaitVisible(`form`, chromedp.ByQuery),
		chromedp.SendKeys(`input[name="Email"]`, "galimjantugelbaev@gmail.com", chromedp.ByQuery),
		chromedp.SendKeys(`input[name="Password"]`, "gakon2006", chromedp.ByQuery),
		chromedp.Click(`button[type="submit"]`, chromedp.ByQuery),
		chromedp.Sleep(4*time.Second),
	)

	if err != nil {
		t.Fatalf("Test failed: %v", err)
	}

}
