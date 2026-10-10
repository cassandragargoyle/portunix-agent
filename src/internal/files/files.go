/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

// Package files provides directory browsing with path whitelist enforcement.
package files

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FileInfo represents metadata about a file or directory.
type FileInfo struct {
	Name    string    `json:"name"`
	Path    string    `json:"path"`
	Size    int64     `json:"size"`
	IsDir   bool      `json:"is_dir"`
	ModTime time.Time `json:"mod_time"`
	Mode    string    `json:"mode"`
}

// Browser provides file browsing restricted to allowed directories.
type Browser struct {
	allowedDirs []string
}

// NewBrowser creates a new file browser with the given allowed directories.
func NewBrowser(allowedDirs []string) (*Browser, error) {
	if len(allowedDirs) == 0 {
		return nil, fmt.Errorf("at least one allowed directory must be configured")
	}

	// Resolve all allowed dirs to absolute paths
	resolved := make([]string, 0, len(allowedDirs))
	for _, dir := range allowedDirs {
		abs, err := filepath.Abs(dir)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve path %q: %w", dir, err)
		}
		resolved = append(resolved, filepath.Clean(abs))
	}

	return &Browser{allowedDirs: resolved}, nil
}

// List returns the contents of a directory, if it is within allowed paths.
func (b *Browser) List(path string) ([]FileInfo, error) {
	cleanPath, err := b.validatePath(path)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %q: %w", cleanPath, err)
	}

	result := make([]FileInfo, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue // skip entries we cannot stat
		}

		result = append(result, FileInfo{
			Name:    entry.Name(),
			Path:    filepath.Join(cleanPath, entry.Name()),
			Size:    info.Size(),
			IsDir:   entry.IsDir(),
			ModTime: info.ModTime(),
			Mode:    info.Mode().String(),
		})
	}

	return result, nil
}

// Info returns metadata for a single file or directory, if within allowed paths.
func (b *Browser) Info(path string) (*FileInfo, error) {
	cleanPath, err := b.validatePath(path)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat %q: %w", cleanPath, err)
	}

	return &FileInfo{
		Name:    info.Name(),
		Path:    cleanPath,
		Size:    info.Size(),
		IsDir:   info.IsDir(),
		ModTime: info.ModTime(),
		Mode:    info.Mode().String(),
	}, nil
}

// IsAllowed checks whether the given path falls within allowed directories.
func (b *Browser) IsAllowed(path string) bool {
	_, err := b.validatePath(path)
	return err == nil
}

// validatePath resolves, cleans, and checks that a path is within allowed dirs.
// Returns the cleaned absolute path or an error.
func (b *Browser) validatePath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path cannot be empty")
	}

	// Resolve to absolute and clean
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path %q: %w", path, err)
	}
	cleanPath := filepath.Clean(abs)

	// Resolve symlinks to prevent symlink-based escapes
	resolved, err := filepath.EvalSymlinks(cleanPath)
	if err != nil {
		// If the path doesn't exist yet, check the parent
		// For files/info this means the file doesn't exist
		if os.IsNotExist(err) {
			return "", fmt.Errorf("path does not exist: %q", cleanPath)
		}
		return "", fmt.Errorf("failed to resolve symlinks for %q: %w", cleanPath, err)
	}

	// Check against allowed directories
	for _, allowed := range b.allowedDirs {
		resolvedAllowed, err := filepath.EvalSymlinks(allowed)
		if err != nil {
			continue
		}
		if resolved == resolvedAllowed || isSubPath(resolved, resolvedAllowed) {
			return resolved, nil
		}
	}

	return "", fmt.Errorf("access denied: path %q is outside allowed directories", path)
}

// isSubPath checks if child is a subdirectory or file within parent.
func isSubPath(child, parent string) bool {
	// Ensure parent ends with separator so "/tmp" doesn't match "/tmpfoo"
	parentPrefix := parent
	if parentPrefix[len(parentPrefix)-1] != filepath.Separator {
		parentPrefix += string(filepath.Separator)
	}
	return strings.HasPrefix(child, parentPrefix)
}
