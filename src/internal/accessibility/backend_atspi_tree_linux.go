//go:build linux

/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package accessibility

import (
	"fmt"
	"strings"
)

// toATSPI extracts the ATSPIRef from an AccessibleRef interface value
func toATSPI(ref AccessibleRef) ATSPIRef {
	return ref.(ATSPIRef)
}

// FindApplication searches the accessibility tree for an application by name.
// The name is matched case-insensitively as a substring.
func (b *atspiBackend) FindApplication(name string) (*Accessible, error) {
	// Get desktop children (registered applications)
	apps, err := b.desktopChildren()
	if err != nil {
		return nil, fmt.Errorf("failed to enumerate applications: %w", err)
	}

	nameLower := strings.ToLower(name)
	for _, app := range apps {
		appName, err := b.NodeName(app.Ref)
		if err != nil {
			continue // skip inaccessible applications
		}
		if strings.ToLower(appName) == nameLower || strings.Contains(strings.ToLower(appName), nameLower) {
			app.Name = appName
			return &app, nil
		}
	}

	return nil, fmt.Errorf("application %q not found in accessibility tree", name)
}

// Children returns the direct children of an accessible node.
func (b *atspiBackend) Children(ref AccessibleRef) ([]Accessible, error) {
	count, err := b.childCount(ref)
	if err != nil {
		return nil, fmt.Errorf("failed to get child count: %w", err)
	}

	children := make([]Accessible, 0, count)
	for i := int32(0); i < count; i++ {
		childRef, err := b.childAtIndex(ref, i)
		if err != nil {
			continue // skip inaccessible children
		}

		child := Accessible{Ref: childRef}
		child.RoleName, _ = b.RoleName(childRef)
		child.Name, _ = b.NodeName(childRef)
		children = append(children, child)
	}

	return children, nil
}

// RoleName returns the role name of an accessible node
func (b *atspiBackend) RoleName(ref AccessibleRef) (string, error) {
	r := toATSPI(ref)
	result, err := b.conn.Call(r.BusName, r.Path, ifaceAccessible, "GetRoleName")
	if err != nil {
		return "", fmt.Errorf("GetRoleName failed: %w", err)
	}
	if len(result) == 0 {
		return "", fmt.Errorf("empty response from GetRoleName")
	}
	role, ok := result[0].(string)
	if !ok {
		return "", fmt.Errorf("unexpected type for role name: %T", result[0])
	}
	return role, nil
}

// NodeName returns the accessible name of a node
func (b *atspiBackend) NodeName(ref AccessibleRef) (string, error) {
	r := toATSPI(ref)
	result, err := b.conn.GetProperty(r.BusName, r.Path, ifaceAccessible, "Name")
	if err != nil {
		return "", fmt.Errorf("get Name property failed: %w", err)
	}
	name, ok := result.(string)
	if !ok {
		return "", fmt.Errorf("unexpected type for Name: %T", result)
	}
	return name, nil
}

// Interfaces returns the list of AT-SPI2 interfaces implemented by a node
func (b *atspiBackend) Interfaces(ref AccessibleRef) ([]string, error) {
	r := toATSPI(ref)
	result, err := b.conn.Call(r.BusName, r.Path, ifaceAccessible, "GetInterfaces")
	if err != nil {
		return nil, fmt.Errorf("GetInterfaces failed: %w", err)
	}
	if len(result) == 0 {
		return nil, nil
	}

	// Handle both []string and []interface{} return types
	switch v := result[0].(type) {
	case []string:
		return v, nil
	case []interface{}:
		ifaces := make([]string, 0, len(v))
		for _, elem := range v {
			if s, ok := elem.(string); ok {
				ifaces = append(ifaces, s)
			}
		}
		return ifaces, nil
	}
	return nil, fmt.Errorf("unexpected type for interfaces: %T", result[0])
}

// desktopChildren returns all applications registered on the accessibility bus
func (b *atspiBackend) desktopChildren() ([]Accessible, error) {
	return b.Children(ATSPIRef{
		BusName: registryDest,
		Path:    registryPath,
	})
}

// childCount returns the number of children for an accessible node
func (b *atspiBackend) childCount(ref AccessibleRef) (int32, error) {
	r := toATSPI(ref)
	result, err := b.conn.GetProperty(r.BusName, r.Path, ifaceAccessible, "ChildCount")
	if err != nil {
		return 0, err
	}
	switch v := result.(type) {
	case int32:
		return v, nil
	case int64:
		return int32(v), nil
	case uint32:
		return int32(v), nil
	}
	return 0, fmt.Errorf("unexpected type for ChildCount: %T", result)
}

// childAtIndex returns the child at the given index
func (b *atspiBackend) childAtIndex(ref AccessibleRef, index int32) (ATSPIRef, error) {
	r := toATSPI(ref)
	result, err := b.conn.Call(r.BusName, r.Path, ifaceAccessible, "GetChildAtIndex", index)
	if err != nil {
		return ATSPIRef{}, fmt.Errorf("GetChildAtIndex(%d) failed: %w", index, err)
	}
	return extractRef(result)
}
