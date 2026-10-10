/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

// Package browser provides browser control capabilities using Playwright.
// It supports headless (background) and headed (interactive) operation modes.
package browser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/playwright-community/playwright-go"
)

// Session manages a single browser instance lifecycle
type Session struct {
	pw        *playwright.Playwright
	browser   playwright.Browser
	context   playwright.BrowserContext
	page      playwright.Page
	mu        sync.Mutex
	headless  bool
	profile   string
	channel   string
	connected bool
}

// LaunchOptions configures browser launch
type LaunchOptions struct {
	Headless bool
	// Profile names a persistent profile stored under the ptx-agent config
	// directory; cookies and logins survive across launches. Empty means a
	// fresh non-persistent context
	Profile string
	// Channel selects a browser distribution channel (chrome, chromium,
	// msedge, ...) to use a system-installed browser instead of the
	// Playwright-bundled build. Empty means the bundled Chromium
	Channel string
}

// Status represents the current browser session state
type Status struct {
	Running   bool   `json:"running"`
	Headless  bool   `json:"headless"`
	Connected bool   `json:"connected,omitempty"`
	Profile   string `json:"profile,omitempty"`
	Channel   string `json:"channel,omitempty"`
	URL       string `json:"url,omitempty"`
	Title     string `json:"title,omitempty"`
	TabCount  int    `json:"tab_count"`
}

// singleton session — one browser per ptx-agent process
var (
	globalSession *Session
	sessionMu     sync.Mutex
)

// GetSession returns the current browser session, or nil if none is active
func GetSession() *Session {
	sessionMu.Lock()
	defer sessionMu.Unlock()
	return globalSession
}

