//go:build linux

/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/CassandraGargoyle/portunix-agent/internal/keyboard"
	"github.com/CassandraGargoyle/portunix-agent/internal/windows"
	"github.com/spf13/cobra"
)

var typingDelay int
var targetWindow string
var noIndent bool

var keyboardCmd = &cobra.Command{
	Use:   "keyboard",
	Short: "Simulate keyboard input",
	Long:  "Send keystrokes and key combinations to the currently focused application.",
}

var keyboardTypeCmd = &cobra.Command{
	Use:   "type <text>",
	Short: "Type a text string",
	Long: `Send a text string as keyboard input.

If --window is specified, the agent will focus that window first
and then type into it without returning control to the terminal.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Focus target window if requested
		if targetWindow != "" {
			if err := focusTargetWindow(targetWindow); err != nil {
				return err
			}
			// Wait for window to gain focus
			time.Sleep(300 * time.Millisecond)
		}

		ctrl, err := keyboard.NewController()
		if err != nil {
			return err
		}

		text := strings.Join(args, " ")
		if err := ctrl.Type(text, typingDelay, noIndent); err != nil {
			return err
		}

		fmt.Fprintf(cmd.ErrOrStderr(), "Typed %d characters.\n", len(text))
		return nil
	},
}

var keyboardKeyCmd = &cobra.Command{
	Use:   "key <combination>",
	Short: "Send a key combination",
	Long: `Send a key combination to the currently focused window.

Examples:
  ptx-agent keyboard key ctrl+c
  ptx-agent keyboard key alt+Tab
  ptx-agent keyboard key Return`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if targetWindow != "" {
			if err := focusTargetWindow(targetWindow); err != nil {
				return err
			}
			time.Sleep(300 * time.Millisecond)
		}

		ctrl, err := keyboard.NewController()
		if err != nil {
			return err
		}

		if err := ctrl.Key(args[0]); err != nil {
			return err
		}

		fmt.Fprintf(cmd.ErrOrStderr(), "Sent key combination: %s\n", args[0])
		return nil
	},
}

// focusTargetWindow focuses a window by ID or by name search.
func focusTargetWindow(target string) error {
	mgr, err := windows.NewManager()
	if err != nil {
		return fmt.Errorf("cannot focus window: %w", err)
	}

	// If target looks like a UUID, use it directly as window ID
	if strings.HasPrefix(target, "{") {
		return mgr.Focus(target)
	}

	// Otherwise search by name and focus the first match
	wins, err := mgr.Find(target)
	if err != nil {
		return err
	}
	if len(wins) == 0 {
		return fmt.Errorf("no window matching %q found", target)
	}
	return mgr.Focus(wins[0].ID)
}

var keyboardReadCmd = &cobra.Command{
	Use:   "read",
	Short: "Read clipboard contents",
	Long:  "Read and print the current system clipboard contents to stdout.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl, err := keyboard.NewController()
		if err != nil {
			return err
		}

		text, err := ctrl.ReadClipboard()
		if err != nil {
			return err
		}

		fmt.Fprint(cmd.OutOrStdout(), text)
		return nil
	},
}

var keyboardPasteCmd = &cobra.Command{
	Use:   "paste <text>",
	Short: "Paste text via clipboard",
	Long: `Copy text to clipboard and paste it into the focused window via Ctrl+V.

This method works correctly regardless of keyboard layout (Czech, German, etc.)
because it bypasses scancode-to-character mapping entirely.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if targetWindow != "" {
			if err := focusTargetWindow(targetWindow); err != nil {
				return err
			}
			time.Sleep(300 * time.Millisecond)
		}

		ctrl, err := keyboard.NewController()
		if err != nil {
			return err
		}

		text := strings.Join(args, " ")
		if err := ctrl.Paste(text); err != nil {
			return err
		}

		fmt.Fprintf(cmd.ErrOrStderr(), "Pasted %d characters.\n", len(text))
		return nil
	},
}

func init() {
	keyboardCmd.PersistentFlags().StringVar(&targetWindow, "window", "",
		"focus this window before typing (window ID or name search)")
	keyboardTypeCmd.Flags().IntVar(&typingDelay, "delay", keyboard.DefaultTypingDelay,
		"delay between keystrokes in milliseconds")
	keyboardTypeCmd.Flags().BoolVar(&noIndent, "no-indent", false,
		"send Home after each newline to counter editor auto-indent")
	keyboardCmd.AddCommand(keyboardTypeCmd)
	keyboardCmd.AddCommand(keyboardKeyCmd)
	keyboardCmd.AddCommand(keyboardPasteCmd)
	keyboardCmd.AddCommand(keyboardReadCmd)
	rootCmd.AddCommand(keyboardCmd)
}
