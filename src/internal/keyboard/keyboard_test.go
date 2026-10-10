//go:build linux

/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package keyboard

import "testing"

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
	lastKey   string
}

func (m *mockBackend) Name() string { return "mock" }

func (m *mockBackend) Type(text string, delayMs int, noIndent bool) error {
	m.lastTyped = text
	return nil
}

func (m *mockBackend) Key(combo string) error {
	m.lastKey = combo
	return nil
}
