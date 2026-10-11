//go:build linux

/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package keyboard

import "testing"

// TestSession_DelegatesToBackend verifies that a session on a backend
// without session support forwards Type and Key to the backend and
// applies the default typing delay
func TestSession_DelegatesToBackend(t *testing.T) {
	mock := &mockBackend{}
	s, err := NewControllerWithBackend(mock).Open()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer s.Close()

	if err := s.Type("hello", 0, false); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if mock.lastTyped != "hello" {
		t.Errorf("expected typed %q, got %q", "hello", mock.lastTyped)
	}
	if mock.lastDelay != DefaultTypingDelay {
		t.Errorf("expected delay %d, got %d", DefaultTypingDelay, mock.lastDelay)
	}

	if err := s.Key("ctrl+c"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if mock.lastKey != "ctrl+c" {
		t.Errorf("expected key %q, got %q", "ctrl+c", mock.lastKey)
	}
}

// TestSession_Validation verifies that a session rejects empty text and
// an empty key combination without calling the backend
func TestSession_Validation(t *testing.T) {
	mock := &mockBackend{}
	s, err := NewControllerWithBackend(mock).Open()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer s.Close()

	if err := s.Type("", 0, false); err == nil {
		t.Error("expected error for empty text")
	}
	if err := s.Key(""); err == nil {
		t.Error("expected error for empty combo")
	}
	if mock.calls != 0 {
		t.Errorf("expected no backend calls, got %d", mock.calls)
	}
}

// TestSession_UseAfterClose verifies that Type and Key return an error
// on a closed session and do not reach the backend
func TestSession_UseAfterClose(t *testing.T) {
	mock := &mockBackend{}
	s, err := NewControllerWithBackend(mock).Open()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := s.Type("hello", 50, false); err == nil {
		t.Error("expected error for Type on a closed session")
	}
	if err := s.Key("ctrl+c"); err == nil {
		t.Error("expected error for Key on a closed session")
	}
	if mock.calls != 0 {
		t.Errorf("expected no backend calls, got %d", mock.calls)
	}
}

// TestSession_DoubleClose verifies that a second Close is safe and that
// the device is released only once
func TestSession_DoubleClose(t *testing.T) {
	mock := &mockSessionBackend{}
	s, err := NewControllerWithBackend(mock).Open()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := s.Close(); err != nil {
		t.Errorf("unexpected error on first Close: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Errorf("unexpected error on second Close: %v", err)
	}
	if mock.closed != 1 {
		t.Errorf("expected 1 device close, got %d", mock.closed)
	}
}

// TestSession_OneDeviceForManyActions verifies that a session on a backend
// with session support opens exactly one device for many actions
func TestSession_OneDeviceForManyActions(t *testing.T) {
	mock := &mockSessionBackend{}
	s, err := NewControllerWithBackend(mock).Open()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i := 0; i < 5; i++ {
		if err := s.Type("a", 50, false); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if err := s.Key("Return"); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	}
	if err := s.Close(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if mock.opened != 1 || mock.closed != 1 {
		t.Errorf("expected 1 open and 1 close, got %d and %d", mock.opened, mock.closed)
	}
	if mock.calls != 10 {
		t.Errorf("expected 10 actions, got %d", mock.calls)
	}
}

// TestController_OpensDevicePerCall verifies that Controller.Type and
// Controller.Key stay one-shot wrappers: each call opens and closes a device
func TestController_OpensDevicePerCall(t *testing.T) {
	mock := &mockSessionBackend{}
	ctrl := NewControllerWithBackend(mock)

	if err := ctrl.Type("hello", 50, false); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err := ctrl.Key("ctrl+c"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if mock.opened != 2 || mock.closed != 2 {
		t.Errorf("expected 2 opens and 2 closes, got %d and %d", mock.opened, mock.closed)
	}

	// Invalid input must not open a device
	if err := ctrl.Type("", 50, false); err == nil {
		t.Error("expected error for empty text")
	}
	if mock.opened != 2 {
		t.Errorf("expected no device for invalid input, got %d opens", mock.opened)
	}
}

// mockSessionBackend implements SessionBackend and counts device sessions.
type mockSessionBackend struct {
	mockBackend
	opened int
	closed int
}

func (m *mockSessionBackend) OpenSession() (Session, error) {
	m.opened++
	return &mockDeviceSession{backend: m}, nil
}

// mockDeviceSession is the device session of mockSessionBackend.
type mockDeviceSession struct {
	backend *mockSessionBackend
}

func (s *mockDeviceSession) Type(text string, delayMs int, noIndent bool) error {
	return s.backend.Type(text, delayMs, noIndent)
}

func (s *mockDeviceSession) Key(combo string) error {
	return s.backend.Key(combo)
}

func (s *mockDeviceSession) Close() error {
	s.backend.closed++
	return nil
}
