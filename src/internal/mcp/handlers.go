/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package mcp

import (
	"fmt"

	"github.com/CassandraGargoyle/portunix-agent/internal/accessibility"
	"github.com/CassandraGargoyle/portunix-agent/internal/files"
	"github.com/CassandraGargoyle/portunix-agent/internal/vscode"
)

// handleFilesList lists files and directories at a given path
func (s *Server) handleFilesList(args map[string]any) (*ToolResult, error) {
	path, ok := getString(args, "path")
	if !ok || path == "" {
		return nil, fmt.Errorf("'path' parameter is required")
	}

	if len(s.config.AllowedDirs) == 0 {
		return nil, fmt.Errorf("no allowed directories configured (use --allowed-dirs flag with 'mcp serve')")
	}

	browser, err := files.NewBrowser(s.config.AllowedDirs)
	if err != nil {
		return nil, err
	}

	entries, err := browser.List(path)
	if err != nil {
		return nil, err
	}

	return jsonResult(entries)
}

// handleVSCodeStatus returns VS Code accessibility diagnostics
func (s *Server) handleVSCodeStatus(args map[string]any) (*ToolResult, error) {
	ctrl, err := vscode.NewController()
	if err != nil {
		return nil, fmt.Errorf("accessibility not available: %w", err)
	}
	defer ctrl.Close()

	status, err := ctrl.Status()
	if err != nil {
		return nil, err
	}

	return jsonResult(status)
}

// handleVSCodeTerminalRead reads text from a VS Code terminal
func (s *Server) handleVSCodeTerminalRead(args map[string]any) (*ToolResult, error) {
	terminalID := getInt(args, "terminal_id", 0)

	ctrl, err := vscode.NewController()
	if err != nil {
		return nil, fmt.Errorf("accessibility not available: %w", err)
	}
	defer ctrl.Close()

	content, err := ctrl.ReadTerminal(terminalID)
	if err != nil {
		return nil, err
	}

	return textResult(content), nil
}

// handleVSCodeTree dumps the VS Code accessibility tree
func (s *Server) handleVSCodeTree(args map[string]any) (*ToolResult, error) {
	depth := getInt(args, "depth", 10)

	ctrl, err := vscode.NewController()
	if err != nil {
		return nil, fmt.Errorf("accessibility not available: %w", err)
	}
	defer ctrl.Close()

	tree, err := ctrl.DumpTree(depth)
	if err != nil {
		return nil, err
	}

	return textResult(tree), nil
}

// handleAppListWindows lists top-level desktop windows via accessibility.
// When verbose=true, includes interfaces and child count per window.
func (s *Server) handleAppListWindows(args map[string]any) (*ToolResult, error) {
	verbose := getBool(args, "verbose", false)

	backend, err := accessibility.Detect()
	if err != nil {
		return nil, fmt.Errorf("accessibility not available: %w", err)
	}
	defer backend.Close()

	children, err := backend.Children(backend.RootRef())
	if err != nil {
		return nil, fmt.Errorf("failed to enumerate desktop windows: %w", err)
	}

	type windowEntry struct {
		Name       string   `json:"name"`
		Role       string   `json:"role"`
		Interfaces []string `json:"interfaces,omitempty"`
		Children   *int     `json:"children,omitempty"`
	}

	windows := make([]windowEntry, 0, len(children))
	for _, child := range children {
		entry := windowEntry{
			Name: child.Name,
			Role: child.RoleName,
		}
		if verbose {
			ifaces, _ := backend.Interfaces(child.Ref)
			entry.Interfaces = ifaces
			grandchildren, _ := backend.Children(child.Ref)
			count := len(grandchildren)
			entry.Children = &count
		}
		windows = append(windows, entry)
	}

	return jsonResult(windows)
}
