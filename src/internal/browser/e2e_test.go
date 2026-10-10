//go:build e2e

/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

// End-to-end tests for issue #013: persistent profile, browser channel,
// automation fingerprint, and CDP connect. These launch real browsers and
// are excluded from the default test run; execute with:
//
//	go test -tags e2e -v ./internal/browser/ -timeout 180s
//
// Requirements: Playwright Chromium installed, X/Wayland display for the
// headed test, google-chrome for the channel and CDP tests
package browser

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
)

// isolateConfig redirects profile storage to a temp dir so tests never
// touch the real ~/.config/ptx-agent
func isolateConfig(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
}

func closeSession(t *testing.T) {
	if s := GetSession(); s != nil {
		if err := s.Close(); err != nil {
			t.Logf("warning: session close failed: %v", err)
		}
	}
}

// T4: cookie set in a persistent profile survives a relaunch
func TestE2EPersistentProfileSurvivesRelaunch(t *testing.T) {
	isolateConfig(t)
	defer closeSession(t)

	s, err := Launch(LaunchOptions{Headless: true, Profile: "e2e-test"})
	if err != nil {
		t.Fatalf("first launch failed: %v", err)
	}

	expires := float64(time.Now().Add(24 * time.Hour).Unix())
	err = s.Context().AddCookies([]playwright.OptionalCookie{{
		Name:    "ptx_e2e",
		Value:   "issue-013",
		Domain:  playwright.String("example.com"),
		Path:    playwright.String("/"),
		Expires: playwright.Float(expires),
	}})
	if err != nil {
		t.Fatalf("failed to set cookie: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}

	s, err = Launch(LaunchOptions{Headless: true, Profile: "e2e-test"})
	if err != nil {
		t.Fatalf("second launch failed: %v", err)
	}
	cookies, err := s.Context().Cookies("https://example.com")
	if err != nil {
		t.Fatalf("failed to read cookies: %v", err)
	}
	for _, c := range cookies {
		if c.Name == "ptx_e2e" && c.Value == "issue-013" {
			return // cookie survived the relaunch
		}
	}
	t.Errorf("cookie ptx_e2e not found after relaunch, got %d cookies", len(cookies))
}

// T5: navigator.webdriver is false/undefined in headed mode (acceptance
// criterion 3); control case verifies headless mode stays transparent
func TestE2EWebdriverFingerprint(t *testing.T) {
	isolateConfig(t)

	cases := []struct {
		name       string
		headless   bool
		wantMasked bool
	}{
		{"headed-masked", false, true},
		{"headless-transparent", true, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			defer closeSession(t)
			s, err := Launch(LaunchOptions{Headless: tc.headless})
			if err != nil {
				t.Fatalf("launch failed: %v", err)
			}
			val, err := s.ActivePage().Evaluate("() => navigator.webdriver")
			if err != nil {
				t.Fatalf("evaluate failed: %v", err)
			}
			masked := val == nil || val == false
			if masked != tc.wantMasked {
				t.Errorf("navigator.webdriver = %v (masked=%v), want masked=%v", val, masked, tc.wantMasked)
			}
		})
	}
}

// T6: channel option launches the system-installed Chrome
func TestE2EChannelChrome(t *testing.T) {
	if _, err := exec.LookPath("google-chrome"); err != nil {
		t.Skip("google-chrome not installed")
	}
	isolateConfig(t)
	defer closeSession(t)

	s, err := Launch(LaunchOptions{Headless: true, Channel: "chrome"})
	if err != nil {
		t.Fatalf("launch with channel=chrome failed: %v", err)
	}
	ua, err := s.ActivePage().Evaluate("() => navigator.userAgent")
	if err != nil {
		t.Fatalf("evaluate failed: %v", err)
	}
	uaStr, _ := ua.(string)
	if uaStr == "" {
		t.Fatalf("empty user agent")
	}
	t.Logf("channel=chrome user agent: %s", uaStr)
	st := s.GetStatus()
	if st.Channel != "chrome" {
		t.Errorf("status channel = %q, want chrome", st.Channel)
	}
}

// T7: connect to a user-started browser over CDP, drive it, and verify
// disconnecting leaves the browser running
func TestE2ECDPConnect(t *testing.T) {
	if _, err := exec.LookPath("google-chrome"); err != nil {
		t.Skip("google-chrome not installed")
	}
	isolateConfig(t)
	defer closeSession(t)

	const endpoint = "http://localhost:9223"
	// Manage the user-data dir manually: t.TempDir cleanup races with
	// Chrome's own shutdown and fails with "directory not empty"
	userDataDir, err := os.MkdirTemp("", "ptx-cdp-")
	if err != nil {
		t.Fatalf("failed to create user data dir: %v", err)
	}
	cmd := exec.Command("google-chrome", "--headless=new",
		"--remote-debugging-port=9223", "--user-data-dir="+userDataDir,
		"--no-first-run", "about:blank")
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start chrome: %v", err)
	}
	defer func() {
		cmd.Process.Kill()
		cmd.Wait()
		// Best-effort: give Chrome's child processes a moment to
		// release the profile dir before removing it
		time.Sleep(300 * time.Millisecond)
		os.RemoveAll(userDataDir)
	}()

	if err := waitForEndpoint(endpoint, 15*time.Second); err != nil {
		t.Fatalf("CDP endpoint not ready: %v", err)
	}

	s, err2 := Connect(endpoint)
	if err2 != nil {
		t.Fatalf("Connect failed: %v", err2)
	}
	st := s.GetStatus()
	if !st.Connected {
		t.Errorf("status connected = false, want true")
	}
	if st.TabCount < 1 {
		t.Errorf("tab count = %d, want >= 1", st.TabCount)
	}

	// Drive the attached browser with an existing primitive
	if _, err := s.ActivePage().Goto("data:text/html,<title>ptx-cdp</title>"); err != nil {
		t.Fatalf("goto on attached browser failed: %v", err)
	}
	title, _ := s.ActivePage().Title()
	if title != "ptx-cdp" {
		t.Errorf("title = %q, want ptx-cdp", title)
	}

	// Close must only disconnect — the user-owned browser keeps serving CDP
	if err := s.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}
	if err := waitForEndpoint(endpoint, 5*time.Second); err != nil {
		t.Errorf("browser no longer reachable after disconnect: %v", err)
	}
}

// T8: failure injection
func TestE2EFailures(t *testing.T) {
	isolateConfig(t)

	t.Run("invalid-profile-name", func(t *testing.T) {
		defer closeSession(t)
		if _, err := Launch(LaunchOptions{Headless: true, Profile: "../escape"}); err == nil {
			t.Error("launch with traversal profile name succeeded, want error")
		}
	})

	t.Run("double-launch", func(t *testing.T) {
		defer closeSession(t)
		if _, err := Launch(LaunchOptions{Headless: true}); err != nil {
			t.Fatalf("first launch failed: %v", err)
		}
		if _, err := Launch(LaunchOptions{Headless: true}); err == nil {
			t.Error("second launch succeeded, want error")
		}
	})

	t.Run("dead-cdp-endpoint", func(t *testing.T) {
		defer closeSession(t)
		if _, err := Connect("http://localhost:9; 254"); err == nil {
			t.Error("connect to invalid endpoint succeeded, want error")
		}
	})
}

func waitForEndpoint(endpoint string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(endpoint + "/json/version")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	return fmt.Errorf("endpoint %s not reachable within %s", endpoint, timeout)
}
