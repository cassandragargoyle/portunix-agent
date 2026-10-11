//go:build linux

/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package keyboard

import (
	"fmt"
	"sync"
)

// Session is an open keyboard device that accepts many actions
type Session interface {
	// Type sends a text string as keyboard input
	Type(text string, delayMs int, noIndent bool) error
	// Key sends a key combination (e.g. "ctrl+c", "alt+Tab")
	Key(combo string) error
	// Close releases the device
	Close() error
}

// SessionBackend is an optional interface for a Backend that keeps
// a device open between actions.
type SessionBackend interface {
	Backend
	// OpenSession opens a device that the caller must close
	OpenSession() (Session, error)
}

// Open creates a session on the active backend. A backend that does not
// implement SessionBackend gets a session that delegates every call to it.
func (c *Controller) Open() (Session, error) {
	sb, ok := c.backend.(SessionBackend)
	if !ok {
		return &session{inner: backendSession{backend: c.backend}}, nil
	}
	inner, err := sb.OpenSession()
	if err != nil {
		return nil, err
	}
	return &session{inner: inner}, nil
}

// session adds input validation and closed-state handling to a backend session.
type session struct {
	mu     sync.Mutex
	inner  Session
	closed bool
}

func (s *session) Type(text string, delayMs int, noIndent bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return fmt.Errorf("keyboard session is closed")
	}
	delayMs, err := checkType(text, delayMs)
	if err != nil {
		return err
	}
	return s.inner.Type(text, delayMs, noIndent)
}

func (s *session) Key(combo string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return fmt.Errorf("keyboard session is closed")
	}
	if err := checkKey(combo); err != nil {
		return err
	}
	return s.inner.Key(combo)
}

func (s *session) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	return s.inner.Close()
}

// backendSession delegates to a backend that has no device to keep open.
type backendSession struct {
	backend Backend
}

func (s backendSession) Type(text string, delayMs int, noIndent bool) error {
	return s.backend.Type(text, delayMs, noIndent)
}

func (s backendSession) Key(combo string) error {
	return s.backend.Key(combo)
}

func (s backendSession) Close() error { return nil }
