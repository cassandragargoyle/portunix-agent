/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package accessibility

import (
	"fmt"
	"unsafe"
)

// AccessibleRef is an opaque reference to a node in the accessibility tree.
// Each platform backend provides its own concrete type implementing this interface.
// Implementations must be comparable (usable as map keys).
type AccessibleRef interface {
	// String returns a human-readable representation for debugging
	String() string
}

// ATSPIRef identifies a node on the AT-SPI2/D-Bus accessibility bus (Linux)
type ATSPIRef struct {
	BusName string
	Path    string
}

func (r ATSPIRef) String() string {
	return fmt.Sprintf("atspi:%s:%s", r.BusName, r.Path)
}

// UIARef identifies a node via Windows UI Automation.
// Element holds an IUIAutomationElement COM pointer.
type UIARef struct {
	Element unsafe.Pointer
}

func (r UIARef) String() string {
	return fmt.Sprintf("uia:%p", r.Element)
}
