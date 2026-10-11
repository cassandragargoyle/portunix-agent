/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package cli

import (
	"encoding/json"
	"io"
	"strconv"
	"strings"

	"github.com/CassandraGargoyle/portunix-agent/internal/version"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// helpAIFlag is the name of the flag that prints the machine-readable help
const helpAIFlag = "help-ai"

// aiHelp is the root document of the --help-ai output (Portunix ecosystem schema)
type aiHelp struct {
	Tool        string      `json:"tool"`
	Version     string      `json:"version"`
	Description string      `json:"description"`
	Commands    []aiCommand `json:"commands"`
}

// aiCommand describes one command, nested commands are listed in Subcommands
type aiCommand struct {
	Name        string        `json:"name"`
	Brief       string        `json:"brief"`
	Description string        `json:"description,omitempty"`
	Parameters  []aiParameter `json:"parameters,omitempty"`
	Subcommands []aiCommand   `json:"subcommands,omitempty"`
}

// aiParameter describes a positional argument or a flag,
// Positional tells the two apart because flag names carry no -- prefix
type aiParameter struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
	Default     string `json:"default,omitempty"`
	Positional  bool   `json:"positional,omitempty"`
}

// positionalArgs describes the positional arguments, the key is the command path
// and the placeholder from the Use line ("files list/path").
// Cobra keeps no type or description for positional arguments
var positionalArgs = map[string]aiParameter{
	"files list/path":          {Type: "string", Description: "Directory path to list"},
	"files info/path":          {Type: "string", Description: "File path"},
	"app tree/name":            {Type: "string", Description: "Application name or a part of it"},
	"app read/name":            {Type: "string", Description: "Application name or a part of it"},
	"keyboard type/text":       {Type: "string", Description: "Text to type"},
	"keyboard key/combination": {Type: "string", Description: "Key combination, for example ctrl+c"},
	"keyboard paste/text":      {Type: "string", Description: "Text to paste"},
	"windows find/query":       {Type: "string", Description: "Window name or a part of it"},
	"windows focus/window-id":  {Type: "string", Description: "Identifier of the window to focus"},
	"scenario run/file.json":   {Type: "string", Description: "Path to the scenario JSON file"},
	"browser goto/url":         {Type: "string", Description: "URL to navigate to"},
	"browser click/index":      {Type: "integer", Description: "Index of the element to click, from the elements command"},
	"browser type/index":       {Type: "integer", Description: "Index of the element to type into, from the elements command"},
	"browser type/text":        {Type: "string", Description: "Text to type into the element"},
	"browser keys/keys":        {Type: "string", Description: "Key combination, for example Enter or Control+a"},
	"browser scroll/direction": {Type: "string", Description: "Scroll direction: up, down, left, right"},
	"browser select/index":     {Type: "integer", Description: "Index of the dropdown element, from the elements command"},
	"browser select/value":     {Type: "string", Description: "Option value to select"},
	"browser tab switch/id":    {Type: "integer", Description: "ID of the tab to switch to, from the tabs command"},
	"browser tab close/id":     {Type: "integer", Description: "ID of the tab to close, the current tab if omitted"},
	"browser tab new/url":      {Type: "string", Description: "URL to open in the new tab"},
}

// undescribedArgPrefix starts the description of an argument missing in positionalArgs
const undescribedArgPrefix = "Positional argument "

// helpAIRequested reports whether the arguments contain the --help-ai flag.
// The check runs before cobra, so the flag works at any position and
// no argument validation or subcommand runs
func helpAIRequested(args []string) bool {
	for _, arg := range args {
		if arg == "--" {
			return false
		}
		if arg == "--"+helpAIFlag {
			return true
		}
		if value, ok := strings.CutPrefix(arg, "--"+helpAIFlag+"="); ok {
			enabled, err := strconv.ParseBool(value)
			return err == nil && enabled
		}
	}
	return false
}

// printHelpAI writes the command tree as a single JSON document
func printHelpAI(w io.Writer) error {
	help := aiHelp{
		Tool:        rootCmd.Name(),
		Version:     version.Version,
		Description: rootCmd.Short,
		Commands:    []aiCommand{},
	}
	for _, cmd := range rootCmd.Commands() {
		if !describable(cmd) {
			continue
		}
		info := describeCommand(cmd, cmd.Name())
		info.Subcommands = describeSubcommands(cmd, "")
		help.Commands = append(help.Commands, info)
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(help)
}

// describable filters out hidden commands and the commands generated by cobra
func describable(cmd *cobra.Command) bool {
	return !cmd.Hidden && cmd.Name() != "help" && cmd.Name() != "completion"
}

// describeSubcommands returns the runnable descendants of cmd as a flat list,
// a nested command is named by its path (for example "terminal read")
func describeSubcommands(cmd *cobra.Command, prefix string) []aiCommand {
	var result []aiCommand
	for _, sub := range cmd.Commands() {
		if !describable(sub) {
			continue
		}
		name := prefix + sub.Name()
		if sub.Runnable() {
			result = append(result, describeCommand(sub, name))
		}
		result = append(result, describeSubcommands(sub, name+" ")...)
	}
	return result
}

// describeCommand describes a single command without its subcommands
func describeCommand(cmd *cobra.Command, name string) aiCommand {
	info := aiCommand{
		Name:        name,
		Brief:       cmd.Short,
		Description: cmd.Long,
	}
	if !cmd.Runnable() {
		return info
	}
	path := strings.TrimPrefix(cmd.CommandPath(), rootCmd.Name()+" ")
	info.Parameters = positionalParameters(path, cmd.Use)
	addFlag := func(f *pflag.Flag) {
		if f.Hidden || f.Name == "help" || f.Name == helpAIFlag {
			return
		}
		info.Parameters = append(info.Parameters, flagParameter(f))
	}
	cmd.NonInheritedFlags().VisitAll(addFlag)
	cmd.InheritedFlags().VisitAll(addFlag)
	return info
}

// positionalParameters derives the positional arguments from a Use line,
// <name> is a required argument and [name] is an optional one
func positionalParameters(path, use string) []aiParameter {
	var params []aiParameter
	for _, token := range strings.Fields(use)[1:] {
		var required bool
		switch {
		case strings.HasPrefix(token, "<") && strings.HasSuffix(token, ">"):
			required = true
		case strings.HasPrefix(token, "[") && strings.HasSuffix(token, "]"):
			required = false
		default:
			continue
		}
		name := token[1 : len(token)-1]
		param := positionalArgs[path+"/"+name]
		if param.Type == "" {
			param = aiParameter{Type: "string", Description: undescribedArgPrefix + name}
		}
		param.Name = name
		param.Required = required
		param.Positional = true
		params = append(params, param)
	}
	return params
}

// flagParameter describes a flag, the name has no -- prefix (ecosystem convention)
func flagParameter(f *pflag.Flag) aiParameter {
	_, required := f.Annotations[cobra.BashCompOneRequiredFlag]
	param := aiParameter{
		Name:        f.Name,
		Type:        flagType(f.Value.Type()),
		Required:    required,
		Description: f.Usage,
	}
	if f.DefValue != "" && f.DefValue != "[]" {
		param.Default = f.DefValue
	}
	return param
}

// flagType maps a pflag type name to the type name used in the schema
func flagType(pflagType string) string {
	switch {
	case pflagType == "bool":
		return "boolean"
	case strings.HasPrefix(pflagType, "int"), strings.HasPrefix(pflagType, "uint"):
		return "integer"
	case strings.HasSuffix(pflagType, "Slice"), strings.HasSuffix(pflagType, "Array"):
		return "array"
	default:
		return pflagType
	}
}
