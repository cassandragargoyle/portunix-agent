# portunix-agent Build Automation
ifeq ($(OS),Windows_NT)
    SHELL := cmd
endif

.PHONY: all build build-debug clean dist fmt help lint test test-coverage vet

BINARY_NAME := ptx-agent
SRC_DIR := src
DIST_DIR := dist

# Detect OS and set executable extension and commands
ifeq ($(OS),Windows_NT)
    EXE_EXT := .exe
    RM := del /f /q
    RMDIR := rmdir /s /q
else
    EXE_EXT :=
    RM := rm -f
    RMDIR := rm -rf
endif

BINARY := $(BINARY_NAME)$(EXE_EXT)

# Read the version from plugin.json with make functions only (works with SHELL := cmd)
# Expects the form "version": "X.Y.Z" and falls back to dev when it is not found
VERSION := $(firstword $(subst ", ,$(patsubst PTXVER=%,%,$(firstword $(filter PTXVER=%,$(subst "version": ",PTXVER=,$(file < plugin.json)))))))
ifeq ($(VERSION),)
    VERSION := dev
endif
VERSION_PKG := github.com/CassandraGargoyle/portunix-agent/internal/version
LDFLAGS := -X $(VERSION_PKG).Version=$(VERSION)

# Default target
help:
	@echo portunix-agent Build Commands
	@echo ================================
	@echo.
	@echo   build             Build the ptx-agent binary
	@echo   build-debug       Build the UIA debug tool (Windows only)
	@echo   clean             Clean build artifacts
	@echo   dist              Cross-compile and package for all platforms
	@echo   fmt               Format code
	@echo   lint              Run all linters (fmt + vet)
	@echo   test              Run all tests
	@echo   test-coverage     Run tests with coverage report
	@echo   vet               Run go vet
	@echo.
	@echo Usage: make [target]

all: build

build:
	@echo Building $(BINARY) version $(VERSION)...
	cd $(SRC_DIR) && go build -ldflags "$(LDFLAGS)" -o ../$(BINARY) ./cmd/ptx-agent/
	@echo Built $(BINARY)

build-debug:
	@echo Building uia-debug$(EXE_EXT)...
	cd $(SRC_DIR) && go build -o ../uia-debug$(EXE_EXT) ./cmd/uia-debug/
	@echo Built uia-debug$(EXE_EXT)

dist:
	uv run scripts/make-release.py

clean:
	@echo Cleaning up...
	-$(RM) $(BINARY)
	-$(RM) uia-debug$(EXE_EXT)
	-$(RMDIR) $(DIST_DIR)
	cd $(SRC_DIR) && go clean -testcache
	@echo Cleanup complete

fmt:
	@echo Formatting code...
	cd $(SRC_DIR) && go fmt ./...

lint: fmt vet
	@echo Lint passed

test:
	@echo Running tests...
	cd $(SRC_DIR) && go test ./...

test-coverage:
	@echo Running tests with coverage...
	cd $(SRC_DIR) && go test -race -coverprofile=coverage.out -covermode=atomic ./...
	cd $(SRC_DIR) && go tool cover -func=coverage.out
	-$(RM) $(SRC_DIR)\coverage.out

vet:
	@echo Running go vet...
	cd $(SRC_DIR) && go vet ./...
