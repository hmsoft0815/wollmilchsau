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
	Main    string `json:"main,omitempty"` // entry point relative to package dir
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
		if strings.HasPrefix(name, "@") {
			subEntries, err := os.ReadDir(filepath.Join(nodeModules, name))
			if err != nil {
				continue
			}
			for _, subEntry := range subEntries {
				if !subEntry.IsDir() || strings.HasPrefix(subEntry.Name(), ".") {
					continue
				}
				subPkgPath := filepath.Join(nodeModules, name, subEntry.Name())
				if pkg, ok := readPackageJSON(subPkgPath, name+"/"+subEntry.Name()); ok {
					pkgs = append(pkgs, pkg)
				}
			}
			continue
		}

		pkgPath := filepath.Join(nodeModules, name)
		if pkg, ok := readPackageJSON(pkgPath, name); ok {
			pkgs = append(pkgs, pkg)
		}
	}

	return pkgs, nil
}

func readPackageJSON(pkgPath, defaultName string) (Package, bool) {
	pkgJSON, err := os.ReadFile(filepath.Join(pkgPath, "package.json"))
	if err != nil {
		return Package{}, false
	}

	var meta struct {
		Name    string `json:"name"`
		Version string `json:"version"`
		Main    string `json:"main"`
		Type    string `json:"type"`
	}
	if err := json.Unmarshal(pkgJSON, &meta); err != nil {
		return Package{}, false
	}

	name := meta.Name
	if name == "" {
		name = defaultName
	}

	return Package{
		Name:    name,
		Version: meta.Version,
		Main:    meta.Main,
		Type:    meta.Type,
	}, true
}
