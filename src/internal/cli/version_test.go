/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/CassandraGargoyle/portunix-agent/internal/version"
)

// runRoot executes rootCmd with the given arguments and returns stdout, stderr and the error
func runRoot(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs(args)
	t.Cleanup(func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
		// Flag values persist between Execute calls on the shared rootCmd
		if f := rootCmd.Flags().Lookup("version"); f != nil {
			_ = f.Value.Set("false")
			f.Changed = false
		}
	})
	err := rootCmd.Execute()
	return stdout.String(), stderr.String(), err
}

func TestVersionOutput(t *testing.T) {
	want := "ptx-agent version " + version.Version + "\n"

	for _, arg := range []string{"--version", "version"} {
		t.Run(arg, func(t *testing.T) {
			stdout, stderr, err := runRoot(t, arg)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if stdout != want {
				t.Errorf("expected stdout %q, got %q", want, stdout)
			}
			if stderr != "" {
				t.Errorf("expected empty stderr, got %q", stderr)
			}
		})
	}
}

func TestVersionDefaultIsDev(t *testing.T) {
	// Tests are built without -ldflags, the same as a plain go build
	if version.Version != "dev" {
		t.Errorf("expected default version dev, got %q", version.Version)
	}
}

func TestVersionSubcommandRejectsArgs(t *testing.T) {
	if _, _, err := runRoot(t, "version", "extra"); err == nil {
		t.Error("expected error for unexpected argument")
	}
}

func TestVersionFlagIsNotPersistent(t *testing.T) {
	stdout, _, err := runRoot(t, "files", "list", ".", "--version")
	if err == nil {
		t.Fatal("expected unknown flag error")
	}
	if !strings.Contains(err.Error(), "unknown flag: --version") {
		t.Errorf("expected unknown flag error, got %v", err)
	}
	if strings.Contains(stdout, "ptx-agent version") {
		t.Errorf("subcommand must not print the version, got %q", stdout)
	}
}

func TestHelpListsVersion(t *testing.T) {
	stdout, _, err := runRoot(t, "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "--version") {
		t.Error("expected --version in the Flags section")
	}
	if !strings.Contains(stdout, "\n  version ") {
		t.Error("expected version in Available Commands")
	}
}
