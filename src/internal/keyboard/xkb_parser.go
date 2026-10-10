//go:build linux

/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package keyboard

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// xkbBasePath is the root directory for XKB configuration files.
// Overridable via setXKBBasePath for testing.
var xkbBasePath = "/usr/share/X11/xkb"

const (
	maxIncDepth = 10
	x11Offset   = 8 // X11 keycode = Linux keycode + 8
)

// setXKBBasePath overrides the XKB base path (for testing)
func setXKBBasePath(path string) {
	xkbBasePath = path
}

// parseKeycodes parses an XKB keycodes file (e.g. keycodes/evdev) and returns
// a map from symbolic key name (e.g. "AE01") to Linux keycode.
// It also resolves alias lines (e.g. alias <ALGR> = <RALT>).
func parseKeycodes(path string) (map[string]uint16, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open keycodes: %w", err)
	}
	defer f.Close()

	codes := make(map[string]uint16)
	aliases := make(map[string]string) // alias target -> source

	// <NAME> = number;
	reKeycode := regexp.MustCompile(`<(\w+)>\s*=\s*(\d+)\s*;`)
	// alias <NAME> = <NAME>;
	reAlias := regexp.MustCompile(`alias\s+<(\w+)>\s*=\s*<(\w+)>\s*;`)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if m := reKeycode.FindStringSubmatch(line); m != nil {
			x11code, _ := strconv.ParseUint(m[2], 10, 16)
			if x11code >= x11Offset {
				codes[m[1]] = uint16(x11code - x11Offset)
			}
			continue
		}

		if m := reAlias.FindStringSubmatch(line); m != nil {
			aliases[m[1]] = m[2]
		}
	}

	// Resolve aliases
	for alias, target := range aliases {
		if code, ok := codes[target]; ok {
			codes[alias] = code
		}
	}

	return codes, scanner.Err()
}

// symbolEntry holds the keysym names at each level for a key
type symbolEntry struct {
	Levels [4]string // [normal, Shift, AltGr, Shift+AltGr]
}

// parseSymbols parses XKB symbols for a given layout and variant.
// Returns a map from symbolic key name to keysym levels.
// Handles include chains recursively (later definitions override earlier ones).
func parseSymbols(layout, variant string) (map[string]symbolEntry, error) {
	result := make(map[string]symbolEntry)
	visited := make(map[string]bool)

	target := layout
	if variant != "" {
		target = layout + "(" + variant + ")"
	}

	if err := parseSymbolsFile(target, result, visited, 0); err != nil {
		return nil, err
	}

	return result, nil
}

// parseSymbolsFile recursively parses a symbols file/variant.
// target format: "filename" or "filename(variant)"
func parseSymbolsFile(target string, result map[string]symbolEntry, visited map[string]bool, depth int) error {
	if depth > maxIncDepth {
		return fmt.Errorf("include depth exceeded for %s", target)
	}
	if visited[target] {
		return nil // cycle detection
	}
	visited[target] = true

	filename, variant := parseTarget(target)
	path := filepath.Join(xkbBasePath, "symbols", filename)

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open symbols/%s: %w", filename, err)
	}
	defer f.Close()

	// Find the correct section (variant) in the file
	scanner := bufio.NewScanner(f)
	section := findSection(scanner, variant)
	if !section {
		return fmt.Errorf("variant %q not found in symbols/%s", variant, filename)
	}

	// Parse the section content
	braceDepth := 1
	reInclude := regexp.MustCompile(`include\s+"([^"]+)"`)
	reKey := regexp.MustCompile(`key\s+<(\w+)>\s*\{(.+)\}`)

	for scanner.Scan() && braceDepth > 0 {
		line := strings.TrimSpace(scanner.Text())

		// Track brace nesting to know when section ends
		braceDepth += strings.Count(line, "{") - strings.Count(line, "}")
		if braceDepth <= 0 {
			break
		}

		// Skip comments
		if strings.HasPrefix(line, "//") {
			continue
		}

		// Handle includes (process before key definitions so they can be overridden)
		if m := reInclude.FindStringSubmatch(line); m != nil {
			if err := parseSymbolsFile(m[1], result, visited, depth+1); err != nil {
				// Non-fatal: log and continue
				continue
			}
			continue
		}

		// Handle key definitions
		if m := reKey.FindStringSubmatch(line); m != nil {
			keyName := m[1]
			entry := parseKeyLevels(m[2])
			result[keyName] = entry
		}
	}

	return scanner.Err()
}

// parseTarget splits "filename(variant)" into filename and variant.
// "filename" alone returns filename with empty variant (meaning default section).
func parseTarget(target string) (string, string) {
	if idx := strings.Index(target, "("); idx != -1 {
		end := strings.Index(target, ")")
		if end == -1 {
			end = len(target)
		}
		return target[:idx], target[idx+1 : end]
	}
	return target, ""
}

// findSection scans forward in the file until it finds the xkb_symbols section
// matching the given variant. Empty variant matches the default/first section.
func findSection(scanner *bufio.Scanner, variant string) bool {
	reSectionNamed := regexp.MustCompile(`xkb_symbols\s+"(\w[\w-]*)"`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		m := reSectionNamed.FindStringSubmatch(line)
		if m == nil {
			continue
		}

		sectionName := m[1]

		// If variant is empty, match the first section (usually "basic" or "default")
		if variant == "" || sectionName == variant {
			// Scan to the opening brace of this section
			if !strings.Contains(line, "{") {
				for scanner.Scan() {
					if strings.Contains(scanner.Text(), "{") {
						break
					}
				}
			}
			return true
		}
	}
	return false
}

// parseKeyLevels extracts keysym names from a key definition body.
// Input format: "[ sym1, sym2, sym3, sym4 ]" (with possible surrounding text
// like "type[FOUR_LEVEL] = [ sym1, ... ]")
func parseKeyLevels(body string) symbolEntry {
	var entry symbolEntry

	// Find the last [ ... ] pair — the keysym list is always the last bracket group.
	// This skips type[...] annotations.
	start := strings.LastIndex(body, "[")
	end := -1
	if start != -1 {
		end = strings.Index(body[start:], "]")
		if end != -1 {
			end += start
		}
	}
	if start == -1 || end == -1 || end <= start {
		return entry
	}

	inner := body[start+1 : end]
	parts := strings.Split(inner, ",")

	for i, part := range parts {
		if i >= 4 {
			break
		}
		entry.Levels[i] = strings.TrimSpace(part)
	}

	return entry
}
