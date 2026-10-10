/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package cli

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"

	"github.com/CassandraGargoyle/portunix-agent/internal/browser"
	"github.com/spf13/cobra"
)

var browserHeadless bool
var browserProfile string
var browserChannel string
var browserEndpoint string

var browserCmd = &cobra.Command{
	Use:   "browser",
	Short: "Control a web browser via Playwright",
	Long:  "Launch, navigate, and interact with web pages using Playwright browser automation.",
}

// Lifecycle commands

var browserInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Playwright browser binaries via Portunix",
	Long: `Installs Playwright browser binaries using the Portunix package system.
Requires portunix to be installed and the playwright-browsers package to be available.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Use standard portunix package install for browser binaries
		portunixCmd := exec.Command("portunix", "install", "playwright-browsers")
		portunixCmd.Stdout = cmd.OutOrStdout()
		portunixCmd.Stderr = cmd.ErrOrStderr()
		if err := portunixCmd.Run(); err != nil {
			return fmt.Errorf("failed to install browser binaries via portunix: %w\n\nEnsure portunix is installed and the playwright-browsers package is available", err)
		}
		return nil
	},
}

var browserLaunchCmd = &cobra.Command{
	Use:   "launch",
	Short: "Launch a browser instance",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := browser.Launch(browser.LaunchOptions{
			Headless: browserHeadless,
			Profile:  browserProfile,
			Channel:  browserChannel,
		})
		if err != nil {
			return err
		}
		mode := "headed"
		if browserHeadless {
			mode = "headless"
		}
		fmt.Printf("Browser launched (%s mode)\n", mode)
		if browserProfile != "" {
			fmt.Printf("Persistent profile: %s\n", browserProfile)
		}
		return nil
	},
}

var browserConnectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Attach to a user-started browser via CDP",
	Long: `Attaches to a browser the user started with remote debugging enabled,
e.g. google-chrome --remote-debugging-port=9222. Closing the session only
disconnects; the browser keeps running.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := browser.Connect(browserEndpoint)
		if err != nil {
			return err
		}
		st := s.GetStatus()
		fmt.Printf("Connected to browser via CDP\nURL: %s\nTabs: %d\n", st.URL, st.TabCount)
		return nil
	},
}

var browserCloseCmd = &cobra.Command{
	Use:   "close",
	Short: "Close the browser instance",
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := browser.RequireSession()
		if err != nil {
			return err
		}
		if err := s.Close(); err != nil {
			return err
		}
		fmt.Println("Browser closed")
		return nil
	},
}

var browserStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show running browser state",
	RunE: func(cmd *cobra.Command, args []string) error {
		s := browser.GetSession()
		if s == nil {
			fmt.Println("No browser session active")
			return nil
		}
		st := s.GetStatus()
		data, _ := json.MarshalIndent(st, "", "  ")
		fmt.Println(string(data))
		return nil
	},
}

// Navigation commands

var browserGotoCmd = &cobra.Command{
	Use:   "goto <url>",
	Short: "Navigate to a URL",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := browser.RequireSession()
		if err != nil {
			return err
		}
		if err := browser.Navigate(s.ActivePage(), args[0]); err != nil {
			return err
		}
		title, _ := s.ActivePage().Title()
		fmt.Printf("Navigated to: %s\nTitle: %s\n", args[0], title)
		return nil
	},
}

var browserBackCmd = &cobra.Command{
	Use:   "back",
	Short: "Go back in history",
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := browser.RequireSession()
		if err != nil {
			return err
		}
		if err := browser.GoBack(s.ActivePage()); err != nil {
			return err
		}
		fmt.Printf("Navigated back to: %s\n", s.ActivePage().URL())
		return nil
	},
}

var browserForwardCmd = &cobra.Command{
	Use:   "forward",
	Short: "Go forward in history",
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := browser.RequireSession()
		if err != nil {
			return err
		}
		if err := browser.GoForward(s.ActivePage()); err != nil {
			return err
		}
		fmt.Printf("Navigated forward to: %s\n", s.ActivePage().URL())
		return nil
	},
}

var browserReloadCmd = &cobra.Command{
	Use:   "reload",
	Short: "Reload current page",
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := browser.RequireSession()
		if err != nil {
			return err
		}
		if err := browser.Reload(s.ActivePage()); err != nil {
			return err
		}
		fmt.Println("Page reloaded")
		return nil
	},
}

