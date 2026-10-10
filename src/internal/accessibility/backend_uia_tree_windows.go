//go:build windows

/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package accessibility

import (
	"fmt"
	"strings"
	"syscall"
	"unsafe"

	"github.com/go-ole/go-ole"
)

// IUIAutomationElement vtable layout (subset of methods used)
type iUIAutomationElementVtbl struct {
	// IUnknown (0-2)
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr

	// IUIAutomationElement methods
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

// IUIAutomationElementArray vtable layout
type iUIAutomationElementArrayVtbl struct {
	QueryInterface uintptr // 0
	AddRef         uintptr // 1
	Release        uintptr // 2
	Length         uintptr // 3
	GetElement     uintptr // 4
}

// comIUnknownVtbl is the minimal IUnknown vtable for releasing COM objects
type comIUnknownVtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr
}

// UIA TreeScope constants
const (
	treeScopeChildren = 0x2
)

// UIA Property IDs
const (
	uiaNamePropertyId         = 30005
	uiaControlTypePropertyId  = 30003
	uiaClassNamePropertyId    = 30012
	uiaAutomationIdPropertyId = 30011
)

// elemPtr wraps an unsafe.Pointer to a UIA element for internal use.
// COM element pointers are stored as unsafe.Pointer to satisfy go vet.
type elemPtr = unsafe.Pointer

// getRootElement returns the UIA desktop root element
func (b *uiaBackend) getRootElement() (elemPtr, error) {
	var element elemPtr
	hr, _, _ := syscall.SyscallN(
		b.automationVtbl.GetRootElement,
		uintptr(unsafe.Pointer(b.automation)),
		uintptr(unsafe.Pointer(&element)),
	)
	if hr != 0 {
		return nil, fmt.Errorf("GetRootElement failed: HRESULT 0x%08X", hr)
	}
	return element, nil
}

// createTrueCondition creates a UIA condition that matches all elements
func (b *uiaBackend) createTrueCondition() (unsafe.Pointer, error) {
	var cond unsafe.Pointer
	hr, _, _ := syscall.SyscallN(
		b.automationVtbl.CreateTrueCondition,
		uintptr(unsafe.Pointer(b.automation)),
		uintptr(unsafe.Pointer(&cond)),
	)
	if hr != 0 {
		return nil, fmt.Errorf("CreateTrueCondition failed: HRESULT 0x%08X", hr)
	}
	return cond, nil
}

// findChildren returns direct children of an element using FindAll
func (b *uiaBackend) findChildren(elem elemPtr) ([]elemPtr, error) {
	evtbl := getElementVtbl(elem)

	var arrayPtr unsafe.Pointer
	hr, _, _ := syscall.SyscallN(
		evtbl.FindAll,
		uintptr(elem),
		uintptr(treeScopeChildren),
		uintptr(b.trueCond),
		uintptr(unsafe.Pointer(&arrayPtr)),
	)
	if hr != 0 {
		return nil, fmt.Errorf("FindAll failed: HRESULT 0x%08X", hr)
	}
	if arrayPtr == nil {
		return nil, nil
	}
	defer releaseCOMObject(arrayPtr)

	avt := getElementArrayVtbl(arrayPtr)

	var length int32
	hr, _, _ = syscall.SyscallN(avt.Length, uintptr(arrayPtr), uintptr(unsafe.Pointer(&length)))
	if hr != 0 {
		return nil, fmt.Errorf("ElementArray.Length failed: HRESULT 0x%08X", hr)
	}

	children := make([]elemPtr, 0, length)
	for i := int32(0); i < length; i++ {
		var child elemPtr
		hr, _, _ = syscall.SyscallN(avt.GetElement, uintptr(arrayPtr), uintptr(i), uintptr(unsafe.Pointer(&child)))
		if hr != 0 || child == nil {
			continue
		}
		children = append(children, child)
	}
	return children, nil
}

// getElementName returns the Name property of a UIA element via GetCurrentPropertyValue
func getElementName(elem elemPtr) (string, error) {
	return getStringProperty(elem, uiaNamePropertyId)
}

// getElementControlType returns the ControlType of a UIA element via GetCurrentPropertyValue
func getElementControlType(elem elemPtr) (int32, error) {
	return getIntProperty(elem, uiaControlTypePropertyId)
}

// getElementClassName returns the ClassName property of a UIA element
func getElementClassName(elem elemPtr) (string, error) {
	return getStringProperty(elem, uiaClassNamePropertyId)
}

// getElementAutomationID returns the AutomationId property of a UIA element
func getElementAutomationID(elem elemPtr) (string, error) {
	return getStringProperty(elem, uiaAutomationIdPropertyId)
}

// getStringProperty reads a string property from a UIA element using GetCurrentPropertyValue
func getStringProperty(elem elemPtr, propID int) (string, error) {
	vtbl := getElementVtbl(elem)
	var v ole.VARIANT
	ole.VariantInit(&v)
	hr, _, _ := syscall.SyscallN(
		vtbl.GetCurrentPropertyValue,
		uintptr(elem),
		uintptr(propID),
		uintptr(unsafe.Pointer(&v)),
	)
	if hr != 0 {
		return "", fmt.Errorf("GetCurrentPropertyValue(%d) failed: HRESULT 0x%08X", propID, hr)
	}
	defer v.Clear()
	if v.VT == ole.VT_BSTR {
		return v.ToString(), nil
	}
	return "", nil
}

