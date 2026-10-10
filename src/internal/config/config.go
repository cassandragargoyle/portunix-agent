/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

// Package config provides configuration for ptx-agent.
package config

// DefaultAllowedDirs returns the default set of allowed directories
// for file browsing. In production this would be loaded from a
// configuration file.
func DefaultAllowedDirs() []string {
	return []string{}
}
