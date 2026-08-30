// Copyright (c) 2026 Michael Lechner. All rights reserved.
// Package npminstall handles installing npm packages and converting them to
// VirtualFiles for injection into esbuild's virtual filesystem.
package npminstall

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/hmsoft0815/wollmilchsau/bundle"
	"github.com/hmsoft0815/wollmilchsau/internal/parser"
)

// DefaultPackages are the bundled JS packages installed by default at server
// startup (when -bundled-js-deps is not set).
var DefaultPackages = []string{
	"crypto-js",
	"lodash",
	"@types/lodash",
	"mathjs",
	"zod",
}

// pkgDescriptions maps a package name to a human-readable one-liner that the
// list_js_packages tool returns so LLM agents know what each package does.
var pkgDescriptions = map[string]string{
	"crypto-js":     "Cryptographic functions (SHA-256, AES, MD5, HMAC, etc.) in JavaScript",
	"lodash":        "Utility library for working with arrays, numbers, objects, strings, etc.",
	"zod":           "TypeScript-first schema validation with static type inference",
	"@types/lodash": "TypeScript type definitions for lodash",
	"mathjs":        "Mathematics engine with matrices, fractions, units, and expressions",
	"uuid":          "Generate RFC-compliant UUIDs (v1, v4, etc.)",
}

// Manager handles npm package installation and cataloging.
type Manager struct {
	mu   sync.Mutex
	dir  string // shared temp dir where all bundled packages live
	pkgs []bundle.Package
}

// NewManager creates a manager for the given temp directory containing
// pre-installed node_modules. Pass an empty string to auto-create one.
func NewManager(tmpDir string) (*Manager, error) {
	if tmpDir == "" {
		var err error
		tmpDir, err = os.MkdirTemp("", "npm-bundle-*")
		if err != nil {
			return nil, fmt.Errorf("create temp dir: %w", err)
		}
	}

	m := &Manager{dir: tmpDir}

	// Initialize if node_modules doesn't exist yet.
	pkgJSON := filepath.Join(tmpDir, "package.json")
	if _, err := os.Stat(pkgJSON); os.IsNotExist(err) {
		initCmd := exec.Command("npm", "init", "-y")
		initCmd.Dir = tmpDir
		initCmd.Stdout = nil
		initCmd.Stderr = nil
		if err := initCmd.Run(); err != nil {
			return nil, fmt.Errorf("npm init: %w", err)
		}
	}

	if err := os.MkdirAll(filepath.Join(tmpDir, "node_modules"), 0o755); err != nil {
		return nil, fmt.Errorf("create node_modules: %w", err)
	}

	return m, nil
}

// Install ensures the given packages are installed in the manager's directory.
func (m *Manager) Install(packages []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(packages) == 0 {
		return nil
	}

	args := []string{"install", "--save-exact", "--no-audit", "--log=warn"}
	args = append(args, packages...)

	cmd := exec.Command("npm", args...)
	cmd.Dir = m.dir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("npm install: %w\n%s", err, truncate(out, 512))
	}

	// Read installed packages.
	pkgs, err := bundle.PackagesFromDir(m.dir)
	if err != nil {
		return fmt.Errorf("read packages: %w", err)
	}
	m.pkgs = pkgs

	return nil
}

// Packages returns the list of bundled packages with metadata.
func (m *Manager) Packages() []bundle.Package {
	return m.pkgs
}

// PackageSpecs returns package names+versions as "pkg@ver" specs.
func (m *Manager) PackageSpecs() []string {
	specs := make([]string, 0, len(m.pkgs))
	for _, p := range m.pkgs {
		if p.Version != "" {
			specs = append(specs, fmt.Sprintf("%s@%s", p.Name, p.Version))
		} else {
			specs = append(specs, p.Name)
		}
	}
	return specs
}

// PackageInfos returns a list of package info objects for the list_js_packages tool.
func (m *Manager) PackageInfos() []map[string]any {
	infos := make([]map[string]any, 0, len(m.pkgs))
	for _, p := range m.pkgs {
		info := map[string]any{
			"name":    p.Name,
			"version": p.Version,
			"type":    p.Type,
			"main":    p.Main,
		}
		if desc, ok := pkgDescriptions[p.Name]; ok {
			info["description"] = desc
		}
		infos = append(infos, info)
	}
	return infos
}

// NodeModulesPath returns the node_modules directory path.
func (m *Manager) NodeModulesPath() string {
	return filepath.Join(m.dir, "node_modules")
}

// ToVirtualFiles walks the manager's node_modules and returns VirtualFile
// entries keyed by "node_modules/<scope>/..." or "node_modules/<pkg>/...".
func (m *Manager) ToVirtualFiles() ([]parser.VirtualFile, error) {
	nodeModules := m.NodeModulesPath()
	var vfs []parser.VirtualFile

	err := fs.WalkDir(os.DirFS(nodeModules), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}

		fullPath := filepath.Join(nodeModules, path)
		data, readErr := os.ReadFile(fullPath)
		if readErr != nil {
			return nil // skip unreadable files
		}

		vfs = append(vfs, parser.VirtualFile{
			Name:    filepath.ToSlash(filepath.Join("node_modules", path)),
			Content: string(data),
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk node_modules: %w", err)
	}

	return vfs, nil
}

func truncate(b []byte, max int) string {
	s := string(b)
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}

// MustJSON marshals v to JSON and panics on error (only for test / init helpers).
func MustJSON(v any) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}
