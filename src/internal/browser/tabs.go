/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package browser

import (
	"fmt"

	"github.com/playwright-community/playwright-go"
)

// TabInfo describes an open browser tab
type TabInfo struct {
	ID    int    `json:"id"`
	URL   string `json:"url"`
	Title string `json:"title"`
}

// ListTabs returns information about all open tabs
func ListTabs(s *Session) ([]TabInfo, error) {
	ctx := s.Context()
	pages := ctx.Pages()

	tabs := make([]TabInfo, 0, len(pages))
	for i, p := range pages {
		title, _ := p.Title()
		tabs = append(tabs, TabInfo{
			ID:    i,
			URL:   p.URL(),
			Title: title,
		})
	}

	return tabs, nil
}

// SwitchTab switches to the tab with the given ID (0-based index)
func SwitchTab(s *Session, tabID int) error {
	ctx := s.Context()
	pages := ctx.Pages()

	if tabID < 0 || tabID >= len(pages) {
		return fmt.Errorf("tab %d not found (have %d tabs)", tabID, len(pages))
	}

	s.SetActivePage(pages[tabID])
	pages[tabID].BringToFront()
	return nil
}

// CloseTab closes the tab with the given ID (0-based index).
// If closing the active tab, switches to the previous tab.
func CloseTab(s *Session, tabID int) error {
	ctx := s.Context()
	pages := ctx.Pages()

	if tabID < 0 || tabID >= len(pages) {
		return fmt.Errorf("tab %d not found (have %d tabs)", tabID, len(pages))
	}

	if len(pages) <= 1 {
		return fmt.Errorf("cannot close the last tab")
	}

	closingPage := pages[tabID]
	activePage := s.ActivePage()

	if err := closingPage.Close(); err != nil {
		return fmt.Errorf("failed to close tab %d: %w", tabID, err)
	}

	// If we closed the active tab, switch to another
	if closingPage == activePage {
		remaining := ctx.Pages()
		if len(remaining) > 0 {
			newActive := remaining[0]
			if tabID > 0 && tabID-1 < len(remaining) {
				newActive = remaining[tabID-1]
			}
			s.SetActivePage(newActive)
		}
	}

	return nil
}

// NewTab opens a new tab, optionally navigating to a URL
func NewTab(s *Session, url string) (*TabInfo, error) {
	ctx := s.Context()

	var page playwright.Page
	var err error

	page, err = ctx.NewPage()
	if err != nil {
		return nil, fmt.Errorf("failed to open new tab: %w", err)
	}

	if url != "" {
		if err := Navigate(page, url); err != nil {
			return nil, err
		}
	}

	s.SetActivePage(page)

	title, _ := page.Title()
	pages := ctx.Pages()
	return &TabInfo{
		ID:    len(pages) - 1,
		URL:   page.URL(),
		Title: title,
	}, nil
}
