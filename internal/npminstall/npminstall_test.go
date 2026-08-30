// Copyright (c) 2026 Michael Lechner. All rights reserved.
package npminstall

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewManager_AutoTmpDir(t *testing.T) {
	mgr, err := NewManager("")
	if err != nil {
		t.Fatalf("NewManager(\"\") failed: %v", err)
	}
	defer os.RemoveAll(mgr.dir)

	if mgr.NodeModulesPath() != filepath.Join(mgr.dir, "node_modules") {
		t.Errorf("unexpected node_modules path: %s", mgr.NodeModulesPath())
	}

	if _, err := os.Stat(filepath.Join(mgr.dir, "package.json")); err != nil {
		t.Errorf("expected package.json to exist: %v", err)
	}
}

func TestNewManager_ExplicitDir(t *testing.T) {
	dir := t.TempDir()
	mgr, err := NewManager(dir)
	if err != nil {
		t.Fatalf("NewManager(dir) failed: %v", err)
	}

	if mgr.dir != dir {
		t.Errorf("expected dir %s, got %s", dir, mgr.dir)
	}
}

func TestManager_PackageMetadata(t *testing.T) {
	dir := t.TempDir()
	mgr, err := NewManager(dir)
	if err != nil {
		t.Fatal(err)
	}

	// Setup mock packages in node_modules
	nm := mgr.NodeModulesPath()
	lodashDir := filepath.Join(nm, "lodash")
	if err = os.MkdirAll(lodashDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(lodashDir, "package.json"), []byte(`{
		"name": "lodash",
		"version": "4.17.21",
		"main": "lodash.js"
	}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(lodashDir, "lodash.js"), []byte(`module.exports = {};`), 0o644); err != nil {
		t.Fatal(err)
	}

	typesDir := filepath.Join(nm, "@types", "lodash")
	if err = os.MkdirAll(typesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(typesDir, "package.json"), []byte(`{
		"name": "@types/lodash",
		"version": "4.14.202"
	}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(typesDir, "index.d.ts"), []byte(`export = _;`), 0o644); err != nil {
		t.Fatal(err)
	}

	// Mock .bin directory (must be skipped in ToVirtualFiles)
	binDir := filepath.Join(nm, ".bin")
	if err = os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(binDir, "some-cli"), []byte(`#!/bin/sh`), 0o755); err != nil {
		t.Fatal(err)
	}

	// Test ToVirtualFiles
	vfs, err := mgr.ToVirtualFiles()
	if err != nil {
		t.Fatalf("ToVirtualFiles failed: %v", err)
	}

	foundLodash := false
	foundTypesLodash := false
	for _, vf := range vfs {
		if strings.HasPrefix(vf.Name, "node_modules/.bin") {
			t.Errorf("hidden directory .bin should have been skipped, found: %s", vf.Name)
		}
		if vf.Name == "node_modules/lodash/lodash.js" {
			foundLodash = true
		}
		if vf.Name == "node_modules/@types/lodash/index.d.ts" {
			foundTypesLodash = true
		}
	}

	if !foundLodash {
		t.Error("expected lodash.js in virtual files")
	}
	if !foundTypesLodash {
		t.Error("expected index.d.ts in virtual files")
	}
}

func TestManager_PackageSpecsAndInfos(t *testing.T) {
	dir := t.TempDir()
	mgr, err := NewManager(dir)
	if err != nil {
		t.Fatal(err)
	}

	// Manually set packages
	nm := mgr.NodeModulesPath()
	zodDir := filepath.Join(nm, "zod")
	if err := os.MkdirAll(zodDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(zodDir, "package.json"), []byte(`{
		"name": "zod",
		"version": "3.22.4",
		"main": "lib/index.js"
	}`), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := mgr.Reload(); err != nil {
		t.Fatalf("Reload failed: %v", err)
	}

	specs := mgr.PackageSpecs()
	if len(specs) != 1 || specs[0] != "zod@3.22.4" {
		t.Errorf("expected [\"zod@3.22.4\"], got %v", specs)
	}

	infos := mgr.PackageInfos()
	if len(infos) != 1 {
		t.Fatalf("expected 1 package info, got %d", len(infos))
	}
	if infos[0]["name"] != "zod" || infos[0]["version"] != "3.22.4" {
		t.Errorf("unexpected info: %+v", infos[0])
	}
	if infos[0]["description"] != pkgDescriptions["zod"] {
		t.Errorf("expected zod description, got %v", infos[0]["description"])
	}
}

func TestHelpers(t *testing.T) {
	s := truncate([]byte("hello world"), 5)
	if s != "hello…" {
		t.Errorf("expected 'hello…', got %q", s)
	}

	short := truncate([]byte("hi"), 5)
	if short != "hi" {
		t.Errorf("expected 'hi', got %q", short)
	}

	jsonStr := MustJSON(map[string]string{"foo": "bar"})
	if !strings.Contains(jsonStr, `"foo": "bar"`) {
		t.Errorf("MustJSON output unexpected: %s", jsonStr)
	}
}
