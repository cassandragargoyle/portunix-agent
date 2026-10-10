//go:build linux

/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package windows

import (
	"fmt"
	"os"
)

// writeTempScript writes a KWin script to a temporary file and returns the path.
func writeTempScript(script string) (string, error) {
	f, err := os.CreateTemp("", "ptx-kwin-*.js")
	if err != nil {
		return "", fmt.Errorf("failed to create temp script: %w", err)
	}
	if _, err := f.WriteString(script); err != nil {
		f.Close()
		os.Remove(f.Name())
		return "", fmt.Errorf("failed to write temp script: %w", err)
	}
	f.Close()
	return f.Name(), nil
}

// removeTempFile removes a temporary file, ignoring errors.
func removeTempFile(path string) {
	os.Remove(path)
}
