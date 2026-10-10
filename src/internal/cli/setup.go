/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package cli

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"

	"github.com/spf13/cobra"
)

const udevRule = `KERNEL=="uinput", GROUP="input", MODE="0660"`
const udevRulePath = "/etc/udev/rules.d/99-ptx-agent-uinput.rules"

var uninstall bool

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Configure system for ptx-agent",
	Long: `Set up udev rules and group membership so ptx-agent can simulate
keyboard input without root privileges. Requires sudo.

After running this command, log out and log back in for changes to take effect.

Use --uninstall to reverse all changes made by setup.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if uninstall {
			return runUninstall()
		}
		return runSetup()
	},
}

func runSetup() error {
	u, err := user.Current()
	if err != nil {
		return err
	}

	fmt.Println("Setting up ptx-agent system access...")

	// Check if udev rule already exists
	if _, err := os.Stat(udevRulePath); err == nil {
		fmt.Println("  udev rule already exists, skipping.")
	} else {
		fmt.Printf("  Creating udev rule: %s\n", udevRulePath)
		writeCmd := fmt.Sprintf("echo '%s' > %s", udevRule, udevRulePath)
		if err := exec.Command("sudo", "sh", "-c", writeCmd).Run(); err != nil {
			return fmt.Errorf("failed to create udev rule (need sudo): %w", err)
		}

		fmt.Println("  Reloading udev rules...")
		_ = exec.Command("sudo", "udevadm", "control", "--reload-rules").Run()
		_ = exec.Command("sudo", "udevadm", "trigger").Run()
	}

	// Add user to input group
	fmt.Printf("  Adding user %s to input group...\n", u.Username)
	if err := exec.Command("sudo", "usermod", "-aG", "input", u.Username).Run(); err != nil {
		return fmt.Errorf("failed to add user to input group: %w", err)
	}

	fmt.Println()
	fmt.Println("Setup complete. Please LOG OUT and LOG BACK IN for changes to take effect.")
	return nil
}

func runUninstall() error {
	u, err := user.Current()
	if err != nil {
		return err
	}

	fmt.Println("Removing ptx-agent system configuration...")

	// Remove udev rule
	if _, err := os.Stat(udevRulePath); err == nil {
		fmt.Printf("  Removing udev rule: %s\n", udevRulePath)
		if err := exec.Command("sudo", "rm", udevRulePath).Run(); err != nil {
			return fmt.Errorf("failed to remove udev rule (need sudo): %w", err)
		}

		fmt.Println("  Reloading udev rules...")
		_ = exec.Command("sudo", "udevadm", "control", "--reload-rules").Run()
		_ = exec.Command("sudo", "udevadm", "trigger").Run()
	} else {
		fmt.Println("  udev rule not found, skipping.")
	}

	// Remove user from input group
	fmt.Printf("  Removing user %s from input group...\n", u.Username)
	if err := exec.Command("sudo", "gpasswd", "-d", u.Username, "input").Run(); err != nil {
		return fmt.Errorf("failed to remove user from input group: %w", err)
	}

	fmt.Println()
	fmt.Println("Uninstall complete. Please LOG OUT and LOG BACK IN for changes to take effect.")
	return nil
}

func init() {
	setupCmd.Flags().BoolVar(&uninstall, "uninstall", false,
		"reverse all changes made by setup")
	rootCmd.AddCommand(setupCmd)
}
