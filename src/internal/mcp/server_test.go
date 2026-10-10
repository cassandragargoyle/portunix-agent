/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
)

// testServer creates a server with captured output for testing
func testServer() (*Server, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	s := NewServer(ServerConfig{})
	s.writer = buf
	return s, buf
}

func parseResponse(t *testing.T, buf *bytes.Buffer) jsonRPCResponse {
	t.Helper()
	var resp jsonRPCResponse
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v\nraw: %s", err, buf.String())
	}
	return resp
}

func TestHandleInitialize(t *testing.T) {
	s, buf := testServer()

	msg := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","clientInfo":{"name":"test","version":"1.0"}}}`
	s.handleMessage(msg)

	resp := parseResponse(t, buf)
	if resp.Error != nil {
		t.Fatalf("unexpected error: %s", resp.Error.Message)
	}
	if string(resp.ID) != "1" {
		t.Errorf("expected id 1, got %s", resp.ID)
	}

	result, err := json.Marshal(resp.Result)
	if err != nil {
		t.Fatal(err)
	}

	var initResult initializeResult
	if err := json.Unmarshal(result, &initResult); err != nil {
		t.Fatal(err)
	}

	if initResult.ProtocolVersion != "2024-11-05" {
		t.Errorf("expected protocol version 2024-11-05, got %s", initResult.ProtocolVersion)
	}
	if initResult.ServerInfo.Name != "ptx-agent" {
		t.Errorf("expected server name ptx-agent, got %s", initResult.ServerInfo.Name)
	}
	if initResult.Capabilities.Tools == nil {
		t.Error("expected tools capability to be set")
	}
}

func TestHandleToolsList(t *testing.T) {
	s, buf := testServer()

	msg := `{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`
	s.handleMessage(msg)

	resp := parseResponse(t, buf)
	if resp.Error != nil {
		t.Fatalf("unexpected error: %s", resp.Error.Message)
	}

	result, _ := json.Marshal(resp.Result)
	var listResult toolsListResult
	if err := json.Unmarshal(result, &listResult); err != nil {
		t.Fatal(err)
	}

	if len(listResult.Tools) != len(allTools) {
		t.Errorf("expected %d tools, got %d", len(allTools), len(listResult.Tools))
	}

	// Verify all expected tools are present
	expected := map[string]bool{
		"agent_files_list":           false,
		"agent_vscode_status":        false,
		"agent_vscode_terminal_read": false,
		"agent_vscode_tree":          false,
		"agent_app_type":             false,
		"agent_app_key":              false,
		"agent_app_focus":            false,
		"agent_app_list_windows":     false,
		"agent_browser_launch":       false,
		"agent_browser_connect":      false,
	}

	for _, tool := range listResult.Tools {
		if _, ok := expected[tool.Name]; ok {
			expected[tool.Name] = true
		}
	}

	for name, found := range expected {
		if !found {
			t.Errorf("tool %q not found in tools/list response", name)
		}
	}
}

func TestHandleNotification(t *testing.T) {
	s, buf := testServer()

	// notifications/initialized should produce no response
	msg := `{"jsonrpc":"2.0","method":"notifications/initialized"}`
	s.handleMessage(msg)

	if buf.Len() != 0 {
		t.Errorf("expected no response for notification, got: %s", buf.String())
	}
}

func TestHandlePing(t *testing.T) {
	s, buf := testServer()

	msg := `{"jsonrpc":"2.0","id":3,"method":"ping"}`
	s.handleMessage(msg)

	resp := parseResponse(t, buf)
	if resp.Error != nil {
		t.Fatalf("unexpected error: %s", resp.Error.Message)
	}
}

func TestHandleUnknownMethod(t *testing.T) {
	s, buf := testServer()

	msg := `{"jsonrpc":"2.0","id":4,"method":"unknown/method"}`
	s.handleMessage(msg)

	resp := parseResponse(t, buf)
	if resp.Error == nil {
		t.Fatal("expected error for unknown method")
	}
	if resp.Error.Code != errMethodNotFound {
		t.Errorf("expected error code %d, got %d", errMethodNotFound, resp.Error.Code)
	}
}

