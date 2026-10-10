/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package browser

import (
	"fmt"
	"time"

	"github.com/playwright-community/playwright-go"
)

// Navigate goes to a URL and waits for the page to load
func Navigate(page playwright.Page, url string) error {
	opts := playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
		Timeout:   playwright.Float(30000),
	}
	_, err := page.Goto(url, opts)
	if err != nil {
		return fmt.Errorf("failed to navigate to %s: %w", url, err)
	}
	return nil
}

// GoBack navigates back in browser history
func GoBack(page playwright.Page) error {
	opts := playwright.PageGoBackOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
		Timeout:   playwright.Float(10000),
	}
	_, err := page.GoBack(opts)
	if err != nil {
		return fmt.Errorf("failed to go back: %w", err)
	}
	return nil
}

// GoForward navigates forward in browser history
func GoForward(page playwright.Page) error {
	opts := playwright.PageGoForwardOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
		Timeout:   playwright.Float(10000),
	}
	_, err := page.GoForward(opts)
	if err != nil {
		return fmt.Errorf("failed to go forward: %w", err)
	}
	return nil
}

// Reload reloads the current page
func Reload(page playwright.Page) error {
	opts := playwright.PageReloadOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
		Timeout:   playwright.Float(30000),
	}
	_, err := page.Reload(opts)
	if err != nil {
		return fmt.Errorf("failed to reload: %w", err)
	}
	return nil
}

// ClickElement clicks an element identified by its index
func ClickElement(page playwright.Page, index int) error {
	elements, err := DetectElements(page, false)
	if err != nil {
		return err
	}

	el, err := FindElementByIndex(elements, index)
	if err != nil {
		return err
	}

	loc := page.Locator(el.Selector)
	if err := loc.First().Click(); err != nil {
		return fmt.Errorf("failed to click element [%d]: %w", index, err)
	}

	// Brief wait for any navigation or DOM updates
	time.Sleep(200 * time.Millisecond)
	return nil
}

// TypeIntoElement types text into an element identified by its index
func TypeIntoElement(page playwright.Page, index int, text string, clear bool) error {
	elements, err := DetectElements(page, false)
	if err != nil {
		return err
	}

	el, err := FindElementByIndex(elements, index)
	if err != nil {
		return err
	}

	loc := page.Locator(el.Selector).First()

	if clear {
		if err := loc.Fill(""); err != nil {
			return fmt.Errorf("failed to clear element [%d]: %w", index, err)
		}
	}

	if err := loc.Fill(text); err != nil {
		return fmt.Errorf("failed to type into element [%d]: %w", index, err)
	}

	return nil
}

// SendKeys sends a key combination to the page
func SendKeys(page playwright.Page, keys string) error {
	if err := page.Keyboard().Press(keys); err != nil {
		return fmt.Errorf("failed to send keys %q: %w", keys, err)
	}
	return nil
}

// ScrollPage scrolls the page in the specified direction
func ScrollPage(page playwright.Page, direction string, amount int) error {
	if amount <= 0 {
		amount = 3
	}

	var deltaX, deltaY int
	scrollPixels := amount * 100

	switch direction {
	case "up":
		deltaY = -scrollPixels
	case "down":
		deltaY = scrollPixels
	case "left":
		deltaX = -scrollPixels
	case "right":
		deltaX = scrollPixels
	default:
		return fmt.Errorf("invalid scroll direction %q (use: up, down, left, right)", direction)
	}

	js := fmt.Sprintf("window.scrollBy(%d, %d)", deltaX, deltaY)
	_, err := page.Evaluate(js)
	if err != nil {
		return fmt.Errorf("failed to scroll: %w", err)
	}

	time.Sleep(200 * time.Millisecond)
	return nil
}

// SelectOption selects a dropdown option by value on the element at the given index
func SelectOption(page playwright.Page, index int, value string) error {
	elements, err := DetectElements(page, false)
	if err != nil {
		return err
	}

	el, err := FindElementByIndex(elements, index)
	if err != nil {
		return err
	}

	loc := page.Locator(el.Selector).First()
	_, err = loc.SelectOption(playwright.SelectOptionValues{Values: playwright.StringSlice(value)})
	if err != nil {
		return fmt.Errorf("failed to select option on element [%d]: %w", index, err)
	}

	return nil
}

// EvaluateJS executes JavaScript on the page and returns the result as a string
func EvaluateJS(page playwright.Page, expression string) (interface{}, error) {
	result, err := page.Evaluate(expression)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate JS: %w", err)
	}
	return result, nil
}

// UploadFile sets a file input element to the given file path
func UploadFile(page playwright.Page, index int, path string) error {
	elements, err := DetectElements(page, false)
	if err != nil {
		return err
	}

	el, err := FindElementByIndex(elements, index)
	if err != nil {
		return err
	}

	loc := page.Locator(el.Selector).First()
	if err := loc.SetInputFiles(path); err != nil {
		return fmt.Errorf("failed to upload file to element [%d]: %w", index, err)
	}

	return nil
}

// WaitForNetworkIdle waits for network requests to settle
func WaitForNetworkIdle(page playwright.Page) error {
	return page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	})
}
