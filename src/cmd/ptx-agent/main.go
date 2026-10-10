/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

// Package main is the entry point for ptx-agent binary.
package main

import (
	"fmt"
	"os"

	"github.com/CassandraGargoyle/portunix-agent/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
