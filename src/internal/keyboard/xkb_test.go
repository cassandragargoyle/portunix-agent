//go:build linux

/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package keyboard

import (
	"os"
	"path/filepath"
	"testing"
)

// Embedded test data to avoid dependency on system XKB files

const testKeycodes = `
xkb_keycodes "evdev" {
	minimum = 8;
	maximum = 255;

	<ESC> = 9;
	<AE01> = 10;
	<AE02> = 11;
	<AE03> = 12;
	<AE04> = 13;
	<AE05> = 14;
	<AE06> = 15;
	<AE07> = 16;
	<AE08> = 17;
	<AE09> = 18;
	<AE10> = 19;
	<AE11> = 20;
	<AE12> = 21;

	<AD01> = 24;
	<AD02> = 25;
	<AD03> = 26;
	<AD04> = 27;
	<AD05> = 28;
	<AD06> = 29;
	<AD07> = 30;
	<AD08> = 31;
	<AD09> = 32;
	<AD10> = 33;

	<AC01> = 38;
	<AC02> = 39;
	<AC03> = 40;
	<AC04> = 41;
	<AC05> = 42;
	<AC06> = 43;
	<AC07> = 44;
	<AC08> = 45;
	<AC09> = 46;
	<AC10> = 47;
	<AC11> = 48;

	<AB01> = 52;
	<AB02> = 53;
	<AB03> = 54;
	<AB04> = 55;
	<AB05> = 56;
	<AB06> = 57;
	<AB07> = 58;
	<AB08> = 59;
	<AB09> = 60;
	<AB10> = 61;

	<TLDE> = 49;
	<BKSL> = 51;
	<SPCE> = 65;
	<RTRN> = 36;

	<LALT> = 64;
	<RALT> = 108;

	alias <ALGR> = <RALT>;
};
`

const testSymbolsLatin = `
default partial
xkb_symbols "basic" {
	key <AE01> { [ 1, exclam, onesuperior, exclamdown ] };
	key <AE02> { [ 2, at, twosuperior, oneeighth ] };
	key <AE03> { [ 3, numbersign, threesuperior, sterling ] };
	key <AE04> { [ 4, dollar, onequarter, dollar ] };
	key <AE05> { [ 5, percent, onehalf, threeeighths ] };
	key <AE06> { [ 6, asciicircum, threequarters, fiveeighths ] };
	key <AE07> { [ 7, ampersand, braceleft, seveneighths ] };
	key <AE08> { [ 8, asterisk, bracketleft, trademark ] };
	key <AE09> { [ 9, parenleft, bracketright, plusminus ] };
	key <AE10> { [ 0, parenright, braceright, degree ] };

	key <AD01> { [ q, Q, at, Greek_OMEGA ] };
	key <AD02> { [ w, W, U017F, section ] };
	key <AD03> { [ e, E, e, E ] };
	key <AD04> { [ r, R, paragraph, registered ] };
	key <AD05> { [ t, T, tslash, Tslash ] };
	key <AD06> { [ y, Y, leftarrow, yen ] };
	key <AD07> { [ u, U, downarrow, uparrow ] };
	key <AD08> { [ i, I, rightarrow, idotless ] };
	key <AD09> { [ o, O, oslash, Oslash ] };
	key <AD10> { [ p, P, thorn, THORN ] };

	key <AC01> { [ a, A, ae, AE ] };
	key <AC02> { [ s, S, ssharp, U1E9E ] };
	key <AC03> { [ d, D, eth, ETH ] };
	key <AC04> { [ f, F, dstroke, ordfeminine ] };
	key <AC05> { [ g, G, eng, ENG ] };
	key <AC06> { [ h, H, hstroke, Hstroke ] };
	key <AC07> { [ j, J, dead_hook, dead_horn ] };
	key <AC08> { [ k, K, kra, ampersand ] };
	key <AC09> { [ l, L, lstroke, Lstroke ] };

	key <AB01> { [ z, Z, guillemotleft, less ] };
	key <AB02> { [ x, X, guillemotright, greater ] };
	key <AB03> { [ c, C, cent, copyright ] };
	key <AB04> { [ v, V, doublelowquotemark, singlelowquotemark ] };
	key <AB05> { [ b, B, leftdoublequotemark, leftsinglequotemark ] };
	key <AB06> { [ n, N, rightdoublequotemark, rightsinglequotemark ] };
	key <AB07> { [ m, M, mu, masculine ] };
	key <AB08> { [ comma, less, U2022, multiply ] };
	key <AB09> { [ period, greater, periodcentered, division ] };
	key <AB10> { [ slash, question, dead_belowdot, dead_abovedot ] };

	key <SPCE> { [ space, space, space, nobreakspace ] };
};
`

