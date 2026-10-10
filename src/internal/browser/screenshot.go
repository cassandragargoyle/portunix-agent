/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package browser

import (
	"encoding/base64"
	"fmt"

	"github.com/playwright-community/playwright-go"
)

// ScreenshotResult contains the captured screenshot data
type ScreenshotResult struct {
	Data     string `json:"data"`
	MimeType string `json:"mime_type"`
	Width    int    `json:"width,omitempty"`
	Height   int    `json:"height,omitempty"`
}

// CaptureScreenshot takes a screenshot of the current page.
// Returns base64-encoded PNG data.
func CaptureScreenshot(page playwright.Page, fullPage bool) (*ScreenshotResult, error) {
	// Wait for network to settle before capturing
	_ = WaitForNetworkIdle(page)

	opts := playwright.PageScreenshotOptions{
		FullPage: playwright.Bool(fullPage),
		Type:     playwright.ScreenshotTypePng,
	}

	data, err := page.Screenshot(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to capture screenshot: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(data)

	return &ScreenshotResult{
		Data:     encoded,
		MimeType: "image/png",
	}, nil
}
