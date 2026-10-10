//go:build linux

/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package accessibility

import "fmt"

// Text reads text content from an accessible node implementing the Text interface.
// Use startOffset=0 and endOffset=-1 to read all available text.
func (b *atspiBackend) Text(ref AccessibleRef, startOffset, endOffset int32) (string, error) {
	r := toATSPI(ref)
	result, err := b.conn.Call(r.BusName, r.Path, ifaceText, "GetText", startOffset, endOffset)
	if err != nil {
		return "", fmt.Errorf("GetText failed: %w", err)
	}
	if len(result) == 0 {
		return "", nil
	}
	text, ok := result[0].(string)
	if !ok {
		return "", fmt.Errorf("unexpected type for text content: %T", result[0])
	}
	return text, nil
}

// CharacterCount returns the number of characters in a text accessible node
func (b *atspiBackend) CharacterCount(ref AccessibleRef) (int32, error) {
	r := toATSPI(ref)
	result, err := b.conn.GetProperty(r.BusName, r.Path, ifaceText, "CharacterCount")
	if err != nil {
		return 0, fmt.Errorf("get CharacterCount failed: %w", err)
	}
	switch v := result.(type) {
	case int32:
		return v, nil
	case int64:
		return int32(v), nil
	case uint32:
		return int32(v), nil
	}
	return 0, fmt.Errorf("unexpected type for CharacterCount: %T", result)
}
