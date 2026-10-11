/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/CassandraGargoyle/portunix-agent/internal/version"
	"github.com/spf13/cobra"
)

// helpAIOutput runs printHelpAI and decodes its output
func helpAIOutput(t *testing.T) aiHelp {
	t.Helper()
	var buf bytes.Buffer
	if err := printHelpAI(&buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	dec := json.NewDecoder(&buf)
	dec.DisallowUnknownFields()
	var help aiHelp
	if err := dec.Decode(&help); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if dec.More() {
		t.Error("expected a single JSON document")
	}
	return help
}

// findCommand returns the command with the given name
func findCommand(t *testing.T, commands []aiCommand, name string) aiCommand {
	t.Helper()
	for _, c := range commands {
		if c.Name == name {
			return c
		}
	}
	t.Fatalf("command %q not found", name)
	return aiCommand{}
}

// findParameter returns the parameter with the given name
func findParameter(t *testing.T, cmd aiCommand, name string) aiParameter {
	t.Helper()
	for _, p := range cmd.Parameters {
		if p.Name == name {
			return p
		}
	}
	t.Fatalf("parameter %q not found in command %q", name, cmd.Name)
	return aiParameter{}
}

func TestHelpAIRequested(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{"no arguments", nil, false},
		{"flag only", []string{"--help-ai"}, true},
		{"flag after subcommand", []string{"files", "list", ".", "--help-ai"}, true},
		{"flag before subcommand", []string{"--help-ai", "files"}, true},
		{"explicit true", []string{"--help-ai=true"}, true},
		{"explicit false", []string{"--help-ai=false"}, false},
		{"invalid value", []string{"--help-ai=maybe"}, false},
		{"other flag", []string{"--help"}, false},
		{"after terminator", []string{"keyboard", "type", "--", "--help-ai"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := helpAIRequested(tt.args); got != tt.want {
				t.Errorf("helpAIRequested(%v) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}

func TestHelpAITopLevelFields(t *testing.T) {
	help := helpAIOutput(t)

	if help.Tool != "ptx-agent" {
		t.Errorf("expected tool ptx-agent, got %q", help.Tool)
	}
	if help.Version != version.Version {
		t.Errorf("expected version %q, got %q", version.Version, help.Version)
	}
	if help.Description != rootCmd.Short {
		t.Errorf("expected description %q, got %q", rootCmd.Short, help.Description)
	}
	if len(help.Commands) == 0 {
		t.Error("expected commands")
	}
}

func TestHelpAIListsAllCommands(t *testing.T) {
	help := helpAIOutput(t)

	// Every runnable command of the cobra tree must be present under its path
	var walk func(cmd *cobra.Command, top *aiCommand, path string)
	walk = func(cmd *cobra.Command, top *aiCommand, path string) {
		for _, sub := range cmd.Commands() {
			name := strings.TrimSpace(path + " " + sub.Name())
			if top == nil {
				info := findCommand(t, help.Commands, name)
				walk(sub, &info, "")
				continue
			}
			if sub.Runnable() {
				findCommand(t, top.Subcommands, name)
			}
			walk(sub, top, name)
		}
	}
	walk(rootCmd, nil, "")

	// Platform independent commands named in the issue
	for _, name := range []string{"version", "setup", "files", "vscode", "app", "mcp", "browser"} {
		findCommand(t, help.Commands, name)
	}
	for _, name := range []string{"help", "completion"} {
		for _, c := range help.Commands {
			if c.Name == name {
				t.Errorf("generated command %q must not be listed", name)
			}
		}
	}
}

func TestHelpAINestedSubcommands(t *testing.T) {
	help := helpAIOutput(t)

	vscode := findCommand(t, help.Commands, "vscode")
	findCommand(t, vscode.Subcommands, "terminal read")
	findCommand(t, vscode.Subcommands, "terminal list")
	for _, c := range vscode.Subcommands {
		if c.Name == "terminal" {
			t.Error("command group without an action must not be listed")
		}
	}

	browser := findCommand(t, help.Commands, "browser")
	findCommand(t, browser.Subcommands, "tab switch")
}

func TestHelpAIParameters(t *testing.T) {
	help := helpAIOutput(t)

	files := findCommand(t, help.Commands, "files")
	if len(files.Parameters) != 0 {
		t.Errorf("command group must have no parameters, got %v", files.Parameters)
	}
	list := findCommand(t, files.Subcommands, "list")
	path := findParameter(t, list, "path")
	if path.Type != "string" || !path.Required {
		t.Errorf("unexpected path parameter: %+v", path)
	}
	// Persistent flag of the parent command
	allowed := findParameter(t, list, "--allowed-dirs")
	if allowed.Type != "array" || allowed.Required || allowed.Default != "" {
		t.Errorf("unexpected --allowed-dirs parameter: %+v", allowed)
	}

	browser := findCommand(t, help.Commands, "browser")
	headless := findParameter(t, findCommand(t, browser.Subcommands, "launch"), "--headless")
	if headless.Type != "boolean" || headless.Default != "true" {
		t.Errorf("unexpected --headless parameter: %+v", headless)
	}
	amount := findParameter(t, findCommand(t, browser.Subcommands, "scroll"), "--amount")
	if amount.Type != "integer" || amount.Default != "3" {
		t.Errorf("unexpected --amount parameter: %+v", amount)
	}
	index := findParameter(t, findCommand(t, browser.Subcommands, "click"), "index")
	if index.Type != "integer" || !index.Required {
		t.Errorf("unexpected index parameter: %+v", index)
	}
	url := findParameter(t, findCommand(t, browser.Subcommands, "tab new"), "url")
	if url.Required {
		t.Errorf("expected optional url parameter: %+v", url)
	}
}

func TestHelpAIParametersAreComplete(t *testing.T) {
	help := helpAIOutput(t)

	var check func(path string, commands []aiCommand)
	check = func(path string, commands []aiCommand) {
		for _, c := range commands {
			name := strings.TrimSpace(path + " " + c.Name)
			if c.Brief == "" {
				t.Errorf("command %q has no brief", name)
			}
			for _, p := range c.Parameters {
				if p.Name == "" || p.Type == "" || p.Description == "" {
					t.Errorf("command %q has an incomplete parameter: %+v", name, p)
				}
				if p.Name == "--help" || p.Name == "--help-ai" {
					t.Errorf("command %q must not list %s", name, p.Name)
				}
				// A new placeholder needs an entry in positionalArgs
				if _, known := positionalArgs[p.Name]; !known && !strings.HasPrefix(p.Name, "--") {
					t.Errorf("command %q has an undescribed positional argument %q", name, p.Name)
				}
			}
			check(name, c.Subcommands)
		}
	}
	check("", help.Commands)
}

func TestHelpAIFlagIsPersistent(t *testing.T) {
	if rootCmd.PersistentFlags().Lookup(helpAIFlag) == nil {
		t.Fatal("expected --help-ai on the root command")
	}
	// Subcommands list the flag in their help
	if filesListCmd.InheritedFlags().Lookup(helpAIFlag) == nil {
		t.Error("expected --help-ai to be inherited by subcommands")
	}
}
