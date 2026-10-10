/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package vscode

import (
	"fmt"
	"testing"

	"github.com/CassandraGargoyle/portunix-agent/internal/accessibility"
)

// mockAccessibility implements accessibility.Backend for testing.
type mockAccessibility struct {
	apps       map[string]*accessibility.Accessible
	children   map[accessibility.AccessibleRef][]accessibility.Accessible
	texts      map[accessibility.AccessibleRef]string
	charCounts map[accessibility.AccessibleRef]int32
	interfaces map[accessibility.AccessibleRef][]string
}

func newMockAccessibility() *mockAccessibility {
	return &mockAccessibility{
		apps:       make(map[string]*accessibility.Accessible),
		children:   make(map[accessibility.AccessibleRef][]accessibility.Accessible),
		texts:      make(map[accessibility.AccessibleRef]string),
		charCounts: make(map[accessibility.AccessibleRef]int32),
		interfaces: make(map[accessibility.AccessibleRef][]string),
	}
}

func (m *mockAccessibility) FindApplication(name string) (*accessibility.Accessible, error) {
	if app, ok := m.apps[name]; ok {
		return app, nil
	}
	return nil, fmt.Errorf("application %q not found", name)
}

func (m *mockAccessibility) Children(ref accessibility.AccessibleRef) ([]accessibility.Accessible, error) {
	if c, ok := m.children[ref]; ok {
		return c, nil
	}
	return nil, nil
}

func (m *mockAccessibility) RoleName(ref accessibility.AccessibleRef) (string, error) {
	return "", nil
}

func (m *mockAccessibility) NodeName(ref accessibility.AccessibleRef) (string, error) {
	return "", nil
}

func (m *mockAccessibility) Text(ref accessibility.AccessibleRef, startOffset, endOffset int32) (string, error) {
	if t, ok := m.texts[ref]; ok {
		return t, nil
	}
	return "", fmt.Errorf("no text for %v", ref)
}

func (m *mockAccessibility) CharacterCount(ref accessibility.AccessibleRef) (int32, error) {
	if c, ok := m.charCounts[ref]; ok {
		return c, nil
	}
	return 0, fmt.Errorf("no count for %v", ref)
}

func (m *mockAccessibility) Interfaces(ref accessibility.AccessibleRef) ([]string, error) {
	if ifaces, ok := m.interfaces[ref]; ok {
		return ifaces, nil
	}
	return nil, nil
}

func (m *mockAccessibility) RootRef() accessibility.AccessibleRef {
	return accessibility.ATSPIRef{BusName: "mock", Path: "/root"}
}

func (m *mockAccessibility) Name() string { return "mock" }
func (m *mockAccessibility) Close() error { return nil }

// buildVSCodeTree sets up a mock accessibility tree resembling VS Code structure
func buildVSCodeTree(mock *mockAccessibility) {
	appRef := accessibility.ATSPIRef{BusName: ":1.100", Path: "/app"}
	frameRef := accessibility.ATSPIRef{BusName: ":1.100", Path: "/frame"}
	panelRef := accessibility.ATSPIRef{BusName: ":1.100", Path: "/panel"}
	docWebRef := accessibility.ATSPIRef{BusName: ":1.100", Path: "/docweb"}
	termPanelRef := accessibility.ATSPIRef{BusName: ":1.100", Path: "/termpanel"}
	listRef := accessibility.ATSPIRef{BusName: ":1.100", Path: "/list"}
	row1Ref := accessibility.ATSPIRef{BusName: ":1.100", Path: "/row1"}
	row2Ref := accessibility.ATSPIRef{BusName: ":1.100", Path: "/row2"}
	row3Ref := accessibility.ATSPIRef{BusName: ":1.100", Path: "/row3"}

	mock.apps["Code"] = &accessibility.Accessible{
		Ref: appRef, Name: "Code", RoleName: "application",
	}

	mock.children[appRef] = []accessibility.Accessible{
		{Ref: frameRef, Name: "main.go - project - VS Code", RoleName: "frame"},
	}
	mock.children[frameRef] = []accessibility.Accessible{
		{Ref: panelRef, Name: "", RoleName: "panel"},
	}
	mock.children[panelRef] = []accessibility.Accessible{
		{Ref: docWebRef, Name: "", RoleName: "document web"},
	}
	mock.children[docWebRef] = []accessibility.Accessible{
		{Ref: termPanelRef, Name: "", RoleName: "panel"},
	}
	mock.children[termPanelRef] = []accessibility.Accessible{
		{Ref: listRef, Name: "xterm-accessibility-tree", RoleName: "list"},
	}
	mock.children[listRef] = []accessibility.Accessible{
		{Ref: row1Ref, Name: "", RoleName: "list item"},
		{Ref: row2Ref, Name: "", RoleName: "list item"},
		{Ref: row3Ref, Name: "", RoleName: "list item"},
	}

	// Terminal rows implement the Text interface
	for _, ref := range []accessibility.AccessibleRef{row1Ref, row2Ref, row3Ref} {
		mock.interfaces[ref] = []string{"org.a11y.atspi.Accessible", "org.a11y.atspi.Text"}
	}
	mock.texts[row1Ref] = "zdenek@dev:~/project$ make build"
	mock.texts[row2Ref] = "go build -o ptx-agent ./src/cmd/ptx-agent/"
	mock.texts[row3Ref] = "zdenek@dev:~/project$ "
}

