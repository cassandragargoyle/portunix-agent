/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package cli

import (
	"fmt"
	"strings"

	"github.com/CassandraGargoyle/portunix-agent/internal/accessibility"
	"github.com/spf13/cobra"
)

var appTreeDepth int
var appListVerbose bool

var appCmd = &cobra.Command{
	Use:   "app",
	Short: "Interact with desktop applications via accessibility",
	Long:  "List, inspect, and read text from desktop applications using the accessibility layer (AT-SPI2 on Linux, UIA on Windows).",
}

var appListCmd = &cobra.Command{
	Use:   "list",
	Short: "List desktop windows",
	Long:  "List all top-level desktop windows visible in the accessibility tree.",
	RunE: func(cmd *cobra.Command, args []string) error {
		backend, err := accessibility.Detect()
		if err != nil {
			return err
		}
		defer backend.Close()

		children, err := backend.Children(backend.RootRef())
		if err != nil {
			return fmt.Errorf("failed to enumerate desktop windows: %w", err)
		}

		if len(children) == 0 {
			fmt.Println("No windows found.")
			return nil
		}

		for _, child := range children {
			fmt.Printf("[%s] %q\n", child.RoleName, child.Name)
			if appListVerbose {
				ifaces, _ := backend.Interfaces(child.Ref)
				if len(ifaces) > 0 {
					fmt.Printf("  Interfaces: %s\n", strings.Join(ifaces, ", "))
				}
				grandchildren, _ := backend.Children(child.Ref)
				fmt.Printf("  Children: %d\n", len(grandchildren))
				fmt.Println()
			}
		}
		return nil
	},
}

var appTreeCmd = &cobra.Command{
	Use:   "tree <name>",
	Short: "Dump accessibility tree of an application",
	Long:  "Dump the accessibility tree of an application matched by name substring.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		backend, err := accessibility.Detect()
		if err != nil {
			return err
		}
		defer backend.Close()

		app, err := backend.FindApplication(args[0])
		if err != nil {
			return err
		}

		fmt.Printf("[%s] %q\n", app.RoleName, app.Name)
		dumpAppSubtree(backend, app.Ref, 1, appTreeDepth)
		return nil
	},
}

var appReadCmd = &cobra.Command{
	Use:   "read <name>",
	Short: "Read text content from an application",
	Long:  "Read text content from an application by walking its accessibility tree and extracting text via TextPattern, ValuePattern, or Name property.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		backend, err := accessibility.Detect()
		if err != nil {
			return err
		}
		defer backend.Close()

		app, err := backend.FindApplication(args[0])
		if err != nil {
			return err
		}

		readAppText(backend, app.Ref, 0, appTreeDepth)
		return nil
	},
}

// dumpAppSubtree recursively prints the accessibility tree
func dumpAppSubtree(backend accessibility.Backend, ref accessibility.AccessibleRef, depth, maxDepth int) {
	if depth > maxDepth {
		return
	}

	children, err := backend.Children(ref)
	if err != nil {
		return
	}

	indent := strings.Repeat("  ", depth)
	for _, child := range children {
		ifaces, _ := backend.Interfaces(child.Ref)
		ifaceStr := ""
		if len(ifaces) > 0 {
			ifaceStr = fmt.Sprintf(" {%s}", strings.Join(ifaces, ", "))
		}
		fmt.Printf("%s[%s] %q%s\n", indent, child.RoleName, child.Name, ifaceStr)
		dumpAppSubtree(backend, child.Ref, depth+1, maxDepth)
	}
}

// readAppText walks the tree and prints text from elements that support Text or Value interfaces
func readAppText(backend accessibility.Backend, ref accessibility.AccessibleRef, depth, maxDepth int) {
	if depth > maxDepth {
		return
	}

	// Try reading text from this element
	ifaces, _ := backend.Interfaces(ref)
	hasText := false
	for _, iface := range ifaces {
		if iface == "Text" || iface == "Value" || iface == "org.a11y.atspi.Text" {
			hasText = true
			break
		}
	}

	if hasText {
		text, err := backend.Text(ref, 0, -1)
		if err == nil && text != "" {
			name, _ := backend.NodeName(ref)
			role, _ := backend.RoleName(ref)
			fmt.Printf("[%s] %q:\n%s\n\n", role, name, text)
			return // don't recurse into children if we got text
		}
	}

	// Recurse into children
	children, err := backend.Children(ref)
	if err != nil {
		return
	}
	for _, child := range children {
		readAppText(backend, child.Ref, depth+1, maxDepth)
	}
}

func init() {
	appListCmd.Flags().BoolVarP(&appListVerbose, "verbose", "v", false, "show interfaces and child count for each window")
	appTreeCmd.Flags().IntVar(&appTreeDepth, "depth", 15, "maximum tree depth")
	appReadCmd.Flags().IntVar(&appTreeDepth, "depth", 15, "maximum tree depth")

	appCmd.AddCommand(appListCmd)
	appCmd.AddCommand(appTreeCmd)
	appCmd.AddCommand(appReadCmd)
}
