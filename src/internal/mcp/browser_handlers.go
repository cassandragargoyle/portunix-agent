/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package mcp

import (
	"encoding/json"
	"fmt"

	"github.com/CassandraGargoyle/portunix-agent/internal/browser"
)

func registerBrowserHandlers(s *Server) {
	s.handlers["agent_browser_launch"] = s.handleBrowserLaunch
	s.handlers["agent_browser_connect"] = s.handleBrowserConnect
	s.handlers["agent_browser_close"] = s.handleBrowserClose
	s.handlers["agent_browser_status"] = s.handleBrowserStatus
	s.handlers["agent_browser_goto"] = s.handleBrowserGoto
	s.handlers["agent_browser_back"] = s.handleBrowserBack
	s.handlers["agent_browser_forward"] = s.handleBrowserForward
	s.handlers["agent_browser_reload"] = s.handleBrowserReload
	s.handlers["agent_browser_click"] = s.handleBrowserClick
	s.handlers["agent_browser_type"] = s.handleBrowserType
	s.handlers["agent_browser_keys"] = s.handleBrowserKeys
	s.handlers["agent_browser_scroll"] = s.handleBrowserScroll
	s.handlers["agent_browser_select"] = s.handleBrowserSelect
	s.handlers["agent_browser_elements"] = s.handleBrowserElements
	s.handlers["agent_browser_screenshot"] = s.handleBrowserScreenshot
	s.handlers["agent_browser_content"] = s.handleBrowserContent
	s.handlers["agent_browser_tabs"] = s.handleBrowserTabs
	s.handlers["agent_browser_tab_switch"] = s.handleBrowserTabSwitch
	s.handlers["agent_browser_tab_close"] = s.handleBrowserTabClose
	s.handlers["agent_browser_tab_new"] = s.handleBrowserTabNew
	s.handlers["agent_browser_evaluate"] = s.handleBrowserEvaluate
	s.handlers["agent_browser_upload"] = s.handleBrowserUpload
}

func (s *Server) handleBrowserLaunch(args map[string]any) (*ToolResult, error) {
	headless := getBool(args, "headless", true)
	profile, _ := getString(args, "profile")
	channel, _ := getString(args, "channel")
	sess, err := browser.Launch(browser.LaunchOptions{
		Headless: headless,
		Profile:  profile,
		Channel:  channel,
	})
	if err != nil {
		return nil, err
	}
	st := sess.GetStatus()
	return jsonResult(st)
}

func (s *Server) handleBrowserConnect(args map[string]any) (*ToolResult, error) {
	endpoint, _ := getString(args, "endpoint")
	sess, err := browser.Connect(endpoint)
	if err != nil {
		return nil, err
	}
	st := sess.GetStatus()
	return jsonResult(st)
}

func (s *Server) handleBrowserClose(args map[string]any) (*ToolResult, error) {
	sess, err := browser.RequireSession()
	if err != nil {
		return nil, err
	}
	if err := sess.Close(); err != nil {
		return nil, err
	}
	return textResult("Browser closed"), nil
}

func (s *Server) handleBrowserStatus(args map[string]any) (*ToolResult, error) {
	sess := browser.GetSession()
	if sess == nil {
		return jsonResult(browser.Status{Running: false})
	}
	return jsonResult(sess.GetStatus())
}

func (s *Server) handleBrowserGoto(args map[string]any) (*ToolResult, error) {
	url, ok := getString(args, "url")
	if !ok || url == "" {
		return nil, fmt.Errorf("'url' parameter is required")
	}

	sess, err := browser.RequireSession()
	if err != nil {
		return nil, err
	}

	newTab := getBool(args, "new_tab", false)
	if newTab {
		tab, err := browser.NewTab(sess, url)
		if err != nil {
			return nil, err
		}
		return jsonResult(tab)
	}

	if err := browser.Navigate(sess.ActivePage(), url); err != nil {
		return nil, err
	}

	title, _ := sess.ActivePage().Title()
	return jsonResult(map[string]string{
		"url":   sess.ActivePage().URL(),
		"title": title,
	})
}

