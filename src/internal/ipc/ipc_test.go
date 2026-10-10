/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package ipc

import (
	"fmt"
	"testing"
)

// mockConnection implements Connection for testing.
type mockConnection struct {
	callResult []interface{}
	callErr    error
	propResult interface{}
	propErr    error
	closed     bool
}

func (m *mockConnection) Call(busName, objectPath, iface, method string, args ...interface{}) ([]interface{}, error) {
	return m.callResult, m.callErr
}

func (m *mockConnection) GetProperty(busName, objectPath, iface, property string) (interface{}, error) {
	return m.propResult, m.propErr
}

func (m *mockConnection) Close() error {
	m.closed = true
	return nil
}

// mockConnector implements Connector for testing.
type mockConnector struct {
	sessionConn Connection
	sessionErr  error
	busConn     Connection
	busErr      error
}

func (m *mockConnector) SessionBus() (Connection, error) {
	return m.sessionConn, m.sessionErr
}

func (m *mockConnector) BusAt(address string) (Connection, error) {
	return m.busConn, m.busErr
}

func (m *mockConnector) Name() string { return "mock" }

func TestMockConnection_Call(t *testing.T) {
	mock := &mockConnection{
		callResult: []interface{}{"hello", int32(42)},
	}

	result, err := mock.Call("dest", "/path", "iface", "Method")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result))
	}
	if result[0] != "hello" {
		t.Errorf("expected %q, got %q", "hello", result[0])
	}
	if result[1] != int32(42) {
		t.Errorf("expected %d, got %v", 42, result[1])
	}
}

func TestMockConnection_CallError(t *testing.T) {
	mock := &mockConnection{
		callErr: fmt.Errorf("connection failed"),
	}

	_, err := mock.Call("dest", "/path", "iface", "Method")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockConnection_GetProperty(t *testing.T) {
	mock := &mockConnection{
		propResult: "test-value",
	}

	result, err := mock.GetProperty("dest", "/path", "iface", "Prop")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "test-value" {
		t.Errorf("expected %q, got %v", "test-value", result)
	}
}

func TestMockConnection_Close(t *testing.T) {
	mock := &mockConnection{}
	if err := mock.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !mock.closed {
		t.Error("expected connection to be closed")
	}
}

func TestMockConnector_SessionBus(t *testing.T) {
	conn := &mockConnection{}
	connector := &mockConnector{sessionConn: conn}

	got, err := connector.SessionBus()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != conn {
		t.Error("expected session connection")
	}
}

func TestMockConnector_SessionBusError(t *testing.T) {
	connector := &mockConnector{
		sessionErr: fmt.Errorf("no session bus"),
	}

	_, err := connector.SessionBus()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockConnector_BusAt(t *testing.T) {
	conn := &mockConnection{}
	connector := &mockConnector{busConn: conn}

	got, err := connector.BusAt("unix:path=/tmp/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != conn {
		t.Error("expected bus connection")
	}
}

func TestMockConnector_Name(t *testing.T) {
	connector := &mockConnector{}
	if connector.Name() != "mock" {
		t.Errorf("expected %q, got %q", "mock", connector.Name())
	}
}

func TestDetect_ReturnsConnector(t *testing.T) {
	connector, err := Detect()
	if err != nil {
		// On non-Linux, detection may fail
		t.Logf("detection failed (expected on non-Linux): %v", err)
		return
	}
	if connector == nil {
		t.Error("expected non-nil connector")
	}
	if connector.Name() != "dbus" {
		t.Errorf("expected %q backend, got %q", "dbus", connector.Name())
	}
}
