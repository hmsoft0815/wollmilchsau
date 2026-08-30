// Copyright (c) 2026 Michael Lechner. All rights reserved.
package bundle

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPackagesFromDir_ValidAndScoped(t *testing.T) {
	tmpDir := t.TempDir()
	nm := filepath.Join(tmpDir, "node_modules")

	// Create regular package: lodash
	lodashDir := filepath.Join(nm, "lodash")
	if err := os.MkdirAll(lodashDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(lodashDir, "package.json"), []byte(`{
		"name": "lodash",
		"version": "4.17.21",
		"main": "lodash.js"
	}`), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create scoped package: @types/lodash
	typesLodashDir := filepath.Join(nm, "@types", "lodash")
	if err := os.MkdirAll(typesLodashDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(typesLodashDir, "package.json"), []byte(`{
		"name": "@types/lodash",
		"version": "4.14.202",
		"main": "index.d.ts",
		"type": "module"
	}`), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create ignored directory: .bin
	binDir := filepath.Join(nm, ".bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create package without package.json (should be ignored)
	noPkgDir := filepath.Join(nm, "nopkg")
	if err := os.MkdirAll(noPkgDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create package with invalid JSON (should be ignored)
	badPkgDir := filepath.Join(nm, "badpkg")
	if err := os.MkdirAll(badPkgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(badPkgDir, "package.json"), []byte(`{invalid`), 0o644); err != nil {
		t.Fatal(err)
	}

	pkgs, err := PackagesFromDir(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pkgMap := make(map[string]Package)
	for _, p := range pkgs {
		pkgMap[p.Name] = p
	}

	if len(pkgMap) != 2 {
		t.Fatalf("expected 2 packages, got %d (%+v)", len(pkgMap), pkgMap)
	}

	lodash, ok := pkgMap["lodash"]
	if !ok {
		t.Fatal("expected package lodash to be present")
	}
	if lodash.Version != "4.17.21" || lodash.Main != "lodash.js" {
		t.Errorf("unexpected lodash meta: %+v", lodash)
	}

	typesLodash, ok := pkgMap["@types/lodash"]
	if !ok {
		t.Fatal("expected package @types/lodash to be present")
	}
	if typesLodash.Version != "4.14.202" || typesLodash.Type != "module" {
		t.Errorf("unexpected @types/lodash meta: %+v", typesLodash)
	}
}

func TestPackagesFromDir_MissingDir(t *testing.T) {
	tmpDir := t.TempDir()
	_, err := PackagesFromDir(tmpDir)
	if err == nil {
		t.Fatal("expected error for non-existent node_modules directory")
	}
}

func TestPackagesFromDir_EmptyScope(t *testing.T) {
	tmpDir := t.TempDir()
	nm := filepath.Join(tmpDir, "node_modules")
	scopeDir := filepath.Join(nm, "@empty")
	if err := os.MkdirAll(scopeDir, 0o755); err != nil {
		t.Fatal(err)
	}

	pkgs, err := PackagesFromDir(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 0 {
		t.Errorf("expected 0 packages, got %d", len(pkgs))
	}
}
