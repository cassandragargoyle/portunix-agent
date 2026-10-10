//go:build windows

/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package accessibility

import (
	"fmt"
	"runtime"
	"unsafe"

	"github.com/go-ole/go-ole"
)

// UIA COM GUIDs
var (
	clsidCUIAutomation = ole.NewGUID("{FF48DBA4-60EF-4201-AA87-54103EEF594E}")
	iidIUIAutomation   = ole.NewGUID("{30CBE57D-D9D0-452A-AB13-7AC5AC4825EE}")
)

// UIA ControlType IDs
const (
	uiaListControlTypeID     = 50008
	uiaListItemControlTypeID = 50009
)

// controlTypeNames maps UIA ControlType IDs to human-readable role names
// compatible with AT-SPI2 naming for cross-platform consistency
var controlTypeNames = map[int32]string{
	50000: "button",
	50001: "calendar",
	50002: "check box",
	50003: "combo box",
	50004: "edit",
	50005: "hyperlink",
	50006: "image",
	50007: "list item",
	50008: "list",
	50009: "menu",
	50010: "menu bar",
	50011: "menu item",
	50012: "progress bar",
	50013: "radio button",
	50014: "scroll bar",
	50015: "slider",
	50016: "spinner",
	50017: "status bar",
	50018: "tab",
	50019: "tab item",
	50020: "text",
	50021: "tool bar",
	50022: "tool tip",
	50023: "tree",
	50024: "tree item",
	50025: "custom",
	50026: "group",
	50027: "thumb",
	50028: "data grid",
	50029: "data item",
	50030: "document",
	50031: "split button",
	50032: "window",
	50033: "pane",
	50034: "header",
	50035: "header item",
	50036: "table",
	50037: "title bar",
	50038: "separator",
}

// uiaBackend implements Backend using Windows UI Automation via COM
type uiaBackend struct {
	automation     *ole.IUnknown
	automationVtbl *iUIAutomationVtbl
	rootElement    elemPtr        // desktop root element (unsafe.Pointer)
	trueCond       unsafe.Pointer // IUIAutomationCondition for FindAll calls
}

// iUIAutomationVtbl represents the vtable layout for IUIAutomation COM interface
type iUIAutomationVtbl struct {
	// IUnknown methods (offsets 0-2)
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr

	// IUIAutomation methods
	CompareElements                   uintptr // 3
	CompareRuntimeIds                 uintptr // 4
	GetRootElement                    uintptr // 5
	ElementFromHandle                 uintptr // 6
	ElementFromPoint                  uintptr // 7
	GetFocusedElement                 uintptr // 8
	CreateTreeWalker                  uintptr // 9
	ControlViewWalker                 uintptr // 10
	ContentViewWalker                 uintptr // 11
	RawViewWalker                     uintptr // 12
	RawViewCondition                  uintptr // 13
	ControlViewCondition              uintptr // 14
	ContentViewCondition              uintptr // 15
	CreateCacheRequest                uintptr // 16
	CreateTrueCondition               uintptr // 17
	CreateFalseCondition              uintptr // 18
	CreatePropertyCondition           uintptr // 19
	CreatePropertyConditionEx         uintptr // 20
	CreateAndCondition                uintptr // 21
	CreateAndConditionFromArray       uintptr // 22
	CreateAndConditionFromNativeArray uintptr // 23
	CreateOrCondition                 uintptr // 24
	CreateOrConditionFromArray        uintptr // 25
	CreateOrConditionFromNativeArray  uintptr // 26
	CreateNotCondition                uintptr // 27
}

// NewUIABackend creates a Windows UI Automation accessibility backend.
// All COM operations are pinned to a single OS thread.
func NewUIABackend() (*uiaBackend, error) {
	runtime.LockOSThread()

	if err := ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED); err != nil {
		// E_RPC_CHANGED_MODE (0x80010106) means COM is already initialized
		oleErr, ok := err.(*ole.OleError)
		if !ok || oleErr.Code() != 0x80010106 {
			runtime.UnlockOSThread()
			return nil, fmt.Errorf("COM initialization failed: %w", err)
		}
	}

	unknown, err := ole.CreateInstance(clsidCUIAutomation, iidIUIAutomation)
	if err != nil {
		ole.CoUninitialize()
		runtime.UnlockOSThread()
		return nil, fmt.Errorf("failed to create IUIAutomation instance: %w", err)
	}

	b := &uiaBackend{
		automation: unknown,
	}

	// Get vtable pointer
	b.automationVtbl = (*iUIAutomationVtbl)(*((*unsafe.Pointer)(unsafe.Pointer(unknown))))

	// Get root element
	rootElement, err := b.getRootElement()
	if err != nil {
		unknown.Release()
		ole.CoUninitialize()
		runtime.UnlockOSThread()
		return nil, fmt.Errorf("failed to get root element: %w", err)
	}
	b.rootElement = rootElement

	// Create true condition for FindAll calls
	trueCond, err := b.createTrueCondition()
	if err != nil {
		releaseElement(rootElement)
		unknown.Release()
		ole.CoUninitialize()
		runtime.UnlockOSThread()
		return nil, fmt.Errorf("failed to create true condition: %w", err)
	}
	b.trueCond = trueCond

	return b, nil
}

func (b *uiaBackend) Name() string { return "uia" }

// RootRef returns a reference to the desktop root element
func (b *uiaBackend) RootRef() AccessibleRef {
	return UIARef{Element: b.rootElement}
}

func (b *uiaBackend) Close() error {
	if b.trueCond != nil {
		releaseCOMObject(b.trueCond)
		b.trueCond = nil
	}
	if b.rootElement != nil {
		releaseElement(b.rootElement)
		b.rootElement = nil
	}
	if b.automation != nil {
		b.automation.Release()
		b.automation = nil
	}
	ole.CoUninitialize()
	runtime.UnlockOSThread()
	return nil
}
