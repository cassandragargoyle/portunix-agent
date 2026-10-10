//go:build windows

/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"syscall"
	"unsafe"

	"github.com/go-ole/go-ole"
)

var (
	clsidCUIAutomation = ole.NewGUID("{FF48DBA4-60EF-4201-AA87-54103EEF594E}")
	iidIUIAutomation   = ole.NewGUID("{30CBE57D-D9D0-452A-AB13-7AC5AC4825EE}")
)

type iUIAutomationVtbl struct {
	QueryInterface       uintptr // 0
	AddRef               uintptr // 1
	Release              uintptr // 2
	CompareElements      uintptr // 3
	CompareRuntimeIds    uintptr // 4
	GetRootElement       uintptr // 5
	ElementFromHandle    uintptr // 6
	ElementFromPoint     uintptr // 7
	GetFocusedElement    uintptr // 8
	CreateTreeWalker     uintptr // 9
	ControlViewWalker    uintptr // 10
	ContentViewWalker    uintptr // 11
	RawViewWalker        uintptr // 12
	RawViewCondition     uintptr // 13
	ControlViewCondition uintptr // 14
	ContentViewCondition uintptr // 15
	CreateCacheRequest   uintptr // 16
	CreateTrueCondition  uintptr // 17
}

type iElementVtbl struct {
	QueryInterface            uintptr // 0
	AddRef                    uintptr // 1
	Release                   uintptr // 2
	SetFocus                  uintptr // 3
	GetRuntimeId              uintptr // 4
	FindFirst                 uintptr // 5
	FindAll                   uintptr // 6
	FindFirstBuildCache       uintptr // 7
	FindAllBuildCache         uintptr // 8
	BuildUpdatedCache         uintptr // 9
	GetCurrentPropertyValue   uintptr // 10
	GetCurrentPropertyValueEx uintptr // 11
	GetCachedPropertyValue    uintptr // 12
	GetCachedPropertyValueEx  uintptr // 13
	GetCurrentPatternAs       uintptr // 14
	GetCachedPatternAs        uintptr // 15
	GetCurrentPattern         uintptr // 16
	GetCachedPattern          uintptr // 17
}

type iElementArrayVtbl struct {
	QueryInterface uintptr // 0
	AddRef         uintptr // 1
	Release        uintptr // 2
	Length         uintptr // 3
	GetElement     uintptr // 4
}

// IUIAutomationTextPattern vtable
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

// IUIAutomationTextRange vtable
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

const (
	treeScopeChildren    = 0x2
	treeScopeDescendants = 0x4
	treeScopeSubtree     = 0x7

	uiaNamePropertyId        = 30005
	uiaControlTypePropertyId = 30003
	uiaClassNamePropertyId   = 30012

	uiaTextPatternId  = 10014
	uiaValuePatternId = 10002
)

// IUIAutomationValuePattern vtable
type iValuePatternVtbl struct {
	QueryInterface uintptr // 0
	AddRef         uintptr // 1
	Release        uintptr // 2
	SetValue       uintptr // 3
	CurrentValue   uintptr // 4
}

type comPtr = unsafe.Pointer

var trueCond comPtr

func vtblOf(obj comPtr) comPtr {
	return *(*comPtr)(obj)
}

func main() {
	runtime.LockOSThread()

	if err := ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED); err != nil {
		oleErr, ok := err.(*ole.OleError)
		if !ok || oleErr.Code() != 0x80010106 {
			fmt.Fprintf(os.Stderr, "CoInitialize: %v\n", err)
			os.Exit(1)
		}
	}
	defer ole.CoUninitialize()

	unknown, err := ole.CreateInstance(clsidCUIAutomation, iidIUIAutomation)
	if err != nil {
		fmt.Fprintf(os.Stderr, "CreateInstance: %v\n", err)
		os.Exit(1)
	}
	defer unknown.Release()

	thisPtr := comPtr(unsafe.Pointer(unknown))
	vt := (*iUIAutomationVtbl)(unsafe.Pointer(unknown.RawVTable))

	// Get root element
	var rootElem comPtr
	hr, _, _ := syscall.SyscallN(vt.GetRootElement, uintptr(thisPtr), uintptr(unsafe.Pointer(&rootElem)))
	if hr != 0 || rootElem == nil {
		fmt.Fprintln(os.Stderr, "Failed to get root element")
		os.Exit(1)
	}

	// Create true condition
	hr, _, _ = syscall.SyscallN(vt.CreateTrueCondition, uintptr(thisPtr), uintptr(unsafe.Pointer(&trueCond)))
	if hr != 0 || trueCond == nil {
		fmt.Fprintln(os.Stderr, "Failed to create true condition")
		os.Exit(1)
	}

	// Find the target app
	searchTerm := "Notepad"
	if len(os.Args) > 1 {
		searchTerm = os.Args[1]
	}

	fmt.Printf("Searching for: %q\n\n", searchTerm)

	appElem := findApp(rootElem, searchTerm)
	if appElem == nil {
		fmt.Fprintf(os.Stderr, "Application %q not found\n", searchTerm)
		os.Exit(1)
	}

	name := getStr(appElem, uiaNamePropertyId)
	fmt.Printf("Found: %q\n\n", name)

	// Dump tree and try to read text
	fmt.Println("=== Accessibility Tree ===")
	dumpTree(appElem, 0, 15)

	// Try reading text content
	fmt.Println("\n=== Text Content ===")
	readAllText(appElem, 0, 10)
}

