/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package files

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewBrowser_NoAllowedDirs(t *testing.T) {
	_, err := NewBrowser([]string{})
	if err == nil {
		t.Fatal("expected error for empty allowed dirs")
	}
}

func TestNewBrowser_NilAllowedDirs(t *testing.T) {
	_, err := NewBrowser(nil)
	if err == nil {
		t.Fatal("expected error for nil allowed dirs")
	}
}

func TestNewBrowser_ValidDirs(t *testing.T) {
	tmpDir := t.TempDir()
	b, err := NewBrowser([]string{tmpDir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b == nil {
		t.Fatal("expected non-nil browser")
	}
}

func TestList_AllowedDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("hello"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("world"), 0644)
	os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755)

	b, err := NewBrowser([]string{tmpDir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries, err := b.List(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}

	// Check that subdir is marked as directory
	found := false
	for _, e := range entries {
		if e.Name == "subdir" {
			found = true
			if !e.IsDir {
				t.Error("expected subdir to be a directory")
			}
		}
	}
	if !found {
		t.Error("subdir not found in listing")
	}
}

func TestList_DeniedDir(t *testing.T) {
	tmpDir := t.TempDir()
	otherDir := t.TempDir()

	b, err := NewBrowser([]string{tmpDir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = b.List(otherDir)
	if err == nil {
		t.Fatal("expected access denied error")
	}
}

func TestList_EmptyPath(t *testing.T) {
	tmpDir := t.TempDir()
	b, err := NewBrowser([]string{tmpDir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = b.List("")
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestList_TraversalAttempt(t *testing.T) {
	tmpDir := t.TempDir()
	b, err := NewBrowser([]string{tmpDir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Attempt directory traversal
	_, err = b.List(filepath.Join(tmpDir, "..", ".."))
	if err == nil {
		t.Fatal("expected access denied error for traversal attempt")
	}
}

func TestInfo_AllowedFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("test content"), 0644)

	b, err := NewBrowser([]string{tmpDir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	info, err := b.Info(testFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.Name != "test.txt" {
		t.Errorf("expected name 'test.txt', got %q", info.Name)
	}
	if info.Size != 12 {
		t.Errorf("expected size 12, got %d", info.Size)
	}
	if info.IsDir {
		t.Error("expected file, not directory")
	}
}

func TestInfo_DeniedFile(t *testing.T) {
	tmpDir := t.TempDir()
	otherDir := t.TempDir()
	testFile := filepath.Join(otherDir, "secret.txt")
	os.WriteFile(testFile, []byte("secret"), 0644)

	b, err := NewBrowser([]string{tmpDir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = b.Info(testFile)
	if err == nil {
		t.Fatal("expected access denied error")
	}
}

func TestInfo_NonexistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	b, err := NewBrowser([]string{tmpDir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = b.Info(filepath.Join(tmpDir, "nonexistent.txt"))
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestIsAllowed(t *testing.T) {
	tmpDir := t.TempDir()
	otherDir := t.TempDir()

	os.WriteFile(filepath.Join(tmpDir, "allowed.txt"), []byte("ok"), 0644)

	b, err := NewBrowser([]string{tmpDir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		allowed bool
	}{
		{"allowed dir", tmpDir, true},
		{"allowed file", filepath.Join(tmpDir, "allowed.txt"), true},
		{"denied dir", otherDir, false},
		{"root", "/", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := b.IsAllowed(tt.path)
			if got != tt.allowed {
				t.Errorf("IsAllowed(%q) = %v, want %v", tt.path, got, tt.allowed)
			}
		})
	}
}

func TestList_Subdirectory(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "sub")
	os.Mkdir(subDir, 0755)
	os.WriteFile(filepath.Join(subDir, "nested.txt"), []byte("nested"), 0644)

	b, err := NewBrowser([]string{tmpDir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries, err := b.List(subDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Name != "nested.txt" {
		t.Errorf("expected 'nested.txt', got %q", entries[0].Name)
	}
}

func TestIsSubPath(t *testing.T) {
	// Use filepath.Join to produce OS-native paths on both Linux and Windows
	base := filepath.Join(string(filepath.Separator), "tmp", "foo")
	tests := []struct {
		name     string
		child    string
		parent   string
		expected bool
	}{
		{"direct child", filepath.Join(base, "bar"), base, true},
		{"nested child", filepath.Join(base, "bar", "baz"), base, true},
		{"not child", filepath.Join(string(filepath.Separator), "tmp", "bar"), base, false},
		{"prefix collision", filepath.Join(string(filepath.Separator), "tmp", "foobar"), base, false},
		{"same path", base, base, false}, // same path handled by equality check in validatePath
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isSubPath(tt.child, tt.parent)
			if got != tt.expected {
				t.Errorf("isSubPath(%q, %q) = %v, want %v", tt.child, tt.parent, got, tt.expected)
			}
		})
	}
}
