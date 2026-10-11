/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package cli

import (
	"fmt"

	"github.com/CassandraGargoyle/portunix-agent/internal/files"
	"github.com/spf13/cobra"
)

var allowedDirs []string

var filesCmd = &cobra.Command{
	Use:   "files",
	Short: "Browse files and directories",
	Long:  "List files and retrieve metadata within allowed (whitelisted) directories.",
}

var filesListCmd = &cobra.Command{
	Use:   "list <path>",
	Short: "List directory contents",
	Long:  "List files and subdirectories within an allowed directory.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		browser, err := files.NewBrowser(allowedDirs)
		if err != nil {
			return err
		}

		entries, err := browser.List(args[0])
		if err != nil {
			return err
		}

		if len(entries) == 0 {
			fmt.Println("Directory is empty.")
			return nil
		}

		for _, e := range entries {
			fmt.Printf("%-10s %8d  %s  %s\n",
				e.Mode, e.Size,
				e.ModTime.Format("2006-01-02 15:04"),
				e.Name)
		}
		return nil
	},
}

var filesInfoCmd = &cobra.Command{
	Use:   "info <path>",
	Short: "Show file metadata",
	Long:  "Display detailed metadata for a file or directory within allowed paths.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		browser, err := files.NewBrowser(allowedDirs)
		if err != nil {
			return err
		}

		info, err := browser.Info(args[0])
		if err != nil {
			return err
		}

		fmt.Printf("Name:     %s\n", info.Name)
		fmt.Printf("Path:     %s\n", info.Path)
		fmt.Printf("Size:     %d bytes\n", info.Size)
		fmt.Printf("Type:     %s\n", fileType(info.IsDir))
		fmt.Printf("Mode:     %s\n", info.Mode)
		fmt.Printf("Modified: %s\n", info.ModTime.Format("2006-01-02 15:04:05"))
		return nil
	},
}

func fileType(isDir bool) string {
	if isDir {
		return "directory"
	}
	return "file"
}

func init() {
	// Register --allowed-dirs flag on the files command, inherited by subcommands
	filesCmd.PersistentFlags().StringSliceVar(&allowedDirs, "allowed-dirs", nil,
		"comma-separated list of allowed directory paths (required)")
	_ = filesCmd.MarkPersistentFlagRequired("allowed-dirs")

	filesCmd.AddCommand(filesListCmd)
	filesCmd.AddCommand(filesInfoCmd)
}
