//go:build !linux

/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package mcp

import "fmt"

// registerPlatformHandlers registers stub handlers for tools that require
// platform-specific packages not available on this OS.
// See handlers_linux.go for the real implementations.
func registerPlatformHandlers(s *Server) {
	s.handlers["agent_app_type"] = handleUnsupported
	s.handlers["agent_app_key"] = handleUnsupported
	s.handlers["agent_app_focus"] = handleUnsupported
}

func handleUnsupported(_ map[string]any) (*ToolResult, error) {
	return nil, fmt.Errorf("this tool is not available on the current platform")
}