func TestController_Status_Found(t *testing.T) {
	mock := newMockAccessibility()
	buildVSCodeTree(mock)
	ctrl := NewControllerWithBackend(mock)

	status, err := ctrl.Status()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !status.Found {
		t.Error("expected VS Code to be found")
	}
	if status.AppName != "Code" {
		t.Errorf("expected app name %q, got %q", "Code", status.AppName)
	}
	if status.TerminalCount != 1 {
		t.Errorf("expected 1 terminal, got %d", status.TerminalCount)
	}
}

func TestController_Status_NotFound(t *testing.T) {
	mock := newMockAccessibility()
	ctrl := NewControllerWithBackend(mock)

	status, err := ctrl.Status()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Found {
		t.Error("expected VS Code not to be found")
	}
}

func TestController_ListTerminals(t *testing.T) {
	mock := newMockAccessibility()
	buildVSCodeTree(mock)
	ctrl := NewControllerWithBackend(mock)

	terminals, err := ctrl.ListTerminals()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(terminals) != 1 {
		t.Fatalf("expected 1 terminal, got %d", len(terminals))
	}
	if terminals[0].Index != 1 {
		t.Errorf("expected index 1, got %d", terminals[0].Index)
	}
	if terminals[0].Name != "xterm-accessibility-tree" {
		t.Errorf("expected name %q, got %q", "xterm-accessibility-tree", terminals[0].Name)
	}
}

func TestController_ListTerminals_NoVSCode(t *testing.T) {
	mock := newMockAccessibility()
	ctrl := NewControllerWithBackend(mock)

	_, err := ctrl.ListTerminals()
	if err == nil {
		t.Fatal("expected error when VS Code not found")
	}
}

func TestController_ReadTerminal(t *testing.T) {
	mock := newMockAccessibility()
	buildVSCodeTree(mock)
	ctrl := NewControllerWithBackend(mock)

	content, err := ctrl.ReadTerminal(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "zdenek@dev:~/project$ make build\ngo build -o ptx-agent ./src/cmd/ptx-agent/\nzdenek@dev:~/project$ "
	if content != expected {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, content)
	}
}

func TestController_ReadTerminal_DefaultIndex(t *testing.T) {
	mock := newMockAccessibility()
	buildVSCodeTree(mock)
	ctrl := NewControllerWithBackend(mock)

	// Index 0 should read the first terminal
	content, err := ctrl.ReadTerminal(0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content == "" {
		t.Error("expected non-empty terminal content")
	}
}

func TestController_ReadTerminal_OutOfRange(t *testing.T) {
	mock := newMockAccessibility()
	buildVSCodeTree(mock)
	ctrl := NewControllerWithBackend(mock)

	_, err := ctrl.ReadTerminal(99)
	if err == nil {
		t.Fatal("expected error for out-of-range index")
	}
}

func TestController_BackendName(t *testing.T) {
	mock := newMockAccessibility()
	ctrl := NewControllerWithBackend(mock)

	if ctrl.BackendName() != "mock" {
		t.Errorf("expected %q, got %q", "mock", ctrl.BackendName())
	}
}

func TestController_DumpTree(t *testing.T) {
	mock := newMockAccessibility()
	buildVSCodeTree(mock)
	ctrl := NewControllerWithBackend(mock)

	tree, err := ctrl.DumpTree(5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tree == "" {
		t.Error("expected non-empty tree dump")
	}
	// Tree should contain VS Code's accessible name
	if !contains(tree, "Code") {
		t.Error("expected tree to contain 'Code'")
	}
	// Tree should contain terminal list
	if !contains(tree, "xterm-accessibility-tree") {
		t.Error("expected tree to contain terminal list")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
