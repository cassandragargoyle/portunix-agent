/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

// Package accessibility provides an abstraction layer for accessibility technologies.
// Supported backends: AT-SPI2 on Linux (via the ipc package),
// UI Automation on Windows (via COM/go-ole).
package accessibility

// Accessible represents a node in the accessibility tree with its properties.
type Accessible struct {
	Ref      AccessibleRef
	Name     string
	RoleName string
}

// Backend defines the interface for accessibility tree operations.
type Backend interface {
	// FindApplication searches for an application by name in the accessibility tree
	FindApplication(name string) (*Accessible, error)

	// Children returns the direct children of an accessible node
	Children(ref AccessibleRef) ([]Accessible, error)

	// RoleName returns the role name of an accessible node (e.g. "application", "frame", "list")
	RoleName(ref AccessibleRef) (string, error)

	// NodeName returns the accessible name of a node
	NodeName(ref AccessibleRef) (string, error)

	// Text reads text content from a node implementing the Text interface.
	// Use startOffset=0 and endOffset=-1 to read all text.
	Text(ref AccessibleRef, startOffset, endOffset int32) (string, error)

	// CharacterCount returns the number of characters in a text node
	CharacterCount(ref AccessibleRef) (int32, error)

	// Interfaces returns the list of accessibility interfaces a node implements
	Interfaces(ref AccessibleRef) ([]string, error)

	// RootRef returns a reference to the desktop root element.
	// Used to enumerate top-level windows via Children(RootRef()).
	RootRef() AccessibleRef

	// Name returns the backend identifier (e.g. "atspi")
	Name() string

	// Close releases backend resources
	Close() error
}
