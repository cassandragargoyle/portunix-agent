/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

// Package cli defines the cobra command tree for ptx-agent.
package cli

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ptx-agent",
	Short: "Portunix agent for remote device control",
	Long: `ptx-agent is a helper binary for the Portunix system that enables
remote device control on machines where it is installed.

It provides application switching, keyboard input simulation,
and file browsing capabilities within configured security boundaries.`,
}

func init() {
	rootCmd.AddCommand(filesCmd)
	rootCmd.AddCommand(vscodeCmd)
	rootCmd.AddCommand(appCmd)
	rootCmd.AddCommand(mcpCmd)
	rootCmd.AddCommand(browserCmd)
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
