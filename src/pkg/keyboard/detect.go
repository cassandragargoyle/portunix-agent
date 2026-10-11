//go:build linux

/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package keyboard

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// detectBackend selects the best available keyboard backend.
// Priority: uinput (works everywhere) > xdotool (X11 only).
func detectBackend() (Backend, error) {
	// uinput works on both Wayland and X11
	if b, err := NewUinputBackend(); err == nil {
		return b, nil
	}

	// xdotool as fallback for X11
	if b, err := NewXdotoolBackend(); err == nil {
		return b, nil
	}

	return nil, fmt.Errorf("no supported keyboard backend found (tried: uinput, xdotool)")
}

func isKWinWayland() bool {
	if os.Getenv("WAYLAND_DISPLAY") != "" {
		return hasProcess("kwin_wayland")
	}
	return hasProcess("kwin_wayland")
}

func hasProcess(name string) bool {
	out, err := exec.Command("pgrep", "-x", name).Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) != ""
}
