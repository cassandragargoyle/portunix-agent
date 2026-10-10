/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/CassandraGargoyle/portunix-agent/internal/mcp"
	"github.com/spf13/cobra"
)

var (
	mcpAllowedDirs []string
	mcpHTTPAddr    string
	mcpToken       string
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "MCP server for AI assistant integration",
	Long:  "Manage the MCP (Model Context Protocol) server that exposes ptx-agent capabilities to AI assistants.",
}

var mcpServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start MCP server",
	Long:  "Start the MCP server for AI assistant integration.\n\nBy default uses stdio transport (JSON-RPC over stdin/stdout).\nUse --http to start an HTTP server with token authentication.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := mcp.ServerConfig{
			AllowedDirs: mcpAllowedDirs,
		}
		server := mcp.NewServer(cfg)

		if mcpHTTPAddr != "" {
			token := mcpToken
			if token == "" {
				var err error
				token, err = mcp.GenerateToken()
				if err != nil {
					return err
				}
			}
			return server.RunHTTP(context.Background(), mcpHTTPAddr, token)
		}

		return server.Run(context.Background())
	},
}

var mcpConfigureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Register with Claude Code",
	Long:  "Register ptx-agent as an MCP server with Claude Code.",
	RunE: func(cmd *cobra.Command, args []string) error {
		claudePath, err := exec.LookPath("claude")
		if err != nil {
			return fmt.Errorf("Claude Code not found in PATH (is it installed?)")
		}

		agentPath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("failed to determine ptx-agent binary path: %w", err)
		}

		// Build the serve command arguments
		serveArgs := []string{agentPath, "mcp", "serve"}
		if len(mcpAllowedDirs) > 0 {
			serveArgs = append(serveArgs, "--allowed-dirs", strings.Join(mcpAllowedDirs, ","))
		}

		// Register with Claude Code: claude mcp add <name> -- <command> [args...]
		registerArgs := append([]string{"mcp", "add", "ptx-agent", "--"}, serveArgs...)
		register := exec.Command(claudePath, registerArgs...)
		register.Stdout = os.Stdout
		register.Stderr = os.Stderr

		fmt.Printf("Registering: claude %s\n", strings.Join(registerArgs, " "))
		if err := register.Run(); err != nil {
			return fmt.Errorf("failed to register MCP server: %w", err)
		}

		fmt.Println("ptx-agent registered as MCP server with Claude Code.")
		return nil
	},
}

var mcpRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Unregister from Claude Code",
	Long:  "Remove ptx-agent MCP server registration from Claude Code.",
	RunE: func(cmd *cobra.Command, args []string) error {
		claudePath, err := exec.LookPath("claude")
		if err != nil {
			return fmt.Errorf("Claude Code not found in PATH (is it installed?)")
		}

		remove := exec.Command(claudePath, "mcp", "remove", "ptx-agent")
		remove.Stdout = os.Stdout
		remove.Stderr = os.Stderr

		if err := remove.Run(); err != nil {
			return fmt.Errorf("failed to remove MCP server: %w", err)
		}

		fmt.Println("ptx-agent MCP server removed from Claude Code.")
		return nil
	},
}

var mcpStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check MCP registration status",
	Long:  "Check whether ptx-agent is registered as an MCP server with Claude Code.",
	RunE: func(cmd *cobra.Command, args []string) error {
		claudePath, err := exec.LookPath("claude")
		if err != nil {
			fmt.Println("Claude Code: not found in PATH")
			return nil
		}
		fmt.Printf("Claude Code: %s\n", claudePath)

		list := exec.Command(claudePath, "mcp", "list")
		output, err := list.Output()
		if err != nil {
			return fmt.Errorf("failed to list MCP servers: %w", err)
		}

		if strings.Contains(string(output), "ptx-agent") {
			fmt.Println("ptx-agent: registered")
		} else {
			fmt.Println("ptx-agent: not registered")
		}
		return nil
	},
}

func init() {
	mcpServeCmd.Flags().StringSliceVar(&mcpAllowedDirs, "allowed-dirs", nil,
		"comma-separated list of allowed directory paths for file browsing")
	mcpServeCmd.Flags().StringVar(&mcpHTTPAddr, "http", "",
		"listen address for HTTP transport (e.g. :9100)")
	mcpServeCmd.Flags().StringVar(&mcpToken, "token", "",
		"fixed bearer token for HTTP auth (auto-generated if not set)")
	mcpConfigureCmd.Flags().StringSliceVar(&mcpAllowedDirs, "allowed-dirs", nil,
		"comma-separated list of allowed directory paths (forwarded to serve)")

	mcpCmd.AddCommand(mcpServeCmd)
	mcpCmd.AddCommand(mcpConfigureCmd)
	mcpCmd.AddCommand(mcpRemoveCmd)
	mcpCmd.AddCommand(mcpStatusCmd)
}
