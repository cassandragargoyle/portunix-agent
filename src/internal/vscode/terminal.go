/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package vscode

import (
	"fmt"
	"strings"

	"github.com/CassandraGargoyle/portunix-agent/internal/accessibility"
)

// Terminal represents a VS Code terminal instance.
type Terminal struct {
	Index int
	Name  string
}

// ListTerminals returns all open terminal instances in VS Code.
func (c *Controller) ListTerminals() ([]Terminal, error) {
	app, err := c.backend.FindApplication(appName)
	if err != nil {
		return nil, fmt.Errorf("VS Code not found in accessibility tree: %w", err)
	}

	lists, err := c.findTerminalLists(app.Ref)
	if err != nil {
		return nil, fmt.Errorf("failed to find terminal panels: %w", err)
	}

	if len(lists) == 0 {
		return nil, fmt.Errorf("no terminals found (is VS Code accessibility enabled? set editor.accessibilitySupport to 'on' in VS Code settings)")
	}

	terminals := make([]Terminal, len(lists))
	for i, info := range lists {
		name := info.name
		if name == "" {
			name = fmt.Sprintf("Terminal %d", i+1)
		}
		terminals[i] = Terminal{
			Index: i + 1,
			Name:  name,
		}
	}

	return terminals, nil
}

// ReadTerminal reads the text content from a specific terminal instance.
// Index is 1-based. Use index 0 or -1 to read the first (active) terminal.
func (c *Controller) ReadTerminal(index int) (string, error) {
	app, err := c.backend.FindApplication(appName)
	if err != nil {
		return "", fmt.Errorf("VS Code not found in accessibility tree: %w", err)
	}

	lists, err := c.findTerminalLists(app.Ref)
	if err != nil {
		return "", fmt.Errorf("failed to find terminal panels: %w", err)
	}

	if len(lists) == 0 {
		return "", fmt.Errorf("no terminals found (is VS Code accessibility enabled? set editor.accessibilitySupport to 'on' in VS Code settings)")
	}

	// Resolve index (1-based to 0-based)
	idx := 0
	if index > 0 {
		idx = index - 1
	}
	if idx >= len(lists) {
		return "", fmt.Errorf("terminal index %d out of range (have %d terminals)", index, len(lists))
	}

	return c.readTerminalContent(lists[idx].ref)
}

// readTerminalContent reads all text rows from a terminal list node.
// Each list item in xterm.js accessibility tree is one terminal row
// implementing the Text interface.
func (c *Controller) readTerminalContent(listRef accessibility.AccessibleRef) (string, error) {
	children, err := c.backend.Children(listRef)
	if err != nil {
		return "", fmt.Errorf("failed to get terminal rows: %w", err)
	}

	var lines []string
	for _, child := range children {
		if child.RoleName != "list item" {
			continue
		}

		// Check if this list item implements the Text interface
		hasText, err := c.hasTextInterface(child.Ref)
		if err != nil || !hasText {
			// Fall back to using the accessible name as text content
			if child.Name != "" {
				lines = append(lines, child.Name)
			}
			continue
		}

		text, err := c.backend.Text(child.Ref, 0, -1)
		if err != nil {
			// Fall back to accessible name
			if child.Name != "" {
				lines = append(lines, child.Name)
			}
			continue
		}
		lines = append(lines, text)
	}

	return strings.Join(lines, "\n"), nil
}

// hasTextInterface checks if a node implements the Text interface.
// Handles both AT-SPI2 ("org.a11y.atspi.Text") and UIA ("Text") interface names.
func (c *Controller) hasTextInterface(ref accessibility.AccessibleRef) (bool, error) {
	ifaces, err := c.backend.Interfaces(ref)
	if err != nil {
		return false, err
	}
	for _, iface := range ifaces {
		if iface == "org.a11y.atspi.Text" || iface == "Text" {
			return true, nil
		}
	}
	return false, nil
}
