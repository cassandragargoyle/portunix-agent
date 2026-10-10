/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package mcp

import "encoding/json"

// ToolHandler handles an MCP tool invocation
type ToolHandler func(args map[string]any) (*ToolResult, error)

// ToolResult is the result of a tool call
type ToolResult struct {
	Content []ContentBlock
	IsError bool
}

// ContentBlock represents a content block in a tool result
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// toolDef defines an MCP tool with its JSON Schema
type toolDef struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

// allTools defines all MCP tools exposed by ptx-agent
var allTools = []toolDef{
	{
		Name:        "agent_files_list",
		Description: "List files and directories at a given path within allowed directories",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "Directory path to list",
				},
			},
			"required": []string{"path"},
		},
	},
	{
		Name:        "agent_vscode_status",
		Description: "Get VS Code accessibility status and diagnostics",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	},
	{
		Name:        "agent_vscode_terminal_read",
		Description: "Read text content from a VS Code integrated terminal",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"terminal_id": map[string]any{
					"type":        "integer",
					"description": "Terminal index (1-based, 0 for active terminal)",
				},
			},
		},
	},
	{
		Name:        "agent_vscode_tree",
		Description: "Dump the VS Code accessibility tree for debugging",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"depth": map[string]any{
					"type":        "integer",
					"description": "Maximum tree depth (default: 10)",
				},
			},
		},
	},
	{
		Name:        "agent_app_type",
		Description: "Simulate keyboard text input in the focused application",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"text": map[string]any{
					"type":        "string",
					"description": "Text to type",
				},
			},
			"required": []string{"text"},
		},
	},
	{
		Name:        "agent_app_key",
		Description: "Send a key combination to the focused application (e.g. 'ctrl+c', 'alt+Tab')",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"keys": map[string]any{
					"type":        "string",
					"description": "Key combination (e.g. 'ctrl+c', 'alt+Tab', 'Return')",
				},
			},
			"required": []string{"keys"},
		},
	},
	{
		Name:        "agent_app_focus",
		Description: "Focus a window by searching for it by name",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{
					"type":        "string",
					"description": "Window name to search for",
				},
			},
			"required": []string{"name"},
		},
	},
	{
		Name:        "agent_app_list_windows",
		Description: "List all top-level desktop windows visible in the accessibility tree",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"verbose": map[string]any{
					"type":        "boolean",
					"description": "Include interfaces and child count for each window (default: false)",
				},
			},
		},
	},
	// Browser control tools
	{
		Name:        "agent_browser_launch",
		Description: "Launch a browser instance for web automation",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"headless": map[string]any{
					"type":        "boolean",
					"description": "Run in headless mode (default: true)",
				},
				"profile": map[string]any{
					"type":        "string",
					"description": "Named persistent profile; cookies and logins survive across launches (default: fresh non-persistent context)",
				},
				"channel": map[string]any{
					"type":        "string",
					"description": "Browser distribution channel to use a system-installed browser: chrome, chromium, msedge (default: Playwright-bundled Chromium)",
				},
			},
		},
	},
	{
		Name:        "agent_browser_connect",
		Description: "Attach to a user-started browser via Chrome DevTools Protocol (start it with --remote-debugging-port=9222); closing only disconnects, the browser keeps running",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"endpoint": map[string]any{
					"type":        "string",
					"description": "CDP endpoint URL (default: http://localhost:9222)",
				},
			},
		},
	},
	{
		Name:        "agent_browser_close",
		Description: "Close the browser instance",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	},
	{
		Name:        "agent_browser_status",
		Description: "Get browser and page state (running, URL, title, tab count)",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	},
	{
		Name:        "agent_browser_goto",
		Description: "Navigate to a URL",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"url": map[string]any{
					"type":        "string",
					"description": "URL to navigate to",
				},
				"new_tab": map[string]any{
					"type":        "boolean",
					"description": "Open URL in a new tab (default: false)",
				},
			},
			"required": []string{"url"},
		},
	},
	{
		Name:        "agent_browser_back",
		Description: "Navigate back in browser history",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	},
	{
		Name:        "agent_browser_forward",
		Description: "Navigate forward in browser history",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	},
	{
		Name:        "agent_browser_reload",
		Description: "Reload the current page",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	},
	{
		Name:        "agent_browser_click",
		Description: "Click an interactive element by its index (use agent_browser_elements to get indices)",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"index": map[string]any{
					"type":        "integer",
					"description": "Element index from agent_browser_elements",
				},
			},
			"required": []string{"index"},
		},
	},
	{
		Name:        "agent_browser_type",
		Description: "Type text into a form element by its index",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"index": map[string]any{
					"type":        "integer",
					"description": "Element index from agent_browser_elements",
				},
				"text": map[string]any{
					"type":        "string",
					"description": "Text to type",
				},
				"clear": map[string]any{
					"type":        "boolean",
					"description": "Clear the field before typing (default: false)",
				},
			},
			"required": []string{"index", "text"},
		},
	},
	{
		Name:        "agent_browser_keys",
		Description: "Send a key combination to the page (e.g. 'Enter', 'Control+a')",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"keys": map[string]any{
					"type":        "string",
					"description": "Key combination (e.g. 'Enter', 'Control+a', 'Tab')",
				},
			},
			"required": []string{"keys"},
		},
	},
	{
		Name:        "agent_browser_scroll",
		Description: "Scroll the page in a direction",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"direction": map[string]any{
					"type":        "string",
					"description": "Scroll direction: up, down, left, right",
				},
				"amount": map[string]any{
					"type":        "integer",
					"description": "Scroll amount multiplier (default: 3)",
				},
			},
			"required": []string{"direction"},
		},
	},
	{
		Name:        "agent_browser_select",
		Description: "Select a dropdown option by value",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"index": map[string]any{
					"type":        "integer",
					"description": "Element index of the select element",
				},
				"value": map[string]any{
					"type":        "string",
					"description": "Option value to select",
				},
			},
			"required": []string{"index", "value"},
		},
	},
	{
		Name:        "agent_browser_elements",
		Description: "List interactive elements on the page with numeric indices for use with click, type, and select tools",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"viewport_only": map[string]any{
					"type":        "boolean",
					"description": "Only show elements visible in viewport (default: false)",
				},
			},
		},
	},
	{
		Name:        "agent_browser_screenshot",
		Description: "Capture a screenshot of the current page (returns base64 PNG)",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"full_page": map[string]any{
					"type":        "boolean",
					"description": "Capture entire page including scrolled content (default: false)",
				},
			},
		},
	},
	{
		Name:        "agent_browser_content",
		Description: "Extract text or HTML content from the page",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"selector": map[string]any{
					"type":        "string",
					"description": "CSS selector to extract from (default: entire page)",
				},
				"format": map[string]any{
					"type":        "string",
					"description": "Output format: text or html (default: text)",
				},
			},
		},
	},
	{
		Name:        "agent_browser_tabs",
		Description: "List open browser tabs with their IDs, URLs, and titles",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	},
	{
		Name:        "agent_browser_tab_switch",
		Description: "Switch to a different browser tab",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"tab_id": map[string]any{
					"type":        "integer",
					"description": "Tab ID to switch to",
				},
			},
			"required": []string{"tab_id"},
		},
	},
	{
		Name:        "agent_browser_tab_close",
		Description: "Close a browser tab",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"tab_id": map[string]any{
					"type":        "integer",
					"description": "Tab ID to close (current tab if not specified)",
				},
			},
		},
	},
	{
		Name:        "agent_browser_tab_new",
		Description: "Open a new browser tab, optionally navigating to a URL",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"url": map[string]any{
					"type":        "string",
					"description": "URL to open in the new tab",
				},
			},
		},
	},
	{
		Name:        "agent_browser_evaluate",
		Description: "Execute JavaScript on the current page and return the result",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"expression": map[string]any{
					"type":        "string",
					"description": "JavaScript expression to evaluate",
				},
			},
			"required": []string{"expression"},
		},
	},
	{
		Name:        "agent_browser_upload",
		Description: "Upload a file to a file input element",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"index": map[string]any{
					"type":        "integer",
					"description": "Element index of the file input",
				},
				"path": map[string]any{
					"type":        "string",
					"description": "Local file path to upload",
				},
			},
			"required": []string{"index", "path"},
		},
	},
}

// Helper functions for tool handlers

func textResult(text string) *ToolResult {
	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: text}},
	}
}

func jsonResult(v any) (*ToolResult, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}
	return textResult(string(data)), nil
}

func getString(args map[string]any, key string) (string, bool) {
	v, ok := args[key]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

func getBool(args map[string]any, key string, defaultVal bool) bool {
	v, ok := args[key]
	if !ok {
		return defaultVal
	}
	b, ok := v.(bool)
	if !ok {
		return defaultVal
	}
	return b
}

func getInt(args map[string]any, key string, defaultVal int) int {
	v, ok := args[key]
	if !ok {
		return defaultVal
	}
	f, ok := v.(float64)
	if !ok {
		return defaultVal
	}
	return int(f)
}
