# Issue #16: uinput backend drops characters that need a dead key without an error

**Status**: ✅ Implemented
**Priority**: Medium
**Type**: Bug
**Created**: 2026-10-11
**GitHub**: #16
**Related**: #14, portunix-archon #1

## Summary

With the uinput backend, a character that is not in the XKB reverse keymap is skipped without
an error. On the `cz` layout the characters that need a dead key (`ť`, `ň`, `ď`, `ó`) are lost
and the caller gets no error.

## Problem Description

Found during the acceptance test of portunix-archon issue #1 (Ubuntu 26.04, X11 session, layout
`cz`, variant `winkeys`, uinput backend).

Input text:

```text
Příliš žluťoučký kůň úpěl ďábelské ódy
```

Text received by the focused window:

```text
Příliš žluoučký ků úpěl ábelské dy
```

- The characters `ť`, `ň`, `ď`, `ó` are missing. On the `cz` layout they have no key of their
  own. A user types them with a dead key (`dead_caron` or `dead_acute`) and a base letter.
- Characters with their own key (`ř`, `í`, `š`, `ž`, `č`, `ý`, `ů`, `ú`, `ě`, `á`, `é`) are correct.
- `Type` returns `nil`. The caller reports 38 typed characters, but only 34 were sent.
- The xdotool backend types the same text correctly.

Cause:

- `typeOn` in `src/pkg/keyboard/backend_uinput.go` skips a character when `keysForChar` returns
  `nil` (`continue // skip unmapped characters`).
- `src/pkg/keyboard/xkb_keysyms.go` maps all dead keys to 0, so the reverse keymap never
  contains a dead key sequence.

The same silent loss happens for every character that the active layout cannot produce, for
example an emoji or a letter of another alphabet.

## Proposed Solution

Two steps. The first one is the minimum.

1. **Do not lose input without an error.** When a character has no mapping, `Type` returns an
   error that names the character and its position. Check the whole text before the first key
   is sent, so that a text is typed completely or not at all.
2. **Type dead key sequences.** Build the reverse keymap also for characters that the layout
   produces with a dead key and a base letter (`dead_caron` + `t` gives `ť`). For characters
   that the layout cannot produce at all, consider a fallback (clipboard paste, as `Paste`
   does today).

## Acceptance Criteria

1. `Type` with a character that has no mapping returns an error that names the character. No
   key is sent in that case.
2. On the `cz` layout the text `Příliš žluťoučký kůň úpěl ďábelské ódy` is typed completely with
   the uinput backend.
3. The behavior is the same for `Controller.Type` and `Session.Type`.
4. Unit tests cover an unmapped character and a dead key sequence with a `cz` reverse keymap.
5. Typing of characters with a direct mapping does not change.

## Testing

- **Unit tests**: reverse keymap built from a `cz` layout fixture; a text with dead key
  characters; a text with a character that the layout cannot produce.
- **Manual test on Linux**: layout `cz`, uinput backend, type the text above into a focused
  window and compare the result. `xev` can be used as the receiver.
- uinput needs write access to `/dev/uinput`.

---

**Closed**: 2026-10-11 (implemented in `cf147d5`)
