//go:build linux

/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package windows

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// KWinBackend implements Backend using KWin Scripting API via DBus (Wayland).
type KWinBackend struct {
	gdbusPath string
}

// NewKWinBackend creates a new KWin backend.
// Returns an error if gdbus or KWin is not available.
func NewKWinBackend() (*KWinBackend, error) {
	path, err := exec.LookPath("gdbus")
	if err != nil {
		return nil, fmt.Errorf("gdbus is not available: %w", err)
	}

	// Verify KWin Scripting interface is accessible
	if err := exec.Command(path, "introspect", "--session",
		"--dest", "org.kde.KWin",
		"--object-path", "/Scripting").Run(); err != nil {
		return nil, fmt.Errorf("KWin Scripting DBus interface not available: %w", err)
	}

	return &KWinBackend{gdbusPath: path}, nil
}

func (b *KWinBackend) Name() string { return "kwin" }

func (b *KWinBackend) List() ([]Window, error) {
	return b.runWindowScript(listAllScript)
}

func (b *KWinBackend) Find(query string) ([]Window, error) {
	all, err := b.runWindowScript(listAllScript)
	if err != nil {
		return nil, err
	}

	queryLower := strings.ToLower(query)
	var matched []Window
	for _, w := range all {
		if strings.Contains(strings.ToLower(w.Name), queryLower) ||
			strings.Contains(strings.ToLower(w.Class), queryLower) {
			matched = append(matched, w)
		}
	}
	return matched, nil
}

func (b *KWinBackend) Focus(windowID string) error {
	script := fmt.Sprintf(focusScript, windowID)
	_, err := b.runScript(script)
	return err
}

// runWindowScript loads a KWin script that prints window info, runs it,
// reads output from journal, and cleans up.
func (b *KWinBackend) runWindowScript(script string) ([]Window, error) {
	output, err := b.runScript(script)
	if err != nil {
		return nil, err
	}
	return parseKWinOutput(output), nil
}

// runScript executes a KWin script via DBus and captures output from journal.
func (b *KWinBackend) runScript(script string) (string, error) {
	// Record timestamp before running so we only read new journal entries
	since := time.Now().Format("2006-01-02 15:04:05")

	// Write script to temp file
	tmpFile, err := writeTempScript(script)
	if err != nil {
		return "", err
	}
	defer removeTempFile(tmpFile)

	// Load script
	loadOut, err := exec.Command(b.gdbusPath, "call", "--session",
		"--dest", "org.kde.KWin",
		"--object-path", "/Scripting",
		"--method", "org.kde.kwin.Scripting.loadScript",
		tmpFile).Output()
	if err != nil {
		return "", fmt.Errorf("failed to load KWin script: %w", err)
	}

	scriptID := extractScriptID(string(loadOut))
	scriptPath := fmt.Sprintf("/Scripting/Script%s", scriptID)

	// Run script
	if err := exec.Command(b.gdbusPath, "call", "--session",
		"--dest", "org.kde.KWin",
		"--object-path", scriptPath,
		"--method", "org.kde.kwin.Script.run").Run(); err != nil {
		return "", fmt.Errorf("failed to run KWin script: %w", err)
	}

	// Brief pause for journal to flush
	time.Sleep(200 * time.Millisecond)

	// Read output from journal since our timestamp
	out, err := exec.Command("journalctl", "--user",
		"--since", since,
		"--no-pager", "--output=cat").Output()
	if err != nil {
		return "", fmt.Errorf("failed to read KWin script output: %w", err)
	}

	// Stop and unload script
	_ = exec.Command(b.gdbusPath, "call", "--session",
		"--dest", "org.kde.KWin",
		"--object-path", scriptPath,
		"--method", "org.kde.kwin.Script.stop").Run()

	return string(out), nil
}

// extractScriptID parses "(N,)" gdbus output to get the script ID.
func extractScriptID(output string) string {
	s := strings.TrimSpace(output)
	s = strings.Trim(s, "()")
	s = strings.TrimSuffix(s, ",")
	return strings.TrimSpace(s)
}

// parseKWinOutput parses tab-separated lines: UUID\tCaption\tClass
func parseKWinOutput(output string) []Window {
	var windows []Window
	for _, line := range strings.Split(output, "\n") {
		// Only parse lines starting with "js: " (KWin script output prefix)
		if !strings.HasPrefix(line, "js: ") {
			continue
		}
		line = strings.TrimPrefix(line, "js: ")
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) < 2 {
			continue
		}
		w := Window{
			ID:   strings.TrimSpace(parts[0]),
			Name: strings.TrimSpace(parts[1]),
		}
		if len(parts) == 3 {
			w.Class = strings.TrimSpace(parts[2])
		}
		if w.Name != "" {
			windows = append(windows, w)
		}
	}
	return windows
}

const listAllScript = `var clients = workspace.windowList();
for (var i = 0; i < clients.length; i++) {
    var c = clients[i];
    if (c.caption && c.caption !== "") {
        print(c.internalId + "\t" + c.caption + "\t" + c.resourceClass);
    }
}
`

// focusScript expects one %s argument: the window internalId (UUID).
const focusScript = `var clients = workspace.windowList();
for (var i = 0; i < clients.length; i++) {
    var c = clients[i];
    if (String(c.internalId) === "%s") {
        workspace.activeWindow = c;
        break;
    }
}
`