func findApp(root comPtr, search string) comPtr {
	children := findChildren(root)
	searchLower := strings.ToLower(search)
	for _, child := range children {
		name := strings.ToLower(getStr(child, uiaNamePropertyId))
		className := strings.ToLower(getStr(child, uiaClassNamePropertyId))
		if strings.Contains(name, searchLower) || strings.Contains(className, searchLower) {
			return child
		}
	}
	return nil
}

func findChildren(elem comPtr) []comPtr {
	evt := (*iElementVtbl)(vtblOf(elem))
	var arrayPtr comPtr
	hr, _, _ := syscall.SyscallN(evt.FindAll, uintptr(elem), uintptr(treeScopeChildren), uintptr(trueCond), uintptr(unsafe.Pointer(&arrayPtr)))
	if hr != 0 || arrayPtr == nil {
		return nil
	}

	avt := (*iElementArrayVtbl)(vtblOf(arrayPtr))
	var length int32
	syscall.SyscallN(avt.Length, uintptr(arrayPtr), uintptr(unsafe.Pointer(&length)))

	result := make([]comPtr, 0, length)
	for i := int32(0); i < length; i++ {
		var child comPtr
		hr, _, _ = syscall.SyscallN(avt.GetElement, uintptr(arrayPtr), uintptr(i), uintptr(unsafe.Pointer(&child)))
		if hr == 0 && child != nil {
			result = append(result, child)
		}
	}
	return result
}

func dumpTree(elem comPtr, depth, maxDepth int) {
	if depth > maxDepth {
		return
	}
	children := findChildren(elem)
	indent := strings.Repeat("  ", depth)
	for _, child := range children {
		name := getStr(child, uiaNamePropertyId)
		ct := getInt(child, uiaControlTypePropertyId)
		className := getStr(child, uiaClassNamePropertyId)

		patterns := ""
		if hasPattern(child, uiaTextPatternId) {
			patterns += " {Text}"
		}
		if hasPattern(child, uiaValuePatternId) {
			patterns += " {Value}"
		}

		role := controlTypeName(ct)
		fmt.Printf("%s[%s] %q class=%q%s\n", indent, role, name, className, patterns)
		dumpTree(child, depth+1, maxDepth)
	}
}

func readAllText(elem comPtr, depth, maxDepth int) {
	if depth > maxDepth {
		return
	}

	// Try TextPattern on this element
	text := readTextPattern(elem)
	if text != "" {
		name := getStr(elem, uiaNamePropertyId)
		role := controlTypeName(getInt(elem, uiaControlTypePropertyId))
		fmt.Printf("[%s] %q → TextPattern:\n%s\n\n", role, name, text)
		return
	}

	// Try ValuePattern on this element
	value := readValuePattern(elem)
	if value != "" {
		name := getStr(elem, uiaNamePropertyId)
		role := controlTypeName(getInt(elem, uiaControlTypePropertyId))
		fmt.Printf("[%s] %q → ValuePattern:\n%s\n\n", role, name, value)
		return
	}

	// Recurse into children
	children := findChildren(elem)
	for _, child := range children {
		readAllText(child, depth+1, maxDepth)
	}
}