// ProfileDir resolves the user-data directory for a named persistent profile.
// Profiles live under the ptx-agent config directory, e.g.
// ~/.config/ptx-agent/browser-profiles/<name> on Linux
func ProfileDir(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("profile name must not be empty")
	}
	if name != filepath.Base(name) || strings.HasPrefix(name, ".") || strings.ContainsAny(name, `/\`) {
		return "", fmt.Errorf("invalid profile name %q: must be a plain directory name", name)
	}
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to resolve user config directory: %w", err)
	}
	return filepath.Join(cfgDir, "ptx-agent", "browser-profiles", name), nil
}

// fingerprintArgs returns browser flags that reduce the automation fingerprint.
// Applied in headed collaborative mode only: drops the --enable-automation
// default flag and disables the Blink AutomationControlled feature so
// navigator.webdriver is not set
func fingerprintArgs(headless bool) (args []string, ignoreDefaultArgs []string) {
	if headless {
		return nil, nil
	}
	return []string{"--disable-blink-features=AutomationControlled"},
		[]string{"--enable-automation"}
}

// Launch starts a new browser instance
func Launch(opts LaunchOptions) (*Session, error) {
	sessionMu.Lock()
	defer sessionMu.Unlock()

	if globalSession != nil {
		return nil, fmt.Errorf("browser already running, close it first")
	}

	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to start playwright: %w", err)
	}

	args, ignoreDefaultArgs := fingerprintArgs(opts.Headless)
	var channel *string
	if opts.Channel != "" {
		channel = playwright.String(opts.Channel)
	}

	s := &Session{
		pw:       pw,
		headless: opts.Headless,
		profile:  opts.Profile,
		channel:  opts.Channel,
	}

	if opts.Profile != "" {
		// Persistent profile: cookies, logins, and anti-bot clearance
		// tokens survive across launches
		dir, err := ProfileDir(opts.Profile)
		if err != nil {
			pw.Stop()
			return nil, err
		}
		if err := os.MkdirAll(dir, 0o700); err != nil {
			pw.Stop()
			return nil, fmt.Errorf("failed to create profile directory: %w", err)
		}
		ctx, err := pw.Chromium.LaunchPersistentContext(dir, playwright.BrowserTypeLaunchPersistentContextOptions{
			Headless:          playwright.Bool(opts.Headless),
			Channel:           channel,
			Args:              args,
			IgnoreDefaultArgs: ignoreDefaultArgs,
		})
		if err != nil {
			pw.Stop()
			return nil, fmt.Errorf("failed to launch persistent context: %w", err)
		}
		// Persistent context starts with an initial blank page
		pages := ctx.Pages()
		var page playwright.Page
		if len(pages) > 0 {
			page = pages[0]
		} else {
			page, err = ctx.NewPage()
			if err != nil {
				ctx.Close()
				pw.Stop()
				return nil, fmt.Errorf("failed to create page: %w", err)
			}
		}
		s.context = ctx
		s.page = page
	} else {
		browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
			Headless:          playwright.Bool(opts.Headless),
			Channel:           channel,
			Args:              args,
			IgnoreDefaultArgs: ignoreDefaultArgs,
		})
		if err != nil {
			pw.Stop()
			return nil, fmt.Errorf("failed to launch browser: %w", err)
		}

		ctx, err := browser.NewContext()
		if err != nil {
			browser.Close()
			pw.Stop()
			return nil, fmt.Errorf("failed to create browser context: %w", err)
		}

		page, err := ctx.NewPage()
		if err != nil {
			ctx.Close()
			browser.Close()
			pw.Stop()
			return nil, fmt.Errorf("failed to create page: %w", err)
		}
		s.browser = browser
		s.context = ctx
		s.page = page
	}

	globalSession = s
	return s, nil
}

// Connect attaches to a user-started browser over the Chrome DevTools
// Protocol (e.g. chrome --remote-debugging-port=9222). The browser remains
// owned by the user; Close only disconnects, it does not terminate it
func Connect(endpoint string) (*Session, error) {
	sessionMu.Lock()
	defer sessionMu.Unlock()

	if globalSession != nil {
		return nil, fmt.Errorf("browser already running, close it first")
	}

	if endpoint == "" {
		endpoint = "http://localhost:9222"
	}

	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to start playwright: %w", err)
	}

	b, err := pw.Chromium.ConnectOverCDP(endpoint)
	if err != nil {
		pw.Stop()
		return nil, fmt.Errorf("failed to connect over CDP to %s: %w", endpoint, err)
	}

	// Reuse the browser's existing default context and page when present
	var ctx playwright.BrowserContext
	if contexts := b.Contexts(); len(contexts) > 0 {
		ctx = contexts[0]
	} else {
		ctx, err = b.NewContext()
		if err != nil {
			b.Close()
			pw.Stop()
			return nil, fmt.Errorf("failed to create browser context: %w", err)
		}
	}

	var page playwright.Page
	if pages := ctx.Pages(); len(pages) > 0 {
		page = pages[0]
	} else {
		page, err = ctx.NewPage()
		if err != nil {
			b.Close()
			pw.Stop()
			return nil, fmt.Errorf("failed to create page: %w", err)
		}
	}

	s := &Session{
		pw:        pw,
		browser:   b,
		context:   ctx,
		page:      page,
		headless:  false,
		connected: true,
	}
	globalSession = s
	return s, nil
}

// Close shuts down the browser and cleans up resources. For a CDP-connected
// session this only disconnects; the user-started browser keeps running
func (s *Session) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sessionMu.Lock()
	defer sessionMu.Unlock()

	var firstErr error

	if s.browser != nil {
		// For a connected browser, Close disconnects without
		// terminating the user-owned process
		if err := s.browser.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	} else if s.context != nil {
		// Persistent context launch has no separate browser handle
		if err := s.context.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	if s.pw != nil {
		if err := s.pw.Stop(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	if globalSession == s {
		globalSession = nil
	}

	return firstErr
}

// GetStatus returns the current session status
func (s *Session) GetStatus() Status {
	s.mu.Lock()
	defer s.mu.Unlock()

	st := Status{
		Running:   true,
		Headless:  s.headless,
		Connected: s.connected,
		Profile:   s.profile,
		Channel:   s.channel,
	}

	pages := s.context.Pages()
	st.TabCount = len(pages)

	if s.page != nil {
		st.URL = s.page.URL()
		st.Title, _ = s.page.Title()
	}

	return st
}

// ActivePage returns the currently active page
func (s *Session) ActivePage() playwright.Page {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.page
}

// SetActivePage sets the currently active page
func (s *Session) SetActivePage(p playwright.Page) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.page = p
}

// Context returns the browser context
func (s *Session) Context() playwright.BrowserContext {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.context
}

// RequireSession returns the current session or an error if none is active
func RequireSession() (*Session, error) {
	s := GetSession()
	if s == nil {
		return nil, fmt.Errorf("no browser session active, launch one first")
	}
	return s, nil
}
