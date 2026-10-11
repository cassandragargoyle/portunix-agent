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
	// xdotool ignores an unknown key name and still exits with 0,
	// so the combination is validated before anything is sent
	if _, err := comboToKeys(combo); err != nil {
		return err
	}
	normalized := NormalizeKeyCombination(combo)
	if err := exec.Command(b.path, "key", "--", normalized).Run(); err != nil {
		return fmt.Errorf("failed to send key combination %q: %w", combo, err)
	}
	return nil
}

// NormalizeKeyCombination converts common key notation to xdotool format.
// Key names are matched without regard to case; a name that is not in the
// table (a letter, a digit) keeps its original casing.
func NormalizeKeyCombination(combo string) string {
	parts := strings.Split(combo, "+")
	for i, part := range parts {
		parts[i] = normalizeKeyName(part)
	}
	return strings.Join(parts, "+")
}

// keyNameMap maps the lowercase key names accepted by comboToKeys to the
// names that xdotool knows
var keyNameMap = map[string]string{
	"ctrl":      "ctrl",
	"alt":       "alt",
	"shift":     "shift",
	"super":     "super",
	"return":    "Return",
	"enter":     "Return",
	"escape":    "Escape",
	"esc":       "Escape",
	"tab":       "Tab",
	"backspace": "BackSpace",
	"delete":    "Delete",
	"del":       "Delete",
	"insert":    "Insert",
	"home":      "Home",
	"end":       "End",
	"pageup":    "Prior",
	"pagedown":  "Next",
	"up":        "Up",
	"down":      "Down",
	"left":      "Left",
	"right":     "Right",
	"space":     "space",
	"f1":        "F1",
	"f2":        "F2",
	"f3":        "F3",
	"f4":        "F4",
	"f5":        "F5",
	"f6":        "F6",
	"f7":        "F7",
	"f8":        "F8",
	"f9":        "F9",
	"f10":       "F10",
	"f11":       "F11",
	"f12":       "F12",
	"-":         "minus",
	"=":         "equal",
}

func normalizeKeyName(key string) string {
	key = strings.TrimSpace(key)
	if mapped, ok := keyNameMap[strings.ToLower(key)]; ok {
		return mapped
	}
	return key
}
