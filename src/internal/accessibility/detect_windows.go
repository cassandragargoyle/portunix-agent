//go:build windows

/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package accessibility

import "fmt"

// Detect selects the best available accessibility backend.
// On Windows, UI Automation (UIA) via COM is used.
func Detect() (Backend, error) {
	b, err := NewUIABackend()
	if err != nil {
		return nil, fmt.Errorf("UIA backend failed: %w", err)
	}
	return b, nil
}
