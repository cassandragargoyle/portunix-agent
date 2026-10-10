/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

// Package ipc provides an abstraction layer for inter-process communication.
// The MVP backend uses D-Bus (godbus/dbus/v5) on Linux.
// Future backends may include COM (Windows) or named pipes.
package ipc

// Connection represents a connection to an IPC message bus.
// Implementations must convert transport-specific types to Go primitives
// (e.g. D-Bus ObjectPath to string).
type Connection interface {
	// Call invokes a method on a remote object and returns the response values.
	// Transport-specific types are converted to Go primitives.
	Call(busName, objectPath, iface, method string, args ...interface{}) ([]interface{}, error)

	// GetProperty retrieves a property value from a remote object.
	// Transport-specific types are converted to Go primitives.
	GetProperty(busName, objectPath, iface, property string) (interface{}, error)

	// Close releases connection resources
	Close() error
}

// Connector creates IPC connections.
type Connector interface {
	// SessionBus returns a connection to the default session message bus
	SessionBus() (Connection, error)

	// BusAt connects to a message bus at a specific address
	BusAt(address string) (Connection, error)

	// Name returns the backend identifier (e.g. "dbus")
	Name() string
}
