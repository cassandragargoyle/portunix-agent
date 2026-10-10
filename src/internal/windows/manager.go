//go:build linux

/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

// Package windows provides window management with pluggable backends.
package windows

import "fmt"

// Window represents a desktop window.
type Window struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Class string `json:"class"`
}

// Backend defines the interface for platform-specific window operations.
type Backend interface {
	// List returns all visible windows
	List() ([]Window, error)
	// Find searches for windows matching the given query by name
	Find(query string) ([]Window, error)
	// Focus activates the window with the given ID
	Focus(windowID string) error
	// Name returns the backend identifier (e.g. "xdotool", "kwin", "win32")
	Name() string
}

// Manager delegates window operations to a detected backend.
type Manager struct {
	backend Backend
}

// NewManager auto-detects the best available backend and creates a Manager.
func NewManager() (*Manager, error) {
	b, err := detectBackend()
	if err != nil {
		return nil, err
	}
	return &Manager{backend: b}, nil
}

// NewManagerWithBackend creates a Manager with a specific backend.
func NewManagerWithBackend(b Backend) *Manager {
	return &Manager{backend: b}
}

// List returns all visible windows.
func (m *Manager) List() ([]Window, error) {
	return m.backend.List()
}

// Find searches for windows matching the given query by name.
func (m *Manager) Find(query string) ([]Window, error) {
	if query == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}
	return m.backend.Find(query)
}

// Focus activates the window with the given ID.
func (m *Manager) Focus(windowID string) error {
	if windowID == "" {
		return fmt.Errorf("window id cannot be empty")
	}
	return m.backend.Focus(windowID)
}

// BackendName returns the name of the active backend.
func (m *Manager) BackendName() string {
	return m.backend.Name()
}
