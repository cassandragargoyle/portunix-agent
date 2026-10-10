/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package cli

import (
	"fmt"

	"github.com/CassandraGargoyle/portunix-agent/internal/version"
	"github.com/spf13/cobra"
)

// versionTemplate is shared by the --version flag and the version subcommand
const versionTemplate = "ptx-agent version %s\n"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the ptx-agent version",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := fmt.Fprintf(cmd.OutOrStdout(), versionTemplate, version.Version)
		return err
	},
}
