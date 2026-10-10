/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package accessibility

import (
	"fmt"
	"testing"
)

// mockBackend implements Backend for testing.
type mockBackend struct {
	apps       map[string]*Accessible
	children   map[AccessibleRef][]Accessible
	roles      map[AccessibleRef]string
	names      map[AccessibleRef]string
	texts      map[AccessibleRef]string
	charCounts map[AccessibleRef]int32
	interfaces map[AccessibleRef][]string
}

func newMockBackend() *mockBackend {
	return &mockBackend{
		apps:       make(map[string]*Accessible),
		children:   make(map[AccessibleRef][]Accessible),
		roles:      make(map[AccessibleRef]string),
		names:      make(map[AccessibleRef]string),
		texts:      make(map[AccessibleRef]string),
		charCounts: make(map[AccessibleRef]int32),
		interfaces: make(map[AccessibleRef][]string),
	}
}

func (m *mockBackend) FindApplication(name string) (*Accessible, error) {
	if app, ok := m.apps[name]; ok {
		return app, nil
	}
	return nil, fmt.Errorf("application %q not found", name)
}

func (m *mockBackend) Children(ref AccessibleRef) ([]Accessible, error) {
	if c, ok := m.children[ref]; ok {
		return c, nil
	}
	return nil, nil
}

func (m *mockBackend) RoleName(ref AccessibleRef) (string, error) {
	if r, ok := m.roles[ref]; ok {
		return r, nil
	}
	return "", fmt.Errorf("no role for %v", ref)
}

func (m *mockBackend) NodeName(ref AccessibleRef) (string, error) {
	if n, ok := m.names[ref]; ok {
		return n, nil
	}
	return "", fmt.Errorf("no name for %v", ref)
}

func (m *mockBackend) Text(ref AccessibleRef, startOffset, endOffset int32) (string, error) {
	if t, ok := m.texts[ref]; ok {
		return t, nil
	}
	return "", fmt.Errorf("no text for %v", ref)
}

func (m *mockBackend) CharacterCount(ref AccessibleRef) (int32, error) {
	if c, ok := m.charCounts[ref]; ok {
		return c, nil
	}
	return 0, fmt.Errorf("no character count for %v", ref)
}

func (m *mockBackend) Interfaces(ref AccessibleRef) ([]string, error) {
	if ifaces, ok := m.interfaces[ref]; ok {
		return ifaces, nil
	}
	return nil, nil
}

func (m *mockBackend) Name() string { return "mock" }
func (m *mockBackend) Close() error { return nil }

func TestMockBackend_FindApplication(t *testing.T) {
	mock := newMockBackend()
	ref := ATSPIRef{BusName: ":1.100", Path: "/org/a11y/atspi/accessible/root"}
	mock.apps["Code"] = &Accessible{
		Ref:      ref,
		Name:     "Code",
		RoleName: "application",
	}

	app, err := mock.FindApplication("Code")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if app.Name != "Code" {
		t.Errorf("expected name %q, got %q", "Code", app.Name)
	}
	if app.Ref != ref {
		t.Errorf("expected ref %v, got %v", ref, app.Ref)
	}
}

func TestMockBackend_FindApplicationNotFound(t *testing.T) {
	mock := newMockBackend()

	_, err := mock.FindApplication("NotExist")
	if err == nil {
		t.Fatal("expected error for missing application")
	}
}

func TestMockBackend_Children(t *testing.T) {
	mock := newMockBackend()
	parentRef := ATSPIRef{BusName: ":1.100", Path: "/root"}
	mock.children[parentRef] = []Accessible{
		{Ref: ATSPIRef{BusName: ":1.100", Path: "/child1"}, Name: "child1", RoleName: "frame"},
		{Ref: ATSPIRef{BusName: ":1.100", Path: "/child2"}, Name: "child2", RoleName: "panel"},
	}

	children, err := mock.Children(parentRef)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(children))
	}
	if children[0].Name != "child1" {
		t.Errorf("expected first child %q, got %q", "child1", children[0].Name)
	}
}

func TestMockBackend_Text(t *testing.T) {
	mock := newMockBackend()
	ref := ATSPIRef{BusName: ":1.100", Path: "/terminal/row1"}
	mock.texts[ref] = "$ echo hello"
	mock.charCounts[ref] = 12

	text, err := mock.Text(ref, 0, -1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != "$ echo hello" {
		t.Errorf("expected %q, got %q", "$ echo hello", text)
	}

	count, err := mock.CharacterCount(ref)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 12 {
		t.Errorf("expected character count 12, got %d", count)
	}
}

func TestMockBackend_Interfaces(t *testing.T) {
	mock := newMockBackend()
	ref := ATSPIRef{BusName: ":1.100", Path: "/node"}
	mock.interfaces[ref] = []string{
		"org.a11y.atspi.Accessible",
		"org.a11y.atspi.Text",
	}

	ifaces, err := mock.Interfaces(ref)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ifaces) != 2 {
		t.Fatalf("expected 2 interfaces, got %d", len(ifaces))
	}
}
