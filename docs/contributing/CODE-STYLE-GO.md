# Go Code Style Guide for portunix-agent

## Overview

This document defines the Go coding standards for the portunix-agent project.

## File Header

Every `.go` file must start with the following license header:

```go
/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
```

Place the header at the very top of the file, before the `package` declaration.

## General Principles

- Follow the official [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- Use `gofmt` for automatic formatting
- Run `go vet` and `golint` regularly
- Keep it simple and readable

## Naming Conventions

### Packages

- Use lowercase, single-word names: `http`, `util`
- Avoid underscores or mixedCaps
- Package name should be short and clear
- Avoid generic names like `util`, `common`, `misc` when possible

### Files

- Use lowercase with underscores: `http_server.go`
- Test files: `<name>_test.go`
- Platform-specific: `<name>_linux.go`

### Variables and Functions

- Use camelCase for private: `myVariable`
- Use PascalCase for public: `MyFunction`
- Short names for short scopes: `i`, `err`, `ctx`
- Descriptive names for package-level declarations

### Constants

- Use PascalCase or camelCase depending on visibility
```go
const MaxRetries = 3        // Exported
const defaultTimeout = 30   // Unexported
```

### Interfaces

- Single-method interfaces: method name + "er" suffix: `Reader`, `Writer`
- Multiple methods: descriptive name: `FileSystem`, `Database`

## Code Organization

### Package Structure

```text
portunix-agent/
├── cmd/                    # Main applications
│   └── app/
│       └── main.go
├── internal/              # Private packages
│   ├── config/
│   └── database/
├── pkg/                   # Public packages
│   ├── api/
│   └── models/
├── test/                  # Additional test data
├── docs/                  # Documentation
├── go.mod
└── go.sum
```

### Import Grouping
```go
import (
    // Standard library
    "fmt"
    "os"
    
    // Third-party packages
    "github.com/gorilla/mux"
    "github.com/sirupsen/logrus"
    
    // Local packages
    "portunix-agent/pkg/api"
    "portunix-agent/internal/config"
)
```

## Code Formatting

### Line Length

- No hard limit, but aim for readability
- Break long lines at logical points

### Comments

```go
// Package http provides HTTP client and server implementations.
package http

// Server represents an HTTP server.
// It handles incoming requests and routes them to appropriate handlers.
type Server struct {
    // router handles request routing
    router *mux.Router
    // config holds server configuration
    config *Config
}

// Start begins listening for HTTP requests.
// It returns an error if the server cannot start.
func (s *Server) Start() error {
    // Implementation here
    return nil
}
```

## Error Handling

### Error Messages

- Start with lowercase
- Don't end with punctuation
- Be specific and actionable

```go
// Good
return fmt.Errorf("cannot open file %s: %w", filename, err)

// Bad
return fmt.Errorf("Error opening file!")
```

### Error Checking

```go
// Always check errors immediately
result, err := doSomething()
if err != nil {
    return fmt.Errorf("failed to do something: %w", err)
}

// For deferred functions
defer func() {
    if err := file.Close(); err != nil {
        log.Printf("failed to close file: %v", err)
    }
}()
```

## Best Practices

### Interfaces

- Accept interfaces, return structs
- Keep interfaces small
- Define interfaces where they're used

```go
// Good - defined by consumer
type Storage interface {
    Save(key string, data []byte) error
    Load(key string) ([]byte, error)
}

func ProcessData(s Storage) error {
    // Use storage
}
```

### Concurrency

- Don't communicate by sharing memory; share memory by communicating
- Use channels for coordination
- Protect shared state with mutexes

```go
// Channel usage
type Worker struct {
    tasks chan Task
    done  chan bool
}

// Mutex usage
type Cache struct {
    mu    sync.RWMutex
    items map[string]interface{}
}
```

### Testing

- Table-driven tests for multiple cases
- Use subtests for better organization
- Mock interfaces, not concrete types

```go
func TestCalculate(t *testing.T) {
    tests := []struct {
        name     string
        input    int
        expected int
    }{
        {"positive", 5, 10},
        {"zero", 0, 0},
        {"negative", -5, 10},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Calculate(tt.input)
            if result != tt.expected {
                t.Errorf("Calculate(%d) = %d, want %d", 
                    tt.input, result, tt.expected)
            }
        })
    }
}
```

## Project-Specific Conventions

### Logging

```go
// Use structured logging
log.WithFields(log.Fields{
    "user_id": userID,
    "action":  "login",
}).Info("user logged in successfully")
```

### Configuration

- Use environment variables or config files
- Validate configuration at startup
- Use struct tags for parsing

```go
type Config struct {
    Port     int    `env:"PORT" default:"8080"`
    Database string `env:"DATABASE_URL" required:"true"`
}
```

### HTTP Handlers

```go
// Use http.HandlerFunc signature
func HandleUser(w http.ResponseWriter, r *http.Request) {
    // Set content type
    w.Header().Set("Content-Type", "application/json")
    
    // Handle errors consistently
    if err := processRequest(r); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // Write response
    json.NewEncoder(w).Encode(response)
}
```

## Platform Build Separation

portunix-agent targets multiple platforms (Linux, Windows). Platform-specific code MUST be isolated using Go build tags and file naming conventions so that each platform compiles only the code it needs.

### Build Tag Rules

1. **Always use both** file suffix AND explicit build tag (belt and suspenders):

```go
//go:build linux

/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package accessibility
```

2. **Build tag goes on the very first line**, before the license header and package declaration.

3. **Tag every file that imports or references platform-specific code:**
   - Imports `github.com/godbus/dbus/v5` → `//go:build linux`
   - Imports `github.com/go-ole/go-ole` → `//go:build windows`
   - References platform-specific types (e.g. `dbus.Conn`, `ole.IUnknown`) → tag required

4. **Never tag these files** — they must compile on all platforms:
   - Interface definitions (e.g. `accessibility.go` defining `Backend` interface)
   - Shared types and constants used by all backends
   - Test files that use only mocks/fakes (no real platform dependencies)

### File Naming Convention

```text
package_name.go                  # shared — no tag, compiles everywhere
package_name_linux.go            # Linux only — auto-selected by Go toolchain
package_name_windows.go          # Windows only — auto-selected by Go toolchain
package_name_test.go             # shared tests (mocks only)
```

### Platform Backend Pattern

Each platform-specific subsystem follows a common pattern:

```text
mypackage/
  ├── mypackage.go               # interface definition (no tag)
  ├── backend_foo_linux.go       # //go:build linux — Linux implementation
  ├── backend_bar_windows.go     # //go:build windows — Windows implementation
  ├── detect_linux.go            # //go:build linux — selects Linux backend
  ├── detect_windows.go          # //go:build windows — selects Windows backend
  └── mypackage_test.go          # mock-based tests (no tag)
```

The `Detect()` function in each platform file returns the appropriate backend for that OS.

### Verification

After any change to build tags or platform files, verify cross-compilation:

```bash
# Must succeed on all target platforms
GOOS=linux   go build ./...
GOOS=windows go build ./...
```

### What NOT To Do

- Do not use `runtime.GOOS` checks as a substitute for build tags — the unused platform code still gets compiled and its imports must resolve
- Do not put platform-specific imports behind `if runtime.GOOS == ...` — the compiler resolves all imports regardless of runtime conditions
- Do not leave platform-specific files without build tags — they will break builds on other platforms

## Tools and Automation

### Linting and Formatting

```bash
# Format code
go fmt ./...

# Run linter
golangci-lint run

# Run tests with coverage
go test -cover ./...
```

### Pre-commit Hooks

- Run `gofmt` and `goimports`
- Run linters
- Run tests

### Continuous Integration

- Build for multiple platforms
- Run all tests
- Check code coverage (aim for >80%)
- Run security scanners

## Code Example

```go
// Package user handles user-related operations for portunix-agent.
package user

import (
    "context"
    "fmt"
    "time"
)

// User represents a user in the system.
type User struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"created_at"`
}

// Service provides user-related operations.
type Service interface {
    GetUser(ctx context.Context, id string) (*User, error)
    CreateUser(ctx context.Context, user *User) error
    UpdateUser(ctx context.Context, user *User) error
    DeleteUser(ctx context.Context, id string) error
}

// service implements the Service interface.
type service struct {
    repo Repository
}

// NewService creates a new user service.
func NewService(repo Repository) Service {
    return &service{
        repo: repo,
    }
}

// GetUser retrieves a user by ID.
func (s *service) GetUser(ctx context.Context, id string) (*User, error) {
    if id == "" {
        return nil, fmt.Errorf("user id cannot be empty")
    }
    
    user, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to get user %s: %w", id, err)
    }
    
    return user, nil
}
```

## References

- [Effective Go](https://golang.org/doc/effective_go.html)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Go Proverbs](https://go-proverbs.github.io/)
- Project-specific requirements in CONTRIBUTING.md