//go:build linux

/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

// Package keyboard provides keyboard input simulation with pluggable backends.
package keyboard

import "fmt"

// DefaultTypingDelay is the default delay between keystrokes in milliseconds.
const DefaultTypingDelay = 50

// Backend defines the interface for platform-specific keyboard operations.
type Backend interface {
	// Type sends a text string as keyboard input
	Type(text string, delayMs int, noIndent bool) error
	// Key sends a key combination (e.g. "ctrl+c", "alt+Tab")
	Key(combo string) error
	// Name returns the backend identifier
	Name() string
}

// Controller delegates keyboard operations to a detected backend.
type Controller struct {
	backend Backend
}

// NewController auto-detects the best available backend and creates a Controller.
func NewController() (*Controller, error) {
	b, err := detectBackend()
	if err != nil {
		return nil, err
	}
	return &Controller{backend: b}, nil
}

// NewControllerWithBackend creates a Controller with a specific backend.
func NewControllerWithBackend(b Backend) *Controller {
	return &Controller{backend: b}
}

// Type sends a text string as keyboard input to the currently focused window.
// When noIndent is true, a Home key is sent after each newline to reset cursor
// position, preventing auto-indent from stacking with typed whitespace.
func (c *Controller) Type(text string, delayMs int, noIndent bool) error {
	if text == "" {
		return fmt.Errorf("text cannot be empty")
	}
	if delayMs <= 0 {
		delayMs = DefaultTypingDelay
	}
	return c.backend.Type(text, delayMs, noIndent)
}

// Key sends a key combination to the currently focused window.
func (c *Controller) Key(combo string) error {
	if combo == "" {
		return fmt.Errorf("key combination cannot be empty")
	}
	return c.backend.Key(combo)
}

// BackendName returns the name of the active backend.
func (c *Controller) BackendName() string {
	return c.backend.Name()
}
