//go:build linux

/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/CassandraGargoyle/portunix-agent/internal/ipc"
)

const (
	a11yBusDest     = "org.a11y.Bus"
	a11yBusPath     = "/org/a11y/bus"
	a11yBusIface    = "org.a11y.Bus"
	registryDest    = "org.a11y.atspi.Registry"
	registryPath    = "/org/a11y/atspi/accessible/root"
	ifaceAccessible = "org.a11y.atspi.Accessible"
	ifaceText       = "org.a11y.atspi.Text"
)

func main() {
	connector := ipc.NewDBusConnector()
	sesConn, err := connector.SessionBus()
	if err != nil {
		fmt.Fprintf(os.Stderr, "session bus: %v\n", err)
		os.Exit(1)
	}
	defer sesConn.Close()

	result, err := sesConn.Call(a11yBusDest, a11yBusPath, a11yBusIface, "GetAddress")
	if err != nil {
		fmt.Fprintf(os.Stderr, "GetAddress: %v\n", err)
		os.Exit(1)
	}
	address := result[0].(string)

	a11yConn, err := connector.BusAt(address)
	if err != nil {
		fmt.Fprintf(os.Stderr, "a11y bus connect: %v\n", err)
		os.Exit(1)
	}
	defer a11yConn.Close()

	// Direct query: VS Code bus :1.17, terminal area starts around node 488
	// Query the terminal content sections from node 488 downward
	bus := findCodeBus(a11yConn)
	if bus == "" {
		fmt.Fprintln(os.Stderr, "VS Code not found")
		os.Exit(1)
	}
	fmt.Printf("VS Code bus: %s\n\n", bus)

	// Dump from node 488 (terminal content area) with deep depth
	fmt.Println("=== Terminal content area (from node 488) ===")
	dumpNode(a11yConn, bus, "/org/a11y/atspi/accessible/488", 0, 15)

	fmt.Println("\n=== Try reading full text from key nodes ===")
	for _, nodeID := range []string{"488", "489", "491", "492", "493", "494"} {
		path := "/org/a11y/atspi/accessible/" + nodeID
		text := readFullText(a11yConn, bus, path)
		if text != "" {
			preview := text
			if len(preview) > 200 {
				preview = preview[:200] + "..."
			}
			fmt.Printf("Node %s text (%d chars): %q\n", nodeID, len(text), preview)
		}
	}
}

func findCodeBus(conn ipc.Connection) string {
	countResult, _ := conn.GetProperty(registryDest, registryPath, ifaceAccessible, "ChildCount")
	count := toInt32(countResult)
	for i := int32(0); i < count; i++ {
		childResult, err := conn.Call(registryDest, registryPath, ifaceAccessible, "GetChildAtIndex", i)
		if err != nil {
			continue
		}
		childBus, childPath := extractRef(childResult)
		nameResult, _ := conn.GetProperty(childBus, childPath, ifaceAccessible, "Name")
		name, _ := nameResult.(string)
		if strings.ToLower(name) == "code" {
			return childBus
		}
	}
	return ""
}

func readFullText(conn ipc.Connection, bus, path string) string {
	// Try GetText(0, -1) first
	textResult, err := conn.Call(bus, path, ifaceText, "GetText", int32(0), int32(-1))
	if err != nil {
		return ""
	}
	if len(textResult) > 0 {
		if t, ok := textResult[0].(string); ok {
			return t
		}
	}
	return ""
}

func dumpNode(conn ipc.Connection, busName, path string, depth, maxDepth int) {
	if depth > maxDepth {
		return
	}
	countResult, err := conn.GetProperty(busName, path, ifaceAccessible, "ChildCount")
	if err != nil {
		return
	}
	count := toInt32(countResult)

	for i := int32(0); i < count; i++ {
		childResult, err := conn.Call(busName, path, ifaceAccessible, "GetChildAtIndex", i)
		if err != nil {
			continue
		}
		childBus, childPath := extractRef(childResult)
		if childBus == "" {
			continue
		}

		nameResult, _ := conn.GetProperty(childBus, childPath, ifaceAccessible, "Name")
		name, _ := nameResult.(string)
		roleResult, _ := conn.Call(childBus, childPath, ifaceAccessible, "GetRoleName")
		role := ""
		if len(roleResult) > 0 {
			role, _ = roleResult[0].(string)
		}

		text := readFullText(conn, childBus, childPath)
		textInfo := ""
		if text != "" {
			preview := text
			if len(preview) > 80 {
				preview = preview[:80] + "..."
			}
			textInfo = fmt.Sprintf(" TEXT=%q", preview)
		}

		cc := childCount(conn, childBus, childPath)
		indent := strings.Repeat("  ", depth)
		fmt.Printf("%s[%s] %q%s (path=%s children=%d)\n", indent, role, name, textInfo, childPath, cc)
		dumpNode(conn, childBus, childPath, depth+1, maxDepth)
	}
}

func childCount(conn ipc.Connection, busName, path string) int32 {
	result, err := conn.GetProperty(busName, path, ifaceAccessible, "ChildCount")
	if err != nil {
		return -1
	}
	return toInt32(result)
}

func extractRef(result []interface{}) (string, string) {
	if len(result) >= 2 {
		b, ok1 := result[0].(string)
		p, ok2 := result[1].(string)
		if ok1 && ok2 {
			return b, p
		}
	}
	if len(result) >= 1 {
		if arr, ok := result[0].([]interface{}); ok && len(arr) >= 2 {
			b, _ := arr[0].(string)
			p, _ := arr[1].(string)
			return b, p
		}
	}
	return "", ""
}

func toInt32(v interface{}) int32 {
	switch val := v.(type) {
	case int32:
		return val
	case int64:
		return int32(val)
	case uint32:
		return int32(val)
	}
	return 0
}
