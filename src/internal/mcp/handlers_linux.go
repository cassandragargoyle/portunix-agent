//go:build linux

/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package mcp

import (
	"fmt"

	"github.com/CassandraGargoyle/portunix-agent/internal/windows"
	"github.com/CassandraGargoyle/portunix-agent/pkg/keyboard"
)

// registerPlatformHandlers registers tool handlers that depend on
// Linux-only packages (keyboard, windows). On Linux these handlers
// use the real implementations; on other platforms handlers_other.go
// provides stubs that return "not available" errors.
func registerPlatformHandlers(s *Server) {
	s.handlers["agent_app_type"] = s.handleAppType
	s.handlers["agent_app_key"] = s.handleAppKey
	s.handlers["agent_app_focus"] = s.handleAppFocus
}

// handleAppType simulates keyboard text input
func (s *Server) handleAppType(args map[string]any) (*ToolResult, error) {
	text, ok := getString(args, "text")
	if !ok || text == "" {
		return nil, fmt.Errorf("'text' parameter is required")
	}

	ctrl, err := keyboard.NewController()
	if err != nil {
		return nil, err
	}

	if err := ctrl.Type(text, keyboard.DefaultTypingDelay, false); err != nil {
		return nil, err
	}

	return textResult(fmt.Sprintf("Typed %d characters", len(text))), nil
}

// handleAppKey sends a key combination
func (s *Server) handleAppKey(args map[string]any) (*ToolResult, error) {
	keys, ok := getString(args, "keys")
	if !ok || keys == "" {
		return nil, fmt.Errorf("'keys' parameter is required")
	}

	ctrl, err := keyboard.NewController()
	if err != nil {
		return nil, err
	}

	if err := ctrl.Key(keys); err != nil {
		return nil, err
	}

	return textResult(fmt.Sprintf("Sent key combination: %s", keys)), nil
}

// handleAppFocus focuses a window by name
func (s *Server) handleAppFocus(args map[string]any) (*ToolResult, error) {
	name, ok := getString(args, "name")
	if !ok || name == "" {
		return nil, fmt.Errorf("'name' parameter is required")
	}

	mgr, err := windows.NewManager()
	if err != nil {
		return nil, err
	}

	wins, err := mgr.Find(name)
	if err != nil {
		return nil, err
	}
	if len(wins) == 0 {
		return nil, fmt.Errorf("no window matching %q found", name)
	}

	if err := mgr.Focus(wins[0].ID); err != nil {
		return nil, err
	}

	return textResult(fmt.Sprintf("Focused window: %s", wins[0].Name)), nil
}