func (s *Server) handleBrowserBack(args map[string]any) (*ToolResult, error) {
	sess, err := browser.RequireSession()
	if err != nil {
		return nil, err
	}
	if err := browser.GoBack(sess.ActivePage()); err != nil {
		return nil, err
	}
	return textResult(fmt.Sprintf("Navigated back to: %s", sess.ActivePage().URL())), nil
}

func (s *Server) handleBrowserForward(args map[string]any) (*ToolResult, error) {
	sess, err := browser.RequireSession()
	if err != nil {
		return nil, err
	}
	if err := browser.GoForward(sess.ActivePage()); err != nil {
		return nil, err
	}
	return textResult(fmt.Sprintf("Navigated forward to: %s", sess.ActivePage().URL())), nil
}

func (s *Server) handleBrowserReload(args map[string]any) (*ToolResult, error) {
	sess, err := browser.RequireSession()
	if err != nil {
		return nil, err
	}
	if err := browser.Reload(sess.ActivePage()); err != nil {
		return nil, err
	}
	return textResult("Page reloaded"), nil
}

func (s *Server) handleBrowserClick(args map[string]any) (*ToolResult, error) {
	index := getInt(args, "index", -1)
	if index < 0 {
		return nil, fmt.Errorf("'index' parameter is required")
	}
	sess, err := browser.RequireSession()
	if err != nil {
		return nil, err
	}
	if err := browser.ClickElement(sess.ActivePage(), index); err != nil {
		return nil, err
	}
	return textResult(fmt.Sprintf("Clicked element [%d]", index)), nil
}

func (s *Server) handleBrowserType(args map[string]any) (*ToolResult, error) {
	index := getInt(args, "index", -1)
	if index < 0 {
		return nil, fmt.Errorf("'index' parameter is required")
	}
	text, ok := getString(args, "text")
	if !ok || text == "" {
		return nil, fmt.Errorf("'text' parameter is required")
	}
	clear := getBool(args, "clear", false)

	sess, err := browser.RequireSession()
	if err != nil {
		return nil, err
	}
	if err := browser.TypeIntoElement(sess.ActivePage(), index, text, clear); err != nil {
		return nil, err
	}
	return textResult(fmt.Sprintf("Typed into element [%d]", index)), nil
}

func (s *Server) handleBrowserKeys(args map[string]any) (*ToolResult, error) {
	keys, ok := getString(args, "keys")
	if !ok || keys == "" {
		return nil, fmt.Errorf("'keys' parameter is required")
	}
	sess, err := browser.RequireSession()
	if err != nil {
		return nil, err
	}
	if err := browser.SendKeys(sess.ActivePage(), keys); err != nil {
		return nil, err
	}
	return textResult(fmt.Sprintf("Sent keys: %s", keys)), nil
}

func (s *Server) handleBrowserScroll(args map[string]any) (*ToolResult, error) {
	direction, ok := getString(args, "direction")
	if !ok || direction == "" {
		return nil, fmt.Errorf("'direction' parameter is required")
	}
	amount := getInt(args, "amount", 3)

	sess, err := browser.RequireSession()
	if err != nil {
		return nil, err
	}
	if err := browser.ScrollPage(sess.ActivePage(), direction, amount); err != nil {
		return nil, err
	}
	return textResult(fmt.Sprintf("Scrolled %s", direction)), nil
}

func (s *Server) handleBrowserSelect(args map[string]any) (*ToolResult, error) {
	index := getInt(args, "index", -1)
	if index < 0 {
		return nil, fmt.Errorf("'index' parameter is required")
	}
	value, ok := getString(args, "value")
	if !ok {
		return nil, fmt.Errorf("'value' parameter is required")
	}
	sess, err := browser.RequireSession()
	if err != nil {
		return nil, err
	}
	if err := browser.SelectOption(sess.ActivePage(), index, value); err != nil {
		return nil, err
	}
	return textResult(fmt.Sprintf("Selected option %q on element [%d]", value, index)), nil
}

func (s *Server) handleBrowserElements(args map[string]any) (*ToolResult, error) {
	viewportOnly := getBool(args, "viewport_only", false)
	sess, err := browser.RequireSession()
	if err != nil {
		return nil, err
	}
	elements, err := browser.DetectElements(sess.ActivePage(), viewportOnly)
	if err != nil {
		return nil, err
	}
	return jsonResult(elements)
}

