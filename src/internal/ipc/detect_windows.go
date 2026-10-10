//go:build windows

/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package ipc

import "fmt"

// Detect returns an error on Windows because the IPC layer is not used.
// Windows accessibility uses UI Automation via COM directly, bypassing IPC.
func Detect() (Connector, error) {
	return nil, fmt.Errorf("IPC layer not used on Windows (accessibility uses UIA/COM directly)")
}
