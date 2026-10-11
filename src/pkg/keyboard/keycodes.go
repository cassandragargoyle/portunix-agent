//go:build linux

/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package keyboard

import (
	"fmt"
	"strings"
)

// Linux key codes from linux/input-event-codes.h
const (
	keyReserved   uint16 = 0
	keyEsc        uint16 = 1
	key1          uint16 = 2
	key2          uint16 = 3
	key3          uint16 = 4
	key4          uint16 = 5
	key5          uint16 = 6
	key6          uint16 = 7
	key7          uint16 = 8
	key8          uint16 = 9
	key9          uint16 = 10
	key0          uint16 = 11
	keyMinus      uint16 = 12
	keyEqual      uint16 = 13
	keyBackspace  uint16 = 14
	keyTab        uint16 = 15
	keyQ          uint16 = 16
	keyW          uint16 = 17
	keyE          uint16 = 18
	keyR          uint16 = 19
	keyT          uint16 = 20
	keyY          uint16 = 21
	keyU          uint16 = 22
	keyI          uint16 = 23
	keyO          uint16 = 24
	keyP          uint16 = 25
	keyLeftBrace  uint16 = 26
	keyRightBrace uint16 = 27
	keyEnter      uint16 = 28
	keyLeftCtrl   uint16 = 29
	keyA          uint16 = 30
	keyS          uint16 = 31
	keyD          uint16 = 32
	keyF          uint16 = 33
	keyG          uint16 = 34
	keyH          uint16 = 35
	keyJ          uint16 = 36
	keyK          uint16 = 37
	keyL          uint16 = 38
	keySemicolon  uint16 = 39
	keyApostrophe uint16 = 40
	keyGrave      uint16 = 41
	keyLeftShift  uint16 = 42
	keyBackslash  uint16 = 43
	keyZ          uint16 = 44
	keyX          uint16 = 45
	keyC          uint16 = 46
	keyV          uint16 = 47
	keyB          uint16 = 48
	keyN          uint16 = 49
	keyM          uint16 = 50
	keyComma      uint16 = 51
	keyDot        uint16 = 52
	keySlash      uint16 = 53
	keyRightShift uint16 = 54
	keyLeftAlt    uint16 = 56
	keySpace      uint16 = 57
	keyCapsLock   uint16 = 58
	keyF1         uint16 = 59
	keyF2         uint16 = 60
	keyF3         uint16 = 61
	keyF4         uint16 = 62
	keyF5         uint16 = 63
	keyF6         uint16 = 64
	keyF7         uint16 = 65
	keyF8         uint16 = 66
	keyF9         uint16 = 67
	keyF10        uint16 = 68
	keyF11        uint16 = 87
	keyF12        uint16 = 88
	keyRightCtrl  uint16 = 97
	keyRightAlt   uint16 = 100
	keyHome       uint16 = 102
	keyUp         uint16 = 103
	keyPageUp     uint16 = 104
	keyLeft       uint16 = 105
	keyRight      uint16 = 106
	keyEnd        uint16 = 107
	keyDown       uint16 = 108
	keyPageDown   uint16 = 109
	keyInsert     uint16 = 110
	keyDelete     uint16 = 111
	keyLeftMeta   uint16 = 125
	keyRightMeta  uint16 = 126
)