func readTextPattern(elem comPtr) string {
	evt := (*iElementVtbl)(vtblOf(elem))
	var pattern comPtr
	hr, _, _ := syscall.SyscallN(evt.GetCurrentPattern, uintptr(elem), uintptr(uiaTextPatternId), uintptr(unsafe.Pointer(&pattern)))
	if hr != 0 || pattern == nil {
		return ""
	}

	// Get DocumentRange
	tpVtbl := (*iTextPatternVtbl)(vtblOf(pattern))
	var docRange comPtr
	hr, _, _ = syscall.SyscallN(tpVtbl.DocumentRange, uintptr(pattern), uintptr(unsafe.Pointer(&docRange)))
	if hr != 0 || docRange == nil {
		return ""
	}

	// GetText(-1) to get all text
	trVtbl := (*iTextRangeVtbl)(vtblOf(docRange))
	var bstr uintptr
	hr, _, _ = syscall.SyscallN(trVtbl.GetText, uintptr(docRange), uintptr(0xFFFFFFFF), uintptr(unsafe.Pointer(&bstr)))
	if hr != 0 || bstr == 0 {
		return ""
	}

	text := bstrToString(bstr)
	sysFreeString(bstr)
	return text
}

func readValuePattern(elem comPtr) string {
	evt := (*iElementVtbl)(vtblOf(elem))
	var pattern comPtr
	hr, _, _ := syscall.SyscallN(evt.GetCurrentPattern, uintptr(elem), uintptr(uiaValuePatternId), uintptr(unsafe.Pointer(&pattern)))
	if hr != 0 || pattern == nil {
		return ""
	}

	vpVtbl := (*iValuePatternVtbl)(vtblOf(pattern))
	var bstr uintptr
	hr, _, _ = syscall.SyscallN(vpVtbl.CurrentValue, uintptr(pattern), uintptr(unsafe.Pointer(&bstr)))
	if hr != 0 || bstr == 0 {
		return ""
	}

	text := bstrToString(bstr)
	sysFreeString(bstr)
	return text
}

func hasPattern(elem comPtr, patternID int) bool {
	evt := (*iElementVtbl)(vtblOf(elem))
	var pattern comPtr
	hr, _, _ := syscall.SyscallN(evt.GetCurrentPattern, uintptr(elem), uintptr(patternID), uintptr(unsafe.Pointer(&pattern)))
	if hr == 0 && pattern != nil {
		// Release it
		pvt := (*iElementVtbl)(vtblOf(pattern))
		syscall.SyscallN(pvt.Release, uintptr(pattern))
		return true
	}
	return false
}

func getStr(elem comPtr, propID int) string {
	evt := (*iElementVtbl)(vtblOf(elem))
	var v ole.VARIANT
	ole.VariantInit(&v)
	hr, _, _ := syscall.SyscallN(evt.GetCurrentPropertyValue, uintptr(elem), uintptr(propID), uintptr(unsafe.Pointer(&v)))
	if hr != 0 {
		return ""
	}
	defer v.Clear()
	if v.VT == ole.VT_BSTR {
		return v.ToString()
	}
	return ""
}

func getInt(elem comPtr, propID int) int32 {
	evt := (*iElementVtbl)(vtblOf(elem))
	var v ole.VARIANT
	ole.VariantInit(&v)
	hr, _, _ := syscall.SyscallN(evt.GetCurrentPropertyValue, uintptr(elem), uintptr(propID), uintptr(unsafe.Pointer(&v)))
	if hr != 0 {
		return -1
	}
	defer v.Clear()
	return int32(v.Val)
}

var (
	oleaut32          = syscall.NewLazyDLL("oleaut32.dll")
	procSysFreeString = oleaut32.NewProc("SysFreeString")
)

func sysFreeString(bstr uintptr) {
	procSysFreeString.Call(bstr)
}

func bstrToString(bstr uintptr) string {
	if bstr == 0 {
		return ""
	}
	length := *(*uint32)(unsafe.Pointer(bstr - 4))
	if length == 0 {
		return ""
	}
	wchars := unsafe.Slice((*uint16)(unsafe.Pointer(bstr)), length/2)
	return syscall.UTF16ToString(wchars)
}

func controlTypeName(ct int32) string {
	names := map[int32]string{
		50000: "button", 50004: "edit", 50008: "list", 50007: "list item",
		50020: "text", 50025: "custom", 50026: "group", 50030: "document",
		50032: "window", 50033: "pane", 50036: "table", 50037: "title bar",
		50021: "tool bar", 50010: "menu bar", 50017: "status bar",
	}
	if n, ok := names[ct]; ok {
		return n
	}
	return fmt.Sprintf("type(%d)", ct)
}
