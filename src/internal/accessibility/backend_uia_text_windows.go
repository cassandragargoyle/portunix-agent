//go:build windows

/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package accessibility

import (
	"fmt"
	"syscall"
	"unsafe"
)

// UIA Pattern IDs
const (
	uiaTextPatternId  = 10014
	uiaValuePatternId = 10002
)

// IUIAutomationTextPattern vtable layout
type iTextPatternVtbl struct {
	QueryInterface         uintptr // 0
	AddRef                 uintptr // 1
	Release                uintptr // 2
	RangeFromPoint         uintptr // 3
	RangeFromChild         uintptr // 4
	GetSelection           uintptr // 5
	GetVisibleRanges       uintptr // 6
	DocumentRange          uintptr // 7
	SupportedTextSelection uintptr // 8
}

// IUIAutomationTextRange vtable layout
type iTextRangeVtbl struct {
	QueryInterface        uintptr // 0
	AddRef                uintptr // 1
	Release               uintptr // 2
	Clone                 uintptr // 3
	Compare               uintptr // 4
	CompareEndpoints      uintptr // 5
	ExpandToEnclosingUnit uintptr // 6
	FindAttribute         uintptr // 7
	FindText              uintptr // 8
	GetAttributeValue     uintptr // 9
	GetBoundingRectangles uintptr // 10
	GetEnclosingElement   uintptr // 11
	GetText               uintptr // 12
}

// IUIAutomationValuePattern vtable layout
type iValuePatternVtbl struct {
	QueryInterface uintptr // 0
	AddRef         uintptr // 1
	Release        uintptr // 2
	SetValue       uintptr // 3
	CurrentValue   uintptr // 4
}

var (
	oleaut32          = syscall.NewLazyDLL("oleaut32.dll")
	procSysFreeString = oleaut32.NewProc("SysFreeString")
)

// Text reads text content from a UIA element.
// Tries UIA TextPattern first (full document text via DocumentRange.GetText),
// then ValuePattern, then falls back to the Name property.
func (b *uiaBackend) Text(ref AccessibleRef, startOffset, endOffset int32) (string, error) {
	elem := toElemPtr(ref)

	// Try TextPattern (e.g. Notepad document, rich text editors)
	if text := readTextPattern(elem); text != "" {
		return applyOffsets(text, startOffset, endOffset), nil
	}

	// Try ValuePattern (e.g. single-line edit controls)
	if text := readValuePattern(elem); text != "" {
		return applyOffsets(text, startOffset, endOffset), nil
	}

	// Fall back to Name property (e.g. xterm.js terminal list items)
	name, err := getElementName(elem)
	if err != nil {
		return "", fmt.Errorf("failed to read text: %w", err)
	}
	return applyOffsets(name, startOffset, endOffset), nil
}

// CharacterCount returns the character count of a UIA element's text content
func (b *uiaBackend) CharacterCount(ref AccessibleRef) (int32, error) {
	elem := toElemPtr(ref)

	if text := readTextPattern(elem); text != "" {
		return int32(len([]rune(text))), nil
	}
	if text := readValuePattern(elem); text != "" {
		return int32(len([]rune(text))), nil
	}

	name, err := getElementName(elem)
	if err != nil {
		return 0, fmt.Errorf("failed to get character count: %w", err)
	}
	return int32(len([]rune(name))), nil
}

// readTextPattern reads text via IUIAutomationTextPattern.DocumentRange.GetText
func readTextPattern(elem elemPtr) string {
	vtbl := getElementVtbl(elem)
	var pattern unsafe.Pointer
	hr, _, _ := syscall.SyscallN(
		vtbl.GetCurrentPattern,
		uintptr(elem),
		uintptr(uiaTextPatternId),
		uintptr(unsafe.Pointer(&pattern)),
	)
	if hr != 0 || pattern == nil {
		return ""
	}
	defer releaseCOMObject(pattern)

	// Get DocumentRange
	tpVtbl := (*iTextPatternVtbl)(*((*unsafe.Pointer)(pattern)))
	var docRange unsafe.Pointer
	hr, _, _ = syscall.SyscallN(
		tpVtbl.DocumentRange,
		uintptr(pattern),
		uintptr(unsafe.Pointer(&docRange)),
	)
	if hr != 0 || docRange == nil {
		return ""
	}
	defer releaseCOMObject(docRange)

	// GetText(-1) to read all text
	trVtbl := (*iTextRangeVtbl)(*((*unsafe.Pointer)(docRange)))
	var bstr unsafe.Pointer
	hr, _, _ = syscall.SyscallN(
		trVtbl.GetText,
		uintptr(docRange),
		uintptr(0xFFFFFFFF), // -1 as uint = read all
		uintptr(unsafe.Pointer(&bstr)),
	)
	if hr != 0 || bstr == nil {
		return ""
	}
	defer procSysFreeString.Call(uintptr(bstr))

	return bstrToString(bstr)
}

// readValuePattern reads text via IUIAutomationValuePattern.CurrentValue
func readValuePattern(elem elemPtr) string {
	vtbl := getElementVtbl(elem)
	var pattern unsafe.Pointer
	hr, _, _ := syscall.SyscallN(
		vtbl.GetCurrentPattern,
		uintptr(elem),
		uintptr(uiaValuePatternId),
		uintptr(unsafe.Pointer(&pattern)),
	)
	if hr != 0 || pattern == nil {
		return ""
	}
	defer releaseCOMObject(pattern)

	vpVtbl := (*iValuePatternVtbl)(*((*unsafe.Pointer)(pattern)))
	var bstr unsafe.Pointer
	hr, _, _ = syscall.SyscallN(
		vpVtbl.CurrentValue,
		uintptr(pattern),
		uintptr(unsafe.Pointer(&bstr)),
	)
	if hr != 0 || bstr == nil {
		return ""
	}
	defer procSysFreeString.Call(uintptr(bstr))

	return bstrToString(bstr)
}

// bstrToString converts a BSTR pointer to a Go string.
// BSTR layout: [uint32 byte-length][wchar_t data...]
// The pointer points to the data; byte-length is at offset -4.
func bstrToString(bstr unsafe.Pointer) string {
	if bstr == nil {
		return ""
	}
	length := *(*uint32)(unsafe.Add(bstr, -4))
	if length == 0 {
		return ""
	}
	wchars := unsafe.Slice((*uint16)(bstr), length/2)
	return syscall.UTF16ToString(wchars)
}

// applyOffsets applies start/end offsets to a text string using rune indexing
func applyOffsets(text string, startOffset, endOffset int32) string {
	if startOffset <= 0 && endOffset < 0 {
		return text
	}
	runes := []rune(text)
	start := int(startOffset)
	if start < 0 {
		start = 0
	}
	end := len(runes)
	if endOffset >= 0 && int(endOffset) < end {
		end = int(endOffset)
	}
	if start >= end {
		return ""
	}
	return string(runes[start:end])
}