// charToKeys maps a rune to the key sequence needed to type it.
// Returns (keys, ok). keys may include keyLeftShift for uppercase/symbols.
func charToKeys(ch rune) ([]uint16, bool) {
	// Lowercase letters
	if ch >= 'a' && ch <= 'z' {
		return []uint16{keyA + uint16(ch-'a')}, true
	}

	// Uppercase letters
	if ch >= 'A' && ch <= 'Z' {
		return []uint16{keyLeftShift, keyA + uint16(ch-'A')}, true
	}

	// Digits
	if ch >= '1' && ch <= '9' {
		return []uint16{key1 + uint16(ch-'1')}, true
	}
	if ch == '0' {
		return []uint16{key0}, true
	}

	// Common characters (US layout)
	switch ch {
	case ' ':
		return []uint16{keySpace}, true
	case '\n':
		return []uint16{keyEnter}, true
	case '\t':
		return []uint16{keyTab}, true
	case '-':
		return []uint16{keyMinus}, true
	case '=':
		return []uint16{keyEqual}, true
	case '[':
		return []uint16{keyLeftBrace}, true
	case ']':
		return []uint16{keyRightBrace}, true
	case ';':
		return []uint16{keySemicolon}, true
	case '\'':
		return []uint16{keyApostrophe}, true
	case '`':
		return []uint16{keyGrave}, true
	case '\\':
		return []uint16{keyBackslash}, true
	case ',':
		return []uint16{keyComma}, true
	case '.':
		return []uint16{keyDot}, true
	case '/':
		return []uint16{keySlash}, true

	// Shifted symbols (US layout)
	case '!':
		return []uint16{keyLeftShift, key1}, true
	case '@':
		return []uint16{keyLeftShift, key2}, true
	case '#':
		return []uint16{keyLeftShift, key3}, true
	case '$':
		return []uint16{keyLeftShift, key4}, true
	case '%':
		return []uint16{keyLeftShift, key5}, true
	case '^':
		return []uint16{keyLeftShift, key6}, true
	case '&':
		return []uint16{keyLeftShift, key7}, true
	case '*':
		return []uint16{keyLeftShift, key8}, true
	case '(':
		return []uint16{keyLeftShift, key9}, true
	case ')':
		return []uint16{keyLeftShift, key0}, true
	case '_':
		return []uint16{keyLeftShift, keyMinus}, true
	case '+':
		return []uint16{keyLeftShift, keyEqual}, true
	case '{':
		return []uint16{keyLeftShift, keyLeftBrace}, true
	case '}':
		return []uint16{keyLeftShift, keyRightBrace}, true
	case ':':
		return []uint16{keyLeftShift, keySemicolon}, true
	case '"':
		return []uint16{keyLeftShift, keyApostrophe}, true
	case '~':
		return []uint16{keyLeftShift, keyGrave}, true
	case '|':
		return []uint16{keyLeftShift, keyBackslash}, true
	case '<':
		return []uint16{keyLeftShift, keyComma}, true
	case '>':
		return []uint16{keyLeftShift, keyDot}, true
	case '?':
		return []uint16{keyLeftShift, keySlash}, true
	}

	return nil, false
}

// comboToKeys parses a key combination string like "ctrl+shift+a" into keycodes.
func comboToKeys(combo string) ([]uint16, error) {
	parts := strings.Split(strings.ToLower(combo), "+")
	keys := make([]uint16, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		k, ok := comboKeyMap[part]
		if !ok {
			return nil, fmt.Errorf("unknown key %q in combination %q", part, combo)
		}
		keys = append(keys, k)
	}

	return keys, nil
}

var comboKeyMap = map[string]uint16{
	"ctrl":      keyLeftCtrl,
	"alt":       keyLeftAlt,
	"shift":     keyLeftShift,
	"super":     keyLeftMeta,
	"return":    keyEnter,
	"enter":     keyEnter,
	"escape":    keyEsc,
	"esc":       keyEsc,
	"tab":       keyTab,
	"backspace": keyBackspace,
	"delete":    keyDelete,
	"del":       keyDelete,
	"insert":    keyInsert,
	"home":      keyHome,
	"end":       keyEnd,
	"pageup":    keyPageUp,
	"pagedown":  keyPageDown,
	"up":        keyUp,
	"down":      keyDown,
	"left":      keyLeft,
	"right":     keyRight,
	"space":     keySpace,
	"f1":        keyF1,
	"f2":        keyF2,
	"f3":        keyF3,
	"f4":        keyF4,
	"f5":        keyF5,
	"f6":        keyF6,
	"f7":        keyF7,
	"f8":        keyF8,
	"f9":        keyF9,
	"f10":       keyF10,
	"f11":       keyF11,
	"f12":       keyF12,
	"a":         keyA,
	"b":         keyB,
	"c":         keyC,
	"d":         keyD,
	"e":         keyE,
	"f":         keyF,
	"g":         keyG,
	"h":         keyH,
	"i":         keyI,
	"j":         keyJ,
	"k":         keyK,
	"l":         keyL,
	"m":         keyM,
	"n":         keyN,
	"o":         keyO,
	"p":         keyP,
	"q":         keyQ,
	"r":         keyR,
	"s":         keyS,
	"t":         keyT,
	"u":         keyU,
	"v":         keyV,
	"w":         keyW,
	"x":         keyX,
	"y":         keyY,
	"z":         keyZ,
	"0":         key0,
	"1":         key1,
	"2":         key2,
	"3":         key3,
	"4":         key4,
	"5":         key5,
	"6":         key6,
	"7":         key7,
	"8":         key8,
	"9":         key9,
	"-":         keyMinus,
	"=":         keyEqual,
}
