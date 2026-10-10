/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

// Package version holds the ptx-agent version shared by the CLI and the MCP server.
package version

// Version is set at build time with
// -ldflags "-X github.com/CassandraGargoyle/portunix-agent/internal/version.Version=<version>"
var Version = "dev"