const testSymbolsCz = `
default partial
xkb_symbols "basic" {
	include "latin(basic)"

	key <AE01> { [ plus, 1, exclam, dead_tilde ] };
	key <AE02> { [ ecaron, 2, at, dead_caron ] };
	key <AE03> { [ scaron, 3, numbersign, dead_circumflex ] };
	key <AE04> { [ ccaron, 4, dollar, dead_breve ] };
	key <AE05> { [ rcaron, 5, percent, dead_abovering ] };
	key <AE06> { [ zcaron, 6, asciicircum, dead_ogonek ] };
	key <AE07> { [ yacute, 7, ampersand, dead_grave ] };
	key <AE08> { [ aacute, 8, asterisk, dead_abovedot ] };
	key <AE09> { [ iacute, 9, parenleft, dead_acute ] };
	key <AE10> { [ eacute, 0, parenright, dead_doubleacute ] };

	key <AD01> { [ q, Q, backslash, Greek_OMEGA ] };
	key <AD06> { [ z, Z, leftarrow, yen ] };
	key <AB01> { [ y, Y, guillemotleft, less ] };

	key <AC10> { [ uacute, quotedbl, dollar, ampersand ] };
	key <AC11> { [ section, exclam, ssharp, U1E9E ] };
};

partial
xkb_symbols "winkeys" {
	include "cz(basic)"

	key <AE11> { [ equal, percent, dead_diaeresis, dead_cedilla ] };
	key <AE12> { [ dead_acute, dead_caron, dead_circumflex, dead_tilde ] };
};
`

// setupTestXKB creates temporary XKB files for testing
func setupTestXKB(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	keycodesDir := filepath.Join(dir, "keycodes")
	symbolsDir := filepath.Join(dir, "symbols")
	if err := os.MkdirAll(keycodesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(symbolsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(keycodesDir, "evdev"), []byte(testKeycodes), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(symbolsDir, "latin"), []byte(testSymbolsLatin), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(symbolsDir, "cz"), []byte(testSymbolsCz), 0o644); err != nil {
		t.Fatal(err)
	}

	return dir
}

func TestParseKeycodes(t *testing.T) {
	dir := setupTestXKB(t)
	path := filepath.Join(dir, "keycodes", "evdev")

	codes, err := parseKeycodes(path)
	if err != nil {
		t.Fatalf("parseKeycodes: %v", err)
	}

	tests := []struct {
		name     string
		keyName  string
		expected uint16
	}{
		{"AE01 digit row", "AE01", 2},
		{"AE02 digit row", "AE02", 3},
		{"AC01 home row", "AC01", 30},
		{"SPCE space bar", "SPCE", 57},
		{"RTRN return", "RTRN", 28},
		{"RALT right alt", "RALT", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := codes[tt.keyName]
			if !ok {
				t.Fatalf("key %s not found in keycodes", tt.keyName)
			}
			if got != tt.expected {
				t.Errorf("key %s: got %d, want %d", tt.keyName, got, tt.expected)
			}
		})
	}

	// Verify alias resolution
	t.Run("ALGR alias", func(t *testing.T) {
		algr, ok := codes["ALGR"]
		if !ok {
			t.Fatal("ALGR alias not resolved")
		}
		ralt := codes["RALT"]
		if algr != ralt {
			t.Errorf("ALGR=%d != RALT=%d", algr, ralt)
		}
	})
}

