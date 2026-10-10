//go:build linux

/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package cli

import (
	"fmt"

	"github.com/CassandraGargoyle/portunix-agent/internal/windows"
	"github.com/spf13/cobra"
)

var windowsCmd = &cobra.Command{
	Use:   "windows",
	Short: "Manage application windows",
	Long:  "List, search, and switch focus between application windows on the desktop.",
}

var windowsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List open windows",
	Long:  "List all visible windows on the current desktop.",
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, err := windows.NewManager()
		if err != nil {
			return err
		}

		wins, err := mgr.List()
		if err != nil {
			return err
		}

		if len(wins) == 0 {
			fmt.Println("No windows found.")
			return nil
		}

		printWindowTable(wins)
		return nil
	},
}

var windowsFindCmd = &cobra.Command{
	Use:   "find <query>",
	Short: "Search for a window by name",
	Long:  "Search for visible windows matching the given name query.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, err := windows.NewManager()
		if err != nil {
			return err
		}

		wins, err := mgr.Find(args[0])
		if err != nil {
			return err
		}

		if len(wins) == 0 {
			fmt.Printf("No windows matching %q found.\n", args[0])
			return nil
		}

		printWindowTable(wins)
		return nil
	},
}

var windowsFocusCmd = &cobra.Command{
	Use:   "focus <window-id>",
	Short: "Switch focus to a window",
	Long:  "Activate and bring the specified window to the foreground.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, err := windows.NewManager()
		if err != nil {
			return err
		}

		if err := mgr.Focus(args[0]); err != nil {
			return err
		}

		fmt.Printf("Focused window %s.\n", args[0])
		return nil
	},
}

func printWindowTable(wins []windows.Window) {
	fmt.Printf("%-40s %-30s %s\n", "ID", "NAME", "CLASS")
	for _, w := range wins {
		fmt.Printf("%-40s %-30s %s\n", w.ID, w.Name, w.Class)
	}
}

func init() {
	windowsCmd.AddCommand(windowsListCmd)
	windowsCmd.AddCommand(windowsFindCmd)
	windowsCmd.AddCommand(windowsFocusCmd)
	rootCmd.AddCommand(windowsCmd)
}
