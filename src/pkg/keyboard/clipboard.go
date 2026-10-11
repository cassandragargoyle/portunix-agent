//go:build linux

/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package keyboard

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// ReadClipboard returns the current system clipboard contents
func (c *Controller) ReadClipboard() (string, error) {
	return getClipboard()
}

// Paste sets the clipboard content and sends Ctrl+V to paste it
// into the currently focused window. Works on Wayland and X11
// regardless of keyboard layout.
func (c *Controller) Paste(text string) error {
	if text == "" {
		return fmt.Errorf("text cannot be empty")
	}

	if err := setClipboard(text); err != nil {
		return fmt.Errorf("failed to set clipboard: %w", err)
	}

	// Brief pause for clipboard to propagate
	time.Sleep(50 * time.Millisecond)

	return c.Key("ctrl+v")
}

// setClipboard writes text to the system clipboard.
// Tries Klipper DBus (KDE), wl-copy (Wayland), xclip (X11) in order.
func setClipboard(text string) error {
	// Try Klipper DBus (KDE — works on Wayland without extra tools)
	// gdbus interprets GVariant string escapes, so backslashes must be doubled
	escaped := strings.ReplaceAll(text, "\\", "\\\\")
	if err := exec.Command("gdbus", "call", "--session",
		"--dest", "org.kde.klipper",
		"--object-path", "/klipper",
		"--method", "org.kde.klipper.klipper.setClipboardContents",
		escaped).Run(); err == nil {
		return nil
	}

	// Try wl-copy (generic Wayland)
	if path, err := exec.LookPath("wl-copy"); err == nil {
		cmd := exec.Command(path)
		cmd.Stdin = strings.NewReader(text)
		if err := cmd.Run(); err == nil {
			return nil
		}
	}

	// Try xclip (X11)
	if path, err := exec.LookPath("xclip"); err == nil {
		cmd := exec.Command(path, "-selection", "clipboard")
		cmd.Stdin = strings.NewReader(text)
		if err := cmd.Run(); err == nil {
			return nil
		}
	}

	return fmt.Errorf("no clipboard tool available (tried: klipper dbus, wl-copy, xclip)")
}

// getClipboard reads text from the system clipboard.
// Tries Klipper DBus (KDE), wl-paste (Wayland), xclip (X11) in order.
func getClipboard() (string, error) {
	// Try Klipper DBus (KDE — works on Wayland without extra tools)
	if out, err := exec.Command("gdbus", "call", "--session",
		"--dest", "org.kde.klipper",
		"--object-path", "/klipper",
		"--method", "org.kde.klipper.klipper.getClipboardContents").Output(); err == nil {
		return stripDBusTupleWrapper(string(out)), nil
	}

	// Try wl-paste (generic Wayland)
	if path, err := exec.LookPath("wl-paste"); err == nil {
		if out, err := exec.Command(path, "--no-newline").Output(); err == nil {
			return string(out), nil
		}
	}

	// Try xclip (X11)
	if path, err := exec.LookPath("xclip"); err == nil {
		if out, err := exec.Command(path, "-selection", "clipboard", "-o").Output(); err == nil {
			return string(out), nil
		}
	}

	return "", fmt.Errorf("no clipboard tool available (tried: klipper dbus, wl-paste, xclip)")
}

// stripDBusTupleWrapper removes the tuple wrapping that gdbus returns.
// gdbus may use single quotes ('...',) or double quotes ("...",)
// depending on content. Both use the same GVariant escape sequences.
func stripDBusTupleWrapper(s string) string {
	s = strings.TrimSpace(s)
	switch {
	case strings.HasPrefix(s, "(\"") && strings.HasSuffix(s, "\",)"):
		s = s[2 : len(s)-3]
	case strings.HasPrefix(s, "('") && strings.HasSuffix(s, "',)"):
		s = s[2 : len(s)-3]
	default:
		return s
	}
	// Unescape GVariant string escapes
	s = strings.ReplaceAll(s, "\\\\", "\x00")
	s = strings.ReplaceAll(s, "\\'", "'")
	s = strings.ReplaceAll(s, "\\\"", "\"")
	s = strings.ReplaceAll(s, "\\n", "\n")
	s = strings.ReplaceAll(s, "\\t", "\t")
	s = strings.ReplaceAll(s, "\x00", "\\")
	return s
}
