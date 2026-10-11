//go:build linux

/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package keyboard

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// XdotoolBackend implements Backend using xdotool (X11).
type XdotoolBackend struct {
	path string
}

// NewXdotoolBackend creates a new xdotool keyboard backend.
func NewXdotoolBackend() (*XdotoolBackend, error) {
	path, err := exec.LookPath("xdotool")
	if err != nil {
		return nil, fmt.Errorf("xdotool is not installed or not in PATH: %w", err)
	}
	return &XdotoolBackend{path: path}, nil
}

func (b *XdotoolBackend) Name() string { return "xdotool" }

func (b *XdotoolBackend) Type(text string, delayMs int, noIndent bool) error {
	delay := strconv.Itoa(delayMs)
	if err := exec.Command(b.path, "type", "--delay", delay, "--", text).Run(); err != nil {
		return fmt.Errorf("failed to type text: %w", err)
	}
	return nil
}

func (b *XdotoolBackend) Key(combo string) error {
	normalized := NormalizeKeyCombination(combo)
	if err := exec.Command(b.path, "key", "--", normalized).Run(); err != nil {
		return fmt.Errorf("failed to send key combination %q: %w", combo, err)
	}
	return nil
}

// NormalizeKeyCombination converts common key notation to xdotool format.
func NormalizeKeyCombination(combo string) string {
	parts := strings.Split(combo, "+")
	for i, part := range parts {
		parts[i] = normalizeKeyName(part)
	}
	return strings.Join(parts, "+")
}

var keyNameMap = map[string]string{
	"Ctrl":      "ctrl",
	"Alt":       "alt",
	"Shift":     "shift",
	"Super":     "super",
	"Enter":     "Return",
	"Esc":       "Escape",
	"Del":       "Delete",
	"Backspace": "BackSpace",
}

func normalizeKeyName(key string) string {
	if mapped, ok := keyNameMap[key]; ok {
		return mapped
	}
	return key
}
