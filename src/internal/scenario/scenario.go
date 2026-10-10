/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

// Package scenario provides JSON-based typing scenario execution.
package scenario

import (
	"encoding/json"
	"fmt"
	"os"
)

// Scenario defines a typing automation scenario.
type Scenario struct {
	Metadata Metadata  `json:"metadata"`
	Commands []Command `json:"commands"`
}

// Metadata holds scenario-level configuration.
type Metadata struct {
	Title              string `json:"title"`
	Description        string `json:"description"`
	DefaultSpeed       string `json:"defaultSpeed"`
	DefaultPauseBefore int    `json:"defaultPauseBefore"`
}

// Command defines a single typing action.
type Command struct {
	Command       string `json:"command"`
	Description   string `json:"description"`
	PauseBefore   int    `json:"pauseBefore"`
	Speed         string `json:"speed"`
	SendEnter     bool   `json:"sendEnter"`
	WaitForHotkey bool   `json:"waitForHotkey"`
}

// SpeedToDelay converts a speed name to keystroke delay in milliseconds.
func SpeedToDelay(speed string) int {
	switch speed {
	case "fast":
		return 20
	case "medium":
		return 50
	case "normal":
		return 80
	case "slow":
		return 150
	default:
		return 50
	}
}

// Load reads a scenario from a JSON file.
func Load(path string) (*Scenario, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read scenario file: %w", err)
	}

	var s Scenario
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("failed to parse scenario file: %w", err)
	}

	// Apply defaults
	for i := range s.Commands {
		if s.Commands[i].Speed == "" {
			s.Commands[i].Speed = s.Metadata.DefaultSpeed
		}
		if s.Commands[i].PauseBefore == 0 {
			s.Commands[i].PauseBefore = s.Metadata.DefaultPauseBefore
		}
	}

	return &s, nil
}
