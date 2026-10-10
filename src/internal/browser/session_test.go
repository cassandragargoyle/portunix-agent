/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package browser

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestProfileDirValid(t *testing.T) {
	dir, err := ProfileDir("default")
	if err != nil {
		t.Fatalf("ProfileDir(default) returned error: %v", err)
	}
	want := filepath.Join("ptx-agent", "browser-profiles", "default")
	if !strings.HasSuffix(dir, want) {
		t.Errorf("ProfileDir(default) = %q, want suffix %q", dir, want)
	}
	if !filepath.IsAbs(dir) {
		t.Errorf("ProfileDir(default) = %q, want absolute path", dir)
	}
}

func TestProfileDirInvalid(t *testing.T) {
	cases := []string{
		"",
		".",
		"..",
		".hidden",
		"../escape",
		"a/b",
		`a\b`,
		"/abs",
	}
	for _, name := range cases {
		if _, err := ProfileDir(name); err == nil {
			t.Errorf("ProfileDir(%q) succeeded, want error", name)
		}
	}
}

func TestFingerprintArgsHeaded(t *testing.T) {
	args, ignore := fingerprintArgs(false)
	if len(args) != 1 || args[0] != "--disable-blink-features=AutomationControlled" {
		t.Errorf("fingerprintArgs(false) args = %v, want AutomationControlled flag", args)
	}
	if len(ignore) != 1 || ignore[0] != "--enable-automation" {
		t.Errorf("fingerprintArgs(false) ignoreDefaultArgs = %v, want --enable-automation", ignore)
	}
}

func TestFingerprintArgsHeadless(t *testing.T) {
	args, ignore := fingerprintArgs(true)
	if args != nil || ignore != nil {
		t.Errorf("fingerprintArgs(true) = (%v, %v), want (nil, nil) — no masking in headless mode", args, ignore)
	}
}
