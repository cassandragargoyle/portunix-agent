//go:build linux

/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package accessibility

import "testing"

func TestExtractRef_Flat(t *testing.T) {
	result := []interface{}{":1.100", "/org/a11y/atspi/accessible/root"}
	ref, err := extractRef(result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ref.BusName != ":1.100" {
		t.Errorf("expected bus name %q, got %q", ":1.100", ref.BusName)
	}
	if ref.Path != "/org/a11y/atspi/accessible/root" {
		t.Errorf("expected path %q, got %q", "/org/a11y/atspi/accessible/root", ref.Path)
	}
}

func TestExtractRef_Struct(t *testing.T) {
	result := []interface{}{
		[]interface{}{":1.200", "/path/to/node"},
	}
	ref, err := extractRef(result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ref.BusName != ":1.200" {
		t.Errorf("expected bus name %q, got %q", ":1.200", ref.BusName)
	}
}

func TestExtractRef_Invalid(t *testing.T) {
	_, err := extractRef([]interface{}{})
	if err == nil {
		t.Fatal("expected error for empty result")
	}
}
