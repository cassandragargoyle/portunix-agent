//go:build linux

/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package keyboard

// keysymToRune maps XKB keysym names to Unicode runes.
// Dead keys and NoSymbol map to 0 (ignored during reverse map building).
var keysymToRune = map[string]rune{
	// Special
	"NoSymbol": 0,
	"any":      0,

	// Dead keys (compose sequences, not direct characters)
	"dead_acute":       0,
	"dead_grave":       0,
	"dead_circumflex":  0,
	"dead_tilde":       0,
	"dead_diaeresis":   0,
	"dead_caron":       0,
	"dead_breve":       0,
	"dead_ogonek":      0,
	"dead_cedilla":     0,
	"dead_macron":      0,
	"dead_abovering":   0,
	"dead_abovedot":    0,
	"dead_doubleacute": 0,
	"dead_belowdot":    0,
	"dead_hook":        0,
	"dead_horn":        0,

	// ASCII letters (lowercase)
	"a": 'a', "b": 'b', "c": 'c', "d": 'd', "e": 'e',
	"f": 'f', "g": 'g', "h": 'h', "i": 'i', "j": 'j',
	"k": 'k', "l": 'l', "m": 'm', "n": 'n', "o": 'o',
	"p": 'p', "q": 'q', "r": 'r', "s": 's', "t": 't',
	"u": 'u', "v": 'v', "w": 'w', "x": 'x', "y": 'y',
	"z": 'z',

	// ASCII letters (uppercase)
	"A": 'A', "B": 'B', "C": 'C', "D": 'D', "E": 'E',
	"F": 'F', "G": 'G', "H": 'H', "I": 'I', "J": 'J',
	"K": 'K', "L": 'L', "M": 'M', "N": 'N', "O": 'O',
	"P": 'P', "Q": 'Q', "R": 'R', "S": 'S', "T": 'T',
	"U": 'U', "V": 'V', "W": 'W', "X": 'X', "Y": 'Y',
	"Z": 'Z',

	// Digits
	"0": '0', "1": '1', "2": '2', "3": '3', "4": '4',
	"5": '5', "6": '6', "7": '7', "8": '8', "9": '9',

	// Whitespace and control
	"space":     ' ',
	"Return":    '\n',
	"Tab":       '\t',
	"BackSpace": '\b',

	// ASCII punctuation and symbols
	"exclam":       '!',
	"at":           '@',
	"numbersign":   '#',
	"dollar":       '$',
	"percent":      '%',
	"asciicircum":  '^',
	"ampersand":    '&',
	"asterisk":     '*',
	"parenleft":    '(',
	"parenright":   ')',
	"minus":        '-',
	"underscore":   '_',
	"equal":        '=',
	"plus":         '+',
	"bracketleft":  '[',
	"bracketright": ']',
	"braceleft":    '{',
	"braceright":   '}',
	"backslash":    '\\',
	"bar":          '|',
	"semicolon":    ';',
	"colon":        ':',
	"apostrophe":   '\'',
	"quotedbl":     '"',
	"grave":        '`',
	"asciitilde":   '~',
	"comma":        ',',
	"period":       '.',
	"slash":        '/',
	"question":     '?',
	"less":         '<',
	"greater":      '>',

	// Czech accented characters (lowercase)
	"ecaron": '\u011B', // ě
	"scaron": '\u0161', // š
	"ccaron": '\u010D', // č
	"rcaron": '\u0159', // ř
	"zcaron": '\u017E', // ž
	"yacute": '\u00FD', // ý
	"aacute": '\u00E1', // á
	"iacute": '\u00ED', // í
	"eacute": '\u00E9', // é
	"uacute": '\u00FA', // ú
	"uring":  '\u016F', // ů
	"dcaron": '\u010F', // ď
	"tcaron": '\u0165', // ť
	"ncaron": '\u0148', // ň
	"oacute": '\u00F3', // ó

	// Czech accented characters (uppercase)
	"Ecaron": '\u011A', // Ě
	"Scaron": '\u0160', // Š
	"Ccaron": '\u010C', // Č
	"Rcaron": '\u0158', // Ř
	"Zcaron": '\u017D', // Ž
	"Yacute": '\u00DD', // Ý
	"Aacute": '\u00C1', // Á
	"Iacute": '\u00CD', // Í
	"Eacute": '\u00C9', // É
	"Uacute": '\u00DA', // Ú
	"Uring":  '\u016E', // Ů
	"Dcaron": '\u010E', // Ď
	"Tcaron": '\u0164', // Ť
	"Ncaron": '\u0147', // Ň
	"Oacute": '\u00D3', // Ó

	// Common European characters
	"EuroSign":    '\u20AC', // €
	"section":     '\u00A7', // §
	"degree":      '\u00B0', // °
	"ssharp":      '\u00DF', // ß
	"mu":          '\u00B5', // µ
	"sterling":    '\u00A3', // £
	"cent":        '\u00A2', // ¢
	"yen":         '\u00A5', // ¥
	"copyright":   '\u00A9', // ©
	"registered":  '\u00AE', // ®
	"trademark":   '\u2122', // ™
	"notsign":     '\u00AC', // ¬
	"brokenbar":   '\u00A6', // ¦
	"currency":    '\u00A4', // ¤
	"paragraph":   '\u00B6', // ¶
	"ordfeminine": '\u00AA', // ª
	"masculine":   '\u00BA', // º

	// French/Spanish/Portuguese accented
	"agrave":      '\u00E0', // à
	"egrave":      '\u00E8', // è
	"igrave":      '\u00EC', // ì
	"ograve":      '\u00F2', // ò
	"ugrave":      '\u00F9', // ù
	"Agrave":      '\u00C0', // À
	"Egrave":      '\u00C8', // È
	"Igrave":      '\u00CC', // Ì
	"Ograve":      '\u00D2', // Ò
	"Ugrave":      '\u00D9', // Ù
	"acircumflex": '\u00E2', // â
	"ecircumflex": '\u00EA', // ê
	"icircumflex": '\u00EE', // î
	"ocircumflex": '\u00F4', // ô
	"ucircumflex": '\u00FB', // û
	"Acircumflex": '\u00C2', // Â
	"Ecircumflex": '\u00CA', // Ê
	"Icircumflex": '\u00CE', // Î
	"Ocircumflex": '\u00D4', // Ô
	"Ucircumflex": '\u00DB', // Û
	"atilde":      '\u00E3', // ã
	"ntilde":      '\u00F1', // ñ
	"otilde":      '\u00F5', // õ
	"Atilde":      '\u00C3', // Ã
	"Ntilde":      '\u00D1', // Ñ
	"Otilde":      '\u00D5', // Õ
	"ccedilla":    '\u00E7', // ç
	"Ccedilla":    '\u00C7', // Ç

	// German
	"adiaeresis": '\u00E4', // ä
	"odiaeresis": '\u00F6', // ö
	"udiaeresis": '\u00FC', // ü
	"Adiaeresis": '\u00C4', // Ä
	"Odiaeresis": '\u00D6', // Ö
	"Udiaeresis": '\u00DC', // Ü

	// Nordic
	"oslash": '\u00F8', // ø
	"Oslash": '\u00D8', // Ø
	"aring":  '\u00E5', // å
	"Aring":  '\u00C5', // Å
	"ae":     '\u00E6', // æ
	"AE":     '\u00C6', // Æ
	"eth":    '\u00F0', // ð
	"ETH":    '\u00D0', // Ð
	"thorn":  '\u00FE', // þ
	"THORN":  '\u00DE', // Þ

	// Polish
	"aogonek":   '\u0105', // ą
	"Aogonek":   '\u0104', // Ą
	"eogonek":   '\u0119', // ę
	"Eogonek":   '\u0118', // Ę
	"lstroke":   '\u0142', // ł
	"Lstroke":   '\u0141', // Ł
	"zabovedot": '\u017C', // ż
	"Zabovedot": '\u017B', // Ż
	"sacute":    '\u015B', // ś
	"Sacute":    '\u015A', // Ś
	"zacute":    '\u017A', // ź
	"Zacute":    '\u0179', // Ź
	"cacute":    '\u0107', // ć
	"Cacute":    '\u0106', // Ć
	"nacute":    '\u0144', // ń
	"Nacute":    '\u0143', // Ń

	// Hungarian
	"lacute":       '\u013A', // ĺ
	"Lacute":       '\u0139', // Ĺ
	"racute":       '\u0155', // ŕ
	"Racute":       '\u0154', // Ŕ
	"lcaron":       '\u013E', // ľ
	"Lcaron":       '\u013D', // Ľ
	"odoubleacute": '\u0151', // ő
	"Odoubleacute": '\u0150', // Ő
	"udoubleacute": '\u0171', // ű
	"Udoubleacute": '\u0170', // Ű
	"ediaeresis":   '\u00EB', // ë
	"Ediaeresis":   '\u00CB', // Ë
	"eabovedot":    '\u0117', // ė
	"Eabovedot":    '\u0116', // Ė

	// Croatian/Slovenian
	"dstroke": '\u0111', // đ
	"Dstroke": '\u0110', // Đ

	// Misc Latin extended
	"hstroke":  '\u0127', // ħ
	"Hstroke":  '\u0126', // Ħ
	"tslash":   '\u0167', // ŧ
	"Tslash":   '\u0166', // Ŧ
	"eng":      '\u014B', // ŋ
	"ENG":      '\u014A', // Ŋ
	"kra":      '\u0138', // ĸ
	"idotless": '\u0131', // ı

	// Math and special symbols
	"multiply":       '\u00D7', // ×
	"division":       '\u00F7', // ÷
	"plusminus":      '\u00B1', // ±
	"onesuperior":    '\u00B9', // ¹
	"twosuperior":    '\u00B2', // ²
	"threesuperior":  '\u00B3', // ³
	"onequarter":     '\u00BC', // ¼
	"onehalf":        '\u00BD', // ½
	"threequarters":  '\u00BE', // ¾
	"exclamdown":     '\u00A1', // ¡
	"questiondown":   '\u00BF', // ¿
	"nobreakspace":   '\u00A0', // non-breaking space
	"periodcentered": '\u00B7', // ·
	"oneeighth":      '\u215B', // ⅛
	"threeeighths":   '\u215C', // ⅜
	"fiveeighths":    '\u215D', // ⅝
	"seveneighths":   '\u215E', // ⅞
	"numerosign":     '\u2116', // №

	// Typographic quotes and dashes
	"guillemotleft":        '\u00AB', // «
	"guillemotright":       '\u00BB', // »
	"singlelowquotemark":   '\u201A', // ‚
	"doublelowquotemark":   '\u201E', // „
	"leftsinglequotemark":  '\u2018', // '
	"rightsinglequotemark": '\u2019', // '
	"leftdoublequotemark":  '\u201C', // "
	"rightdoublequotemark": '\u201D', // "
	"endash":               '\u2013', // –
	"emdash":               '\u2014', // —
	"ellipsis":             '\u2026', // …
	"enfilledcircbullet":   '\u2022', // •
	"dagger":               '\u2020', // †
	"doubledagger":         '\u2021', // ‡
	"minutes":              '\u2032', // ′
	"seconds":              '\u2033', // ″

	// Arrows
	"leftarrow":  '\u2190', // ←
	"uparrow":    '\u2191', // ↑
	"rightarrow": '\u2192', // →
	"downarrow":  '\u2193', // ↓

	// Greek (used in some layouts)
	"Greek_OMEGA": '\u03A9', // Ω

	// Misc symbols from XKB files
	"lessthanequal":     '\u2264', // ≤
	"greaterthanequal":  '\u2265', // ≥
	"notequal":          '\u2260', // ≠
	"squareroot":        '\u221A', // √
	"permille":          '\u2030', // ‰
	"partialderivative": '\u2202', // ∂
}

// resolveKeysym converts a keysym name to a Unicode rune.
// Returns (rune, true) for known keysyms, (0, false) for unknown ones.
// Single-character keysyms (like "a", "1") are resolved directly.
// Unicode code point notation (UXXXX) is also handled.
func resolveKeysym(name string) (rune, bool) {
	if r, ok := keysymToRune[name]; ok {
		return r, r != 0
	}

	// Handle UXXXX unicode notation (e.g. "U2212" -> '−')
	if len(name) >= 2 && name[0] == 'U' {
		var codepoint rune
		valid := true
		for _, c := range name[1:] {
			codepoint <<= 4
			switch {
			case c >= '0' && c <= '9':
				codepoint |= rune(c - '0')
			case c >= 'a' && c <= 'f':
				codepoint |= rune(c-'a') + 10
			case c >= 'A' && c <= 'F':
				codepoint |= rune(c-'A') + 10
			default:
				valid = false
			}
		}
		if valid && codepoint > 0 {
			return codepoint, true
		}
	}

	return 0, false
}
