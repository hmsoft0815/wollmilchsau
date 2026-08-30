// Copyright (c) 2026 Michael Lechner. All rights reserved.
// Package bundle provides bundled JavaScript packages ready for injection into
// the execution environment without requiring runtime npm installs.
package bundle

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Package represents a JS package with its metadata and file list.
type Package struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Type    string `json:"type,omitempty"` // "module" for ESM, empty for CJS
	Main    string `json:"main,omitempty"`   // entry point relative to package dir
}

// PackagesFromDir reads all installed npm packages from a node_modules directory.
func PackagesFromDir(baseDir string) ([]Package, error) {
	nodeModules := filepath.Join(baseDir, "node_modules")
	entries, err := os.ReadDir(nodeModules)
	if err != nil {
		return nil, fmt.Errorf("read node_modules: %w", err)
	}

	var pkgs []Package
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		name := entry.Name()
		pkgPath := filepath.Join(nodeModules, name)

		pkgJSON, err := os.ReadFile(filepath.Join(pkgPath, "package.json"))
		if err != nil {
			continue // skip packages without package.json
		}

		var meta struct {
			Name    string `json:"name"`
			Version string `json:"version"`
			Main    string `json:"main"`
			Type    string `json:"type"`
		}
		if err := json.Unmarshal(pkgJSON, &meta); err != nil {
			continue // skip invalid packages
		}

		pkgs = append(pkgs, Package{
			Name:    meta.Name,
			Version: meta.Version,
			Main:    meta.Main,
			Type:    meta.Type,
		})
	}

	return pkgs, nil
}
