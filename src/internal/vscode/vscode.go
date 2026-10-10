/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

// Package vscode provides VS Code control via the accessibility interface.
// It depends only on the accessibility.Backend interface, never on AT-SPI2 or D-Bus directly.
package vscode

import (
	"fmt"
	"strings"

	"github.com/CassandraGargoyle/portunix-agent/internal/accessibility"
)

// appName is the accessible name used by VS Code
const appName = "Code"

// Status represents VS Code accessibility diagnostics.
type Status struct {
	Found         bool
	AppName       string
	Accessible    bool
	TerminalCount int
}

// Controller provides VS Code control via the accessibility tree.
type Controller struct {
	backend accessibility.Backend
}

// NewController auto-detects the accessibility backend and creates a Controller.
func NewController() (*Controller, error) {
	b, err := accessibility.Detect()
	if err != nil {
		return nil, err
	}
	return &Controller{backend: b}, nil
}

// NewControllerWithBackend creates a Controller with a specific backend.
func NewControllerWithBackend(b accessibility.Backend) *Controller {
	return &Controller{backend: b}
}

// Status returns VS Code accessibility diagnostics.
func (c *Controller) Status() (*Status, error) {
	app, err := c.backend.FindApplication(appName)
	if err != nil {
		return &Status{Found: false}, nil
	}

	status := &Status{
		Found:      true,
		AppName:    app.Name,
		Accessible: true,
	}

	terminals, err := c.findTerminalLists(app.Ref)
	if err == nil {
		status.TerminalCount = len(terminals)
	}

	return status, nil
}

// Close releases backend resources.
func (c *Controller) Close() error {
	return c.backend.Close()
}

// BackendName returns the name of the active accessibility backend.
func (c *Controller) BackendName() string {
	return c.backend.Name()
}

// findTerminalLists walks the VS Code accessibility tree to find terminal list nodes.
// VS Code terminal uses xterm.js which creates a list with "list item" children
// implementing the Text interface. The list is deeply nested (~30 levels) inside
// the VS Code accessibility tree.
func (c *Controller) findTerminalLists(appRef accessibility.AccessibleRef) ([]terminalInfo, error) {
	var terminals []terminalInfo
	err := c.walkTree(appRef, 0, func(node accessibility.Accessible, depth int) bool {
		if node.RoleName == "list" && c.isTerminalList(node) {
			terminals = append(terminals, terminalInfo{
				ref:  node.Ref,
				name: node.Name,
			})
			return false // no need to recurse into the list itself
		}
		// VS Code terminal list is deeply nested (~30 levels)
		return depth < 40
	})
	if err != nil {
		return nil, err
	}
	return terminals, nil
}

// terminalInfo holds the reference and name of a terminal list node
type terminalInfo struct {
	ref  accessibility.AccessibleRef
	name string
}

// walkTree performs a depth-first traversal of the accessibility tree.
// The visitor function returns true to continue into children, false to skip.
func (c *Controller) walkTree(ref accessibility.AccessibleRef, depth int, visitor func(accessibility.Accessible, int) bool) error {
	children, err := c.backend.Children(ref)
	if err != nil {
		return err
	}

	for _, child := range children {
		if visitor(child, depth) {
			if err := c.walkTree(child.Ref, depth+1, visitor); err != nil {
				// Continue traversal even if a subtree fails
				continue
			}
		}
	}
	return nil
}

// isTerminalList checks if a list node is an xterm.js terminal accessibility tree.
// It verifies by checking if the list contains "list item" children with text content,
// which is how xterm.js exposes terminal rows.
func (c *Controller) isTerminalList(node accessibility.Accessible) bool {
	// Named lists with "xterm" or "terminal" are definitely terminal lists
	nameLower := strings.ToLower(node.Name)
	if strings.Contains(nameLower, "xterm") || strings.Contains(nameLower, "terminal") {
		return true
	}

	// For unnamed lists (common in newer VS Code versions), verify by checking
	// if children are "list item" nodes — this distinguishes terminal lists
	// from other lists (e.g. file explorer, which uses "tree item")
	children, err := c.backend.Children(node.Ref)
	if err != nil || len(children) == 0 {
		return false
	}
	for _, child := range children {
		if child.RoleName == "list item" {
			return true
		}
	}
	return false
}

// DumpTree returns a text representation of the accessibility tree for debugging.
// Use maxDepth >= 35 to see the full terminal area.
func (c *Controller) DumpTree(maxDepth int) (string, error) {
	app, err := c.backend.FindApplication(appName)
	if err != nil {
		return "", fmt.Errorf("VS Code not found: %w", err)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[%s] %q\n", app.RoleName, app.Name))
	c.dumpSubtree(&sb, app.Ref, 1, maxDepth)
	return sb.String(), nil
}

func (c *Controller) dumpSubtree(sb *strings.Builder, ref accessibility.AccessibleRef, depth, maxDepth int) {
	if depth > maxDepth {
		return
	}

	children, err := c.backend.Children(ref)
	if err != nil {
		return
	}

	indent := strings.Repeat("  ", depth)
	for _, child := range children {
		ifaces, _ := c.backend.Interfaces(child.Ref)
		ifaceStr := ""
		if len(ifaces) > 0 {
			ifaceStr = fmt.Sprintf(" {%s}", strings.Join(ifaces, ", "))
		}
		sb.WriteString(fmt.Sprintf("%s[%s] %q%s\n", indent, child.RoleName, child.Name, ifaceStr))
		c.dumpSubtree(sb, child.Ref, depth+1, maxDepth)
	}
}