// Interaction commands

var browserClickCmd = &cobra.Command{
	Use:   "click <index>",
	Short: "Click an element by its index",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		index, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid element index: %s", args[0])
		}
		s, err := browser.RequireSession()
		if err != nil {
			return err
		}
		if err := browser.ClickElement(s.ActivePage(), index); err != nil {
			return err
		}
		fmt.Printf("Clicked element [%d]\n", index)
		return nil
	},
}

var browserTypeCmd = &cobra.Command{
	Use:   "type <index> <text>",
	Short: "Type text into an element",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		index, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid element index: %s", args[0])
		}
		s, err := browser.RequireSession()
		if err != nil {
			return err
		}
		if err := browser.TypeIntoElement(s.ActivePage(), index, args[1], false); err != nil {
			return err
		}
		fmt.Printf("Typed into element [%d]\n", index)
		return nil
	},
}

var browserKeysCmd = &cobra.Command{
	Use:   "keys <keys>",
	Short: "Send key combination (e.g. 'Enter', 'Control+a')",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := browser.RequireSession()
		if err != nil {
			return err
		}
		if err := browser.SendKeys(s.ActivePage(), args[0]); err != nil {
			return err
		}
		fmt.Printf("Sent keys: %s\n", args[0])
		return nil
	},
}

var browserScrollAmount int

var browserScrollCmd = &cobra.Command{
	Use:   "scroll <direction>",
	Short: "Scroll page (up, down, left, right)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := browser.RequireSession()
		if err != nil {
			return err
		}
		if err := browser.ScrollPage(s.ActivePage(), args[0], browserScrollAmount); err != nil {
			return err
		}
		fmt.Printf("Scrolled %s\n", args[0])
		return nil
	},
}

var browserSelectCmd = &cobra.Command{
	Use:   "select <index> <value>",
	Short: "Select a dropdown option",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		index, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid element index: %s", args[0])
		}
		s, err := browser.RequireSession()
		if err != nil {
			return err
		}
		if err := browser.SelectOption(s.ActivePage(), index, args[1]); err != nil {
			return err
		}
		fmt.Printf("Selected option %q on element [%d]\n", args[1], index)
		return nil
	},
}

// State capture commands

var browserElementsViewportOnly bool

var browserElementsCmd = &cobra.Command{
	Use:   "elements",
	Short: "List interactive elements with indices",
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := browser.RequireSession()
		if err != nil {
			return err
		}
		elements, err := browser.DetectElements(s.ActivePage(), browserElementsViewportOnly)
		if err != nil {
			return err
		}
		if len(elements) == 0 {
			fmt.Println("No interactive elements found")
			return nil
		}
		for _, el := range elements {
			fmt.Println(el.String())
		}
		return nil
	},
}

var browserScreenshotFull bool

var browserScreenshotCmd = &cobra.Command{
	Use:   "screenshot",
	Short: "Capture a screenshot",
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := browser.RequireSession()
		if err != nil {
			return err
		}
		result, err := browser.CaptureScreenshot(s.ActivePage(), browserScreenshotFull)
		if err != nil {
			return err
		}
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
		return nil
	},
}

var browserContentSelector string
var browserContentFormat string

var browserContentCmd = &cobra.Command{
	Use:   "content",
	Short: "Extract page text content",
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := browser.RequireSession()
		if err != nil {
			return err
		}
		content, err := browser.ExtractContent(s.ActivePage(), browserContentSelector, browserContentFormat)
		if err != nil {
			return err
		}
		fmt.Println(content)
		return nil
	},
}

// Tab commands

var browserTabsCmd = &cobra.Command{
	Use:   "tabs",
	Short: "List open tabs",
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := browser.RequireSession()
		if err != nil {
			return err
		}
		tabs, err := browser.ListTabs(s)
		if err != nil {
			return err
		}
		for _, tab := range tabs {
			fmt.Printf("[%d] %s - %s\n", tab.ID, tab.Title, tab.URL)
		}
		return nil
	},
}

var browserTabCmd = &cobra.Command{
	Use:   "tab",
	Short: "Tab management (switch, close, new)",
}