// getIntProperty reads an integer property from a UIA element using GetCurrentPropertyValue
func getIntProperty(elem elemPtr, propID int) (int32, error) {
	vtbl := getElementVtbl(elem)
	var v ole.VARIANT
	ole.VariantInit(&v)
	hr, _, _ := syscall.SyscallN(
		vtbl.GetCurrentPropertyValue,
		uintptr(elem),
		uintptr(propID),
		uintptr(unsafe.Pointer(&v)),
	)
	if hr != 0 {
		return 0, fmt.Errorf("GetCurrentPropertyValue(%d) failed: HRESULT 0x%08X", propID, hr)
	}
	defer v.Clear()
	return int32(v.Val), nil
}

// releaseElement releases a UIA element COM object
func releaseElement(elem elemPtr) {
	if elem != nil {
		vtbl := getElementVtbl(elem)
		syscall.SyscallN(vtbl.Release, uintptr(elem))
	}
}

// releaseCOMObject releases any COM object via its IUnknown.Release
func releaseCOMObject(obj unsafe.Pointer) {
	if obj != nil {
		vtbl := (*comIUnknownVtbl)(*((*unsafe.Pointer)(obj)))
		syscall.SyscallN(vtbl.Release, uintptr(obj))
	}
}

// getElementVtbl extracts the vtable from a UIA element pointer
func getElementVtbl(elem elemPtr) *iUIAutomationElementVtbl {
	return (*iUIAutomationElementVtbl)(*((*unsafe.Pointer)(elem)))
}

// getElementArrayVtbl extracts the vtable from a UIA element array pointer
func getElementArrayVtbl(arr unsafe.Pointer) *iUIAutomationElementArrayVtbl {
	return (*iUIAutomationElementArrayVtbl)(*((*unsafe.Pointer)(arr)))
}

// toElemPtr converts a UIARef to an internal element pointer
func toElemPtr(ref AccessibleRef) elemPtr {
	return ref.(UIARef).Element
}

// FindApplication searches for VS Code by window name in the UIA tree
func (b *uiaBackend) FindApplication(name string) (*Accessible, error) {
	children, err := b.findChildren(b.rootElement)
	if err != nil {
		return nil, fmt.Errorf("failed to enumerate desktop children: %w", err)
	}

	nameLower := strings.ToLower(name)
	for _, child := range children {
		winName, _ := getElementName(child)
		winNameLower := strings.ToLower(winName)

		if strings.Contains(winNameLower, nameLower) || strings.Contains(winNameLower, "visual studio code") {
			controlType, _ := getElementControlType(child)
			roleName := controlTypeToRole(controlType)

			// Release other children
			for _, c := range children {
				if c != child {
					releaseElement(c)
				}
			}

			return &Accessible{
				Ref:      UIARef{Element: child},
				Name:     winName,
				RoleName: roleName,
			}, nil
		}
	}

	// Release all children since no match found
	for _, c := range children {
		releaseElement(c)
	}

	return nil, fmt.Errorf("application %q not found in accessibility tree", name)
}

// Children returns the direct children of an accessible node
func (b *uiaBackend) Children(ref AccessibleRef) ([]Accessible, error) {
	elem := toElemPtr(ref)
	if elem == nil {
		return nil, fmt.Errorf("invalid element reference")
	}

	childPtrs, err := b.findChildren(elem)
	if err != nil {
		return nil, err
	}

	children := make([]Accessible, 0, len(childPtrs))
	for _, child := range childPtrs {
		name, _ := getElementName(child)
		controlType, _ := getElementControlType(child)
		roleName := controlTypeToRole(controlType)

		children = append(children, Accessible{
			Ref:      UIARef{Element: child},
			Name:     name,
			RoleName: roleName,
		})
	}

	return children, nil
}

// RoleName returns the role name of a UIA element
func (b *uiaBackend) RoleName(ref AccessibleRef) (string, error) {
	elem := toElemPtr(ref)
	controlType, err := getElementControlType(elem)
	if err != nil {
		return "", err
	}
	return controlTypeToRole(controlType), nil
}

// NodeName returns the accessible name of a UIA element
func (b *uiaBackend) NodeName(ref AccessibleRef) (string, error) {
	elem := toElemPtr(ref)
	return getElementName(elem)
}

// Interfaces returns a list of supported capability names
func (b *uiaBackend) Interfaces(ref AccessibleRef) ([]string, error) {
	elem := toElemPtr(ref)
	var ifaces []string

	vtbl := getElementVtbl(elem)

	// Check if element supports TextPattern (pattern ID 10014)
	var pattern unsafe.Pointer
	hr, _, _ := syscall.SyscallN(
		vtbl.GetCurrentPattern,
		uintptr(elem),
		10014, // UIA_TextPatternId
		uintptr(unsafe.Pointer(&pattern)),
	)
	if hr == 0 && pattern != nil {
		ifaces = append(ifaces, "Text")
		releaseCOMObject(pattern)
	}

	// Check ValuePattern (pattern ID 10002)
	hr, _, _ = syscall.SyscallN(
		vtbl.GetCurrentPattern,
		uintptr(elem),
		10002, // UIA_ValuePatternId
		uintptr(unsafe.Pointer(&pattern)),
	)
	if hr == 0 && pattern != nil {
		ifaces = append(ifaces, "Value")
		releaseCOMObject(pattern)
	}

	return ifaces, nil
}

// controlTypeToRole maps UIA ControlType to cross-platform role name
func controlTypeToRole(controlType int32) string {
	if name, ok := controlTypeNames[controlType]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", controlType)
}
