//go:build linux

/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package ipc

// Detect selects the best available IPC backend.
// On Linux, D-Bus is used for AT-SPI2 accessibility communication.
func Detect() (Connector, error) {
	return NewDBusConnector(), nil
}