func TestHandleParseError(t *testing.T) {
	s, buf := testServer()

	s.handleMessage("not valid json")

	resp := parseResponse(t, buf)
	if resp.Error == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if resp.Error.Code != errParseError {
		t.Errorf("expected error code %d, got %d", errParseError, resp.Error.Code)
	}
}

func TestHandleToolsCallUnknownTool(t *testing.T) {
	s, buf := testServer()

	msg := `{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"nonexistent_tool","arguments":{}}}`
	s.handleMessage(msg)

	resp := parseResponse(t, buf)
	if resp.Error == nil {
		t.Fatal("expected error for unknown tool")
	}
	if resp.Error.Code != errMethodNotFound {
		t.Errorf("expected error code %d, got %d", errMethodNotFound, resp.Error.Code)
	}
}

func TestHandleFilesListNoAllowedDirs(t *testing.T) {
	s, buf := testServer()

	msg := `{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"agent_files_list","arguments":{"path":"/tmp"}}}`
	s.handleMessage(msg)

	resp := parseResponse(t, buf)
	if resp.Error != nil {
		t.Fatalf("expected tool error in result, not JSON-RPC error: %s", resp.Error.Message)
	}

	// Should return isError in the tool response
	result, _ := json.Marshal(resp.Result)
	var callResult toolCallResponse
	if err := json.Unmarshal(result, &callResult); err != nil {
		t.Fatal(err)
	}

	if !callResult.IsError {
		t.Error("expected isError to be true when allowed-dirs not configured")
	}
	if len(callResult.Content) == 0 || !strings.Contains(callResult.Content[0].Text, "allowed directories") {
		t.Error("expected error message about allowed directories")
	}
}

func TestHandleFilesListWithAllowedDirs(t *testing.T) {
	s, buf := testServer()
	dir := t.TempDir()
	s.config.AllowedDirs = []string{dir}

	// Marshal the path so Windows backslashes are escaped correctly
	pathJSON, err := json.Marshal(dir)
	if err != nil {
		t.Fatal(err)
	}
	msg := `{"jsonrpc":"2.0","id":7,"method":"tools/call","params":{"name":"agent_files_list","arguments":{"path":` + string(pathJSON) + `}}}`
	s.handleMessage(msg)

	resp := parseResponse(t, buf)
	if resp.Error != nil {
		t.Fatalf("unexpected JSON-RPC error: %s", resp.Error.Message)
	}

	result, _ := json.Marshal(resp.Result)
	var callResult toolCallResponse
	if err := json.Unmarshal(result, &callResult); err != nil {
		t.Fatal(err)
	}

	if callResult.IsError {
		t.Errorf("expected success, got error: %s", callResult.Content[0].Text)
	}
}

func TestToolDefinitionsHaveSchemas(t *testing.T) {
	for _, tool := range allTools {
		if tool.Name == "" {
			t.Error("tool has empty name")
		}
		if tool.Description == "" {
			t.Errorf("tool %q has empty description", tool.Name)
		}
		if tool.InputSchema == nil {
			t.Errorf("tool %q has nil inputSchema", tool.Name)
		}
		schemaType, ok := tool.InputSchema["type"]
		if !ok || schemaType != "object" {
			t.Errorf("tool %q inputSchema type should be 'object', got %v", tool.Name, schemaType)
		}
	}
}

// --- HTTP transport tests ---

func mustFreeAddr(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to find free port: %v", err)
	}
	addr := ln.Addr().String()
	ln.Close()
	return addr
}

func TestGenerateToken(t *testing.T) {
	token1, err := GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	if len(token1) != 32 {
		t.Errorf("expected 32-char hex token, got %d chars: %s", len(token1), token1)
	}

	token2, err := GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	if token1 == token2 {
		t.Error("two generated tokens should not be identical")
	}
}