func TestParseSymbols_LatinBasic(t *testing.T) {
	dir := setupTestXKB(t)
	origBase := xkbBasePath
	setXKBBasePath(dir)
	defer setXKBBasePath(origBase)

	symbols, err := parseSymbols("latin", "")
	if err != nil {
		t.Fatalf("parseSymbols: %v", err)
	}

	tests := []struct {
		name    string
		keyName string
		level   int
		keysym  string
	}{
		{"AE01 level 0", "AE01", 0, "1"},
		{"AE01 level 1 (Shift)", "AE01", 1, "exclam"},
		{"AC01 level 0", "AC01", 0, "a"},
		{"AC01 level 1 (Shift)", "AC01", 1, "A"},
		{"SPCE level 0", "SPCE", 0, "space"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, ok := symbols[tt.keyName]
			if !ok {
				t.Fatalf("key %s not found", tt.keyName)
			}
			if entry.Levels[tt.level] != tt.keysym {
				t.Errorf("got %q, want %q", entry.Levels[tt.level], tt.keysym)
			}
		})
	}
}

func TestParseSymbols_CzechWinkeys(t *testing.T) {
	dir := setupTestXKB(t)
	origBase := xkbBasePath
	setXKBBasePath(dir)
	defer setXKBBasePath(origBase)

	symbols, err := parseSymbols("cz", "winkeys")
	if err != nil {
		t.Fatalf("parseSymbols: %v", err)
	}

	tests := []struct {
		name    string
		keyName string
		level   int
		keysym  string
	}{
		// Czech overrides from cz(basic)
		{"AE02 level 0 (ecaron)", "AE02", 0, "ecaron"},
		{"AE02 level 1 (Shift=2)", "AE02", 1, "2"},
		{"AE02 level 2 (AltGr=@)", "AE02", 2, "at"},
		{"AE03 level 0 (scaron)", "AE03", 0, "scaron"},
		{"AE04 level 0 (ccaron)", "AE04", 0, "ccaron"},

		// winkeys-specific override
		{"AE11 level 0 (equal)", "AE11", 0, "equal"},
		{"AE11 level 1 (percent)", "AE11", 1, "percent"},

		// Inherited from latin(basic) — not overridden by cz
		{"AC01 level 0 (a)", "AC01", 0, "a"},
		{"AC01 level 1 (A)", "AC01", 1, "A"},
		{"AB07 level 0 (m)", "AB07", 0, "m"},

		// Czech z/y swap
		{"AD06 level 0 (z)", "AD06", 0, "z"},
		{"AB01 level 0 (y)", "AB01", 0, "y"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, ok := symbols[tt.keyName]
			if !ok {
				t.Fatalf("key %s not found in symbols", tt.keyName)
			}
			if entry.Levels[tt.level] != tt.keysym {
				t.Errorf("got %q, want %q", entry.Levels[tt.level], tt.keysym)
			}
		})
	}
}

func TestBuildReverseMap(t *testing.T) {
	dir := setupTestXKB(t)
	origBase := xkbBasePath
	setXKBBasePath(dir)
	defer setXKBBasePath(origBase)

	rmap, err := buildReverseMap("cz", "winkeys")
	if err != nil {
		t.Fatalf("buildReverseMap: %v", err)
	}

	tests := []struct {
		name      string
		ch        rune
		keycode   uint16
		modifiers []uint16
	}{
		{"ecaron no mod", '\u011B', 3, nil},               // ě on AE02 level 0
		{"2 with Shift", '2', 3, []uint16{keyLeftShift}},  // 2 on AE02 level 1
		{"at with AltGr", '@', 3, []uint16{keyRightAlt}},  // @ on AE02 level 2
		{"a no mod", 'a', 30, nil},                        // a on AC01 level 0 (from latin)
		{"A with Shift", 'A', 30, []uint16{keyLeftShift}}, // A on AC01 level 1
		{"space", ' ', 57, nil},                           // space on SPCE level 0
		{"scaron no mod", '\u0161', 4, nil},               // š on AE03 level 0
		{"ccaron no mod", '\u010D', 5, nil},               // č on AE04 level 0
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, ok := rmap[tt.ch]
			if !ok {
				t.Fatalf("rune %q (%U) not found in reverse map", string(tt.ch), tt.ch)
			}
			if m.Keycode != tt.keycode {
				t.Errorf("keycode: got %d, want %d", m.Keycode, tt.keycode)
			}
			if len(m.Modifiers) != len(tt.modifiers) {
				t.Fatalf("modifiers count: got %d, want %d", len(m.Modifiers), len(tt.modifiers))
			}
			for i, mod := range tt.modifiers {
				if m.Modifiers[i] != mod {
					t.Errorf("modifier[%d]: got %d, want %d", i, m.Modifiers[i], mod)
				}
			}
		})
	}
}

