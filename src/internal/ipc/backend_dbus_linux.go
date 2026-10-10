//go:build linux

/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package ipc

import (
	"fmt"

	"github.com/godbus/dbus/v5"
)

// dbusConnector creates D-Bus connections.
type dbusConnector struct{}

// NewDBusConnector returns a Connector backed by D-Bus (godbus/dbus/v5).
// This is the only place in the codebase where godbus is imported.
func NewDBusConnector() Connector {
	return &dbusConnector{}
}

func (c *dbusConnector) SessionBus() (Connection, error) {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to session D-Bus: %w", err)
	}
	return &dbusConnection{conn: conn}, nil
}

func (c *dbusConnector) BusAt(address string) (Connection, error) {
	conn, err := dbus.Connect(address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to D-Bus at %s: %w", address, err)
	}
	return &dbusConnection{conn: conn}, nil
}

func (c *dbusConnector) Name() string { return "dbus" }

// dbusConnection wraps a single D-Bus connection.
type dbusConnection struct {
	conn *dbus.Conn
}

func (c *dbusConnection) Call(busName, objectPath, iface, method string, args ...interface{}) ([]interface{}, error) {
	obj := c.conn.Object(busName, dbus.ObjectPath(objectPath))
	call := obj.Call(iface+"."+method, 0, args...)
	if call.Err != nil {
		return nil, call.Err
	}
	return convertBody(call.Body), nil
}

func (c *dbusConnection) GetProperty(busName, objectPath, iface, property string) (interface{}, error) {
	obj := c.conn.Object(busName, dbus.ObjectPath(objectPath))
	variant, err := obj.GetProperty(iface + "." + property)
	if err != nil {
		return nil, err
	}
	return convertValue(variant.Value()), nil
}

func (c *dbusConnection) Close() error {
	return c.conn.Close()
}

// convertBody converts all D-Bus specific types in a body slice to Go primitives
func convertBody(body []interface{}) []interface{} {
	result := make([]interface{}, len(body))
	for i, v := range body {
		result[i] = convertValue(v)
	}
	return result
}

// convertValue recursively converts D-Bus specific types to Go primitives.
// dbus.ObjectPath → string, dbus.Variant → unwrapped value,
// slices and nested structures are converted recursively.
func convertValue(v interface{}) interface{} {
	switch val := v.(type) {
	case dbus.ObjectPath:
		return string(val)
	case dbus.Variant:
		return convertValue(val.Value())
	case []interface{}:
		for i, elem := range val {
			val[i] = convertValue(elem)
		}
		return val
	case [][]interface{}:
		result := make([]interface{}, len(val))
		for i, inner := range val {
			result[i] = convertBody(inner)
		}
		return result
	case []dbus.ObjectPath:
		result := make([]string, len(val))
		for i, p := range val {
			result[i] = string(p)
		}
		return result
	default:
		return v
	}
}
