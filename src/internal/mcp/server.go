/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

// Package mcp implements the MCP (Model Context Protocol) server for ptx-agent.
// It exposes device control capabilities to AI assistants via stdio JSON-RPC.
package mcp

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// JSON-RPC 2.0 types

type jsonRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type jsonRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *jsonRPCError   `json:"error,omitempty"`
}

type jsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// JSON-RPC error codes
const (
	errParseError     = -32700
	errInvalidRequest = -32600
	errMethodNotFound = -32601
	errInvalidParams  = -32602
)

// MCP protocol types

type initializeResult struct {
	ProtocolVersion string     `json:"protocolVersion"`
	Capabilities    mcpCaps    `json:"capabilities"`
	ServerInfo      serverInfo `json:"serverInfo"`
}

type mcpCaps struct {
	Tools *struct{} `json:"tools,omitempty"`
}

type serverInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type toolsListResult struct {
	Tools []toolDef `json:"tools"`
}

type toolsCallParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments,omitempty"`
}

type toolCallResponse struct {
	Content []ContentBlock `json:"content"`
	IsError bool           `json:"isError,omitempty"`
}

// ServerConfig configures the MCP server.
type ServerConfig struct {
	AllowedDirs []string
}

// Server implements the MCP protocol over stdio or HTTP.
type Server struct {
	config   ServerConfig
	handlers map[string]ToolHandler
	writer   io.Writer
	mu       sync.Mutex
	reqMu    sync.Mutex // serializes HTTP request processing
}

// NewServer creates a new MCP server with the given configuration.
func NewServer(cfg ServerConfig) *Server {
	s := &Server{
		config:   cfg,
		handlers: make(map[string]ToolHandler),
		writer:   os.Stdout,
	}
	s.handlers["agent_files_list"] = s.handleFilesList
	s.handlers["agent_vscode_status"] = s.handleVSCodeStatus
	s.handlers["agent_vscode_terminal_read"] = s.handleVSCodeTerminalRead
	s.handlers["agent_vscode_tree"] = s.handleVSCodeTree
	s.handlers["agent_app_list_windows"] = s.handleAppListWindows
	registerPlatformHandlers(s)
	registerBrowserHandlers(s)
	return s
}

// Run starts the MCP server, reading JSON-RPC messages from stdin
// until EOF or a termination signal is received.
func (s *Server) Run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	lines := make(chan string)
	scanErr := make(chan error, 1)

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
		for scanner.Scan() {
			lines <- scanner.Text()
		}
		scanErr <- scanner.Err()
		close(lines)
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-scanErr:
			return err
		case line, ok := <-lines:
			if !ok {
				return nil
			}
			if line == "" {
				continue
			}
			s.handleMessage(line)
		}
	}
}

// GenerateToken returns a cryptographically random 32-character hex string
func GenerateToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// RunHTTP starts the MCP server in HTTP mode, listening on the given address.
// It validates requests using the provided bearer token.
func (s *Server) RunHTTP(ctx context.Context, addr, token string) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		auth := r.Header.Get("Authorization")
		if auth != "Bearer "+token {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}

		line := string(bytes.TrimSpace(body))
		if line == "" {
			http.Error(w, "Empty body", http.StatusBadRequest)
			return
		}

		resp := s.processHTTPRequest(line)
		w.Header().Set("Content-Type", "application/json")
		w.Write(resp)
	})

	srv := &http.Server{Handler: mux}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	fmt.Fprintf(os.Stderr, "MCP server listening on http://%s\n", ln.Addr())
	fmt.Fprintf(os.Stderr, "Token: %s\n", token)

	srvErr := make(chan error, 1)
	go func() {
		srvErr <- srv.Serve(ln)
	}()

	select {
	case <-ctx.Done():
		return srv.Close()
	case err := <-srvErr:
		return err
	}
}

// processHTTPRequest handles a single JSON-RPC message and returns the
// response bytes. It serializes access to ensure thread safety when
// multiple HTTP requests arrive concurrently.
func (s *Server) processHTTPRequest(line string) []byte {
	s.reqMu.Lock()
	defer s.reqMu.Unlock()

	buf := &bytes.Buffer{}

	s.mu.Lock()
	origWriter := s.writer
	s.writer = buf
	s.mu.Unlock()

	s.handleMessage(line)

	s.mu.Lock()
	s.writer = origWriter
	s.mu.Unlock()

	return bytes.TrimSpace(buf.Bytes())
}

func (s *Server) handleMessage(line string) {
	var req jsonRPCRequest
	if err := json.Unmarshal([]byte(line), &req); err != nil {
		s.sendError(nil, errParseError, "Parse error")
		return
	}

	if req.JSONRPC != "2.0" {
		s.sendError(req.ID, errInvalidRequest, "Invalid JSON-RPC version")
		return
	}

	switch req.Method {
	case "initialize":
		s.handleInitialize(req.ID)
	case "notifications/initialized", "initialized":
		// Notification — no response
	case "tools/list":
		s.handleToolsList(req.ID)
	case "tools/call":
		s.handleToolsCall(req.ID, req.Params)
	case "ping":
		s.sendResult(req.ID, map[string]any{})
	default:
		if req.ID != nil {
			s.sendError(req.ID, errMethodNotFound, fmt.Sprintf("Method not found: %s", req.Method))
		}
	}
}

func (s *Server) handleInitialize(id json.RawMessage) {
	s.sendResult(id, initializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities:    mcpCaps{Tools: &struct{}{}},
		ServerInfo: serverInfo{
			Name:    "ptx-agent",
			Version: "0.1.0",
		},
	})
}

func (s *Server) handleToolsList(id json.RawMessage) {
	s.sendResult(id, toolsListResult{Tools: allTools})
}

func (s *Server) handleToolsCall(id json.RawMessage, params json.RawMessage) {
	var p toolsCallParams
	if err := json.Unmarshal(params, &p); err != nil {
		s.sendError(id, errInvalidParams, "Invalid params")
		return
	}

	handler, ok := s.handlers[p.Name]
	if !ok {
		s.sendError(id, errMethodNotFound, fmt.Sprintf("Unknown tool: %s", p.Name))
		return
	}

	result, err := handler(p.Arguments)
	if err != nil {
		s.sendResult(id, toolCallResponse{
			Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error: %s", err)}},
			IsError: true,
		})
		return
	}

	s.sendResult(id, toolCallResponse{
		Content: result.Content,
		IsError: result.IsError,
	})
}

func (s *Server) sendResult(id json.RawMessage, result any) {
	s.send(jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	})
}

func (s *Server) sendError(id json.RawMessage, code int, message string) {
	s.send(jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &jsonRPCError{Code: code, Message: message},
	})
}

func (s *Server) send(resp jsonRPCResponse) {
	data, err := json.Marshal(resp)
	if err != nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	fmt.Fprintf(s.writer, "%s\n", data)
}
