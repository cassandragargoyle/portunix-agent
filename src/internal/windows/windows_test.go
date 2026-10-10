//go:build linux

/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package windows

import "testing"

func TestNewManager_AutoDetect(t *testing.T) {
	mgr, err := NewManager()
	if err != nil {
		// No backend available on this machine
		if mgr != nil {
			t.Error("expected nil manager when no backend is available")
		}
	} else {
		if mgr == nil {
			t.Error("expected non-nil manager")
		}
		t.Logf("detected backend: %s", mgr.BackendName())
	}
}

func TestFocus_EmptyID(t *testing.T) {
	mock := &mockBackend{}
	mgr := NewManagerWithBackend(mock)
	err := mgr.Focus("")
	if err == nil {
		t.Error("expected error for empty window id")
	}
}

func TestFind_EmptyQuery(t *testing.T) {
	mock := &mockBackend{}
	mgr := NewManagerWithBackend(mock)
	_, err := mgr.Find("")
	if err == nil {
		t.Error("expected error for empty query")
	}
}

func TestManager_DelegatesToBackend(t *testing.T) {
	mock := &mockBackend{
		windows: []Window{
			{ID: "1", Name: "Test Window", Class: "test"},
		},
	}
	mgr := NewManagerWithBackend(mock)

	wins, err := mgr.List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(wins) != 1 {
		t.Fatalf("expected 1 window, got %d", len(wins))
	}
	if wins[0].Name != "Test Window" {
		t.Errorf("expected name %q, got %q", "Test Window", wins[0].Name)
	}
}

func TestParseKWinOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"empty", "", 0},
		{"no js prefix", "some random line\nanother line", 0},
		{"single window", "js: {abc-123}\tMy Window\tmy-app", 1},
		{"multiple windows", "js: {a}\tWin1\tapp1\njs: {b}\tWin2\tapp2", 2},
		{"mixed lines", "some log\njs: {a}\tWin1\tapp1\nmore log", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseKWinOutput(tt.input)
			if len(got) != tt.expected {
				t.Errorf("parseKWinOutput() returned %d windows, want %d", len(got), tt.expected)
			}
		})
	}
}

func TestExtractScriptID(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"(0,)", "0"},
		{"(1,)", "1"},
		{"(42,)", "42"},
	}

	for _, tt := range tests {
		got := extractScriptID(tt.input)
		if got != tt.expected {
			t.Errorf("extractScriptID(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

// mockBackend implements Backend for testing.
type mockBackend struct {
	windows []Window
}

func (m *mockBackend) Name() string { return "mock" }

func (m *mockBackend) List() ([]Window, error) {
	return m.windows, nil
}

func (m *mockBackend) Find(query string) ([]Window, error) {
	return m.windows, nil
}

func (m *mockBackend) Focus(windowID string) error {
	return nil
}
