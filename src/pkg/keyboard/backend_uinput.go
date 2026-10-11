//go:build linux

/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package keyboard

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
)

// Linux input event constants
const (
	evSyn = 0x00
	evKey = 0x01

	synReport = 0

	uiSetEvbit   = 0x40045564
	uiSetKeybit  = 0x40045565
	uiDevCreate  = 0x5501
	uiDevDestroy = 0x5502
)

// inputEvent matches struct input_event from linux/input.h
type inputEvent struct {
	Time  unix.Timeval
	Type  uint16
	Code  uint16
	Value int32
}

// uinputUserDev matches struct uinput_user_dev from linux/uinput.h
type uinputUserDev struct {
	Name         [80]byte
	ID           inputID
	FFEffectsMax uint32
	Absmax       [64]int32
	Absmin       [64]int32
	Absfuzz      [64]int32
	Absflat      [64]int32
}

type inputID struct {
	Bustype uint16
	Vendor  uint16
	Product uint16
	Version uint16
}

// UinputBackend implements Backend using /dev/uinput (works on Wayland and X11).
type UinputBackend struct {
	reverseMap ReverseMap
}

// NewUinputBackend creates a new uinput backend.
// It detects the active XKB layout and builds a reverse keymap for
// layout-aware character input. Falls back to US layout on failure.
func NewUinputBackend() (*UinputBackend, error) {
	f, err := os.OpenFile("/dev/uinput", os.O_WRONLY|unix.O_NONBLOCK, 0)
	if err != nil {
		return nil, fmt.Errorf("cannot open /dev/uinput: %w", err)
	}
	f.Close()

	rmap := newReverseMap()
	return &UinputBackend{reverseMap: rmap}, nil
}

func (b *UinputBackend) Name() string { return "uinput" }

func (b *UinputBackend) Type(text string, delayMs int, noIndent bool) error {
	dev, err := b.createDevice()
	if err != nil {
		return err
	}
	defer dev.destroy()

	return b.typeOn(dev, text, delayMs, noIndent)
}

// OpenSession creates the virtual device once and keeps it open until
// the returned session is closed
func (b *UinputBackend) OpenSession() (Session, error) {
	dev, err := b.createDevice()
	if err != nil {
		return nil, err
	}
	return &uinputSession{backend: b, dev: dev}, nil
}

// uinputSession sends all actions through one virtual device.
type uinputSession struct {
	backend *UinputBackend
	dev     *uinputDevice
}

func (s *uinputSession) Type(text string, delayMs int, noIndent bool) error {
	return s.backend.typeOn(s.dev, text, delayMs, noIndent)
}

func (s *uinputSession) Key(combo string) error {
	return s.backend.keyOn(s.dev, combo)
}

func (s *uinputSession) Close() error {
	s.dev.destroy()
	return nil
}

// typeOn types the text on an already created device.
func (b *UinputBackend) typeOn(dev *uinputDevice, text string, delayMs int, noIndent bool) error {
	for _, ch := range text {
		keys := b.keysForChar(ch)
		if keys == nil {
			continue // skip unmapped characters
		}

		if err := dev.typeKeys(keys); err != nil {
			return fmt.Errorf("failed to type character %q: %w", string(ch), err)
		}

		// After newline, send Home to reset cursor position so that
		// editor auto-indent does not stack with typed whitespace
		if ch == '\n' && noIndent {
			time.Sleep(time.Duration(delayMs) * time.Millisecond)
			if err := dev.typeKeys([]uint16{keyHome}); err != nil {
				return fmt.Errorf("failed to send Home after newline: %w", err)
			}
		}

		time.Sleep(time.Duration(delayMs) * time.Millisecond)
	}
	return nil
}

// keysForChar resolves a character to keycodes using the XKB reverse map
// first, then falls back to the hardcoded US layout mapping.
func (b *UinputBackend) keysForChar(ch rune) []uint16 {
	if b.reverseMap != nil {
		if m, ok := b.reverseMap[ch]; ok {
			keys := make([]uint16, 0, len(m.Modifiers)+1)
			keys = append(keys, m.Modifiers...)
			keys = append(keys, m.Keycode)
			return keys
		}
	}

	if keys, ok := charToKeys(ch); ok {
		return keys
	}
	return nil
}

