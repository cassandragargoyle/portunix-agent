//go:build linux

/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package windows

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// detectBackend selects the best available window management backend.
// Priority: KWin (Wayland) > xdotool (X11).
func detectBackend() (Backend, error) {
	if isKWinWayland() {
		b, err := NewKWinBackend()
		if err == nil {
			return b, nil
		}
	}

	if isX11() {
		b, err := NewXdotoolBackend()
		if err == nil {
			return b, nil
		}
	}

	return nil, fmt.Errorf("no supported window backend found (tried: kwin/wayland, xdotool/x11)")
}

// isKWinWayland checks if KWin Wayland compositor is running.
func isKWinWayland() bool {
	if os.Getenv("WAYLAND_DISPLAY") != "" {
		return hasProcess("kwin_wayland")
	}
	// Fallback: check for kwin_wayland process even without env var
	// (env may not be inherited in non-interactive shells)
	return hasProcess("kwin_wayland")
}

// isX11 checks if an X11 display server is available.
func isX11() bool {
	return os.Getenv("DISPLAY") != "" || hasProcess("Xorg") || hasProcess("Xwayland")
}

// hasProcess checks if a process with the given name is running.
func hasProcess(name string) bool {
	out, err := exec.Command("pgrep", "-x", name).Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) != ""
}
