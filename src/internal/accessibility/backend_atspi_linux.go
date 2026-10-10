//go:build linux

/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package accessibility

import (
	"fmt"
	"strings"

	"github.com/CassandraGargoyle/portunix-agent/internal/ipc"
)

// AT-SPI2 D-Bus constants
const (
	a11yBusDest  = "org.a11y.Bus"
	a11yBusPath  = "/org/a11y/bus"
	a11yBusIface = "org.a11y.Bus"

	registryDest = "org.a11y.atspi.Registry"
	registryPath = "/org/a11y/atspi/accessible/root"

	ifaceAccessible = "org.a11y.atspi.Accessible"
	ifaceText       = "org.a11y.atspi.Text"
)

// atspiBackend implements Backend using AT-SPI2 over IPC (D-Bus on Linux).
// It depends only on the ipc.Connection interface, never on godbus directly.
type atspiBackend struct {
	conn    ipc.Connection // accessibility bus connection
	sesConn ipc.Connection // session bus connection (kept for cleanup)
}

// NewATSPIBackend creates an AT-SPI2 backend by connecting to the accessibility bus.
// It first connects to the session bus to discover the accessibility bus address,
// then connects to the accessibility bus for all further operations.
func NewATSPIBackend(connector ipc.Connector) (*atspiBackend, error) {
	// Step 1: connect to session bus
	sesConn, err := connector.SessionBus()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to session bus: %w", err)
	}

	// Step 2: get accessibility bus address
	result, err := sesConn.Call(a11yBusDest, a11yBusPath, a11yBusIface, "GetAddress")
	if err != nil {
		sesConn.Close()
		return nil, fmt.Errorf("failed to get accessibility bus address (is AT-SPI2 running?): %w", err)
	}
	if len(result) == 0 {
		sesConn.Close()
		return nil, fmt.Errorf("empty response from org.a11y.Bus.GetAddress")
	}
	address, ok := result[0].(string)
	if !ok || address == "" {
		sesConn.Close()
		return nil, fmt.Errorf("invalid accessibility bus address: %v", result[0])
	}

	// Step 3: connect to accessibility bus
	a11yConn, err := connector.BusAt(address)
	if err != nil {
		sesConn.Close()
		return nil, fmt.Errorf("failed to connect to accessibility bus at %s: %w", address, err)
	}

	return &atspiBackend{conn: a11yConn, sesConn: sesConn}, nil
}

func (b *atspiBackend) Name() string { return "atspi" }

// RootRef returns a reference to the AT-SPI2 registry root (desktop)
func (b *atspiBackend) RootRef() AccessibleRef {
	return ATSPIRef{BusName: registryDest, Path: registryPath}
}

func (b *atspiBackend) Close() error {
	var errs []string
	if err := b.conn.Close(); err != nil {
		errs = append(errs, fmt.Sprintf("a11y bus: %v", err))
	}
	if err := b.sesConn.Close(); err != nil {
		errs = append(errs, fmt.Sprintf("session bus: %v", err))
	}
	if len(errs) > 0 {
		return fmt.Errorf("close errors: %s", strings.Join(errs, "; "))
	}
	return nil
}

// extractRef extracts an accessible reference from IPC call result.
// Handles both flattened (busName, path as separate values) and struct
// (single []interface{}{busName, path}) return formats.
func extractRef(result []interface{}) (ATSPIRef, error) {
	// Flattened format: two separate return values
	if len(result) >= 2 {
		busName, ok1 := result[0].(string)
		path, ok2 := result[1].(string)
		if ok1 && ok2 {
			return ATSPIRef{BusName: busName, Path: path}, nil
		}
	}
	// Struct format: single value containing [busName, path]
	if len(result) >= 1 {
		if arr, ok := result[0].([]interface{}); ok && len(arr) >= 2 {
			busName, ok1 := arr[0].(string)
			path, ok2 := arr[1].(string)
			if ok1 && ok2 {
				return ATSPIRef{BusName: busName, Path: path}, nil
			}
		}
	}
	return ATSPIRef{}, fmt.Errorf("failed to extract accessible reference from %v", result)
}
