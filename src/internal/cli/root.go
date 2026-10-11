/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

// Package cli defines the cobra command tree for ptx-agent.
package cli

import (
	"fmt"
	"os"

	"github.com/CassandraGargoyle/portunix-agent/internal/version"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ptx-agent",
	Short: "Portunix agent for remote device control",
	Long: `ptx-agent is a helper binary for the Portunix system that enables
remote device control on machines where it is installed.

It provides application switching, keyboard input simulation,
and file browsing capabilities within configured security boundaries.`,
	Version: version.Version,
}

func init() {
	// Define the flag here so cobra does not add the -v shorthand
	rootCmd.Flags().Bool("version", false, "Print the ptx-agent version")
	rootCmd.SetVersionTemplate(fmt.Sprintf(versionTemplate, "{{.Version}}"))
	// Execute handles the flag before cobra validates the arguments
	rootCmd.PersistentFlags().Bool(helpAIFlag, false, "Show machine-readable help in JSON format")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(filesCmd)
	rootCmd.AddCommand(vscodeCmd)
	rootCmd.AddCommand(appCmd)
	rootCmd.AddCommand(mcpCmd)
	rootCmd.AddCommand(browserCmd)
}

// Execute runs the root command.
func Execute() error {
	if args := os.Args[1:]; helpAIRequested(args) && helpAIConfirmed(args) {
		return printHelpAI(os.Stdout)
	}
	return rootCmd.Execute()
}
