/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package browser

import (
	"fmt"
	"strings"

	"github.com/playwright-community/playwright-go"
)

// Element represents an interactive element detected on the page
type Element struct {
	Index       int    `json:"index"`
	Tag         string `json:"tag"`
	Type        string `json:"type"`
	Text        string `json:"text,omitempty"`
	Href        string `json:"href,omitempty"`
	Placeholder string `json:"placeholder,omitempty"`
	AriaLabel   string `json:"aria_label,omitempty"`
	Visible     bool   `json:"visible"`
	Selector    string `json:"-"`
}

// String returns a human-readable representation of the element
func (e Element) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("[%d]", e.Index))
	parts = append(parts, e.Type)
	if e.Text != "" {
		text := e.Text
		if len(text) > 60 {
			text = text[:57] + "..."
		}
		parts = append(parts, fmt.Sprintf("%q", text))
	}
	if e.Href != "" {
		parts = append(parts, fmt.Sprintf("-> %s", e.Href))
	}
	if !e.Visible {
		parts = append(parts, "(offscreen)")
	}
	return strings.Join(parts, " ")
}

// elementDetectionJS is injected into the page to detect interactive elements.
// Inspired by browser-use's approach: finds clickable/interactive elements
// by analyzing tag names, roles, and attributes.
const elementDetectionJS = `() => {
	const interactiveTags = ['A', 'BUTTON', 'INPUT', 'SELECT', 'TEXTAREA', 'DETAILS', 'SUMMARY'];
	const interactiveRoles = ['button', 'link', 'checkbox', 'radio', 'tab', 'menuitem',
		'option', 'switch', 'textbox', 'combobox', 'listbox', 'slider', 'spinbutton',
		'searchbox', 'menu', 'menubar', 'navigation', 'treeitem'];

	function isVisible(el) {
		const style = window.getComputedStyle(el);
		if (style.display === 'none' || style.visibility === 'hidden' || style.opacity === '0') return false;
		const rect = el.getBoundingClientRect();
		return rect.width > 0 && rect.height > 0;
	}

	function isInViewport(el) {
		const rect = el.getBoundingClientRect();
		return rect.top < window.innerHeight && rect.bottom > 0 &&
			rect.left < window.innerWidth && rect.right > 0;
	}

	function isInteractive(el) {
		if (interactiveTags.includes(el.tagName)) return true;
		const role = el.getAttribute('role');
		if (role && interactiveRoles.includes(role)) return true;
		if (el.hasAttribute('onclick') || el.hasAttribute('tabindex')) return true;
		const cursor = window.getComputedStyle(el).cursor;
		if (cursor === 'pointer') return true;
		if (el.contentEditable === 'true') return true;
		return false;
	}

	function getElementText(el) {
		if (el.tagName === 'INPUT' || el.tagName === 'TEXTAREA') {
			return el.value || el.placeholder || '';
		}
		return (el.innerText || el.textContent || '').trim().substring(0, 200);
	}

	function getElementType(el) {
		const role = el.getAttribute('role');
		if (role) return role;
		const tag = el.tagName.toLowerCase();
		if (tag === 'a') return 'link';
		if (tag === 'input') return el.type || 'text';
		if (tag === 'select') return 'select';
		if (tag === 'textarea') return 'textarea';
		if (tag === 'button') return 'button';
		if (tag === 'details') return 'details';
		if (tag === 'summary') return 'summary';
		return tag;
	}

	function buildSelector(el) {
		if (el.id) return '#' + CSS.escape(el.id);
		const tag = el.tagName.toLowerCase();
		const parts = [tag];
		const role = el.getAttribute('role');
		if (role) parts.push('[role="' + role + '"]');
		const name = el.getAttribute('name');
		if (name) parts.push('[name="' + CSS.escape(name) + '"]');
		const ariaLabel = el.getAttribute('aria-label');
		if (ariaLabel) parts.push('[aria-label="' + CSS.escape(ariaLabel) + '"]');
		// Add nth-of-type if needed for uniqueness
		const selector = parts.join('');
		const matches = document.querySelectorAll(selector);
		if (matches.length > 1) {
			const idx = Array.from(matches).indexOf(el);
			return selector + ':nth-of-type(' + (idx + 1) + ')';
		}
		return selector;
	}

	const elements = [];
	const all = document.querySelectorAll('*');
	for (const el of all) {
		if (!isVisible(el)) continue;
		if (!isInteractive(el)) continue;
		// Skip elements nested inside other interactive elements that we already captured
		const parentInteractive = el.parentElement && el.parentElement.closest(
			interactiveTags.map(t => t.toLowerCase()).join(',') + ',[role]'
		);
		if (parentInteractive && elements.some(e => e._el === parentInteractive)) continue;

		elements.push({
			_el: el,
			tag: el.tagName.toLowerCase(),
			type: getElementType(el),
			text: getElementText(el),
			href: el.href || '',
			placeholder: el.placeholder || '',
			ariaLabel: el.getAttribute('aria-label') || '',
			visible: isInViewport(el),
			selector: buildSelector(el),
		});
	}

	// Return without the _el references
	return elements.map(({_el, ...rest}) => rest);
}`

// DetectElements finds all interactive elements on the current page.
// Returns elements with sequential numeric indices starting at 1.
func DetectElements(page playwright.Page, viewportOnly bool) ([]Element, error) {
	result, err := page.Evaluate(elementDetectionJS)
	if err != nil {
		return nil, fmt.Errorf("failed to detect elements: %w", err)
	}

	rawElements, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected element detection result type")
	}

	var elements []Element
	index := 1
	for _, raw := range rawElements {
		m, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}

		visible := toBool(m["visible"])
		if viewportOnly && !visible {
			continue
		}

		el := Element{
			Index:       index,
			Tag:         toString(m["tag"]),
			Type:        toString(m["type"]),
			Text:        toString(m["text"]),
			Href:        toString(m["href"]),
			Placeholder: toString(m["placeholder"]),
			AriaLabel:   toString(m["ariaLabel"]),
			Visible:     visible,
			Selector:    toString(m["selector"]),
		}
		elements = append(elements, el)
		index++
	}

	return elements, nil
}

// FindElementByIndex returns the element with the given index from the detected elements.
// The index is 1-based as presented to the user.
func FindElementByIndex(elements []Element, index int) (*Element, error) {
	for i := range elements {
		if elements[i].Index == index {
			return &elements[i], nil
		}
	}
	return nil, fmt.Errorf("element with index %d not found", index)
}

// ExtractContent extracts text content from the page, optionally filtered by CSS selector
func ExtractContent(page playwright.Page, selector string, format string) (string, error) {
	if selector != "" {
		loc := page.Locator(selector)
		count, err := loc.Count()
		if err != nil || count == 0 {
			return "", fmt.Errorf("no elements matching selector %q", selector)
		}

		switch format {
		case "html":
			return loc.First().InnerHTML()
		default:
			return loc.First().InnerText()
		}
	}

	// Full page content
	switch format {
	case "html":
		return page.Content()
	default:
		loc := page.Locator("body")
		return loc.InnerText()
	}
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}

func toBool(v interface{}) bool {
	if v == nil {
		return false
	}
	b, ok := v.(bool)
	if !ok {
		return false
	}
	return b
}
