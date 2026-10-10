/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package cli

import (
	"fmt"

	"github.com/CassandraGargoyle/portunix-agent/internal/vscode"
	"github.com/spf13/cobra"
)

var terminalIndex int

var vscodeCmd = &cobra.Command{
	Use:   "vscode",
	Short: "Control VS Code via accessibility",
	Long:  "Read and interact with Visual Studio Code through the accessibility layer (AT-SPI2 on Linux, UIA on Windows).",
}

var vscodeTerminalCmd = &cobra.Command{
	Use:   "terminal",
	Short: "VS Code terminal operations",
	Long:  "Read content from VS Code integrated terminal instances.",
}

var vscodeTerminalReadCmd = &cobra.Command{
	Use:   "read",
	Short: "Read terminal content",
	Long:  "Read text content from the VS Code integrated terminal.",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl, err := vscode.NewController()
		if err != nil {
			return err
		}
		defer ctrl.Close()

		content, err := ctrl.ReadTerminal(terminalIndex)
		if err != nil {
			return err
		}

		fmt.Println(content)
		return nil
	},
}

var vscodeTerminalListCmd = &cobra.Command{
	Use:   "list",
	Short: "List open terminals",
	Long:  "List all open terminal instances in VS Code.",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl, err := vscode.NewController()
		if err != nil {
			return err
		}
		defer ctrl.Close()

		terminals, err := ctrl.ListTerminals()
		if err != nil {
			return err
		}

		if len(terminals) == 0 {
			fmt.Println("No terminals found.")
			return nil
		}

		for _, t := range terminals {
			fmt.Printf("Terminal %d: %q\n", t.Index, t.Name)
		}
		return nil
	},
}

var vscodeStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show VS Code accessibility status",
	Long:  "Show diagnostics about VS Code accessibility integration.",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl, err := vscode.NewController()
		if err != nil {
			return fmt.Errorf("accessibility not available: %w", err)
		}
		defer ctrl.Close()

		status, err := ctrl.Status()
		if err != nil {
			return err
		}

		if !status.Found {
			fmt.Println("VS Code: not found")
			fmt.Println("Hint: ensure VS Code is running")
			return nil
		}

		fmt.Printf("VS Code: found (%s)\n", status.AppName)
		fmt.Printf("Accessibility: enabled\n")
		fmt.Printf("Terminal count: %d\n", status.TerminalCount)
		return nil
	},
}

var vscodeTreeCmd = &cobra.Command{
	Use:   "tree",
	Short: "Dump accessibility tree (debug)",
	Long:  "Dump the VS Code accessibility tree for debugging purposes.",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl, err := vscode.NewController()
		if err != nil {
			return err
		}
		defer ctrl.Close()

		tree, err := ctrl.DumpTree(10)
		if err != nil {
			return err
		}

		fmt.Print(tree)
		return nil
	},
}

func init() {
	vscodeTerminalReadCmd.Flags().IntVar(&terminalIndex, "index", 0,
		"terminal index (1-based, 0 for active/first terminal)")

	vscodeTerminalCmd.AddCommand(vscodeTerminalReadCmd)
	vscodeTerminalCmd.AddCommand(vscodeTerminalListCmd)

	vscodeCmd.AddCommand(vscodeTerminalCmd)
	vscodeCmd.AddCommand(vscodeStatusCmd)
	vscodeCmd.AddCommand(vscodeTreeCmd)
}
