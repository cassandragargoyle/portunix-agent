//go:build linux

/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package keyboard

import (
	"os"
	"path/filepath"
	"testing"
)

// TestNormalizeKeyCombination verifies that key combination strings are
// correctly normalized: modifier names are lowercased, common aliases
// (Enter->Return, Esc->Escape, Backspace->BackSpace) are replaced,
// while non-modifier key names keep their original casing
func TestNormalizeKeyCombination(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"lowercase ctrl", "ctrl+c", "ctrl+c"},
		{"uppercase Ctrl", "Ctrl+C", "ctrl+C"},
		{"alt tab", "Alt+Tab", "alt+Tab"},
		{"shift", "Shift+a", "shift+a"},
		{"enter", "Enter", "Return"},
		{"escape", "Esc", "Escape"},
		{"backspace", "Backspace", "BackSpace"},
		{"super", "Super+l", "super+l"},
		{"complex combo", "Ctrl+Shift+Alt+Delete", "ctrl+shift+alt+Delete"},
		{"already normalized", "ctrl+shift+a", "ctrl+shift+a"},
		{"lowercase aliases", "enter", "Return"},
		{"lowercase tab", "alt+tab", "alt+Tab"},
		{"page keys", "PageUp+pagedown", "Prior+Next"},
		{"arrow key", "ctrl+left", "ctrl+Left"},
		{"function key", "alt+f4", "alt+F4"},
		{"symbols", "ctrl+-", "ctrl+minus"},
		{"spaces around parts", "ctrl + s", "ctrl+s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeKeyCombination(tt.input)
			if got != tt.expected {
				t.Errorf("NormalizeKeyCombination(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// TestNormalizeKeyCombination_CoversComboKeyMap verifies that every key name
// accepted by comboToKeys normalizes to a name xdotool knows: a name from
// the xdotool table, or a single letter or digit that xdotool takes as is
func TestNormalizeKeyCombination_CoversComboKeyMap(t *testing.T) {
	for name := range comboKeyMap {
		if _, ok := keyNameMap[name]; ok {
			continue
		}
		if len(name) == 1 && (name[0] >= 'a' && name[0] <= 'z' || name[0] >= '0' && name[0] <= '9') {
			continue
		}
		t.Errorf("key name %q has no xdotool name", name)
	}
}

// TestXdotoolKey_Validation verifies that the xdotool backend rejects a
// combination with an unknown key name before xdotool runs, with the same
// error as the uinput backend, and that valid combinations reach xdotool
func TestXdotoolKey_Validation(t *testing.T) {
	// A fake xdotool that records its arguments, no X server is needed
	dir := t.TempDir()
	log := filepath.Join(dir, "calls")
	fake := filepath.Join(dir, "xdotool")
	script := "#!/bin/sh\necho \"$@\" >> " + log + "\n"
	if err := os.WriteFile(fake, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	b := &XdotoolBackend{path: fake}

	for _, combo := range []string{"ctrl+nosuchkey", "nosuchkey", "ctrl+", "ctrl+Print"} {
		err := b.Key(combo)
		if err == nil {
			t.Errorf("Key(%q): expected error", combo)
			continue
		}
		_, want := comboToKeys(combo)
		if err.Error() != want.Error() {
			t.Errorf("Key(%q) error = %q, want %q", combo, err, want)
		}
	}
	if _, err := os.Stat(log); !os.IsNotExist(err) {
		t.Error("xdotool was called for an invalid combination")
	}

	for _, combo := range []string{"ctrl+s", "alt+Tab", "Return"} {
		if err := b.Key(combo); err != nil {
			t.Errorf("Key(%q): unexpected error: %v", combo, err)
		}
	}
	got, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	want := "key -- ctrl+s\nkey -- alt+Tab\nkey -- Return\n"
	if string(got) != want {
		t.Errorf("xdotool calls = %q, want %q", got, want)
	}
}

// TestNewController_AutoDetect verifies that NewController either returns
// a valid controller with a detected backend or returns an error with nil
// controller when no backend is available on the host system
func TestNewController_AutoDetect(t *testing.T) {
	ctrl, err := NewController()
	if err != nil {
		if ctrl != nil {
			t.Error("expected nil controller when no backend is available")
		}
	} else {
		if ctrl == nil {
			t.Error("expected non-nil controller")
		}
		t.Logf("detected backend: %s", ctrl.BackendName())
	}
}

// TestType_EmptyText verifies that Type rejects empty input with an error
func TestType_EmptyText(t *testing.T) {
	mock := &mockBackend{}
	ctrl := NewControllerWithBackend(mock)
	err := ctrl.Type("", 0, false)
	if err == nil {
		t.Error("expected error for empty text")
	}
}

// TestKey_EmptyCombo verifies that Key rejects empty combo string with an error
func TestKey_EmptyCombo(t *testing.T) {
	mock := &mockBackend{}
	ctrl := NewControllerWithBackend(mock)
	err := ctrl.Key("")
	if err == nil {
		t.Error("expected error for empty combo")
	}
}

// TestController_DelegatesToBackend verifies that Controller correctly
// forwards Type and Key calls to the underlying backend and that the
// backend receives the expected arguments
func TestController_DelegatesToBackend(t *testing.T) {
	mock := &mockBackend{}
	ctrl := NewControllerWithBackend(mock)

	if err := ctrl.Type("hello", 50, false); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if mock.lastTyped != "hello" {
		t.Errorf("expected typed %q, got %q", "hello", mock.lastTyped)
	}

	if err := ctrl.Key("ctrl+c"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if mock.lastKey != "ctrl+c" {
		t.Errorf("expected key %q, got %q", "ctrl+c", mock.lastKey)
	}
}

// mockBackend implements Backend for testing.
type mockBackend struct {
	lastTyped string
	lastDelay int
	lastKey   string
	calls     int
}

func (m *mockBackend) Name() string { return "mock" }

func (m *mockBackend) Type(text string, delayMs int, noIndent bool) error {
	m.lastTyped = text
	m.lastDelay = delayMs
	m.calls++
	return nil
}

func (m *mockBackend) Key(combo string) error {
	m.lastKey = combo
	m.calls++
	return nil
}
