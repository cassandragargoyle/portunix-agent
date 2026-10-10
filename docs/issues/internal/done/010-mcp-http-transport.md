# Issue #010 - MCP Server: HTTP transport with token authentication

**Type**: enhancement
**Priority**: high
**Status**: ✅ Implemented
**Branch**: `feature/010-mcp-http-transport`
**Depends on**: Issue #006
**Blocks**: Issue #009

## Problem

The current MCP server (`ptx-agent mcp serve`) communicates only via stdio (stdin/stdout JSON-RPC). This means:

- The server lives and dies with the pipe — when stdin closes, the browser and all state is lost
- Only one client can connect (the process holding the pipe)
- External tools (e.g., Claude Code calling MCP tools from bash) cannot interact with a running browser session interactively
- No way to keep a browser session alive across multiple command invocations

## Feature Description

Add an HTTP-based transport to the MCP server so that:

1. `ptx-agent mcp serve` can listen on a TCP port instead of (or in addition to) stdio
2. A session token is generated at startup and printed to stdout
3. Clients send JSON-RPC requests as HTTP POST with the token in the `Authorization` header
4. The browser session and all state persist as long as the server process runs
5. Multiple sequential requests from the same or different clients work against the same session

## Proposed Interface

### Server startup

```bash
# HTTP mode — listen on port, print token
ptx-agent mcp serve --http :9100
# Output: MCP server listening on http://localhost:9100
# Output: Token: a1b2c3d4e5f6

# Stdio mode — unchanged (backward compatible)
ptx-agent mcp serve
```

### Client requests

```bash
# Send a single MCP tool call
curl -s -X POST http://localhost:9100 \
  -H "Authorization: Bearer a1b2c3d4e5f6" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"agent_browser_launch","arguments":{"headless":false}}}'
```

### Token management

- Token is a random string generated at server startup
- Printed to stdout (or stderr) so the launching process can capture it
- Every HTTP request must include `Authorization: Bearer <token>`
- Requests without a valid token get HTTP 401
- Optional: `--token <value>` flag to set a specific token instead of generating one

## Technical Requirements

1. **HTTP handler**: Accept POST requests with JSON-RPC body, dispatch to existing `handleMessage()` logic
2. **Token auth**: Generate random token, validate on every request, reject unauthorized requests
3. **Backward compatible**: `ptx-agent mcp serve` without `--http` flag continues to use stdio
4. **Single response per request**: Each HTTP POST contains one JSON-RPC message, response is returned as HTTP response body
5. **CORS**: Not needed initially (localhost only)
6. **Graceful shutdown**: `SIGINT`/`SIGTERM` closes HTTP server and browser session

## Architecture Change

```text
Current:
  stdin → [MCP Server] → stdout

New (HTTP mode):
  HTTP POST :9100 → [Auth middleware] → [MCP Server] → HTTP Response
```

The core `handleMessage()` logic stays the same — only the transport layer changes.

## MCP Specification Compliance

The MCP spec (2024-11-05) defines [Streamable HTTP transport](https://modelcontextprotocol.io/specification/2025-03-26/basic/transports#streamable-http) as an alternative to stdio. This issue aligns with that specification:

- POST endpoint for client-to-server messages
- JSON-RPC request/response over HTTP
- Session management via token

## Acceptance Criteria

1. [ ] `ptx-agent mcp serve --http :9100` starts an HTTP server on port 9100
2. [ ] Server prints a session token at startup
3. [ ] JSON-RPC requests via HTTP POST with valid token are processed correctly
4. [ ] Requests without valid token return HTTP 401
5. [ ] Browser session persists across multiple HTTP requests
6. [ ] `ptx-agent mcp serve` (without `--http`) still works via stdio
7. [ ] Graceful shutdown on SIGINT/SIGTERM
8. [ ] Works on Linux and Windows

## Complexity Estimate

Medium — the JSON-RPC handling logic already exists, this adds an HTTP wrapper with token auth on top.
