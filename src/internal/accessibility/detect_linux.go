//go:build linux

/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package accessibility

import (
	"fmt"

	"github.com/CassandraGargoyle/portunix-agent/internal/ipc"
)

// Detect selects the best available accessibility backend.
// Priority: AT-SPI2 (Linux).
func Detect() (Backend, error) {
	connector, err := ipc.Detect()
	if err != nil {
		return nil, fmt.Errorf("no IPC backend available: %w", err)
	}

	b, err := NewATSPIBackend(connector)
	if err != nil {
		return nil, fmt.Errorf("AT-SPI2 backend failed: %w", err)
	}

	return b, nil
}