func (b *UinputBackend) Key(combo string) error {
	dev, err := b.createDevice()
	if err != nil {
		return err
	}
	defer dev.destroy()

	return b.keyOn(dev, combo)
}

// keyOn sends the key combination on an already created device.
func (b *UinputBackend) keyOn(dev *uinputDevice, combo string) error {
	keys, err := comboToKeys(combo)
	if err != nil {
		return err
	}

	// Press all keys
	for _, k := range keys {
		if err := dev.emit(evKey, k, 1); err != nil {
			return err
		}
		if err := dev.emit(evSyn, synReport, 0); err != nil {
			return err
		}
	}

	time.Sleep(20 * time.Millisecond)

	// Release all keys in reverse order
	for i := len(keys) - 1; i >= 0; i-- {
		if err := dev.emit(evKey, keys[i], 0); err != nil {
			return err
		}
		if err := dev.emit(evSyn, synReport, 0); err != nil {
			return err
		}
	}

	return nil
}

// uinputDevice wraps the /dev/uinput file descriptor.
type uinputDevice struct {
	f *os.File
}

func (b *UinputBackend) createDevice() (*uinputDevice, error) {
	f, err := os.OpenFile("/dev/uinput", os.O_WRONLY|unix.O_NONBLOCK, 0)
	if err != nil {
		return nil, fmt.Errorf("cannot open /dev/uinput: %w", err)
	}

	fd := int(f.Fd())

	// Enable EV_KEY and EV_SYN
	if err := unix.IoctlSetInt(fd, uiSetEvbit, evKey); err != nil {
		f.Close()
		return nil, fmt.Errorf("failed to set EV_KEY: %w", err)
	}
	if err := unix.IoctlSetInt(fd, uiSetEvbit, evSyn); err != nil {
		f.Close()
		return nil, fmt.Errorf("failed to set EV_SYN: %w", err)
	}

	// Enable all keys we might need (0-255 covers standard keyboard)
	for k := 0; k < 256; k++ {
		_ = unix.IoctlSetInt(fd, uiSetKeybit, k)
	}

	// Write device description
	var uidev uinputUserDev
	copy(uidev.Name[:], []byte("ptx-agent virtual keyboard"))
	uidev.ID = inputID{
		Bustype: 0x03, // BUS_USB
		Vendor:  0x1234,
		Product: 0x5678,
		Version: 1,
	}

	devBytes := (*[unsafe.Sizeof(uidev)]byte)(unsafe.Pointer(&uidev))[:]
	if _, err := f.Write(devBytes); err != nil {
		f.Close()
		return nil, fmt.Errorf("failed to write device setup: %w", err)
	}

	// Create device
	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, f.Fd(), uiDevCreate, 0); errno != 0 {
		f.Close()
		return nil, fmt.Errorf("failed to create uinput device: %w", errno)
	}

	// Let the device appear in the system
	time.Sleep(200 * time.Millisecond)

	return &uinputDevice{f: f}, nil
}

func (d *uinputDevice) emit(typ, code uint16, value int32) error {
	ev := inputEvent{
		Type:  typ,
		Code:  code,
		Value: value,
	}
	var buf bytes.Buffer
	if err := binary.Write(&buf, binary.LittleEndian, ev); err != nil {
		return err
	}
	_, err := d.f.Write(buf.Bytes())
	return err
}

// typeKeys presses and releases a set of keys (handles Shift etc).
func (d *uinputDevice) typeKeys(keys []uint16) error {
	// Press all modifier + key
	for _, k := range keys {
		if err := d.emit(evKey, k, 1); err != nil {
			return err
		}
		if err := d.emit(evSyn, synReport, 0); err != nil {
			return err
		}
	}

	time.Sleep(10 * time.Millisecond)

	// Release in reverse order
	for i := len(keys) - 1; i >= 0; i-- {
		if err := d.emit(evKey, keys[i], 0); err != nil {
			return err
		}
		if err := d.emit(evSyn, synReport, 0); err != nil {
			return err
		}
	}

	return nil
}

func (d *uinputDevice) destroy() {
	unix.Syscall(unix.SYS_IOCTL, d.f.Fd(), uiDevDestroy, 0)
	d.f.Close()
}
