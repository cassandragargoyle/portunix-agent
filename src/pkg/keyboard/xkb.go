//go:build linux

/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package keyboard

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
)

// KeyMapping describes how to produce a character via uinput
type KeyMapping struct {
	Keycode   uint16
	Modifiers []uint16
}

// ReverseMap maps Unicode runes to the key sequence needed to type them
type ReverseMap map[rune]KeyMapping

// xkbLayout holds the detected XKB layout and variant
type xkbLayout struct {
	Layout  string
	Variant string
}

// detectXKBLayout detects the active XKB layout using localectl or setxkbmap.
func detectXKBLayout() (xkbLayout, error) {
	// Try localectl first
	if layout, err := detectLayoutLocalectl(); err == nil {
		return layout, nil
	}

	// Fallback to setxkbmap -query
	if layout, err := detectLayoutSetxkbmap(); err == nil {
		return layout, nil
	}

	return xkbLayout{}, fmt.Errorf("cannot detect XKB layout")
}

func detectLayoutLocalectl() (xkbLayout, error) {
	out, err := exec.Command("localectl", "status").Output()
	if err != nil {
		return xkbLayout{}, err
	}

	var result xkbLayout
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if val, ok := strings.CutPrefix(line, "X11 Layout:"); ok {
			result.Layout = strings.TrimSpace(val)
		}
		if val, ok := strings.CutPrefix(line, "X11 Variant:"); ok {
			result.Variant = strings.TrimSpace(val)
		}
	}

	if result.Layout == "" {
		return xkbLayout{}, fmt.Errorf("no X11 Layout found in localectl output")
	}
	return result, nil
}

func detectLayoutSetxkbmap() (xkbLayout, error) {
	out, err := exec.Command("setxkbmap", "-query").Output()
	if err != nil {
		return xkbLayout{}, err
	}

	var result xkbLayout
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if val, ok := strings.CutPrefix(line, "layout:"); ok {
			result.Layout = strings.TrimSpace(val)
		}
		if val, ok := strings.CutPrefix(line, "variant:"); ok {
			result.Variant = strings.TrimSpace(val)
		}
	}

	if result.Layout == "" {
		return xkbLayout{}, fmt.Errorf("no layout found in setxkbmap output")
	}
	return result, nil
}

// buildReverseMap constructs a character-to-keycode map for the given layout.
// It parses XKB keycodes and symbols files, then combines them with the
// keysym-to-rune table to build a reverse lookup.
func buildReverseMap(layout, variant string) (ReverseMap, error) {
	// Parse keycodes from evdev
	keycodesPath := filepath.Join(xkbBasePath, "keycodes", "evdev")
	codes, err := parseKeycodes(keycodesPath)
	if err != nil {
		return nil, fmt.Errorf("parse keycodes: %w", err)
	}

	// Parse symbols for the layout
	symbols, err := parseSymbols(layout, variant)
	if err != nil {
		return nil, fmt.Errorf("parse symbols: %w", err)
	}

	// Build reverse map: for each key and level, map the resulting rune
	// to its keycode + modifiers
	rmap := make(ReverseMap)

	for keyName, entry := range symbols {
		linuxCode, ok := codes[keyName]
		if !ok {
			continue
		}

		for level, keysymName := range entry.Levels {
			if keysymName == "" {
				continue
			}

			ch, ok := resolveKeysym(keysymName)
			if !ok {
				continue
			}

			// Skip if we already have a mapping for this rune at a lower level
			// (prefer simpler key combos)
			if existing, exists := rmap[ch]; exists {
				if len(existing.Modifiers) <= modifiersForLevel(level) {
					continue
				}
			}

			mapping := KeyMapping{
				Keycode:   linuxCode,
				Modifiers: buildModifiers(level),
			}
			rmap[ch] = mapping
		}
	}

	return rmap, nil
}

// modifiersForLevel returns the number of modifier keys for a given level
func modifiersForLevel(level int) int {
	switch level {
	case 0:
		return 0
	case 1:
		return 1 // Shift
	case 2:
		return 1 // AltGr
	case 3:
		return 2 // Shift + AltGr
	default:
		return 0
	}
}

// buildModifiers returns the modifier keycodes for a given XKB level
func buildModifiers(level int) []uint16 {
	switch level {
	case 0:
		return nil
	case 1:
		return []uint16{keyLeftShift}
	case 2:
		return []uint16{keyRightAlt}
	case 3:
		return []uint16{keyLeftShift, keyRightAlt}
	default:
		return nil
	}
}

// newReverseMap detects the layout and builds the reverse map.
// Returns nil on failure (caller should fall back to charToKeys).
func newReverseMap() ReverseMap {
	layout, err := detectXKBLayout()
	if err != nil {
		log.Printf("xkb: layout detection failed: %v", err)
		return nil
	}

	log.Printf("xkb: detected layout=%s variant=%s", layout.Layout, layout.Variant)

	rmap, err := buildReverseMap(layout.Layout, layout.Variant)
	if err != nil {
		log.Printf("xkb: reverse map build failed: %v", err)
		return nil
	}

	log.Printf("xkb: reverse map built with %d entries", len(rmap))
	return rmap
}