func TestProcessHTTPRequest(t *testing.T) {
	server := NewServer(ServerConfig{})

	msg := `{"jsonrpc":"2.0","id":1,"method":"ping"}`
	resp := server.processHTTPRequest(msg)

	var rpcResp jsonRPCResponse
	if err := json.Unmarshal(resp, &rpcResp); err != nil {
		t.Fatalf("failed to parse response: %v\nraw: %s", err, resp)
	}
	if rpcResp.Error != nil {
		t.Fatalf("unexpected error: %s", rpcResp.Error.Message)
	}
	if string(rpcResp.ID) != "1" {
		t.Errorf("expected id 1, got %s", rpcResp.ID)
	}
}

func TestProcessHTTPRequestParseError(t *testing.T) {
	server := NewServer(ServerConfig{})

	resp := server.processHTTPRequest("not valid json")

	var rpcResp jsonRPCResponse
	if err := json.Unmarshal(resp, &rpcResp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if rpcResp.Error == nil {
		t.Fatal("expected parse error")
	}
	if rpcResp.Error.Code != errParseError {
		t.Errorf("expected error code %d, got %d", errParseError, rpcResp.Error.Code)
	}
}

func TestHTTPServerIntegration(t *testing.T) {
	token := "test-token-abc123"
	addr := mustFreeAddr(t)
	baseURL := fmt.Sprintf("http://%s", addr)

	server := NewServer(ServerConfig{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		server.RunHTTP(ctx, addr, token)
	}()
	time.Sleep(150 * time.Millisecond)

	t.Run("valid request", func(t *testing.T) {
		body := `{"jsonrpc":"2.0","id":1,"method":"ping"}`
		req, _ := http.NewRequest("POST", baseURL, strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}

		respBody, _ := io.ReadAll(resp.Body)
		var rpcResp jsonRPCResponse
		if err := json.Unmarshal(respBody, &rpcResp); err != nil {
			t.Fatalf("failed to parse response: %v\nraw: %s", err, respBody)
		}
		if rpcResp.Error != nil {
			t.Errorf("unexpected error: %s", rpcResp.Error.Message)
		}
	})

	t.Run("missing token returns 401", func(t *testing.T) {
		body := `{"jsonrpc":"2.0","id":1,"method":"ping"}`
		resp, err := http.Post(baseURL, "application/json", strings.NewReader(body))
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 401 {
			t.Errorf("expected 401, got %d", resp.StatusCode)
		}
	})

	t.Run("wrong token returns 401", func(t *testing.T) {
		body := `{"jsonrpc":"2.0","id":1,"method":"ping"}`
		req, _ := http.NewRequest("POST", baseURL, strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer wrong-token")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 401 {
			t.Errorf("expected 401, got %d", resp.StatusCode)
		}
	})

	t.Run("GET method returns 405", func(t *testing.T) {
		req, _ := http.NewRequest("GET", baseURL, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 405 {
			t.Errorf("expected 405, got %d", resp.StatusCode)
		}
	})

	t.Run("tools/list via HTTP", func(t *testing.T) {
		body := `{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`
		req, _ := http.NewRequest("POST", baseURL, strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}

		respBody, _ := io.ReadAll(resp.Body)
		var rpcResp jsonRPCResponse
		if err := json.Unmarshal(respBody, &rpcResp); err != nil {
			t.Fatalf("failed to parse: %v", err)
		}
		if rpcResp.Error != nil {
			t.Errorf("unexpected error: %s", rpcResp.Error.Message)
		}
	})
}

func TestHelperFunctions(t *testing.T) {
	// Test textResult
	r := textResult("hello")
	if len(r.Content) != 1 || r.Content[0].Text != "hello" || r.Content[0].Type != "text" {
		t.Error("textResult returned unexpected content")
	}
	if r.IsError {
		t.Error("textResult should not be an error")
	}

	// Test getString
	args := map[string]any{"key": "value", "num": float64(42)}
	s, ok := getString(args, "key")
	if !ok || s != "value" {
		t.Error("getString failed for existing key")
	}
	_, ok = getString(args, "missing")
	if ok {
		t.Error("getString should return false for missing key")
	}

	// Test getInt
	n := getInt(args, "num", 0)
	if n != 42 {
		t.Errorf("getInt expected 42, got %d", n)
	}
	n = getInt(args, "missing", 99)
	if n != 99 {
		t.Errorf("getInt default expected 99, got %d", n)
	}
}