var browserTabSwitchCmd = &cobra.Command{
	Use:   "switch <id>",
	Short: "Switch to a tab by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tabID, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid tab ID: %s", args[0])
		}
		s, err := browser.RequireSession()
		if err != nil {
			return err
		}
		if err := browser.SwitchTab(s, tabID); err != nil {
			return err
		}
		fmt.Printf("Switched to tab %d\n", tabID)
		return nil
	},
}

var browserTabCloseCmd = &cobra.Command{
	Use:   "close [id]",
	Short: "Close a tab (current if no ID given)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := browser.RequireSession()
		if err != nil {
			return err
		}

		tabID := -1
		if len(args) > 0 {
			tabID, err = strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid tab ID: %s", args[0])
			}
		} else {
			// Find current tab ID
			tabs, err := browser.ListTabs(s)
			if err != nil {
				return err
			}
			activePage := s.ActivePage()
			for _, t := range tabs {
				pages := s.Context().Pages()
				if t.ID < len(pages) && pages[t.ID] == activePage {
					tabID = t.ID
					break
				}
			}
		}

		if tabID < 0 {
			return fmt.Errorf("could not determine current tab")
		}

		if err := browser.CloseTab(s, tabID); err != nil {
			return err
		}
		fmt.Printf("Closed tab %d\n", tabID)
		return nil
	},
}

var browserTabNewCmd = &cobra.Command{
	Use:   "new [url]",
	Short: "Open a new tab",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := browser.RequireSession()
		if err != nil {
			return err
		}
		url := ""
		if len(args) > 0 {
			url = args[0]
		}
		tab, err := browser.NewTab(s, url)
		if err != nil {
			return err
		}
		fmt.Printf("Opened new tab [%d]: %s\n", tab.ID, tab.URL)
		return nil
	},
}

func init() {
	// Launch flags
	browserLaunchCmd.Flags().BoolVar(&browserHeadless, "headless", true, "run browser in headless mode")
	browserLaunchCmd.Flags().StringVar(&browserProfile, "profile", "", "named persistent profile (cookies and logins survive across launches)")
	browserLaunchCmd.Flags().StringVar(&browserChannel, "channel", "", "browser distribution channel: chrome, chromium, msedge (default: bundled Chromium)")

	// Connect flags
	browserConnectCmd.Flags().StringVar(&browserEndpoint, "endpoint", "http://localhost:9222", "CDP endpoint URL of a user-started browser")

	// Scroll flags
	browserScrollCmd.Flags().IntVar(&browserScrollAmount, "amount", 3, "scroll amount multiplier")

	// Elements flags
	browserElementsCmd.Flags().BoolVar(&browserElementsViewportOnly, "viewport-only", false, "only show elements in viewport")

	// Screenshot flags
	browserScreenshotCmd.Flags().BoolVar(&browserScreenshotFull, "full", false, "capture full page screenshot")

	// Content flags
	browserContentCmd.Flags().StringVar(&browserContentSelector, "selector", "", "CSS selector to extract from")
	browserContentCmd.Flags().StringVar(&browserContentFormat, "format", "text", "output format: text or html")

	// Tab subcommands
	browserTabCmd.AddCommand(browserTabSwitchCmd)
	browserTabCmd.AddCommand(browserTabCloseCmd)
	browserTabCmd.AddCommand(browserTabNewCmd)

	// Register all subcommands
	browserCmd.AddCommand(browserInstallCmd)
	browserCmd.AddCommand(browserLaunchCmd)
	browserCmd.AddCommand(browserConnectCmd)
	browserCmd.AddCommand(browserCloseCmd)
	browserCmd.AddCommand(browserStatusCmd)
	browserCmd.AddCommand(browserGotoCmd)
	browserCmd.AddCommand(browserBackCmd)
	browserCmd.AddCommand(browserForwardCmd)
	browserCmd.AddCommand(browserReloadCmd)
	browserCmd.AddCommand(browserClickCmd)
	browserCmd.AddCommand(browserTypeCmd)
	browserCmd.AddCommand(browserKeysCmd)
	browserCmd.AddCommand(browserScrollCmd)
	browserCmd.AddCommand(browserSelectCmd)
	browserCmd.AddCommand(browserElementsCmd)
	browserCmd.AddCommand(browserScreenshotCmd)
	browserCmd.AddCommand(browserContentCmd)
	browserCmd.AddCommand(browserTabsCmd)
	browserCmd.AddCommand(browserTabCmd)
}