func (s *Server) handleBrowserScreenshot(args map[string]any) (*ToolResult, error) {
	fullPage := getBool(args, "full_page", false)
	sess, err := browser.RequireSession()
	if err != nil {
		return nil, err
	}
	result, err := browser.CaptureScreenshot(sess.ActivePage(), fullPage)
	if err != nil {
		return nil, err
	}
	return jsonResult(result)
}

func (s *Server) handleBrowserContent(args map[string]any) (*ToolResult, error) {
	selector, _ := getString(args, "selector")
	format, _ := getString(args, "format")
	if format == "" {
		format = "text"
	}

	sess, err := browser.RequireSession()
	if err != nil {
		return nil, err
	}
	content, err := browser.ExtractContent(sess.ActivePage(), selector, format)
	if err != nil {
		return nil, err
	}
	return textResult(content), nil
}

func (s *Server) handleBrowserTabs(args map[string]any) (*ToolResult, error) {
	sess, err := browser.RequireSession()
	if err != nil {
		return nil, err
	}
	tabs, err := browser.ListTabs(sess)
	if err != nil {
		return nil, err
	}
	return jsonResult(tabs)
}

func (s *Server) handleBrowserTabSwitch(args map[string]any) (*ToolResult, error) {
	tabID := getInt(args, "tab_id", -1)
	if tabID < 0 {
		return nil, fmt.Errorf("'tab_id' parameter is required")
	}
	sess, err := browser.RequireSession()
	if err != nil {
		return nil, err
	}
	if err := browser.SwitchTab(sess, tabID); err != nil {
		return nil, err
	}
	return textResult(fmt.Sprintf("Switched to tab %d", tabID)), nil
}

func (s *Server) handleBrowserTabClose(args map[string]any) (*ToolResult, error) {
	tabID := getInt(args, "tab_id", -1)
	sess, err := browser.RequireSession()
	if err != nil {
		return nil, err
	}

	if tabID < 0 {
		// Close current tab — find its index
		pages := sess.Context().Pages()
		active := sess.ActivePage()
		for i, p := range pages {
			if p == active {
				tabID = i
				break
			}
		}
		if tabID < 0 {
			return nil, fmt.Errorf("could not determine current tab")
		}
	}

	if err := browser.CloseTab(sess, tabID); err != nil {
		return nil, err
	}
	return textResult(fmt.Sprintf("Closed tab %d", tabID)), nil
}

func (s *Server) handleBrowserTabNew(args map[string]any) (*ToolResult, error) {
	url, _ := getString(args, "url")
	sess, err := browser.RequireSession()
	if err != nil {
		return nil, err
	}
	tab, err := browser.NewTab(sess, url)
	if err != nil {
		return nil, err
	}
	return jsonResult(tab)
}

func (s *Server) handleBrowserEvaluate(args map[string]any) (*ToolResult, error) {
	expression, ok := getString(args, "expression")
	if !ok || expression == "" {
		return nil, fmt.Errorf("'expression' parameter is required")
	}
	sess, err := browser.RequireSession()
	if err != nil {
		return nil, err
	}
	result, err := browser.EvaluateJS(sess.ActivePage(), expression)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(result)
	if err != nil {
		return textResult(fmt.Sprintf("%v", result)), nil
	}
	return textResult(string(data)), nil
}

func (s *Server) handleBrowserUpload(args map[string]any) (*ToolResult, error) {
	index := getInt(args, "index", -1)
	if index < 0 {
		return nil, fmt.Errorf("'index' parameter is required")
	}
	path, ok := getString(args, "path")
	if !ok || path == "" {
		return nil, fmt.Errorf("'path' parameter is required")
	}
	sess, err := browser.RequireSession()
	if err != nil {
		return nil, err
	}
	if err := browser.UploadFile(sess.ActivePage(), index, path); err != nil {
		return nil, err
	}
	return textResult(fmt.Sprintf("Uploaded %s to element [%d]", path, index)), nil
}
