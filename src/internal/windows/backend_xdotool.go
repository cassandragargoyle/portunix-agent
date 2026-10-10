//go:build linux

/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package windows

import (
	"fmt"
	"os/exec"
	"strings"
)

// XdotoolBackend implements Backend using xdotool (X11).
type XdotoolBackend struct {
	path string
}

// NewXdotoolBackend creates a new xdotool backend.
// Returns an error if xdotool is not available.
func NewXdotoolBackend() (*XdotoolBackend, error) {
	path, err := exec.LookPath("xdotool")
	if err != nil {
		return nil, fmt.Errorf("xdotool is not installed or not in PATH: %w", err)
	}
	return &XdotoolBackend{path: path}, nil
}

func (b *XdotoolBackend) Name() string { return "xdotool" }

func (b *XdotoolBackend) List() ([]Window, error) {
	out, err := exec.Command(b.path, "search", "--onlyvisible", "--name", "").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list windows: %w", err)
	}
	return b.parseWindowIDs(strings.TrimSpace(string(out)))
}

func (b *XdotoolBackend) Find(query string) ([]Window, error) {
	out, err := exec.Command(b.path, "search", "--onlyvisible", "--name", query).Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return []Window{}, nil
		}
		return nil, fmt.Errorf("failed to search windows: %w", err)
	}
	return b.parseWindowIDs(strings.TrimSpace(string(out)))
}

func (b *XdotoolBackend) Focus(windowID string) error {
	if err := exec.Command(b.path, "windowactivate", windowID).Run(); err != nil {
		return fmt.Errorf("failed to focus window %s: %w", windowID, err)
	}
	return nil
}

func (b *XdotoolBackend) getWindowName(id string) string {
	out, err := exec.Command(b.path, "getwindowname", id).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func (b *XdotoolBackend) getWindowClass(id string) string {
	out, err := exec.Command("xprop", "-id", id, "WM_CLASS").Output()
	if err != nil {
		return ""
	}
	parts := strings.SplitN(string(out), "=", 2)
	if len(parts) < 2 {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func (b *XdotoolBackend) parseWindowIDs(output string) ([]Window, error) {
	if output == "" {
		return []Window{}, nil
	}

	lines := strings.Split(output, "\n")
	result := make([]Window, 0, len(lines))

	for _, line := range lines {
		id := strings.TrimSpace(line)
		if id == "" {
			continue
		}
		result = append(result, Window{
			ID:    id,
			Name:  b.getWindowName(id),
			Class: b.getWindowClass(id),
		})
	}
	return result, nil
}