func TestResolveKeysym(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		want   rune
		wantOK bool
	}{
		{"lowercase letter", "a", 'a', true},
		{"uppercase letter", "A", 'A', true},
		{"digit", "1", '1', true},
		{"symbol name", "exclam", '!', true},
		{"czech ecaron", "ecaron", '\u011B', true},
		{"dead key returns false", "dead_caron", 0, false},
		{"NoSymbol returns false", "NoSymbol", 0, false},
		{"unknown returns false", "nonexistent_keysym", 0, false},
		{"Unicode notation U00E1", "U00E1", '\u00E1', true},
		{"Unicode notation U2022", "U2022", '\u2022', true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := resolveKeysym(tt.input)
			if ok != tt.wantOK {
				t.Errorf("resolveKeysym(%q): ok=%v, want %v", tt.input, ok, tt.wantOK)
			}
			if got != tt.want {
				t.Errorf("resolveKeysym(%q): got %U, want %U", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseTarget(t *testing.T) {
	tests := []struct {
		input       string
		wantFile    string
		wantVariant string
	}{
		{"cz", "cz", ""},
		{"cz(winkeys)", "cz", "winkeys"},
		{"latin(basic)", "latin", "basic"},
		{"us", "us", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			f, v := parseTarget(tt.input)
			if f != tt.wantFile {
				t.Errorf("file: got %q, want %q", f, tt.wantFile)
			}
			if v != tt.wantVariant {
				t.Errorf("variant: got %q, want %q", v, tt.wantVariant)
			}
		})
	}
}

func TestKeysForChar_Fallback(t *testing.T) {
	// With nil reverse map, should fall back to charToKeys
	b := &UinputBackend{reverseMap: nil}

	keys := b.keysForChar('a')
	if keys == nil {
		t.Fatal("expected keys for 'a' from fallback")
	}
	if len(keys) != 1 || keys[0] != keyA {
		t.Errorf("got %v, want [%d]", keys, keyA)
	}

	// Shift+symbol from fallback
	keys = b.keysForChar('!')
	if keys == nil {
		t.Fatal("expected keys for '!' from fallback")
	}
	if len(keys) != 2 || keys[0] != keyLeftShift || keys[1] != key1 {
		t.Errorf("got %v, want [%d, %d]", keys, keyLeftShift, key1)
	}
}

func TestKeysForChar_ReverseMapOverride(t *testing.T) {
	// Reverse map should take priority over fallback
	rmap := ReverseMap{
		'a': {Keycode: 99, Modifiers: nil}, // fake mapping
	}
	b := &UinputBackend{reverseMap: rmap}

	keys := b.keysForChar('a')
	if keys == nil {
		t.Fatal("expected keys for 'a'")
	}
	if len(keys) != 1 || keys[0] != 99 {
		t.Errorf("got %v, want [99]", keys)
	}
}

func TestParseKeyLevels(t *testing.T) {
	tests := []struct {
		name   string
		body   string
		levels [4]string
	}{
		{
			"four levels",
			"[ ecaron, 2, at, dead_caron ]",
			[4]string{"ecaron", "2", "at", "dead_caron"},
		},
		{
			"two levels",
			"[ a, A ]",
			[4]string{"a", "A", "", ""},
		},
		{
			"with surrounding text",
			"type[FOUR_LEVEL] = [ q, Q, backslash, Greek_OMEGA ]",
			[4]string{"q", "Q", "backslash", "Greek_OMEGA"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := parseKeyLevels(tt.body)
			for i := 0; i < 4; i++ {
				if entry.Levels[i] != tt.levels[i] {
					t.Errorf("level[%d]: got %q, want %q", i, entry.Levels[i], tt.levels[i])
				}
			}
		})
	}
}
