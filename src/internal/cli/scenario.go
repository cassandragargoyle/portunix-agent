//go:build linux

/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package cli

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/CassandraGargoyle/portunix-agent/internal/scenario"
	"github.com/CassandraGargoyle/portunix-agent/pkg/keyboard"
	"github.com/spf13/cobra"
)

var scenarioWindow string

var scenarioCmd = &cobra.Command{
	Use:   "scenario",
	Short: "Run typing scenarios from JSON files",
	Long:  "Execute automated typing scenarios defined in JSON files.",
}

var scenarioRunCmd = &cobra.Command{
	Use:   "run <file.json>",
	Short: "Run a typing scenario",
	Long: `Execute a typing scenario from a JSON file.

The scenario file defines a sequence of commands to type,
with configurable speed, pauses, and Enter key handling.

Use --window to target a specific window by name or ID.
Use Enter key (or hotkey when waitForHotkey is set) to advance between steps.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := scenario.Load(args[0])
		if err != nil {
			return err
		}

		// Focus target window if specified
		if scenarioWindow != "" {
			if err := focusTargetWindow(scenarioWindow); err != nil {
				return err
			}
			time.Sleep(500 * time.Millisecond)
		}

		ctrl, err := keyboard.NewController()
		if err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "Scenario: %s\n", s.Metadata.Title)
		fmt.Fprintf(os.Stderr, "Commands: %d\n\n", len(s.Commands))

		reader := bufio.NewReader(os.Stdin)

		for i, c := range s.Commands {
			// Show description
			if c.Description != "" {
				fmt.Fprintf(os.Stderr, "[%d/%d] %s\n", i+1, len(s.Commands), c.Description)
			}

			// Wait for hotkey (Enter in terminal) if requested
			if c.WaitForHotkey {
				fmt.Fprintf(os.Stderr, "  Press ENTER to continue...\n")
				reader.ReadString('\n')

				// Re-focus window after user interaction switched to terminal
				if scenarioWindow != "" {
					if err := focusTargetWindow(scenarioWindow); err != nil {
						return err
					}
					time.Sleep(300 * time.Millisecond)
				}
			}

			// Pause before typing
			if c.PauseBefore > 0 {
				time.Sleep(time.Duration(c.PauseBefore) * time.Millisecond)
			}

			// Type the command
			delay := scenario.SpeedToDelay(c.Speed)
			if err := ctrl.Type(c.Command, delay, false); err != nil {
				fmt.Fprintf(os.Stderr, "  ERROR: %v\n", err)
				continue
			}

			// Send Enter if requested
			if c.SendEnter {
				time.Sleep(100 * time.Millisecond)
				if err := ctrl.Key("Return"); err != nil {
					fmt.Fprintf(os.Stderr, "  ERROR sending Enter: %v\n", err)
				}
			}

			fmt.Fprintf(os.Stderr, "  Typed: %s\n", c.Command)
		}

		fmt.Fprintf(os.Stderr, "\nScenario complete.\n")
		return nil
	},
}

func init() {
	scenarioRunCmd.Flags().StringVar(&scenarioWindow, "window", "",
		"focus this window before each command (window ID or name search)")
	scenarioCmd.AddCommand(scenarioRunCmd)
	rootCmd.AddCommand(scenarioCmd)
}
